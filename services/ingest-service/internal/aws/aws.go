package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// AWSClient wraps S3 and SQS clients
type AWSClient struct {
	S3       *s3.Client
	SQS      *sqs.Client
	DynamoDb *dynamodb.Client
}

// NewAWSClient initializes a new AWSClient. While it initializes S3, SQS and DynamoDB clients,
// the function is fast enough to not introduce any noticeable delay or overhead.
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
