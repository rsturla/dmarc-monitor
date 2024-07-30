import { NestedStack, NestedStackProps } from "aws-cdk-lib";
import { Effect, PolicyStatement, ServicePrincipal } from "aws-cdk-lib/aws-iam";
import { RetentionDays } from "aws-cdk-lib/aws-logs";
import { Bucket, BucketPolicy } from "aws-cdk-lib/aws-s3";
import { EmailIdentity, Identity, ReceiptRuleSet } from "aws-cdk-lib/aws-ses";
import * as actions from "aws-cdk-lib/aws-ses-actions";
import * as lambda from "aws-cdk-lib/aws-lambda";
import {
  AwsCustomResource,
  AwsCustomResourcePolicy,
  AwsSdkCall,
  PhysicalResourceId,
} from "aws-cdk-lib/custom-resources";
import { Construct } from "constructs";
import * as path from "path";
import { Queue } from "aws-cdk-lib/aws-sqs";

export interface SESNestedStackProps extends NestedStackProps {
  readonly rawEmailBucketName: string;
  readonly receiverDomain: string;
  readonly rawEmailQueueArn: string;
}

export class SESNestedStack extends NestedStack {
  public readonly receiptRuleSet: ReceiptRuleSet;

  constructor(scope: Construct, id: string, props: SESNestedStackProps) {
    super(scope, id, props);

    const rawEmailBucket = Bucket.fromBucketName(
      this,
      "RawEmailBucket",
      props.rawEmailBucketName
    );
    const rawEmailQueue = Queue.fromQueueArn(
      this,
      "RawEmailQueue",
      props.rawEmailQueueArn
    );

    // Create Lambda function to process incoming emails (../../bin/enqueue-reports)
    const enqueueRawEmailFunction = new lambda.Function(
      this,
      "EnqueueRawEmailFunction",
      {
        runtime: lambda.Runtime.PROVIDED_AL2023,
        handler: "bootstrap",
        code: lambda.Code.fromAsset(
          path.join(__dirname, "../../bin/enqueue-mail")
        ),
        environment: {
          RAW_EMAIL_QUEUE_URL: rawEmailQueue.queueUrl,
          REPORT_STORAGE_BUCKET_NAME: rawEmailBucket.bucketName,
        },
      }
    );

    // Add permissions to HeadObject on raw/ prefix in the bucket
    enqueueRawEmailFunction.addToRolePolicy(
      new PolicyStatement({
        actions: ["s3:GetObject"],
        resources: [`${rawEmailBucket.bucketArn}/raw/*`],
      })
    );

    // Add permissions to send messages to the queue and to send messages to the queue
    rawEmailQueue.grantSendMessages(enqueueRawEmailFunction);
    enqueueRawEmailFunction.role?.addManagedPolicy;

    // Create the email identity
    new EmailIdentity(this, "IngestServiceIdentity", {
      identity: Identity.domain(props.receiverDomain),
    });

    // Create the receipt rule set
    const receiptRuleSet = new ReceiptRuleSet(this, "ReceiptRuleSet", {
      dropSpam: false,
    });

    enqueueRawEmailFunction.addPermission("SESInvokeLambdaPermission", {
      principal: new ServicePrincipal("ses.amazonaws.com"),
      sourceArn: `arn:aws:ses:${this.region}:${this.account}:receipt-rule-set/${receiptRuleSet.receiptRuleSetName}:receipt-rule/*`,
      sourceAccount: this.account,
    });

    const rawEmailBucketPolicy = new BucketPolicy(
      this,
      "RawEmailBucketPolicy",
      {
        bucket: rawEmailBucket,
      }
    );

    rawEmailBucketPolicy.document.addStatements(
      this.createS3BucketPolicyStatement(
        receiptRuleSet.receiptRuleSetName,
        rawEmailBucket.bucketArn
      )
    );

    const receiptRule = receiptRuleSet.addRule("RawEmail", {
      recipients: [props.receiverDomain],
      enabled: true,
      actions: [
        new actions.S3({
          bucket: rawEmailBucket,
          objectKeyPrefix: "raw/",
        }),
        new actions.Lambda({
          function: enqueueRawEmailFunction,
        }),
      ],
    });

    this.createSetActiveReceiptRuleSetCustomResource(
      receiptRuleSet.receiptRuleSetName
    );

    // Add dependency on the policies to ensure they are created before the SES rule
    receiptRule.node.addDependency(rawEmailBucketPolicy);

    this.receiptRuleSet = receiptRuleSet;
  }

  private createS3BucketPolicyStatement(
    receiptRuleSetName: string,
    bucketArn: string
  ): PolicyStatement {
    const { account, region } = NestedStack.of(this);

    return new PolicyStatement({
      sid: "AllowSESPuts",
      actions: ["s3:PutObject"],
      resources: [`${bucketArn}/*`],
      principals: [new ServicePrincipal("ses.amazonaws.com")],
      conditions: {
        StringLike: {
          "AWS:SourceAccount": account,
          "AWS:SourceArn": `arn:aws:ses:${region}:${account}:receipt-rule-set/${receiptRuleSetName}:receipt-rule/*`,
        },
      },
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
        new PolicyStatement({
          sid: "SesCustomResourceSetActiveReceiptRuleSet",
          effect: Effect.ALLOW,
          actions: ["ses:SetActiveReceiptRuleSet"],
          resources: ["*"],
        }),
      ]),
    });
  }
}
