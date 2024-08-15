package models

// IngestMessage represents the message sent to the SQS queue by each stage of the pipeline
type IngestMessage struct {
	// MessageID is the unique identifier for the email message, provided by SES
	MessageID string `json:"messageID"`

	// MessageTimestamp is the time the email message was received, provided by SES
	// Populated by the enqueue-email function
	MessageTimestamp string `json:"timestamp"`

	// TenantID is the unique identifier for the tenant that owns the email messages
	// Populated by the enqueue-email function
	TenantID string `json:"tenantID"`

	// RawS3ObjectPath is the path to the raw email message in the S3 bucket
	// Populated by the enqueue-email function
	RawS3ObjectPath string `json:"s3ObjectPath"`

	// AttachmentS3ObjectPath is the path to the extracted attachment in the S3 bucket
	// Populated by the extract-attachment function
	AttachmentS3ObjectPath string `json:"attachmentS3ObjectPath"`

	// ReportID is the unique identifier for the DMARC report, provided by the report
	// Populated by the parse-report function
	ReportID string `json:"reportID"`

	// ReportTimestamp is the time the DMARC report was received, provided by the report
	// Populated by the parse-report function
	ReportTimestamp string `json:"reportTimestamp"`
}
