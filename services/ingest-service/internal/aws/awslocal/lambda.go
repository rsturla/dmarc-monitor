package awslocal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// Function to read and parse an event JSON file, returning the event and context
// for the local Lambda function.
func CreateLocalEvent[T any](eventFile string) (T, context.Context, error) {
	var event T

	file, err := os.ReadFile(eventFile)
	if err != nil {
		return event, nil, fmt.Errorf("could not read event JSON file: %w", err)
	}

	if err := json.Unmarshal(file, &event); err != nil {
		return event, nil, fmt.Errorf("could not unmarshal event file: %w", err)
	}

	ctx := context.Background()
	return event, ctx, nil
}
