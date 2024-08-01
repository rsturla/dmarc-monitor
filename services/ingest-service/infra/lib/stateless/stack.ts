import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import { SESNestedStack } from "./ses-nested-stack";
import { ExtractAttachmentNestedStack } from "./extract-attachment-nested-stack";

export interface StatelessStackProps extends cdk.StackProps {
  readonly ingestStorageBucketName: string;
  readonly receiverDomain: string;
  readonly rawEmailQueueArn: string;
  readonly attachmentQueueArn: string;
  readonly dmarcReportQueueArn: string;
  readonly dmarcReportTableName: string;
}

export class StatelessStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: StatelessStackProps) {
    super(scope, id, props);

    const sesNestedStack = new SESNestedStack(this, "SESNestedStack", {
      ingestStorageBucketName: props.ingestStorageBucketName,
      receiverDomain: props.receiverDomain,
      rawEmailQueueArn: props.rawEmailQueueArn,
    });

    const extractAttachmentNestedStack = new ExtractAttachmentNestedStack(
      this,
      "ExtractAttachmentNestedStack",
      {
        ingestStorageBucketName: props.ingestStorageBucketName,
        rawEmailQueueArn: props.rawEmailQueueArn,
        attachmentQueueArn: props.attachmentQueueArn,
      }
    );
  }
}
