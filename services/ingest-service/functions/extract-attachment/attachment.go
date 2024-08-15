package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/aws"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/compress"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/email/message"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/errors"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/models"
)

func processEmailAttachment(ctx context.Context, attachment message.Attachment, awsClient *aws.AWSClient, config *Config, sqsMessage models.IngestMessage) error {
	data, err := getAttachmentData(&attachment)
	if err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error getting attachment data: %v", err))
	}

	attachmentS3ObjectPath, err := saveReport(ctx, awsClient, config, sqsMessage.MessageID, sqsMessage.TenantID, data)
	if err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error saving report to S3: %v", err))
	}

	messageJSON, err := json.Marshal(models.IngestMessage{
		TenantID:               sqsMessage.TenantID,
		RawS3ObjectPath:        sqsMessage.RawS3ObjectPath,
		AttachmentS3ObjectPath: attachmentS3ObjectPath,
		MessageTimestamp:       sqsMessage.MessageTimestamp,
		MessageID:              sqsMessage.MessageID,
	})
	if err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error marshalling message: %v", err))
	}

	if err := awsClient.SQSPublishMessage(ctx, config.NextStageQueueURL, string(messageJSON)); err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error publishing message to SQS: %v", err))
	}
	return nil
}

func getAttachmentData(attachment *message.Attachment) ([]byte, error) {
	data, err := io.ReadAll(attachment.Data)
	if err != nil {
		return nil, errors.NewLambdaError(500, fmt.Sprintf("error reading attachment data: %v", err))
	}

	uncompressed, err := compress.Decompress(data, attachment.ContentType)
	if err != nil {
		return nil, errors.NewLambdaError(500, fmt.Sprintf("error decompressing attachment data: %v", err))
	}

	return uncompressed, nil
}

func saveReport(ctx context.Context, awsClient *aws.AWSClient, config *Config, messageID string, tenantID string, data []byte) (string, error) {
	s3Key := fmt.Sprintf("reports/%s/%s/%s.xml", tenantID, time.Now().Format("2006/01/02"), messageID)
	contentType := "application/xml"
	if err := awsClient.S3PutObject(ctx, config.ReportStorageBucketName, s3Key, contentType, data); err != nil {
		return "", errors.NewLambdaError(500, fmt.Sprintf("error saving report to S3: %v", err))
	}

	return s3Key, nil
}
