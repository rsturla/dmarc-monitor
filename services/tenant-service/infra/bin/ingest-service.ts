import * as cdk from "aws-cdk-lib";
import { StatefulStack } from "../lib/stateful-stack";
import { StatelessStack } from "../lib/stateless-stack";

const app = new cdk.App();
const appName = "dmarc-monitor-tenant-service";

const statefulStack = new StatefulStack(app, `${appName}-StatefulStack`);
new StatelessStack(app, `${appName}-StatelessStack`, {
  tenantTableArn: statefulStack.tenantTable.tableArn,
  createTenantQueueArn: statefulStack.createTenantQueue.queueArn,
  deleteTenantQueueArn: statefulStack.deleteTenantQueue.queueArn,
  updateTenantQueueArn: statefulStack.updateTenantQueue.queueArn,
});

app.synth();
