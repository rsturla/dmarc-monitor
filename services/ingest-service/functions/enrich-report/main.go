package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/aws"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/aws/awslocal"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/config"
)

// Main function
func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		event, ctx, err := awslocal.CreateLocalEvent[events.SQSEvent]("./sample-events/SQSEvent.json")
		if err != nil {
			log.Fatalf("Error creating local event: %v", err)
		}
		if err := handleRequest(ctx, event); err != nil {
			log.Fatalf("Error processing local event: %v", err)
		}
	} else {
		lambda.Start(handleRequest)
	}
}

// HandleRequest processes the SQS event
func handleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	awsClient, err := aws.NewAWSClient(ctx)
	if err != nil {
		return fmt.Errorf("error creating AWS client: %w", err)
	}

	for _, record := range sqsEvent.Records {
		if err := processMessage(ctx, awsClient, cfg, record); err != nil {
			log.Printf("Error processing message: %v", err)
			return fmt.Errorf("error processing message: %w", err)
		}
	}

	return nil
}

// LoadConfig loads the configuration
func loadConfig() (*Config, error) {
	return config.NewConfig[Config]()
}

// ProcessMessage processes the SQS message
func processMessage(ctx context.Context, awsClient *aws.AWSClient, cfg *Config, record events.SQSMessage) error {
	// Process the message
	return nil
}
