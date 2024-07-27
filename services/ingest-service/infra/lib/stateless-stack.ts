import { StackProps, Stack } from "aws-cdk-lib";
import { Bucket } from "aws-cdk-lib/aws-s3";
import { ServicePrincipal, PolicyStatement, Effect } from "aws-cdk-lib/aws-iam";
import { ReceiptRuleSet, Identity, EmailIdentity } from "aws-cdk-lib/aws-ses";
import { BucketPolicy } from "aws-cdk-lib/aws-s3";
import {
  AwsSdkCall,
  AwsCustomResource,
  AwsCustomResourcePolicy,
  PhysicalResourceId,
} from "aws-cdk-lib/custom-resources";
import * as actions from "aws-cdk-lib/aws-ses-actions";
import { Function, Runtime, Code } from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";
import { RetentionDays } from "aws-cdk-lib/aws-logs";
import * as path from "path";

export interface StatelessStackProps extends StackProps {
  readonly aggregateReportTableName: string;
  readonly aggregateReportS3BucketName: string;
  readonly receiverDomain: string;
}

export class StatelessStack extends Stack {
  constructor(scope: Construct, id: string, props: StatelessStackProps) {
    super(scope, id, props);

    const { account, region } = Stack.of(this);

    // Import the aggregate report S3 bucket
    const aggregateReportS3Bucket = Bucket.fromBucketName(
      this,
      "AggregateReportS3Bucket",
      props.aggregateReportS3BucketName
    );

    // Create the receipt rule set
    const receiptRuleSet = new ReceiptRuleSet(this, "ReceiptRuleSet", {
      dropSpam: false,
    });

    // Create the S3 bucket policy for SES put object
    const aggregateReportS3BucketPolicy = new BucketPolicy(
      this,
      "AggregateReportS3BucketPolicy",
      {
        bucket: aggregateReportS3Bucket,
      }
    );

    aggregateReportS3BucketPolicy.document.addStatements(
      this.createS3BucketPolicyStatement(
        receiptRuleSet.receiptRuleSetName,
        aggregateReportS3Bucket.bucketArn
      )
    );

    // Create Lambda function to process incoming emails (./bin/enqueue-reports)
    const enqueueReportFunction = new Function(this, "EnqueueReportFunction", {
      runtime: Runtime.PROVIDED_AL2023,
      handler: "bootstrap",
      code: Code.fromAsset(path.join(__dirname, "../../bin/enqueue-report")),
      environment: {},
    });

    enqueueReportFunction.addPermission("InvokeBySES", {
      principal: new ServicePrincipal("ses.amazonaws.com"),
      sourceArn: `arn:aws:ses:${region}:${account}:receipt-rule-set/${receiptRuleSet.receiptRuleSetName}:receipt-rule/*`,
      sourceAccount: account,
    });

    // Create the receipt rule
    const receiptRule = receiptRuleSet.addRule("IngestServiceRule", {
      recipients: [props.receiverDomain],
      enabled: true,
      actions: [
        new actions.S3({
          bucket: aggregateReportS3Bucket,
          objectKeyPrefix: "emails/",
        }),
        new actions.Lambda({
          function: enqueueReportFunction,
        }),
      ],
    });

    // Set the active receipt rule set using a custom resource
    this.createSetActiveReceiptRuleSetCustomResource(
      receiptRuleSet.receiptRuleSetName
    );

    // Add dependency on the bucket policy to ensure it is created before the rule
    receiptRule.node.addDependency(aggregateReportS3BucketPolicy);

    // Create the email identity
    new EmailIdentity(this, "IngestServiceIdentity", {
      identity: Identity.domain(props.receiverDomain),
    });
  }

  private createS3BucketPolicyStatement(
    receiptRuleSetName: string,
    bucketArn: string
  ): PolicyStatement {
    const { account, region } = Stack.of(this);

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

    new AwsCustomResource(this, "setActiveReceiptRuleSetCustomResource", {
      onCreate: setActiveReceiptRuleSetSdkCall,
      onUpdate: setActiveReceiptRuleSetSdkCall,
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
