import * as cdk from "aws-cdk-lib";
import { StatefulStack } from "../lib/stateful-stack";
import { StatelessStack } from "../lib/stateless-stack";

const app = new cdk.App();

const statefulStack = new StatefulStack(app, "StatefulStack");
new StatelessStack(app, "StatelessStack", {
  aggregateReportTableName: statefulStack.aggregateReportTable.tableName,
  aggregateReportS3BucketName: statefulStack.aggregateReportS3Bucket.bucketName,
  receiverDomain: process.env.RECEIVER_DOMAIN!,
});

app.synth();
