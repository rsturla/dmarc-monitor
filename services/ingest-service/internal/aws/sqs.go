package aws

import (
	"context"
	"encoding/json"
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

// ParseSQSMessage parses an SQS message into a given struct.
func ParseSQSMessage(body string, v interface{}) error {
	if err := json.Unmarshal([]byte(body), v); err != nil {
		return fmt.Errorf("error unmarshalling message: %w", err)
	}
	return nil
}
