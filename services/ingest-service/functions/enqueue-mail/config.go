package main

type Config struct {
	ReportStorageBucketName string `env:"REPORT_STORAGE_BUCKET_NAME"`
	RawEmailQueueURL        string `env:"RAW_EMAIL_QUEUE_URL"`
}
