package dmarc

import (
	"os"
	"testing"
)

type TestCases struct {
	FileName        string
	Valid           bool
	PassCount       int
	QuarantineCount int
	RejectCount     int
}

func TestParseXML(t *testing.T) {
	testCases := []TestCases{
		{
			FileName:        "./testdata/00-empty-valid.xml",
			Valid:           true,
			PassCount:       0,
			QuarantineCount: 0,
			RejectCount:     0,
		},
		{
			FileName:        "./testdata/01-multiple-valid.xml",
			Valid:           true,
			PassCount:       2,
			QuarantineCount: 1,
			RejectCount:     4,
		},
		{
			FileName:        "./testdata/02-single-valid.xml",
			Valid:           true,
			PassCount:       1,
			QuarantineCount: 0,
			RejectCount:     0,
		},
		{
			FileName: "./testdata/03-invalid.xml",
			Valid:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.FileName, func(t *testing.T) {
			// Read the XML file
			data, err := os.ReadFile(tc.FileName)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", tc.FileName, err)
			}

			// Create a Feedback object
			var feedback RUA

			// Parse the XML
			err = feedback.ParseXML(data)

			if tc.Valid && err != nil {
				t.Errorf("expected file %s to be valid, but got error: %v", tc.FileName, err)
			} else if !tc.Valid && err == nil {
				t.Errorf("expected file %s to be invalid, but parsing succeeded", tc.FileName)
			}

			if tc.Valid {
				// Get the total count of emails for each disposition
				passCount := 0
				quarantineCount := 0
				rejectCount := 0
				for _, record := range feedback.Records {
					switch record.Row.PolicyEvaluated.Disposition {
					case "none":
						passCount += record.Row.Count
					case "quarantine":
						quarantineCount += record.Row.Count
					case "reject":
						rejectCount += record.Row.Count
					}
				}

				if passCount != tc.PassCount {
					t.Errorf("expected pass count to be %d, got %d", tc.PassCount, passCount)
				}

				if quarantineCount != tc.QuarantineCount {
					t.Errorf("expected quarantine count to be %d, got %d", tc.QuarantineCount, quarantineCount)
				}

				if rejectCount != tc.RejectCount {
					t.Errorf("expected reject count to be %d, got %d", tc.RejectCount, rejectCount)
				}
			}
		})
	}
}
