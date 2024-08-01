import { Stack, StackProps, RemovalPolicy, Duration } from "aws-cdk-lib";
import { Bucket, BlockPublicAccess, StorageClass } from "aws-cdk-lib/aws-s3";
import { QueueEncryption } from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";
import { DynamoDBTable, S3Bucket } from "cdk-constructs/datastores";
import { SQSQueue } from "cdk-constructs/messaging";
import { AttributeType } from "aws-cdk-lib/aws-dynamodb";

export class StatefulStack extends Stack {
  public readonly ingestStorageBucket: Bucket;
  public readonly rawEmailQueue: SQSQueue;
  public readonly attachmentQueue: SQSQueue;
  public readonly dmarcReportQueue: SQSQueue;
  public readonly dmarcReportTable: DynamoDBTable;

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const ingestStorageBucket = new S3Bucket(this, "IngestStorageBucket", {
      removalPolicy: RemovalPolicy.DESTROY,
      blockPublicAccess: BlockPublicAccess.BLOCK_ALL,
      lifecycleRules: [
        {
          prefix: "emails/",
          expiration: Duration.days(180),
          transitions: [
            {
              storageClass: StorageClass.INFREQUENT_ACCESS,
              transitionAfter: Duration.days(30),
            },
            {
              storageClass: StorageClass.GLACIER,
              transitionAfter: Duration.days(90),
            },
          ],
        },
        {
          prefix: "reports/",
          expiration: Duration.days(365),
          transitions: [
            {
              storageClass: StorageClass.INFREQUENT_ACCESS,
              transitionAfter: Duration.days(30),
            },
            {
              storageClass: StorageClass.GLACIER,
              transitionAfter: Duration.days(90),
            },
          ],
        },
      ],
    });

    // rawEmailQueue: Messages added when SES receives an email.
    // Messages point to the S3 object containing the email
    const rawEmailQueue = new SQSQueue(this, "RawEmailQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
      enableDeadLetterQueue: true,
    });

    // attachmentQueue: Messages added when a Lambda function has extracted an attachment from an email
    // Messages point to the S3 object containing the extracted XML attachment
    const attachmentQueue = new SQSQueue(this, "AttachmentQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
      enableDeadLetterQueue: true,
    });

    // dmarcReportQueue: Messages added when a Lambda function has parsed a DMARC report attachment and stored the data in DynamoDB
    // Messages point to the DynamoDB primary key of the stored DMARC report
    const dmarcReportQueue = new SQSQueue(this, "DmarcReportQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
      enableDeadLetterQueue: true,
    });

    const dmarcReportTable = new DynamoDBTable(this, "DmarcReportTable", {
      partitionKey: {
        name: "id",
        type: AttributeType.STRING,
      },
    });

    this.ingestStorageBucket = ingestStorageBucket;
    this.rawEmailQueue = rawEmailQueue;
    this.attachmentQueue = attachmentQueue;
    this.dmarcReportQueue = dmarcReportQueue;
    this.dmarcReportTable = dmarcReportTable;
  }
}
