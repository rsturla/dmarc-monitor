package message

import "testing"

func TestExtractPlusAddressTag(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"test+abc@example.com", "abc"},
		{"user+tag+extra@example.com", "tag"},
		{"noTag@example.com", ""},
		{"@example.com", ""},
		{"invalidemail", ""},
		{"user+@example.com", ""},
		{"user+tag@subdomain.example.com", "tag"},
		{"user+tag@sub.domain.example.com", "tag"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := ExtractPlusAddressTag(tt.email)
			if got != tt.expected {
				t.Errorf("extractPlusAddressTag(%q) = %q; want %q", tt.email, got, tt.expected)
			}
		})
	}
}
