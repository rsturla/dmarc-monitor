package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/aws"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/config"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/errors"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/models"
)

func handler(ctx context.Context, sesEvent events.SimpleEmailEvent) error {
	config, err := config.NewConfig[Config]()
	if err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error loading configuration: %v", err))
	}

	awsClient, err := aws.NewAWSClient(ctx)
	if err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error creating AWS client: %v", err))
	}

	for _, record := range sesEvent.Records {
		if err := processEmail(ctx, awsClient, config, record.SES.Mail); err != nil {
			log.Printf("Error processing email with MessageID %s: %v", record.SES.Mail.MessageID, err)
			// Optionally: continue processing other emails, or return the error
			return errors.NewLambdaError(500, fmt.Sprintf("error processing email with MessageID %s: %v", record.SES.Mail.MessageID, err))
		}
	}

	return nil
}

// processEmail processes an individual SES email message and adds it to the SQS queue for further processing downstream.
func processEmail(ctx context.Context, awsClient *aws.AWSClient, config *Config, mail events.SimpleEmailMessage) error {
	tenantID := strings.Split(mail.Destination[0], "@")[0]
	messageJSON, err := json.Marshal(models.IngestMessage{
		TenantID:         tenantID,
		RawS3ObjectPath:  fmt.Sprintf("raw/%s", mail.MessageID),
		MessageTimestamp: fmt.Sprintf("%d", mail.Timestamp.Unix()),
		MessageID:        mail.MessageID,
	})
	if err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error marshalling message: %v", err))
	}

	if err := awsClient.SQSPublishMessage(ctx, config.NextStageQueueURL, string(messageJSON)); err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error publishing message to SQS: %v", err))
	}

	return nil
}
