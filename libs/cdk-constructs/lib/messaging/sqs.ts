import { AnyPrincipal, Effect, PolicyStatement } from 'aws-cdk-lib/aws-iam';
import { Key } from 'aws-cdk-lib/aws-kms';
import { DeadLetterQueue, Queue, QueueEncryption, QueueProps } from 'aws-cdk-lib/aws-sqs';
import { Duration } from 'aws-cdk-lib/core';
import { Construct } from 'constructs';

const SQS_DEFAULTS = {
    MAX_RETENTION_PERIOD: Duration.days(14),
    DEFAULT_RECEIVE_COUNT: 3,
    MAX_ALLOWED_QUEUE_NAME_LENGTH: 80,
};

export interface SQSQueueProps extends QueueProps {
    readonly enableDeadLetterQueue?: boolean;
    readonly deadLetterQueueMaxReceiveCount?: number;
}

export class SQSQueue extends Queue {
    readonly deadLetterQueue?: DeadLetterQueue;

    constructor(scope: Construct, id: string, props: SQSQueueProps) {
        let {
            encryption,
            encryptionMasterKey,
            enableDeadLetterQueue,
            deadLetterQueue,
        } = props;

        if (!props.encryption) {
            encryption = QueueEncryption.SQS_MANAGED;
        } else if (props.encryption === QueueEncryption.KMS && !props.encryptionMasterKey) {
            encryptionMasterKey = SQSQueue.createEncryptionMasterKey(scope, id);
        }

        let deadLetterQueueName = deadLetterQueue?.queue.queueName;
        if (props.queueName && !deadLetterQueue) {
            deadLetterQueueName = props.fifo ? SQSQueue.getQueueNameForFifoDlq(props.queueName) : `${props.queueName}-dlq`;

            if (deadLetterQueueName.length > SQS_DEFAULTS.MAX_ALLOWED_QUEUE_NAME_LENGTH) {
                deadLetterQueueName = undefined;
            }
        }

        if (enableDeadLetterQueue && !deadLetterQueue) {
            deadLetterQueue = {
                queue: new SQSQueue(scope, `${id}DLQ`, {
                    queueName: deadLetterQueueName,
                    encryption,
                    encryptionMasterKey,
                    enableDeadLetterQueue: false,
                    retentionPeriod: SQS_DEFAULTS.MAX_RETENTION_PERIOD,
                    fifo: props.fifo,
                }),
                maxReceiveCount: props.deadLetterQueueMaxReceiveCount || SQS_DEFAULTS.DEFAULT_RECEIVE_COUNT,
            };
        }

        super(scope, id, {
            ...{
                encryption,
                encryptionMasterKey,
                deadLetterQueue,
                retentionPeriod: props.retentionPeriod || SQS_DEFAULTS.MAX_RETENTION_PERIOD,
            }
        });

        const secureQueuePolicyStatement = new PolicyStatement({
            sid: 'DenyInsecureTransport',
            effect: Effect.DENY,
            actions: ['sqs:*'],
            principals: [new AnyPrincipal()],
            conditions: {
                Bool: {
                    'aws:SecureTransport': 'false',
                },
            }
        });
        this.addToResourcePolicy(secureQueuePolicyStatement);

        this.deadLetterQueue = deadLetterQueue;
    }

    private static createEncryptionMasterKey(scope: Construct, id: string) {
        return new Key(scope, `${id}Key`, {
            description: `Key for ${id} queue`,
            enableKeyRotation: true
        });
    }

    private static getQueueNameForFifoDlq(queueName: string) {
        const fifoSuffix = '.fifo';
        if (!queueName.endsWith(fifoSuffix)) {
            throw new Error(`Fifo queue must have name ending with .fifo suffix. Present queue name : ${queueName}`);
        }
        // Find the last index of suffix occurrence
        const fifoSuffixIndex = queueName.lastIndexOf(fifoSuffix);
        // Final string is [0, fifoSuffixIndex] + '-dlq' + [fifoSuffixIndex+1 till end]
        return `${queueName.substring(0, fifoSuffixIndex)}-dlq${queueName.substring(fifoSuffixIndex)}`;
    }
}
