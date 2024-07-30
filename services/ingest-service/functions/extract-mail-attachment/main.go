package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/aws"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/aws/awslocal"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/compress"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/config"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/pkg/message"
)

type SQSMessage struct {
	MessageId    string `json:"messageID"`
	S3BucketName string `json:"s3BucketName"`
	S3ObjectPath string `json:"s3ObjectPath"`
	S3ObjectFull string `json:"s3BucketFull"`
	ReceivedTime string `json:"receivedTime"`
	Tag          string `json:"tag"`
}

func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		event, ctx, err := awslocal.CreateLocalEvent[events.SQSEvent]("./sample-events/SQSEvent.json")
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

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	config, err := config.NewConfig[Config]()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	awsClient, err := aws.NewAWSClient(ctx)
	if err != nil {
		return fmt.Errorf("error creating AWS client: %w", err)
	}

	for _, record := range sqsEvent.Records {
		if err := processRecord(ctx, awsClient, config, record); err != nil {
			return fmt.Errorf("error processing message: %w", err)
		}
	}

	return nil
}

func processRecord(ctx context.Context, awsClient *aws.AWSClient, config *Config, record events.SQSMessage) error {
	var sqsMessage SQSMessage
	if err := json.Unmarshal([]byte(record.Body), &sqsMessage); err != nil {
		return fmt.Errorf("error unmarshalling message: %w", err)
	}

	log.Printf("Processing message: %s\n", sqsMessage.MessageId)

	body, err := awsClient.GetS3Object(ctx, sqsMessage.S3BucketName, sqsMessage.S3ObjectPath)
	if err != nil {
		return err
	}

	email, err := message.ParseMail(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error parsing email: %w", err)
	}

	for _, attachment := range email.Attachments {
		data, err := getAttachmentData(&attachment)
		if err != nil {
			return fmt.Errorf("error processing attachment: %w", err)
		}

		// Save the report to the S3 bucket - under the reports/<message> key
		if err := saveReport(ctx, awsClient, config, sqsMessage.MessageId, sqsMessage.Tag, data); err != nil {
			return fmt.Errorf("error saving report: %w", err)
		}
	}

	return nil
}

func getAttachmentData(attachment *message.Attachment) ([]byte, error) {
	data, err := io.ReadAll(attachment.Data)
	if err != nil {
		return nil, fmt.Errorf("error reading attachment data: %w", err)
	}

	uncompressed, err := compress.Decompress(data, attachment.ContentType)
	if err != nil {
		return nil, fmt.Errorf("error uncompressing attachment: %w", err)
	}

	return uncompressed, nil
}

func saveReport(ctx context.Context, awsClient *aws.AWSClient, config *Config, messageID string, tenantId string, data []byte) error {
	s3Key := fmt.Sprintf("reports/%s/%s/%s.xml", tenantId, time.Now().Format("2006/01/02"), messageID)
	contentType := "application/xml"
	_, err := awsClient.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &config.ReportStorageBucketName,
		Key:         &s3Key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("error saving report to S3: %w", err)
	}

	return nil
}
