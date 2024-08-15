package main

type Config struct {
	ReportStorageBucketName string `env:"INGEST_STORAGE_BUCKET_NAME"`
	NextStageQueueURL       string `env:"NEXT_STAGE_QUEUE_URL"`
}
