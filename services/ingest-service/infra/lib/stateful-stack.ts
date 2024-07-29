import {
  Stack,
  StackProps,
  RemovalPolicy,
  CfnOutput,
  Duration,
} from "aws-cdk-lib";
import { Bucket, BlockPublicAccess } from "aws-cdk-lib/aws-s3";
import { Queue, QueueEncryption } from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";

export class StatefulStack extends Stack {
  public readonly rawEmailBucket: Bucket;
  public readonly rawEmailQueue: Queue;

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const rawEmailBucket = new Bucket(this, "RawEmailBucket", {
      removalPolicy: RemovalPolicy.DESTROY,
      blockPublicAccess: BlockPublicAccess.BLOCK_ALL,
    });

    // Automatically delete objects from the raw email bucket after 30 days
    rawEmailBucket.addLifecycleRule({
      expiration: Duration.days(30),
      prefix: "raw/",
    });

    const rawEmailDqlQueue = new Queue(this, "RawEmailDqlQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
    });

    const rawEmailQueue = new Queue(this, "RawEmailQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
      deadLetterQueue: {
        queue: rawEmailDqlQueue,
        maxReceiveCount: 3,
      },
    });

    this.rawEmailBucket = rawEmailBucket;
    this.rawEmailQueue = rawEmailQueue;
  }
}
