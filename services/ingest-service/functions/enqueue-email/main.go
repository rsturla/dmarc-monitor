package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/aws"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/aws/awslocal"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/config"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/models"
)

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

func handler(ctx context.Context, sesEvent events.SimpleEmailEvent) error {
	config, err := config.NewConfig[Config]()
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

func processEmail(ctx context.Context, awsClient *aws.AWSClient, config *Config, mail events.SimpleEmailMessage) error {
	messageJSON, err := json.Marshal(models.IngestMessage{
		TenantID:         strings.Split(mail.Destination[0], "@")[0],
		RawS3ObjectPath:  fmt.Sprintf("%s%s", "raw/", mail.MessageID),
		MessageTimestamp: fmt.Sprintf("%d", mail.Timestamp.Unix()),
		MessageID:        mail.MessageID,
	})
	if err != nil {
		return fmt.Errorf("error marshalling message to JSON: %w", err)
	}

	if err := awsClient.SQSPublishMessage(ctx, config.NextStageQueueURL, string(messageJSON)); err != nil {
		return fmt.Errorf("error publishing message to SQS: %w", err)
	}

	return nil
}

func extractPlusAddressTag(email string) (string, error) {
	// Get the value between + and @ in the email address
	// If the input is not a valid email address, return an error
	// If the email address does not contain a +, return an empty string
	emailParts := strings.Split(email, "@")
	if len(emailParts) != 2 {
		return "", fmt.Errorf("invalid email address: %s", email)
	}

	recipient := emailParts[0]
	tagParts := strings.Split(recipient, "+")
	if len(tagParts) != 2 {
		return "", nil
	}
	return tagParts[1], nil
}
