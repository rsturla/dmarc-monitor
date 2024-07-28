package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Environment variables
var (
	QueueURL   = os.Getenv("QUEUE_URL")
	BucketName = os.Getenv("BUCKET_NAME")
)

// Handler function for AWS Lambda
func handler(ctx context.Context, sesEvent events.SimpleEmailEvent) error {
	for _, record := range sesEvent.Records {
		if err := processEmail(ctx, record.SES.Mail); err != nil {
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

// Function to read and parse event.json file and call the handler
func handleLocalEvent() error {
	file, err := os.ReadFile("./sample-events/event.json")
	if err != nil {
		return fmt.Errorf("could not read event.json file: %w", err)
	}

	var sesEvent events.SimpleEmailEvent
	if err := json.Unmarshal(file, &sesEvent); err != nil {
		return fmt.Errorf("could not unmarshal event.json file: %w", err)
	}

	ctx := context.Background()
	return handler(ctx, sesEvent)
}

// Processes the email and publishes a message to SQS
func processEmail(ctx context.Context, mail events.SimpleEmailMessage) error {
	recipientTag := extractPlusAddressTag(mail.Destination[0])
	if recipientTag == "" {
		return fmt.Errorf("no tag found in recipient email address")
	}

	rawEmailLocation, err := getRawEmailLocation(ctx, BucketName, mail.MessageID)
	if err != nil {
		return fmt.Errorf("error getting raw email location: %w", err)
	}

	message := map[string]string{
		"tag":      recipientTag,
		"location": rawEmailLocation,
	}
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message to JSON: %w", err)
	}

	if err := publishSQSMessage(ctx, QueueURL, string(messageJSON)); err != nil {
		return fmt.Errorf("error publishing message to SQS: %w", err)
	}

	fmt.Printf("Processing email with tag %s and location %s\n", recipientTag, rawEmailLocation)
	return nil
}

// Retrieves the raw email location from S3
func getRawEmailLocation(ctx context.Context, bucket, messageId string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("error loading AWS SDK config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	messageLocation := fmt.Sprintf("raw/%s", messageId)
	_, err = s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &messageLocation,
	})
	if err != nil {
		return "", fmt.Errorf("error getting raw email location: %w", err)
	}

	return fmt.Sprintf("s3://%s/raw/%s", bucket, messageId), nil
}

// Publishes a message to an SQS queue
func publishSQSMessage(ctx context.Context, queueURL, message string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("error loading AWS SDK config: %w", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)
	msg, err := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: &message,
		QueueUrl:    &queueURL,
	})
	if err != nil {
		return fmt.Errorf("error sending message to SQS: %w", err)
	}

	fmt.Printf("Message sent with ID %s\n", *msg.MessageId)
	return nil
}

// Extracts the tag from a plus-addressed email
func extractPlusAddressTag(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	localParts := strings.Split(parts[0], "+")
	if len(localParts) < 2 {
		return ""
	}

	return localParts[1]
}
