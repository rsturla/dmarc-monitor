import { NestedStack, NestedStackProps } from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as iam from "aws-cdk-lib/aws-iam";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as sqs from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";
import * as path from "path";

export interface ExtractReportNestedStackProps extends NestedStackProps {
  readonly rawEmailBucketName: string;
  readonly rawEmailQueueArn: string;
}

export class ExtractReportNestedStack extends NestedStack {
  constructor(
    scope: Construct,
    id: string,
    props: ExtractReportNestedStackProps
  ) {
    super(scope, id, props);

    const rawEmailBucket = s3.Bucket.fromBucketName(
      this,
      "RawEmailBucket",
      props.rawEmailBucketName
    );
    const rawEmailQueue = sqs.Queue.fromQueueArn(
      this,
      "RawEmailQueue",
      props.rawEmailQueueArn
    );

    const extractReportLambda = new lambda.Function(
      this,
      "ExtractReportLambda",
      {
        runtime: lambda.Runtime.PROVIDED_AL2023,
        handler: "bootstrap",
        code: lambda.Code.fromAsset(
          path.join(__dirname, "../../bin/extract-report")
        ),
        environment: {
          QUEUE_URL: rawEmailQueue.queueUrl,
          BUCKET_NAME: rawEmailBucket.bucketName,
          DYNAMODB_TABLE_NAME: "DMARCReportTable",
        },
      }
    );

    // Add trigger for the lambda to be invoked when a new email is added to the queue
    extractReportLambda.addEventSourceMapping("EmailEventSource", {
      eventSourceArn: props.rawEmailQueueArn,
      batchSize: 1,
    });

    // Grant the lambda permission to read from the report bucket
    extractReportLambda.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["s3:GetObject"],
        resources: [`${rawEmailBucket.bucketArn}/raw/*`],
      })
    );
    extractReportLambda.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["s3:PutObject"],
        resources: [`${rawEmailBucket.bucketArn}/reports/*`],
      })
    );

    extractReportLambda.addToRolePolicy(
      new iam.PolicyStatement({
        actions: [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
        ],
        resources: [props.rawEmailQueueArn],
      })
    );
  }
}
