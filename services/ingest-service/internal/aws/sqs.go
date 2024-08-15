package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSPublishMessage sends a message to an SQS queue.
func (c *AWSClient) SQSPublishMessage(ctx context.Context, queueURL, message string) error {
	msg, err := c.SQS.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: &message,
		QueueUrl:    &queueURL,
	})
	if err != nil {
		return fmt.Errorf("error sending message to SQS queue %s: %w", queueURL, err)
	}

	fmt.Printf("Message sent with ID %s\n", *msg.MessageId)
	return nil
}
