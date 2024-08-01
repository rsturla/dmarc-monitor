import * as cdk from 'aws-cdk-lib';
import { Billing, BillingMode, Capacity, EnableScalingProps, TableEncryptionV2, TablePropsV2, TableV2, UtilizationScalingProps } from 'aws-cdk-lib/aws-dynamodb';
import { Construct } from 'constructs';

export interface DynamoDBTableProps extends TablePropsV2 {
    readonly readScalingProps?: EnableScalingProps;
    readonly writeScalingProps?: EnableScalingProps;
}

const DEFAULT_UTILIZATION_SCALING_PROPS: UtilizationScalingProps = {
    targetUtilizationPercent: 60,
    scaleInCooldown: cdk.Duration.seconds(60),
    scaleOutCooldown: cdk.Duration.seconds(60),
};

export class DynamoDBTable extends TableV2 {
    constructor(scope: Construct, id: string, props: DynamoDBTableProps) {
        if (!props) {
            throw new Error('DynamoDB table props must be provided');
        }

        let {
            billing,
            pointInTimeRecovery,
            deletionProtection,
            encryption
        } = props;

        if (deletionProtection === undefined) {
            deletionProtection = true;
        }

        if (pointInTimeRecovery === undefined) {
            pointInTimeRecovery = true;
        }

        if (props.billing?.mode == BillingMode.PROVISIONED) {
            if (!props.readScalingProps || !props.writeScalingProps) {
                throw new Error('Read and write scaling properties must be provided for provisioned billing mode');
            }

            billing = Billing.provisioned({
                readCapacity: Capacity.autoscaled({
                    minCapacity: props.readScalingProps.minCapacity,
                    maxCapacity: props.readScalingProps.maxCapacity,
                }),
                writeCapacity: Capacity.autoscaled({
                    minCapacity: props.writeScalingProps.minCapacity,
                    maxCapacity: props.writeScalingProps.maxCapacity,
                }),
            });
        }

        if (!encryption) {
            encryption = TableEncryptionV2.awsManagedKey();
        }

        const updatedProps: DynamoDBTableProps = {
            billing,
            encryption,
            pointInTimeRecovery,
            deletionProtection,
            ...props,
        };

        super(scope, id, updatedProps);
    }
}
