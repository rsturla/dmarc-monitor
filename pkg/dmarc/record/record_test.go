package dmarc

import (
	"reflect"
	"testing"
	"time"
)

func TestParseRecord(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    *Record
		wantErr bool
	}{
		{
			"valid record",
			"v=DMARC1; p=none; adkim=r; aspf=s; fo=0; rf=afrf; ri=86400; rua=mailto:test@example.com",
			&Record{
				Version:             "DMARC1",
				Policy:              PolicyNone,
				DKIMAlignment:       AlignmentModeRelaxed,
				SPFAlignment:        AlignmentModeStrict,
				FailureOptions:      []FailureOption{FailureOptionAll},
				ReportFormats:       []ReportFormat{ReportFormatAFRF},
				ReportInterval:      86400 * time.Second,
				ReportURIsAggregate: []string{"mailto:test@example.com"},
			},
			false,
		},
		{
			"valid record", "v=DMARC1; p=reject; adkim=s; aspf=r; fo=0:1:s; rf=afrf; ri=3600; rua=mailto:test@example.com,mailto:anothertest@example2.uk",
			&Record{
				Version:             "DMARC1",
				Policy:              PolicyReject,
				DKIMAlignment:       AlignmentModeStrict,
				SPFAlignment:        AlignmentModeRelaxed,
				FailureOptions:      []FailureOption{FailureOptionAll, FailureOptionAny, FailureOptionSPF},
				ReportFormats:       []ReportFormat{ReportFormatAFRF},
				ReportInterval:      3600 * time.Second,
				ReportURIsAggregate: []string{"mailto:test@example.com", "mailto:anothertest@example2.uk"},
			},
			false},
		{
			"invalid record",
			"v=DMARC1; p=none; adkim=r; aspf=r; fo=0; rf=afrf; ri=86400; rua=mailto:test@example.com; sp",
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRecord(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid record")
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestParsePolicy(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    Policy
		wantErr bool
	}{
		{"valid p=none", "none", PolicyNone, false},
		{"valid p=quarantine", "quarantine", PolicyQuarantine, false},
		{"valid p=reject", "reject", PolicyReject, false},
		{"invalid p=unknown", "unknown", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePolicy(tc.input, "p")
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid policy")
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if got != tc.want {
				t.Errorf("Expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestParseAlignmentMode(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    AlignmentMode
		wantErr bool
	}{
		{"valid adkim=r", "r", AlignmentModeRelaxed, false},
		{"valid adkim=s", "s", AlignmentModeStrict, false},
		{"invalid adkim=x", "x", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAlignmentMode(tc.input, "adkim")
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid alignment mode")
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if got != tc.want {
				t.Errorf("Expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestParseFailureOptions(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []FailureOption
		wantErr bool
	}{
		{"valid fo=0", "0", []FailureOption{FailureOptionAll}, false},
		{"valid fo=1", "1", []FailureOption{FailureOptionAny}, false},
		{"valid fo=d", "d", []FailureOption{FailureOptionDKIM}, false},
		{"valid fo=s", "s", []FailureOption{FailureOptionSPF}, false},
		{"valid fo=d:s", "d:s", []FailureOption{FailureOptionDKIM, FailureOptionSPF}, false},
		{"invalid fo=x", "x", []FailureOption{FailureOptionAll}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFailureOptions(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid failure option")
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(got) != len(tc.want) {
				t.Errorf("Expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestParseParams(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			"valid params",
			"p=none;adkim=r;aspf=r;fo=0;rf=afrf;ri=86400;rua=mailto:test@example.com;sp=none",
			map[string]string{"p": "none", "adkim": "r", "aspf": "r", "fo": "0", "rf": "afrf", "ri": "86400", "rua": "mailto:test@example.com", "sp": "none"},
			false,
		},
		{
			"invalid params",
			"p=none;adkim=r;aspf=r;fo=0;rf=afrf;ri=86400;rua=mailto:test@example.com;sp",
			map[string]string{"p": "none", "adkim": "r", "aspf": "r", "fo": "0", "rf": "afrf", "ri": "86400", "rua": "mailto:test@example.com"},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseParams(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid params")
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(got) != len(tc.want) {
				t.Errorf("Expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestParseURIList(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"valid uri list", "mailto:test@example.com,mailto:anothertest@example2.com", []string{"mailto:test@example.com", "mailto:anothertest@example2.com"}, false},
		{"invalid uri list", "mailto:a.com,http://b.com", []string{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseURIList(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid URI list")
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(got) != len(tc.want) {
				t.Errorf("Expected %v, got %v", tc.want, got)
			}
		})
	}
}
