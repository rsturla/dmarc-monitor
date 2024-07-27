package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Handler function for AWS Lambda
func handler(ctx context.Context, sesEvent events.SimpleEmailEvent) error {
	for _, record := range sesEvent.Records {
		fmt.Printf("Email received: %s\n", record.SES.Mail.MessageID)
		fmt.Printf("From: %s\n", record.SES.Mail.Source)
		fmt.Printf("To: %s\n", record.SES.Mail.Destination)
		fmt.Printf("Subject: %s\n", record.SES.Mail.CommonHeaders.Subject)
	}

	fmt.Println("Email processing complete.")

	return nil
}

// Function to read and parse event.json file and call the handler
func handleLocalEvent() error {
	file, err := os.ReadFile("./sample-events/event.json")
	if err != nil {
		return fmt.Errorf("could not read event.json file: %v", err)
	}

	var sesEvent events.SimpleEmailEvent
	if err := json.Unmarshal(file, &sesEvent); err != nil {
		return fmt.Errorf("could not unmarshal event.json file: %v", err)
	}

	ctx := context.Background()
	return handler(ctx, sesEvent)
}

func main() {
	// Check if the AWS_LAMBDA_RUNTIME_API environment variable is set
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		if err := handleLocalEvent(); err != nil {
			fmt.Printf("Error processing local event: %v\n", err)
		}
	} else {
		// Start the Lambda handler as usual
		lambda.Start(handler)
	}
}
