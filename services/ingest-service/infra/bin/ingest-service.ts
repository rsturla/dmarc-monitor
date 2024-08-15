import * as cdk from "aws-cdk-lib";
import { StatefulStack } from "../lib/stateful-stack";
import { StatelessStack } from "../lib/stateless";

const app = new cdk.App();
const appName = "dmarc-monitor-ingest-service";

const statefulStack = new StatefulStack(app, `${appName}-StatefulStack`);

new StatelessStack(app, `${appName}-StatelessStack`, {
  ingestStorageBucketName: statefulStack.ingestStorageBucket.bucketName,
  receiverDomain: process.env.RECEIVER_DOMAIN || "dm.sturla.tech",
  extractAttachmentQueueArn: statefulStack.extractAttachmentQueue.queueArn,
  parseReportQueueArn: statefulStack.parseReportQueue.queueArn,
  dmarcReportTableName: statefulStack.dmarcReportTable.tableName,
  dmarcRecordTableName: statefulStack.dmarcRecordTable.tableName,
});

app.synth();
