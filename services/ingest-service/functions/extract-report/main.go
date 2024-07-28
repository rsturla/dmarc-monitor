package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

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

	body, err := getS3ObjectBody(ctx, awsClient, message.S3BucketName, message.S3ObjectPath)
	if err != nil {
		return err
	}

	email, err := Parse(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error parsing email: %w", err)
	}

	for _, attachment := range email.Attachments {
		if err := processAttachment(&attachment); err != nil {
			return err
		}
	}

	return nil
}

func getS3ObjectBody(ctx context.Context, awsClient *AWSClient, bucket, key string) ([]byte, error) {
	contents, err := awsClient.S3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting object from S3: %w", err)
	}
	defer contents.Body.Close()

	return io.ReadAll(contents.Body)
}

func processAttachment(attachment *Attachment) error {
	data, err := io.ReadAll(attachment.Data)
	if err != nil {
		return fmt.Errorf("error reading attachment data: %w", err)
	}

	uncompressed, err := uncompress(data, attachment.ContentType)
	if err != nil {
		return err
	}

	log.Printf("Attachment: %s, MIME type: %s, Uncompressed content: %s\n", attachment.Filename, attachment.ContentType, uncompressed)
	return nil
}

func uncompress(data []byte, mime string) ([]byte, error) {
	switch mime {
	case "application/gzip":
		return uncompressGzip(data)
	case "application/zip":
		return uncompressZip(data)
	default:
		return nil, fmt.Errorf("unsupported MIME type: %s", mime)
	}
}

func uncompressGzip(data []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gzipReader.Close()

	return io.ReadAll(gzipReader)
}

func uncompressZip(data []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("error creating zip reader: %w", err)
	}

	for _, f := range zipReader.File {
		if !f.FileInfo().IsDir() {
			file, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("error opening file: %w", err)
			}
			defer file.Close()

			return io.ReadAll(file)
		}
	}

	return nil, fmt.Errorf("no files found in zip archive")
}
