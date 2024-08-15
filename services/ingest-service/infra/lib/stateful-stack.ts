import { Stack, StackProps, RemovalPolicy, Duration } from "aws-cdk-lib";
import { Bucket, BlockPublicAccess, StorageClass } from "aws-cdk-lib/aws-s3";
import { QueueEncryption } from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";
import { DynamoDBTable, S3Bucket } from "cdk-constructs/datastores";
import { SQSQueue } from "cdk-constructs/messaging";
import { AttributeType } from "aws-cdk-lib/aws-dynamodb";

export class StatefulStack extends Stack {
  public readonly ingestStorageBucket: Bucket;

  public readonly extractAttachmentQueue: SQSQueue;
  public readonly parseReportQueue: SQSQueue;

  public readonly dmarcReportTable: DynamoDBTable;
  public readonly dmarcRecordTable: DynamoDBTable;

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

    // ExtractAttachmentQueue: Messages added when SES receives an email.
    // Messages point to the S3 object containing the email
    const extractAttachmentQueue = new SQSQueue(
      this,
      "ExtractAttachmentQueue",
      {
        encryption: QueueEncryption.SQS_MANAGED,
        enableDeadLetterQueue: true,
      }
    );

    // parseReportQueue: Messages added when a Lambda function has extracted an attachment from an email
    // Messages point to the S3 object containing the extracted XML attachment
    const parseReportQueue = new SQSQueue(this, "ParseReportQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
      enableDeadLetterQueue: true,
    });

    const dmarcReportTable = new DynamoDBTable(this, "DmarcReportTable", {
      partitionKey: {
        name: "id",
        type: AttributeType.STRING,
      },
    });
    const dmarcRecordTable = new DynamoDBTable(this, "DmarcRecordTable", {
      partitionKey: {
        name: "id",
        type: AttributeType.STRING,
      },
    });

    this.ingestStorageBucket = ingestStorageBucket;
    this.extractAttachmentQueue = extractAttachmentQueue;
    this.parseReportQueue = parseReportQueue;
    this.dmarcReportTable = dmarcReportTable;
    this.dmarcRecordTable = dmarcRecordTable;
  }
}
