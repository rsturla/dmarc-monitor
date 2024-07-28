package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// AWSClient wraps S3 and SQS clients
type AWSClient struct {
	S3  *s3.Client
	SQS *sqs.Client
}

// NewAWSClient initializes a new AWSClient
func NewAWSClient(ctx context.Context) (*AWSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS SDK config: %w", err)
	}

	return &AWSClient{
		S3:  s3.NewFromConfig(cfg),
		SQS: sqs.NewFromConfig(cfg),
	}, nil
}

// Publishes a message to an SQS queue
func (c *AWSClient) publishSQSMessage(ctx context.Context, queueURL, message string) error {
	msg, err := c.SQS.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: &message,
		QueueUrl:    &queueURL,
	})
	if err != nil {
		return fmt.Errorf("error sending message to SQS queue %s: %w", queueURL, err)
	}

	log.Printf("Message sent with ID %s\n", *msg.MessageId)
	return nil
}
