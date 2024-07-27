import * as cdk from "aws-cdk-lib";
import { StatefulStack } from "../lib/stateful-stack";
import { StatelessStack } from "../lib/stateless-stack";

const app = new cdk.App();

const statefulStack = new StatefulStack(app, "StatefulStack");
new StatelessStack(app, "StatelessStack", {
  aggregateReportTableName: statefulStack.aggregateReportTable.tableName,
  aggregateReportS3BucketName: statefulStack.aggregateReportS3Bucket.bucketName,
  aggregateReportSNSTopicName: statefulStack.aggregateReportSNSTopic.topicName,
  // Get the receiver domain from environment variables
  receiverDomain: process.env.RECEIVER_DOMAIN!,
});

app.synth();
