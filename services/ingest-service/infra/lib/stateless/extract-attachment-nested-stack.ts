import * as cdk from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as sqs from "aws-cdk-lib/aws-sqs";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as iam from "aws-cdk-lib/aws-iam";
import { Construct } from "constructs";

export interface ExtractAttachmentNestedStackProps
  extends cdk.NestedStackProps {
  readonly ingestStorageBucketName: string;
  readonly rawEmailQueueArn: string;
  readonly attachmentQueueArn: string;
}

export class ExtractAttachmentNestedStack extends cdk.Stack {
  constructor(
    scope: Construct,
    id: string,
    props: ExtractAttachmentNestedStackProps
  ) {
    super(scope, id, props);

    const ingestStorageBucket = s3.Bucket.fromBucketName(
      this,
      "IngestStorageBucket",
      props.ingestStorageBucketName
    );
    const rawEmailQueue = this.getSQSQueue(props.rawEmailQueueArn);
    const attachmentQueue = this.getSQSQueue(props.attachmentQueueArn);

    // Create a Lambda function to extract attachments from incoming emails
    const extractAttachmentFunction = new lambda.Function(
      this,
      "ExtractAttachmentFunction",
      {
        runtime: lambda.Runtime.PROVIDED_AL2023,
        handler: "bootstrap",
        code: lambda.Code.fromAsset("../../bin/extract-attachment"),
        environment: {
          INGEST_STORAGE_BUCKET_NAME: ingestStorageBucket.bucketName,
          ATTACHMENT_QUEUE_URL: attachmentQueue.queueUrl,
        },
      }
    );

    extractAttachmentFunction.addEventSourceMapping("RawEmailEventSource", {
      eventSourceArn: rawEmailQueue.queueArn,
      batchSize: 10,
      bisectBatchOnError: true,
      maxBatchingWindow: cdk.Duration.seconds(60),
    });

    extractAttachmentFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["s3:GetObject"],
        resources: [`${ingestStorageBucket.bucketArn}/raw/*`],
      })
    );

    extractAttachmentFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["s3:PutObject"],
        resources: [`${ingestStorageBucket.bucketArn}/reports/*`],
      })
    );

    extractAttachmentFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["sqs:SendMessage"],
        resources: [attachmentQueue.queueArn],
      })
    );

    extractAttachmentFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: [
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ReceiveMessage",
        ],
        resources: [rawEmailQueue.queueArn],
      })
    );
  }

  private getSQSQueue(queueArn: string): sqs.IQueue {
    return sqs.Queue.fromQueueArn(this, "AttachmentQueue", queueArn);
  }
}
