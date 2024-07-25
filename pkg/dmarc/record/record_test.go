package record

import (
	"reflect"
	"testing"
	"time"
)

func TestParseRecord(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		want        *Record
		wantErr     bool
		expectedErr *DMARCError
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
			nil,
		},
		{
			"valid record",
			"v=DMARC1; p=reject; adkim=s; aspf=r; fo=0:1:s; rf=afrf; ri=3600; rua=mailto:test@example.com,mailto:anothertest@example2.uk",
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
			false,
			nil,
		},
		{
			"invalid record",
			"v=DMARC1; p=none; adkim=r; aspf=r; fo=0; rf=afrf; ri=86400; rua=mailto:test@example.com; sp",
			nil,
			true,
			&DMARCError{Parameter: "sp", Message: ErrMalformedParam.Error()},
		},
		{
			"invalid record",
			"v=DMARC1; p=none; adkim=aaa; aspf=r; fo=0; rf=afrf; ri=86400;",
			nil,
			true,
			&DMARCError{Parameter: "adkim", Message: ErrInvalidParam.Error()},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRecord(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid record")
				} else {
					if dmarcErr, ok := err.(*DMARCError); ok {
						if dmarcErr.Parameter != tc.expectedErr.Parameter || dmarcErr.Message != tc.expectedErr.Message {
							t.Errorf("Expected error %v, got %v", tc.expectedErr, dmarcErr)
						}
					} else {
						t.Errorf("Expected DMARCError, got %v", err)
					}
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
		name        string
		input       string
		want        Policy
		wantErr     bool
		expectedErr *DMARCError
	}{
		{"valid p=none", "none", PolicyNone, false, nil},
		{"valid p=quarantine", "quarantine", PolicyQuarantine, false, nil},
		{"valid p=reject", "reject", PolicyReject, false, nil},
		{"invalid p=unknown", "unknown", "", true, &DMARCError{Parameter: "p", Message: ErrInvalidParam.Error()}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePolicy(tc.input, "p")
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid policy")
				} else {
					if dmarcErr, ok := err.(*DMARCError); ok {
						if dmarcErr.Parameter != tc.expectedErr.Parameter || dmarcErr.Message != tc.expectedErr.Message {
							t.Errorf("Expected error %v, got %v", tc.expectedErr, dmarcErr)
						}
					} else {
						t.Errorf("Expected DMARCError, got %v", err)
					}
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
		name        string
		input       string
		want        AlignmentMode
		wantErr     bool
		expectedErr *DMARCError
	}{
		{"valid adkim=r", "r", AlignmentModeRelaxed, false, nil},
		{"valid adkim=s", "s", AlignmentModeStrict, false, nil},
		{"invalid adkim=x", "x", "", true, &DMARCError{Parameter: "adkim", Message: ErrInvalidParam.Error()}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAlignmentMode(tc.input, "adkim")
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid alignment mode")
				} else {
					if dmarcErr, ok := err.(*DMARCError); ok {
						if dmarcErr.Parameter != tc.expectedErr.Parameter || dmarcErr.Message != tc.expectedErr.Message {
							t.Errorf("Expected error %v, got %v", tc.expectedErr, dmarcErr)
						}
					} else {
						t.Errorf("Expected DMARCError, got %v", err)
					}
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
		name        string
		input       string
		want        []FailureOption
		wantErr     bool
		expectedErr *DMARCError
	}{
		{"valid fo=0", "0", []FailureOption{FailureOptionAll}, false, nil},
		{"valid fo=1", "1", []FailureOption{FailureOptionAny}, false, nil},
		{"valid fo=d", "d", []FailureOption{FailureOptionDKIM}, false, nil},
		{"valid fo=s", "s", []FailureOption{FailureOptionSPF}, false, nil},
		{"valid fo=d:s", "d:s", []FailureOption{FailureOptionDKIM, FailureOptionSPF}, false, nil},
		{"invalid fo=x", "x", []FailureOption{FailureOptionAll}, true, &DMARCError{Parameter: "fo", Message: ErrInvalidParam.Error()}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFailureOptions(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid failure option")
				} else {
					if dmarcErr, ok := err.(*DMARCError); ok {
						if dmarcErr.Parameter != tc.expectedErr.Parameter || dmarcErr.Message != tc.expectedErr.Message {
							t.Errorf("Expected error %v, got %v", tc.expectedErr, dmarcErr)
						}
					} else {
						t.Errorf("Expected DMARCError, got %v", err)
					}
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

func TestParseParams(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		want        map[string]string
		wantErr     bool
		expectedErr *DMARCError
	}{
		{
			"valid params",
			"p=none;adkim=r;aspf=r;fo=0;rf=afrf;ri=86400;rua=mailto:test@example.com;sp=none",
			map[string]string{"p": "none", "adkim": "r", "aspf": "r", "fo": "0", "rf": "afrf", "ri": "86400", "rua": "mailto:test@example.com", "sp": "none"},
			false,
			nil,
		},
		{
			"invalid params",
			"p=none;adkim=r;aspf=r;fo=0;rf=afrf;ri=86400;rua=mailto:test@example.com;sp",
			nil,
			true,
			&DMARCError{Parameter: "sp", Message: ErrMalformedParam.Error()},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseParams(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid params")
				} else {
					if dmarcErr, ok := err.(*DMARCError); ok {
						if dmarcErr.Parameter != tc.expectedErr.Parameter || dmarcErr.Message != tc.expectedErr.Message {
							t.Errorf("Expected error %v, got %v", tc.expectedErr, dmarcErr)
						}
					} else {
						t.Errorf("Expected DMARCError, got %v", err)
					}
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			} else {
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("Expected %v, got %v", tc.want, got)
				}
			}
		})
	}
}

func TestParseURIList(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		want        []string
		wantErr     bool
		expectedErr *DMARCError
	}{
		{"valid uri list", "mailto:test@example.com,mailto:anothertest@example2.com", []string{"mailto:test@example.com", "mailto:anothertest@example2.com"}, false, nil},
		{"invalid uri list", "mailto:a.com,http://b.com", nil, true, &DMARCError{Parameter: "rua", Message: ErrInvalidURI.Error()}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseURIList(tc.input, "rua")
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid URI list")
				} else {
					if dmarcErr, ok := err.(*DMARCError); ok {
						if dmarcErr.Parameter != tc.expectedErr.Parameter || dmarcErr.Message != tc.expectedErr.Message {
							t.Errorf("Expected error %v, got %v", tc.expectedErr, dmarcErr)
						}
					} else {
						t.Errorf("Expected DMARCError, got %v", err)
					}
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
