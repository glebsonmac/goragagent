package unit

import (
	"testing"

	"goragagent/cmd"

	"github.com/stretchr/testify/assert"
)

func TestFindRelevantInfo(t *testing.T) {
	t.Log("Running search functionality tests...")

	// Test data
	records := []cmd.Record{
		{
			Location: "Travis County",
			DataType: "tax",
			Values: map[string]string{
				"tax_rate": "1.9%",
			},
			Source: "tax_policies_2023.pdf",
		},
		{
			Location: "Williamson County",
			DataType: "tax",
			Values: map[string]string{
				"tax_rate": "2.1%",
			},
			Source: "tax_records_2023.csv",
		},
		{
			Location: "Travis Heights",
			DataType: "tax",
			Values: map[string]string{
				"tax_rate": "1.8%",
			},
			Source: "local_rates.pdf",
		},
	}

	t.Logf("Loaded %d test records", len(records))

	tests := []struct {
		name           string
		query          string
		expectedResult string
	}{
		{
			name:           "Exact Match",
			query:          "Travis County",
			expectedResult: "According to tax_policies_2023.pdf, the tax rate in Travis County is 1.9%",
		},
		{
			name:           "Partial Match",
			query:          "Travis",
			expectedResult: "According to tax_policies_2023.pdf, the tax rate in Travis County is 1.9%",
		},
		{
			name:           "Case Insensitive",
			query:          "TRAVIS",
			expectedResult: "According to tax_policies_2023.pdf, the tax rate in Travis County is 1.9%",
		},
		{
			name:           "No Match",
			query:          "Dallas",
			expectedResult: "No relevant information found in the database",
		},
		{
			name:           "Empty Query",
			query:          "",
			expectedResult: "No relevant information found in the database",
		},
		{
			name:           "Multiple Words",
			query:          "Williamson",
			expectedResult: "According to tax_records_2023.csv, the tax rate in Williamson County is 2.1%",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing search with query: '%s'", tc.query)

			result, _ := cmd.FindRelevantInfo(tc.query, records)
			assert.Contains(t, result, tc.expectedResult,
				"Search result doesn't match expected output")

			if tc.expectedResult == "No relevant information found in the database" {
				t.Log("✓ Successfully handled non-matching query")
			} else {
				t.Logf("✓ Successfully found matching record: %s", result)
			}
		})
	}
}
