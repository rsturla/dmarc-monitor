package awslocal

import (
	"os"
	"testing"
)

// Sample event struct for testing
type TestEvent struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestCreateLocalEvent(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   string
		expectError   bool
		expectedEvent TestEvent
	}{
		{
			name:          "Valid event file",
			fileContent:   `{"name":"John Doe","age":30}`,
			expectError:   false,
			expectedEvent: TestEvent{Name: "John Doe", Age: 30},
		},
		{
			name:        "Invalid JSON",
			fileContent: `{"name":"John Doe","age":30`,
			expectError: true,
		},
		{
			name:        "Non-existent file",
			fileContent: ``,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var eventFile string
			var err error

			if tt.fileContent != "" {
				// Create a temporary file with the specified content
				tmpFile, tmpErr := os.CreateTemp("", "event*.json")
				if tmpErr != nil {
					t.Fatalf("could not create temp file: %v", tmpErr)
				}
				defer os.Remove(tmpFile.Name())

				if _, writeErr := tmpFile.Write([]byte(tt.fileContent)); writeErr != nil {
					t.Fatalf("could not write to temp file: %v", writeErr)
				}
				tmpFile.Close()
				eventFile = tmpFile.Name()
			} else {
				eventFile = "nonexistent.json"
			}

			var event TestEvent
			event, ctx, err := CreateLocalEvent[TestEvent](eventFile)

			if (err != nil) != tt.expectError {
				t.Errorf("CreateLocalEvent() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && ctx == nil {
				t.Errorf("expected context to be non-nil")
			}

			if !tt.expectError && event != tt.expectedEvent {
				t.Errorf("CreateLocalEvent() event = %v, expectedEvent %v", event, tt.expectedEvent)
			}
		})
	}
}
