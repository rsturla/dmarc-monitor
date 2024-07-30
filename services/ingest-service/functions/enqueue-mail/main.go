package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/aws"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/aws/awslocal"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/config"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/message"
)

// Handler function for AWS Lambda
func handler(ctx context.Context, sesEvent events.SimpleEmailEvent) error {
	config, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	awsClient, err := aws.NewAWSClient(ctx)
	if err != nil {
		return fmt.Errorf("error creating AWS client: %w", err)
	}

	for _, record := range sesEvent.Records {
		if err := processEmail(ctx, awsClient, config, record.SES.Mail); err != nil {
			return fmt.Errorf("error processing email: %w", err)
		}
	}
	log.Print("Email processing complete.")
	return nil
}

// Main entry point
func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		event, ctx, err := awslocal.CreateLocalEvent[events.SimpleEmailEvent]("./sample-events/SQSEvent.json")
		if err != nil {
			log.Printf("Error creating local event: %v\n", err)
		}
		if err := handler(ctx, event); err != nil {
			log.Printf("Error processing local event: %v\n", err)
		}
	} else {
		lambda.Start(handler)
	}
}

// Process the email and publishes a message to SQS
func processEmail(ctx context.Context, awsClient *aws.AWSClient, config *config.Config, mail events.SimpleEmailMessage) error {
	recipientTag, err := message.ExtractPlusAddress(mail.Destination[0])
	if err != nil {
		return fmt.Errorf("error extracting tag from recipient email address: %w", err)
	}

	messageJSON, err := json.Marshal(map[string]string{
		"tag":          recipientTag,
		"s3BucketName": config.ReportStorageBucketName,
		"s3ObjectPath": fmt.Sprintf("%s%s", "raw/", mail.MessageID),
		"receivedTime": fmt.Sprintf("%d", mail.Timestamp.Unix()),
		"messageID":    mail.MessageID,
	})
	if err != nil {
		return fmt.Errorf("error marshalling message to JSON: %w", err)
	}

	if err := awsClient.PublishSQSMessage(ctx, config.RawEmailQueueURL, string(messageJSON)); err != nil {
		return fmt.Errorf("error publishing message to SQS: %w", err)
	}

	return nil
}
