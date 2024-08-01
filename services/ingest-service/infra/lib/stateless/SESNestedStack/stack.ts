import { aws_ses_actions, NestedStack, NestedStackProps } from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as sqs from "aws-cdk-lib/aws-sqs";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as iam from "aws-cdk-lib/aws-iam";
import { Construct } from "constructs";
import path from "path";
import { EmailIdentity, Identity, ReceiptRuleSet } from "aws-cdk-lib/aws-ses";
import {
  AwsCustomResource,
  AwsCustomResourcePolicy,
  AwsSdkCall,
  PhysicalResourceId,
} from "aws-cdk-lib/custom-resources";
import { RetentionDays } from "aws-cdk-lib/aws-logs";

export interface SESNestedStackProps extends NestedStackProps {
  readonly ingestStorageBucketName: string;
  readonly receiverDomain: string;
  readonly rawEmailQueueArn: string;
}

export class SESNestedStack extends NestedStack {
  constructor(scope: Construct, id: string, props: SESNestedStackProps) {
    super(scope, id, props);

    // Fetch the dependencies from the parent stack
    const ingestStorageBucket = s3.Bucket.fromBucketName(
      this,
      "IngestStorageBucket",
      props.ingestStorageBucketName
    );

    const rawEmailQueue = sqs.Queue.fromQueueArn(
      this,
      "RawEmailQueue",
      props.rawEmailQueueArn
    );

    // Create an email identity for the receiver domain so SES can receive emails
    new EmailIdentity(this, "EmailIdentity", {
      identity: Identity.domain(props.receiverDomain),
    });

    // Create the receipt rule set to define how SES should handle incoming emails
    const receiptRuleSet = this.createReceiptRuleSet();

    // Create a Lambda function to process incoming emails
    const enqueueEmailFunction = this.createEnqueueEmailFunction(
      rawEmailQueue,
      ingestStorageBucket,
      receiptRuleSet
    );

    const ingestStorageBucketPolicy = new s3.BucketPolicy(
      this,
      "RawEmailBucketPolicy",
      {
        bucket: ingestStorageBucket,
      }
    );

    ingestStorageBucketPolicy.document.addStatements(
      this.createIngestBucketPolicyStatement(
        receiptRuleSet.receiptRuleSetName,
        ingestStorageBucket.bucketArn
      )
    );

    const receiptRule = receiptRuleSet.addRule("IngestRule", {
      recipients: [props.receiverDomain],
      enabled: true,
      actions: [
        new aws_ses_actions.S3({
          bucket: ingestStorageBucket,
          objectKeyPrefix: "raw/",
        }),
        new aws_ses_actions.Lambda({
          function: enqueueEmailFunction,
        }),
      ],
    });

    this.createSetActiveReceiptRuleSetCustomResource(
      receiptRuleSet.receiptRuleSetName
    );

    receiptRule.node.addDependency(ingestStorageBucketPolicy);
  }

  private createEnqueueEmailFunction(
    rawEmailQueue: sqs.IQueue,
    ingestStorageBucket: s3.IBucket,
    receiptRuleSet: ReceiptRuleSet
  ) {
    const enqueueEmailFuncton = new lambda.Function(
      this,
      "EnqueueEmailFunction",
      {
        runtime: lambda.Runtime.PROVIDED_AL2023,
        handler: "bootstrap",
        code: lambda.Code.fromAsset(
          path.join(__dirname, "../../bin/enqueue-email")
        ),
        environment: {
          RAW_EMAIL_QUEUE_URL: rawEmailQueue.queueUrl,
          INGEST_STORAGE_BUCKET_NAME: ingestStorageBucket.bucketName,
          RAW_EMAIL_S3_PREFIX: "raw/",
        },
      }
    );

    enqueueEmailFuncton.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["s3:GetObject"],
        resources: [`${ingestStorageBucket.bucketArn}/raw/*`],
      })
    );

    enqueueEmailFuncton.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["sqs:SendMessage"],
        resources: [rawEmailQueue.queueArn],
      })
    );

    enqueueEmailFuncton.addPermission("SESInvokePermission", {
      principal: new iam.ServicePrincipal("ses.amazonaws.com"),
      sourceArn: `arn:aws:ses:${this.region}:${this.account}:receipt-rule-set/${receiptRuleSet.receiptRuleSetName}:receipt-rule/*`,
      sourceAccount: this.account,
    });

    return enqueueEmailFuncton;
  }

  private createReceiptRuleSet() {
    return new ReceiptRuleSet(this, "ReceiptRuleSet", {
      dropSpam: false,
    });
  }

  private createSetActiveReceiptRuleSetCustomResource(
    receiptRuleSetName: string
  ): void {
    const setActiveReceiptRuleSetSdkCall: AwsSdkCall = {
      service: "SES",
      action: "setActiveReceiptRuleSet",
      physicalResourceId: PhysicalResourceId.of("SesCustomResource"),
      parameters: {
        RuleSetName: receiptRuleSetName,
      },
    };

    const setInactiveReceiptRuleSetSdkCall: AwsSdkCall = {
      service: "SES",
      action: "setActiveReceiptRuleSet",
      physicalResourceId: PhysicalResourceId.of("SesCustomResource"),
    };

    new AwsCustomResource(this, "setActiveReceiptRuleSetCustomResource", {
      onCreate: setActiveReceiptRuleSetSdkCall,
      onUpdate: setActiveReceiptRuleSetSdkCall,
      onDelete: setInactiveReceiptRuleSetSdkCall,
      logRetention: RetentionDays.ONE_WEEK,
      installLatestAwsSdk: true,
      policy: AwsCustomResourcePolicy.fromStatements([
        new iam.PolicyStatement({
          sid: "SesCustomResourceSetActiveReceiptRuleSet",
          effect: iam.Effect.ALLOW,
          actions: ["ses:SetActiveReceiptRuleSet"],
          resources: ["*"],
        }),
      ]),
    });
  }

  private createIngestBucketPolicyStatement(
    receiptRuleSetName: string,
    bucketArn: string
  ): iam.PolicyStatement {
    const { account, region } = NestedStack.of(this);

    return new iam.PolicyStatement({
      sid: "AllowSESPuts",
      actions: ["s3:PutObject"],
      resources: [`${bucketArn}/*`],
      principals: [new iam.ServicePrincipal("ses.amazonaws.com")],
      conditions: {
        StringLike: {
          "AWS:SourceAccount": account,
          "AWS:SourceArn": `arn:aws:ses:${region}:${account}:receipt-rule-set/${receiptRuleSetName}:receipt-rule/*`,
        },
      },
    });
  }
}
