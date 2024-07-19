package dmarc

import (
	"io/ioutil"
	"testing"
)

type TestCases struct {
	FileName string
	Valid    bool
}

func TestUnmarshalXml(t *testing.T) {
	testCases := []TestCases{
		{
			FileName: "./testdata/00-empty-valid.xml",
			Valid:    true,
		},
		{
			FileName: "./testdata/01-multiple-valid.xml",
			Valid:    true,
		},
		{
			FileName: "./testdata/02-single-valid.xml",
			Valid:    true,
		},
		{
			FileName: "./testdata/03-invalid.xml",
			Valid:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.FileName, func(t *testing.T) {
			// Read the XML file
			data, err := ioutil.ReadFile(tc.FileName)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", tc.FileName, err)
			}

			// Create a Feedback object
			var feedback Feedback

			// Parse the XML
			err = feedback.Parse(data)

			if tc.Valid && err != nil {
				t.Errorf("expected file %s to be valid, but got error: %v", tc.FileName, err)
			} else if !tc.Valid && err == nil {
				t.Errorf("expected file %s to be invalid, but parsing succeeded", tc.FileName)
			}
		})
	}
}
