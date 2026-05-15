import { RemovalPolicy, Stack, StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { DynamoDBTable } from "cdk-constructs/datastores";
import { SQSQueue } from "cdk-constructs/messaging";
import { AttributeType } from "aws-cdk-lib/aws-dynamodb";
import { QueueEncryption } from "aws-cdk-lib/aws-sqs";

export class StatefulStack extends Stack {
  public readonly tenantTable: DynamoDBTable;
  public readonly createTenantQueue: SQSQueue;
  public readonly deleteTenantQueue: SQSQueue;
  public readonly updateTenantQueue: SQSQueue;

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const tenantTable = new DynamoDBTable(this, "TenantTable", {
      partitionKey: {
        name: "id",
        type: AttributeType.STRING,
      },
      removalPolicy: RemovalPolicy.DESTROY,
    });

    const createTenantQueue = new SQSQueue(this, "CreateTenantQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
      enableDeadLetterQueue: true,
    });

    const deleteTenantQueue = new SQSQueue(this, "DeleteTenantQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
      enableDeadLetterQueue: true,
    });

    const updateTenantQueue = new SQSQueue(this, "UpdateTenantQueue", {
      encryption: QueueEncryption.SQS_MANAGED,
      enableDeadLetterQueue: true,
    });

    this.tenantTable = tenantTable;
    this.createTenantQueue = createTenantQueue;
    this.deleteTenantQueue = deleteTenantQueue;
    this.updateTenantQueue = updateTenantQueue;
  }
}
