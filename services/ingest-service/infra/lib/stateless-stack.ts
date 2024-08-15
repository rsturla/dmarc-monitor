import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as sqs from "aws-cdk-lib/aws-sqs";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as ses from "aws-cdk-lib/aws-ses";
import * as ses_actions from "aws-cdk-lib/aws-ses-actions";
import * as iam from "aws-cdk-lib/aws-iam";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as cr from "aws-cdk-lib/custom-resources";
import * as logs from "aws-cdk-lib/aws-logs";

export interface StatelessStackProps extends cdk.StackProps {
  readonly ingestStorageBucketName: string;
  readonly receiverDomain: string;
  readonly extractAttachmentQueueArn: string;
  readonly parseReportQueueArn: string;
  readonly dmarcReportTableName: string;
  readonly dmarcRecordTableName: string;
}

export class StatelessStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: StatelessStackProps) {
    super(scope, id, props);

    const ingestStorageBucket = this.getS3Bucket(props.ingestStorageBucketName);
    const extractAttachmentQueue = this.getSQSQueue(
      props.extractAttachmentQueueArn
    );
    const parseReportQueue = this.getSQSQueue(props.parseReportQueueArn);
    const dmarcReportTable = this.getDynamoDBTable(props.dmarcReportTableName);
    const dmarcRecordTable = this.getDynamoDBTable(props.dmarcRecordTableName);

    // Create SES identity to for SES to establish trust with
    new ses.EmailIdentity(this, "EmailIdentity", {
      identity: ses.Identity.domain(props.receiverDomain),
    });

    // Create receipt rule set to define how SES should handle incoming emails
    const receiptRuleSet = new ses.ReceiptRuleSet(this, "ReceiptRuleSet");

    // Allow SES to write incoming emails to the ingest storage bucket
    const ingestStorageBucketPolicy = new s3.BucketPolicy(
      this,
      "IngestStorageBucketPolicy",
      {
        bucket: ingestStorageBucket,
      }
    );
    ingestStorageBucketPolicy.document.addStatements(
      new iam.PolicyStatement({
        actions: ["s3:PutObject"],
        resources: [`${ingestStorageBucket.bucketArn}/*`],
        principals: [new iam.ServicePrincipal("ses.amazonaws.com")],
        conditions: {
          StringLike: {
            "AWS:SourceAccount": this.account,
            "AWS:SourceArn": `arn:aws:ses:${this.region}:${this.account}:receipt-rule-set/${receiptRuleSet.receiptRuleSetName}:receipt-rule/*`,
          },
        },
      })
    );

    // Create a Lambda function to process incoming emails
    const enqueueEmailFunction = this.createLambdaFunction(
      "EnqueueEmailFunction",
      "../bin/enqueue-email",
      {
        INGEST_STORAGE_BUCKET_NAME: ingestStorageBucket.bucketName,
        NEXT_STAGE_QUEUE_URL: extractAttachmentQueue.queueUrl,
      }
    );
    const enqueueEmailFunctionPolicies: iam.PolicyStatement[] = [
      new iam.PolicyStatement({
        actions: ["sqs:SendMessage"],
        resources: [extractAttachmentQueue.queueArn],
      }),
      new iam.PolicyStatement({
        actions: ["s3:GetObject"],
        resources: [`${ingestStorageBucket.bucketArn}/raw/*`],
      }),
    ];
    this.attachLambdaPolicies(
      enqueueEmailFunction,
      enqueueEmailFunctionPolicies
    );

    // Create a receipt rule for processing incoming emails
    const receiptRule = new ses.ReceiptRule(this, "ReceiptRule", {
      ruleSet: receiptRuleSet,
      recipients: [props.receiverDomain],
      actions: [
        new ses_actions.S3({
          bucket: ingestStorageBucket,
          objectKeyPrefix: "raw/",
        }),
        new ses_actions.Lambda({
          function: enqueueEmailFunction,
        }),
      ],
    });
    receiptRule.node.addDependency(ingestStorageBucketPolicy);

    // Ensure the receipt rule set is active
    this.setActiveReceiptRuleSet(receiptRuleSet.receiptRuleSetName);

    // Create a Lambda function to extract attachments from incoming emails and put them in S3
    const extractAttachmentFunction = this.createLambdaFunction(
      "ExtractAttachmentFunction",
      "../bin/extract-attachment",
      {
        INGEST_STORAGE_BUCKET_NAME: ingestStorageBucket.bucketName,
        NEXT_STAGE_QUEUE_URL: parseReportQueue.queueUrl,
      }
    );

