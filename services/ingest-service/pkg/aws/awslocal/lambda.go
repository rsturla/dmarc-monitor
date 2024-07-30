package awslocal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// Function to read and parse event.json file and call the handler
func CreateLocalEvent[T any](eventFile string) (T, context.Context, error) {
	var event T

	file, err := os.ReadFile(eventFile)
	if err != nil {
		return event, nil, fmt.Errorf("could not read event.json file: %w", err)
	}

	if err := json.Unmarshal(file, &event); err != nil {
		return event, nil, fmt.Errorf("could not unmarshal event file: %w", err)
	}

	ctx := context.Background()
	return event, ctx, nil
}
