import * as cdk from "aws-cdk-lib";
import { StatefulStack } from "../lib/stateful-stack";
import { StatelessStack } from "../lib/stateless-stack";

const app = new cdk.App();

new StatefulStack(app, "StatefulStack");
new StatelessStack(app, "StatelessStack");

app.synth();
