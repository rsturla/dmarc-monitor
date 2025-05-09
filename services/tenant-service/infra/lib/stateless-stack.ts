import { Size, Stack, StackProps, Aws } from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as apigateway from "aws-cdk-lib/aws-apigateway";
import * as sqs from "aws-cdk-lib/aws-sqs";
import * as iam from "aws-cdk-lib/aws-iam";
import { Construct } from "constructs";

export interface StatelessStackProps extends StackProps {
  tenantTableArn: string;
  createTenantQueueArn: string;
  deleteTenantQueueArn: string;
  updateTenantQueueArn: string;
}

export class StatelessStack extends Stack {
  constructor(scope: Construct, id: string, props: StatelessStackProps) {
    super(scope, id, props);

    const tenantTable = this.getDynamoDBTable(props.tenantTableArn);
    const createTenantQueue = this.getSQSQueue(props.createTenantQueueArn);
    const deleteTenantQueue = this.getSQSQueue(props.deleteTenantQueueArn);
    const updateTenantQueue = this.getSQSQueue(props.updateTenantQueueArn);

    const integrationRole = new iam.Role(this, "IntegrationRole", {
      assumedBy: new iam.ServicePrincipal("apigateway.amazonaws.com"),
    });

    tenantTable.grantReadWriteData(integrationRole);
    createTenantQueue.grantSendMessages(integrationRole);
    deleteTenantQueue.grantSendMessages(integrationRole);
    updateTenantQueue.grantSendMessages(integrationRole);

    const tenantApi = new apigateway.RestApi(this, "TenantApi", {
      restApiName: "Tenant Service",
      deploy: true,
      defaultMethodOptions: {
        authorizationType: apigateway.AuthorizationType.IAM,
      },
      minCompressionSize: Size.kibibytes(512), // 512 KiB = 524,288 bytes
    });

    // Create /tenants/{tenantId} resource
    const tenantResource = tenantApi.root.addResource("tenants");
    const tenantIdResource = tenantResource.addResource("{tenantId}");

    // POST /tenants
    tenantResource.addMethod(
      "POST",
      new apigateway.AwsIntegration({
        service: "sqs",
        path: `${Aws.ACCOUNT_ID}/${createTenantQueue.queueName}`,
        integrationHttpMethod: "POST",
        options: {
          credentialsRole: integrationRole,
          integrationResponses: [
            {
              statusCode: "200",
              responseTemplates: {
                "application/json": JSON.stringify({
                  message: "Tenant created successfully",
                }),
              },
            },
          ],
        },
      })
    );

    // DELETE /tenants/{tenantId}
    tenantIdResource.addMethod(
      "DELETE",
      new apigateway.AwsIntegration({
        service: "sqs",
        path: `${Aws.ACCOUNT_ID}/${deleteTenantQueue.queueName}`,
        integrationHttpMethod: "POST",
        options: {
          credentialsRole: integrationRole,
          integrationResponses: [
            {
              statusCode: "200",
              responseTemplates: {
                "application/json": JSON.stringify({
                  message: "Tenant deleted successfully",
                }),
              },
            },
          ],
        },
      })
    );

    // PUT /tenants/{tenantId}
    tenantIdResource.addMethod(
      "PUT",
      new apigateway.AwsIntegration({
        service: "sqs",
        path: `${Aws.ACCOUNT_ID}/${updateTenantQueue.queueName}`,
        integrationHttpMethod: "POST",
        options: {
          credentialsRole: integrationRole,
          integrationResponses: [
            {
              statusCode: "200",
              responseTemplates: {
                "application/json": JSON.stringify({
                  message: "Tenant updated successfully",
                }),
              },
            },
          ],
        },
      })
    );
  }

  private getSQSQueue(queueArn: string): sqs.IQueue {
    return sqs.Queue.fromQueueArn(this, queueArn, queueArn);
  }

  private getDynamoDBTable(tableName: string): dynamodb.ITable {
    return dynamodb.Table.fromTableName(this, tableName, tableName);
  }
}
