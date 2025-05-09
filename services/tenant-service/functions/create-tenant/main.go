package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rsturla/dmarc-monitor/libs/common/aws/awslocal"
)

func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		event, ctx, err := awslocal.CreateLocalEvent[events.SQSMessage]("./sample-events/SQSEvent.json")
		if err != nil {
			log.Fatalf("Error creating local event: %v\n", err)
		}
		if err := handler(ctx, event); err != nil {
			log.Fatalf("Error processing local event: %v\n", err)
		}
	} else {
		lambda.Start(handler)
	}
}
