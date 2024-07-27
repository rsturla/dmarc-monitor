import * as cdk from "aws-cdk-lib";
import { Bucket } from "aws-cdk-lib/aws-s3";
import * as iam from "aws-cdk-lib/aws-iam";
import * as ses from "aws-cdk-lib/aws-ses";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as cr from "aws-cdk-lib/custom-resources";
import * as actions from "aws-cdk-lib/aws-ses-actions";
import * as sns from "aws-cdk-lib/aws-sns";
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";
import { RetentionDays } from "aws-cdk-lib/aws-logs";
import path from "path";

export interface StatelessStackProps extends cdk.StackProps {
  readonly aggregateReportTableName: string;
  readonly aggregateReportS3BucketName: string;
  readonly aggregateReportSNSTopicName: string;

  readonly receiverDomain: string;
}

export class StatelessStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: StatelessStackProps) {
    super(scope, id, props);

    // Import the aggregate report S3 bucket
    const aggregateReportS3Bucket = Bucket.fromBucketName(
      this,
      "AggregateReportS3Bucket",
      props.aggregateReportS3BucketName
    );

    const aggregateReportSNSTopic = sns.Topic.fromTopicArn(
      this,
      "AggregateReportSNSTopic",
      props.aggregateReportSNSTopicName
    );

    const receiptRuleSet = new ses.ReceiptRuleSet(
      this,
      "IngestServiceRuleSet",
      {
        dropSpam: false,
      }
    );

    const aggregateReportS3BucketPolicy = new s3.BucketPolicy(
      this,
      "AggregateReportS3BucketPolicy",
      {
        bucket: aggregateReportS3Bucket,
      }
    );

    aggregateReportS3BucketPolicy.document.addStatements(
      new iam.PolicyStatement({
        sid: "AllowSESPuts",
        actions: ["s3:PutObject"],
        resources: [`${aggregateReportS3Bucket.bucketArn}/*`],
        principals: [new iam.ServicePrincipal("ses.amazonaws.com")],
        conditions: {
          StringLike: {
            "AWS:SourceAccount": cdk.Stack.of(this).account,
            "AWS:SourceArn": `arn:aws:ses:${cdk.Stack.of(this).region}:${
              cdk.Stack.of(this).account
            }:receipt-rule-set/${
              receiptRuleSet.receiptRuleSetName
            }:receipt-rule/*`,
          },
        },
      })
    );

    // Create Lambda function to process incoming emails (./bin/enqueue-reports)
    const enqueueReportFunction = new lambda.Function(
      this,
      "EnqueueReportFunction",
      {
        runtime: lambda.Runtime.PROVIDED_AL2023,
        handler: "bootstrap",
        code: lambda.Code.fromAsset(
          path.join(__dirname, "../../bin/enqueue-report")
        ),
        environment: {},
      }
    );

    enqueueReportFunction.addPermission("InvokeBySES", {
      principal: new iam.ServicePrincipal("ses.amazonaws.com"),
      sourceArn: `arn:aws:ses:${cdk.Stack.of(this).region}:${
        cdk.Stack.of(this).account
      }:receipt-rule-set/${receiptRuleSet.receiptRuleSetName}:receipt-rule/*`,
      sourceAccount: cdk.Stack.of(this).account,
    });

    const receiptRule = receiptRuleSet.addRule("IngestServiceRule", {
      recipients: [props.receiverDomain],
      enabled: true,
      actions: [
        new actions.S3({
          bucket: aggregateReportS3Bucket,
          objectKeyPrefix: "emails/",
        }),
        new actions.Sns({
          topic: {
            ...aggregateReportSNSTopic,
            // Due to a bug I have not been able to identify, the topicArn is outputting the
            // topic name instead of the ARN. This is a workaround to get the ARN.
            topicArn: `arn:aws:sns:${cdk.Stack.of(this).region}:${
              cdk.Stack.of(this).account
            }:${aggregateReportSNSTopic.topicArn}`,
          },
          encoding: actions.EmailEncoding.BASE64,
        }),
        new actions.Lambda({
          function: enqueueReportFunction,
        }),
      ],
    });

    const setActiveReceiptRuleSetSdkCall: cr.AwsSdkCall = {
      service: "SES",
      action: "setActiveReceiptRuleSet",
      physicalResourceId: cr.PhysicalResourceId.of("SesCustomResource"),
      parameters: {
        RuleSetName: receiptRuleSet.receiptRuleSetName,
      },
    };

    new cr.AwsCustomResource(this, "setActiveReceiptRuleSetCustomResource", {
      onCreate: setActiveReceiptRuleSetSdkCall,
      onUpdate: setActiveReceiptRuleSetSdkCall,
      logRetention: RetentionDays.ONE_WEEK,
      policy: cr.AwsCustomResourcePolicy.fromStatements([
        new iam.PolicyStatement({
          sid: "SesCustomResourceSetActiveReceiptRuleSet",
          effect: iam.Effect.ALLOW,
          actions: ["ses:SetActiveReceiptRuleSet"],
          resources: ["*"],
        }),
      ]),
    });

    // Add dependency on the bucket policy to ensure it is created before the rule
    receiptRule.node.addDependency(aggregateReportS3BucketPolicy);

    new ses.EmailIdentity(this, "IngestServiceIdentity", {
      identity: ses.Identity.domain(props.receiverDomain),
    });
  }
}
