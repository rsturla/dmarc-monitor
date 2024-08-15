package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dynamodbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/aws"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/aws/awslocal"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/config"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/dmarc/rua"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/models"
)

// Main function
func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		event, ctx, err := awslocal.CreateLocalEvent[events.SQSEvent]("./sample-events/SQSEvent.json")
		if err != nil {
			log.Fatalf("Error creating local event: %v", err)
		}
		if err := handleRequest(ctx, event); err != nil {
			log.Fatalf("Error processing local event: %v", err)
		}
	} else {
		lambda.Start(handleRequest)
	}
}

// HandleRequest processes the SQS event
func handleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	cfg, err := loadConfig()
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

// LoadConfig loads the configuration
func loadConfig() (*Config, error) {
	return config.NewConfig[Config]()
}

// ProcessRecord processes an individual SQS record
func processRecord(ctx context.Context, awsClient *aws.AWSClient, cfg *Config, record events.SQSMessage) error {
	var sqsMessage models.IngestMessage
	if err := json.Unmarshal([]byte(record.Body), &sqsMessage); err != nil {
		return fmt.Errorf("error unmarshalling message: %w", err)
	}

	log.Printf("Processing message: %s", sqsMessage.MessageID)

	body, err := awsClient.S3GetObject(ctx, cfg.ReportStorageBucketName, sqsMessage.AttachmentS3ObjectPath)
	if err != nil {
		return err
	}

	ruaReport, err := parseRUAReport(body)
	if err != nil {
		return err
	}

	return storeReports(ctx, awsClient, cfg, sqsMessage, ruaReport)
}

// ParseRUAReport parses the XML body into a RUA report
func parseRUAReport(body []byte) (*rua.RUA, error) {
	var ruaReport rua.RUA
	if err := ruaReport.ParseXML(body); err != nil {
		return nil, fmt.Errorf("error parsing RUA report: %w", err)
	}
	log.Printf("Parsed RUA report: %+v", ruaReport)
	return &ruaReport, nil
}

// StoreReports stores the DMARC reports and records in DynamoDB
func storeReports(ctx context.Context, awsClient *aws.AWSClient, cfg *Config, sqsMessage models.IngestMessage, ruaReport *rua.RUA) error {
	dmarcReportItem := createDmarcReportItem(sqsMessage, ruaReport)
	dmarcRecordItems := createDmarcRecordItems(dmarcReportItem, ruaReport)

	reportStorageObject, err := attributevalue.MarshalMap(dmarcReportItem)
	if err != nil {
		return fmt.Errorf("error marshalling DmarcReportItem: %w", err)
	}

	if err := awsClient.DynamoDBPutItem(ctx, cfg.ReportTableName, &reportStorageObject); err != nil {
		return fmt.Errorf("error putting DmarcReportItem: %w", err)
	}

	return storeDmarcRecordItems(ctx, awsClient, cfg.RecordTableName, dmarcRecordItems)
}

// CreateDmarcReportItem creates a DMARC report item
func createDmarcReportItem(sqsMessage models.IngestMessage, ruaReport *rua.RUA) models.DmarcReportMetadataItem {
	return models.DmarcReportMetadataItem{
		ID:               fmt.Sprintf("%s#%s", sqsMessage.TenantID, ruaReport.ReportMetadata.ReportID),
		ReportId:         ruaReport.ReportMetadata.ReportID,
		OrgName:          ruaReport.ReportMetadata.OrgName,
		Email:            ruaReport.ReportMetadata.Email,
		ExtraContactInfo: ruaReport.ReportMetadata.ExtraContactInfo,
		DateRangeBegin:   ruaReport.ReportMetadata.DateRange.Begin,
		DateRangeEnd:     ruaReport.ReportMetadata.DateRange.End,
		Domain:           ruaReport.PolicyPublished.Domain,
		Adkim:            ruaReport.PolicyPublished.Adkim,
		Aspf:             ruaReport.PolicyPublished.Aspf,
		P:                ruaReport.PolicyPublished.P,
		Sp:               ruaReport.PolicyPublished.Sp,
		Pct:              ruaReport.PolicyPublished.Pct,
		Np:               ruaReport.PolicyPublished.Np,
	}
}

// CreateDmarcRecordItems creates DMARC record items
func createDmarcRecordItems(dmarcReportItem models.DmarcReportMetadataItem, ruaReport *rua.RUA) []models.DmarcRecordItem {
	var dmarcRecordItems []models.DmarcRecordItem
	for i, record := range ruaReport.Records {
		var authResultsDkim []models.DmarcAuthResultNestedAttribute
		for _, dkim := range record.AuthResults.Dkim {
			authResultsDkim = append(authResultsDkim, models.DmarcAuthResultNestedAttribute{
				Domain:   dkim.Domain,
				Result:   dkim.Result,
				Selector: dkim.Selector,
			})
		}

		dmarcRecordItems = append(dmarcRecordItems, models.DmarcRecordItem{
			ID:                         fmt.Sprintf("%s#%d", dmarcReportItem.ID, i),
			ReportId:                   dmarcReportItem.ReportId,
			SourceIp:                   record.Row.SourceIp.String(),
			Count:                      record.Row.Count,
			PolicyEvaluatedDisposition: record.Row.PolicyEvaluated.Disposition,
			PolicyEvaluatedDkim:        record.Row.PolicyEvaluated.Dkim,
			PolicyEvaluatedSpf:         record.Row.PolicyEvaluated.Spf,
			HeaderFrom:                 record.Identifiers.HeaderFrom,
			AuthResultsDkim:            authResultsDkim,
			AuthResultsSpf: models.DmarcAuthResultNestedAttribute{
				Domain: record.AuthResults.Spf.Domain,
				Result: record.AuthResults.Spf.Result,
			},
		})
	}
	return dmarcRecordItems
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

	log.Printf("Successfully stored DMARC records: %+v", items)
	return nil
}
