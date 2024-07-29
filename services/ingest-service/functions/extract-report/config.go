package main

import (
	"fmt"
	"os"
)

// Config holds the configuration for the application
type Config struct {
	QueueURL          string
	BucketName        string
	DynamoDBTableName string
}

// Initialize configuration
func loadConfig() (*Config, error) {
	queueURL := os.Getenv("QUEUE_URL")
	bucketName := os.Getenv("BUCKET_NAME")
	dynamoDBTableName := os.Getenv("DYNAMODB_TABLE_NAME")

	if queueURL == "" || bucketName == "" || dynamoDBTableName == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return &Config{
		QueueURL:          queueURL,
		BucketName:        bucketName,
		DynamoDBTableName: dynamoDBTableName,
	}, nil
}
