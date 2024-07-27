import * as cdk from "aws-cdk-lib";
import * as iam from "aws-cdk-lib/aws-iam";
import { Construct } from "constructs";

export class StatefulStack extends cdk.Stack {
  public readonly aggregateReportTable: cdk.aws_dynamodb.Table;
  public readonly aggregateReportS3Bucket: cdk.aws_s3.Bucket;
  public readonly aggregateReportSNSTopic: cdk.aws_sns.Topic;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const aggregateReportTable = new cdk.aws_dynamodb.Table(
      this,
      "AggregateReportTable",
      {
        partitionKey: {
          name: "id",
          type: cdk.aws_dynamodb.AttributeType.STRING,
        },
        removalPolicy: cdk.RemovalPolicy.DESTROY,
      }
    );

    const aggregateReportS3Bucket = new cdk.aws_s3.Bucket(
      this,
      "AggregateReportS3Bucket",
      {
        removalPolicy: cdk.RemovalPolicy.DESTROY,
        blockPublicAccess: cdk.aws_s3.BlockPublicAccess.BLOCK_ALL,
      }
    );

    const aggregateReportSNSTopic = new cdk.aws_sns.Topic(
      this,
      "AggregateReportSNSTopic"
    );

    new cdk.CfnOutput(this, "AggregateReportTableOutput", {
      value: aggregateReportTable.tableName,
    });

    new cdk.CfnOutput(this, "AggregateReportS3BucketOutput", {
      value: aggregateReportS3Bucket.bucketName,
    });

    new cdk.CfnOutput(this, "AggregateReportSNSTopicOutput", {
      value: aggregateReportSNSTopic.topicArn,
    });

    this.aggregateReportTable = aggregateReportTable;
    this.aggregateReportS3Bucket = aggregateReportS3Bucket;
    this.aggregateReportSNSTopic = aggregateReportSNSTopic;
  }
}
