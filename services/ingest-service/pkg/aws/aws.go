package aws

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// AWSClient wraps S3 and SQS clients
type AWSClient struct {
	S3       *s3.Client
	SQS      *sqs.Client
	DynamoDb *dynamodb.Client
}

// NewAWSClient initializes a new AWSClient
func NewAWSClient(ctx context.Context) (*AWSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS SDK config: %w", err)
	}

	return &AWSClient{
		S3:       s3.NewFromConfig(cfg),
		SQS:      sqs.NewFromConfig(cfg),
		DynamoDb: dynamodb.NewFromConfig(cfg),
	}, nil
}

func (c *AWSClient) S3ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := c.S3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		var notFoundErr *types.NotFound
		if errors.As(err, &notFoundErr) {
			return false, nil
		}
		return false, err // Return the actual error if it's not a NotFound error
	}
	return true, nil
}

func (c *AWSClient) PublishSQSMessage(ctx context.Context, queueURL, message string) error {
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

func (c *AWSClient) GetS3Object(ctx context.Context, bucket, key string) ([]byte, error) {
	obj, err := c.S3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting object from S3: %w", err)
	}
	defer obj.Body.Close()

	return io.ReadAll(obj.Body)
}