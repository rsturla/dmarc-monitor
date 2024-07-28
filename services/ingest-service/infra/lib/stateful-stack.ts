import { Stack, StackProps, RemovalPolicy, CfnOutput } from "aws-cdk-lib";
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

    const rawEmailQueue = new Queue(this, "RawEmailQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
    });

    this.rawEmailBucket = rawEmailBucket;
    this.rawEmailQueue = rawEmailQueue;
  }
}
