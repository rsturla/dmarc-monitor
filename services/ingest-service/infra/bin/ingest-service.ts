import * as cdk from "aws-cdk-lib";
import { StatefulStack } from "../lib/stateful-stack";
import { StatelessStack } from "../lib/stateless-stack";

const app = new cdk.App();
const appName = "dmarc-monitor-ingest-service";

const statefulStack = new StatefulStack(app, `${appName}-StatefulStack`);

new StatelessStack(app, `${appName}-StatelessStack`, {
  rawEmailQueueArn: statefulStack.rawEmailQueue.queueArn,
  rawEmailBucketName: statefulStack.rawEmailBucket.bucketName,
  receiverDomain: process.env.RECEIVER_DOMAIN || "dm.sturla.tech",
});

app.synth();
