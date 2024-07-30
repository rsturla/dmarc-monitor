package config

import (
	"os"
	"testing"
)

func TestGetEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		unset     bool
		expectErr bool
		expected  string
	}{
		{
			name:      "Variable is set",
			key:       "TEST_VAR",
			value:     "test_value",
			unset:     false,
			expectErr: false,
			expected:  "test_value",
		},
		{
			name:      "Variable is not set",
			key:       "MISSING_VAR",
			unset:     true,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.unset {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result, err := getEnvironmentVariable(tt.key)
			if (err != nil) != tt.expectErr {
				t.Errorf("getEnvironmentVariable() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr && result != tt.expected {
				t.Errorf("getEnvironmentVariable() result = %s, expected %s", result, tt.expected)
			}
		})
	}
}
