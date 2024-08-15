package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dynamodbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/aws"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/config"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/dmarc"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/dmarc/rua"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/errors"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/models"
)

// handler processes the SQS event
func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	cfg, err := config.NewConfig[Config]()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	awsClient, err := aws.NewAWSClient(ctx)
	if err != nil {
		return fmt.Errorf("error creating AWS client: %w", err)
	}

	for _, record := range sqsEvent.Records {
		if err := processRecord(ctx, awsClient, cfg, record); err != nil {
			log.Printf("Error processing message: %v", err)
			return fmt.Errorf("error processing message: %w", err)
		}
	}

	return nil
}

// ProcessRecord processes an individual SQS record
func processRecord(ctx context.Context, awsClient *aws.AWSClient, cfg *Config, record events.SQSMessage) error {
	var sqsMessage models.IngestMessage
	if err := aws.ParseSQSMessage(record.Body, &sqsMessage); err != nil {
		return errors.NewLambdaError(500, fmt.Sprintf("error unmarshalling message: %v", err))
	}

	body, err := awsClient.S3GetObject(ctx, cfg.ReportStorageBucketName, sqsMessage.AttachmentS3ObjectPath)
	if err != nil {
		return err
	}

	ruaReport, err := dmarc.ParseRUAReport(body)
	if err != nil {
		return err
	}

	return storeReports(ctx, awsClient, cfg, sqsMessage, ruaReport)
}

// StoreReports stores the DMARC reports and records in DynamoDB
func storeReports(ctx context.Context, awsClient *aws.AWSClient, cfg *Config, sqsMessage models.IngestMessage, ruaReport *rua.RUA) error {
	dmarcReportItem := dmarc.CreateDmarcReportItem(sqsMessage, ruaReport)
	dmarcRecordItems := dmarc.CreateDmarcRecordItems(dmarcReportItem, ruaReport)

	if err := storeDmarcReportItem(ctx, awsClient, cfg.ReportTableName, dmarcReportItem); err != nil {
		return err
	}

	return storeDmarcRecordItems(ctx, awsClient, cfg.RecordTableName, dmarcRecordItems)
}

// StoreDmarcReportItem stores the DMARC report item in DynamoDB
func storeDmarcReportItem(ctx context.Context, awsClient *aws.AWSClient, tableName string, item models.DmarcReportMetadataItem) error {
	reportStorageObject, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("error marshalling DmarcReportItem: %w", err)
	}

	if err := awsClient.DynamoDBPutItem(ctx, tableName, &reportStorageObject); err != nil {
		return fmt.Errorf("error putting DmarcReportItem: %w", err)
	}

	return nil
}

// StoreDmarcRecordItems stores the DMARC record items in DynamoDB
func storeDmarcRecordItems(ctx context.Context, awsClient *aws.AWSClient, tableName string, items []models.DmarcRecordItem) error {
	reportStorageObjects := make([]map[string]dynamodbTypes.AttributeValue, len(items))
	for i, record := range items {
		recordStorageObject, err := attributevalue.MarshalMap(record)
		if err != nil {
			return fmt.Errorf("error marshalling DmarcRecordItem: %w", err)
		}
		reportStorageObjects[i] = recordStorageObject
	}

	if err := awsClient.DynamoDBPutBatchItems(ctx, tableName, reportStorageObjects); err != nil {
		return fmt.Errorf("error putting DmarcRecordItems: %w", err)
	}

	return nil
}
