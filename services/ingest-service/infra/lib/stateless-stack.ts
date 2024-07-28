import { StackProps, Stack } from "aws-cdk-lib";
import { Construct } from "constructs";
import { SESNestedStack } from "./ses-nested-stack";
import { ExtractReportNestedStack } from "./extract-report-nested-stack";

export interface StatelessStackProps extends StackProps {
  readonly rawEmailQueueArn: string;
  readonly rawEmailBucketName: string;
  readonly receiverDomain: string;
}

export class StatelessStack extends Stack {
  constructor(scope: Construct, id: string, props: StatelessStackProps) {
    super(scope, id, props);

    const sesNestedStack = new SESNestedStack(this, "SESNestedStack", {
      rawEmailBucketName: props.rawEmailBucketName,
      receiverDomain: props.receiverDomain,
      rawEmailQueueArn: props.rawEmailQueueArn,
    });

    const extractReportNestedStack = new ExtractReportNestedStack(
      this,
      "ExtractReportNestedStack",
      {
        rawEmailBucketName: props.rawEmailBucketName,
        rawEmailQueueArn: props.rawEmailQueueArn,
      }
    );
  }
}
