package main

import "testing"

func TestExtractPlusAddressTag(t *testing.T) {
	tests := []struct {
		email    string
		expected string
		wantErr  bool
	}{
		{"test+abc@example.com", "abc", false},
		{"user+tag+extra@example.com", "tag", false},
		{"noTag@example.com", "", false},
		{"@example.com", "", true},
		{"invalidemail", "", true},
		{"user+@example.com", "", false},
		{"user+tag@subdomain.example.com", "tag", false},
		{"user+tag@sub.domain.example.com", "tag", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got, err := extractPlusAddressTag(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractPlusAddressTag(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ExtractPlusAddressTag(%q) = %q, want %q", tt.email, got, tt.expected)
			}
		})
	}
}
