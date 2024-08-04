package main

type Config struct {
	ReportStorageBucketName string `env:"INGEST_STORAGE_BUCKET_NAME"`
	ReportTableName         string `env:"INGEST_TABLE_NAME"`
	RecordTableName         string `env:"INGEST_RECORD_TABLE_NAME"`
	ReportQueueUrl          string `env:"REPORT_QUEUE_URL"`
}
