package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/aws"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/config"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/email/message"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/errors"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/models"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	config, err := config.NewConfig[Config]()
	if err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error loading configuration: %v", err))
	}

	awsClient, err := aws.NewAWSClient(ctx)
	if err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error creating AWS client: %v", err))
	}

	for _, record := range sqsEvent.Records {
		if err := processRecord(ctx, awsClient, config, record); err != nil {
			log.Printf("Error processing SQS message with MessageID %s: %v", record.MessageId, err)
			return errors.NewLambdaError(500, fmt.Sprintf("error processing message with SQS MessageID %s: %v", record.MessageId, err))
		}
	}

	return nil
}

func processRecord(ctx context.Context, awsClient *aws.AWSClient, config *Config, record events.SQSMessage) error {
	var sqsMessage models.IngestMessage
	if err := aws.ParseSQSMessage(record.Body, &sqsMessage); err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error unmarshalling message: %v", err))
	}

	rawEmail, err := getRawEmail(ctx, awsClient, config, sqsMessage.MessageID)
	if err != nil {
		return err
	}

	email, err := message.ParseMail(bytes.NewReader(rawEmail))
	if err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error parsing email: %v", err))
	}

	for _, attachment := range email.Attachments {
		// Save the report to the S3 bucket - under the reports/<message> key
		if err := processEmailAttachment(ctx, attachment, awsClient, config, sqsMessage); err != nil {
			return err
		}
	}

	return nil
}

func getRawEmail(ctx context.Context, awsClient *aws.AWSClient, config *Config, messageID string) ([]byte, error) {
	body, err := awsClient.S3GetObject(ctx, config.ReportStorageBucketName, messageID)
	if err != nil {
		return nil, errors.NewLambdaError(500, fmt.Sprintf("error getting raw email from S3: %v", err))
	}

	return body, nil
}
