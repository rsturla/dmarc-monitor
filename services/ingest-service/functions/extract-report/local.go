package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
)

// Function to read and parse event.json file and call the handler
func handleLocalEvent() error {
	file, err := os.ReadFile("./sample-events/SQSEvent.json")
	if err != nil {
		return fmt.Errorf("could not read event.json file: %w", err)
	}

	var sqsEvent events.SQSEvent
	if err := json.Unmarshal(file, &sqsEvent); err != nil {
		return fmt.Errorf("could not unmarshal event.json file: %w", err)
	}

	ctx := context.Background()
	return handler(ctx, sqsEvent)
}
