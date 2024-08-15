package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/aws/awslocal"
)

// Main function
func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		event, ctx, err := awslocal.CreateLocalEvent[events.SQSEvent]("./sample-events/SQSEvent.json")
		if err != nil {
			log.Fatalf("Error creating local event: %v", err)
		}
		if err := handler(ctx, event); err != nil {
			log.Fatalf("Error processing local event: %v", err)
		}
	} else {
		lambda.Start(handler)
	}
}
