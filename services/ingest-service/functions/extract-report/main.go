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
)

type Message struct {
	MessageId    string `json:"messageID"`
	S3BucketName string `json:"s3BucketName"`
	S3ObjectPath string `json:"s3ObjectPath"`
	S3ObjectFull string `json:"s3BucketFull"`
	ReceivedTime string `json:"receivedTime"`
	Tag          string `json:"tag"`
}

// Handler function for AWS Lambda
func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	awsClient, err := NewAWSClient(ctx)
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

// Main entry point
func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		if err := handleLocalEvent(); err != nil {
			log.Printf("Error processing local event: %v\n", err)
		}
	} else {
		lambda.Start(handler)
	}
}

func processRecord(ctx context.Context, awsClient *AWSClient, config *Config, record events.SQSMessage) error {
	var message Message
	if err := json.Unmarshal([]byte(record.Body), &message); err != nil {
		return fmt.Errorf("error unmarshalling message: %w", err)
	}

	log.Printf("Processing message: %s\n", message.MessageId)

	body, err := getS3Object(ctx, awsClient, message.S3BucketName, message.S3ObjectPath)
	if err != nil {
		return err
	}

	email, err := Parse(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error parsing email: %w", err)
	}

	for _, attachment := range email.Attachments {
		data, err := getAttachmentData(&attachment)
		if err != nil {
			return fmt.Errorf("error processing attachment: %w", err)
		}

		// Save the report to the S3 bucket - under the reports/<message> key
		if err := saveReport(ctx, awsClient, config, message.MessageId, message.Tag, data); err != nil {
			return fmt.Errorf("error saving report: %w", err)
		}
	}

	return nil
}

func getAttachmentData(attachment *Attachment) ([]byte, error) {
	data, err := io.ReadAll(attachment.Data)
	if err != nil {
		return nil, fmt.Errorf("error reading attachment data: %w", err)
	}

	uncompressed, err := uncompress(data, attachment.ContentType)
	if err != nil {
		return nil, fmt.Errorf("error uncompressing attachment: %w", err)
	}

	return uncompressed, nil
}

func saveReport(ctx context.Context, awsClient *AWSClient, config *Config, messageID string, tenantId string, data []byte) error {
	s3Key := fmt.Sprintf("reports/%s/%s/%s.xml", tenantId, time.Now().Format("2006/01/02"), messageID)
	contentType := "application/xml"
	_, err := awsClient.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &config.BucketName,
		Key:         &s3Key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("error saving report to S3: %w", err)
	}

	return nil
}
