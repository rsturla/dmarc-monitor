package main

type Config struct {
	ReportStorageBucketName string `env:"INGEST_STORAGE_BUCKET_NAME"`
	ReportQueueURL          string `env:"INGEST_QUEUE_URL"`
}
