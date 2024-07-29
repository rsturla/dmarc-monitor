package main

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// AWSClient wraps S3 and SQS clients
type AWSClient struct {
	S3       *s3.Client
	SQS      *sqs.Client
	DynamoDB *dynamodb.Client
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
		DynamoDB: dynamodb.NewFromConfig(cfg),
	}, nil
}

func getS3Object(ctx context.Context, awsClient *AWSClient, bucket, key string) ([]byte, error) {
	contents, err := awsClient.S3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting object from S3: %w", err)
	}
	defer contents.Body.Close()

	return io.ReadAll(contents.Body)
}
