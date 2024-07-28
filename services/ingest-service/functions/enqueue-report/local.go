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
	file, err := os.ReadFile("./sample-events/event.json")
	if err != nil {
		return fmt.Errorf("could not read event.json file: %w", err)
	}

	var sesEvent events.SimpleEmailEvent
	if err := json.Unmarshal(file, &sesEvent); err != nil {
		return fmt.Errorf("could not unmarshal event.json file: %w", err)
	}

	ctx := context.Background()
	return handler(ctx, sesEvent)
}