    extractAttachmentFunction.addEventSourceMapping(
      "ExtractAttachmentEventSource",
      {
        eventSourceArn: extractAttachmentQueue.queueArn,
        batchSize: 10,
        maxBatchingWindow: cdk.Duration.seconds(10),
      }
    );
    const extractAttachmentFunctionPolicies: iam.PolicyStatement[] = [
      new iam.PolicyStatement({
        actions: ["s3:GetObject"],
        resources: [`${ingestStorageBucket.bucketArn}/raw/*`],
      }),
      new iam.PolicyStatement({
        actions: ["s3:PutObject"],
        resources: [`${ingestStorageBucket.bucketArn}/reports/*`],
      }),
      new iam.PolicyStatement({
        actions: ["sqs:SendMessage"],
        resources: [parseReportQueue.queueArn],
      }),
      new iam.PolicyStatement({
        actions: [
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ReceiveMessage",
        ],
        resources: [extractAttachmentQueue.queueArn],
      }),
    ];
    this.attachLambdaPolicies(
      extractAttachmentFunction,
      extractAttachmentFunctionPolicies
    );

    // Create a Lambda function to parse DMARC reports and store them in DynamoDB
    const parseReportFunction = this.createLambdaFunction(
      "ParseReportFunction",
      "../bin/parse-report",
      {
        INGEST_STORAGE_BUCKET_NAME: ingestStorageBucket.bucketName,
        INGEST_TABLE_NAME: dmarcReportTable.tableName,
        INGEST_RECORD_TABLE_NAME: dmarcRecordTable.tableName,
      }
    );

    parseReportFunction.addEventSourceMapping("ParseReportEventSource", {
      eventSourceArn: parseReportQueue.queueArn,
      batchSize: 10,
      maxBatchingWindow: cdk.Duration.seconds(10),
    });
    const parseReportFunctionPolicies: iam.PolicyStatement[] = [
      new iam.PolicyStatement({
        actions: ["s3:GetObject"],
        resources: [`${ingestStorageBucket.bucketArn}/reports/*`],
      }),
      new iam.PolicyStatement({
        actions: [
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ReceiveMessage",
        ],
        resources: [parseReportQueue.queueArn],
      }),
      new iam.PolicyStatement({
        actions: ["dynamodb:PutItem", "dynamodb:BatchWriteItem"],
        resources: [dmarcReportTable.tableArn, dmarcRecordTable.tableArn],
      }),
    ];
    this.attachLambdaPolicies(parseReportFunction, parseReportFunctionPolicies);
  }

  private getS3Bucket(bucketName: string): s3.IBucket {
    return s3.Bucket.fromBucketName(this, bucketName, bucketName);
  }

  private getSQSQueue(queueArn: string): sqs.IQueue {
    return sqs.Queue.fromQueueArn(this, queueArn, queueArn);
  }

  private getDynamoDBTable(tableName: string): dynamodb.ITable {
    return dynamodb.Table.fromTableName(this, tableName, tableName);
  }

  private setActiveReceiptRuleSet(receiptRuleSetName: string): void {
    const setActiveReceiptRuleSetSdkCall: cr.AwsSdkCall = {
      service: "SES",
      action: "setActiveReceiptRuleSet",
      physicalResourceId: cr.PhysicalResourceId.of("SesCustomResource"),
      parameters: {
        RuleSetName: receiptRuleSetName,
      },
    };

    const setInactiveReceiptRuleSetSdkCall: cr.AwsSdkCall = {
      service: "SES",
      action: "setActiveReceiptRuleSet",
      physicalResourceId: cr.PhysicalResourceId.of("SesCustomResource"),
    };

    new cr.AwsCustomResource(this, "setActiveReceiptRuleSetCustomResource", {
      onCreate: setActiveReceiptRuleSetSdkCall,
      onUpdate: setActiveReceiptRuleSetSdkCall,
      onDelete: setInactiveReceiptRuleSetSdkCall,
      logRetention: logs.RetentionDays.ONE_WEEK,
      installLatestAwsSdk: true,
      policy: cr.AwsCustomResourcePolicy.fromStatements([
        new iam.PolicyStatement({
          actions: ["ses:SetActiveReceiptRuleSet"],
          resources: ["*"],
        }),
      ]),
    });
  }

  private createLambdaFunction(
    name: string,
    codePath: string,
    environment: { [key: string]: string }
  ): lambda.Function {
    return new lambda.Function(this, name, {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: "bootstrap",
      code: lambda.Code.fromAsset(codePath),
      environment,
    });
  }

  private attachLambdaPolicies(
    lambdaFunction: lambda.Function,
    policies: iam.PolicyStatement[]
  ): void {
    policies.forEach((policy) => lambdaFunction.addToRolePolicy(policy));
  }
}
