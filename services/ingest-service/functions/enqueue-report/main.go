package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Handler function for AWS Lambda
func handler(ctx context.Context, sesEvent events.SimpleEmailEvent) error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	awsClient, err := NewAWSClient(ctx)
	if err != nil {
		return fmt.Errorf("error creating AWS client: %w", err)
	}

	for _, record := range sesEvent.Records {
		if err := processEmail(ctx, awsClient, config, record.SES.Mail); err != nil {
			return fmt.Errorf("error processing email: %w", err)
		}
	}
	fmt.Println("Email processing complete.")
	return nil
}

// Main entry point
func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		if err := handleLocalEvent(); err != nil {
			fmt.Printf("Error processing local event: %v\n", err)
		}
	} else {
		lambda.Start(handler)
	}
}

// Process the email and publishes a message to SQS
func processEmail(ctx context.Context, awsClient *AWSClient, config *Config, mail events.SimpleEmailMessage) error {
	recipientTag, err := extractPlusAddressTag(mail.Destination[0])
	if err != nil {
		return fmt.Errorf("error extracting tag from recipient email address: %w", err)
	}

	rawEmailLocation, err := getRawEmailLocation(ctx, awsClient.S3, config.BucketName, "raw/", mail.MessageID)
	if err != nil {
		return fmt.Errorf("error getting raw email location: %w", err)
	}

	messageJSON, err := json.Marshal(map[string]string{
		"tag":          recipientTag,
		"location":     rawEmailLocation,
		"receivedTime": fmt.Sprintf("%d", mail.Timestamp.Unix()),
		"messageID":    mail.MessageID,
	})
	if err != nil {
		return fmt.Errorf("error marshalling message to JSON: %w", err)
	}

	if err := awsClient.publishSQSMessage(ctx, config.QueueURL, string(messageJSON)); err != nil {
		return fmt.Errorf("error publishing message to SQS: %w", err)
	}

	log.Printf("Processed email with tag: %s and location: %s", recipientTag, rawEmailLocation)
	return nil
}
