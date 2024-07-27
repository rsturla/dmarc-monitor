import { Stack, StackProps, RemovalPolicy, CfnOutput } from "aws-cdk-lib";
import { Table, AttributeType } from "aws-cdk-lib/aws-dynamodb";
import { Bucket, BlockPublicAccess } from "aws-cdk-lib/aws-s3";
import { Queue, QueueEncryption } from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";

export class StatefulStack extends Stack {
  public readonly aggregateReportTable: Table;
  public readonly aggregateReportS3Bucket: Bucket;
  public readonly rawEmailsQueue: Queue;

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const aggregateReportTable = new Table(this, "AggregateReportTable", {
      partitionKey: {
        name: "id",
        type: AttributeType.STRING,
      },
      removalPolicy: RemovalPolicy.DESTROY,
    });

    const aggregateReportS3Bucket = new Bucket(
      this,
      "AggregateReportS3Bucket",
      {
        removalPolicy: RemovalPolicy.DESTROY,
        blockPublicAccess: BlockPublicAccess.BLOCK_ALL,
      }
    );

    const rawEmailsQueue = new Queue(this, "RawEmailsQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
    });

    new CfnOutput(this, "AggregateReportTableOutput", {
      value: aggregateReportTable.tableName,
    });

    new CfnOutput(this, "AggregateReportS3BucketOutput", {
      value: aggregateReportS3Bucket.bucketName,
    });

    new CfnOutput(this, "RawEmailsQueueOutput", {
      value: rawEmailsQueue.queueName,
    });

    this.aggregateReportTable = aggregateReportTable;
    this.aggregateReportS3Bucket = aggregateReportS3Bucket;
    this.rawEmailsQueue = rawEmailsQueue;
  }
}
