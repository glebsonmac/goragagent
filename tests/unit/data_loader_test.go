package unit

import (
	"os"
	"path/filepath"
	"testing"

	"goragagent/cmd"

	"github.com/stretchr/testify/assert"
)

// Mock TaxRecord for testing
type TaxRecord struct {
	Location string
	TaxRate  string
	Source   string
}

func TestLoadData(t *testing.T) {
	t.Log("Running data loader tests...")

	// Create test data directory if it doesn't exist
	err := os.MkdirAll("data/test", 0755)
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll("data/test")

	// Test cases
	tests := []struct {
		name        string
		csvContent  string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid CSV",
			csvContent: `location,tax_rate,source
Travis County,1.9%,tax_policies_2023.pdf
Williamson County,2.1%,tax_records_2023.csv`,
			expectError: false,
		},
		{
			name:        "Empty CSV",
			csvContent:  `location,tax_rate,source`,
			expectError: true,
			errorMsg:    "CSV file is empty or missing data rows",
		},
		{
			name: "Invalid Format",
			csvContent: `location,tax_rate,source
Travis County,1.9%`,
			expectError: true,
			errorMsg:    "invalid data at row 2: record on line 2: wrong number of fields",
		},
		{
			name:        "Empty File",
			csvContent:  "",
			expectError: true,
			errorMsg:    "CSV file is empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing %s...", tc.name)

			// Create test file in data/test directory
			testFile := filepath.Join("data", "test", tc.name+".csv")
			err := os.WriteFile(testFile, []byte(tc.csvContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}
			defer os.Remove(testFile)

			// Test loadData function
			records, err := cmd.LoadData(testFile, "tax")

			if tc.expectError {
				assert.Error(t, err, "Expected error but got none")
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg,
						"Error message doesn't match expected")
				}
				t.Logf("✓ Successfully detected invalid input: %v", err)
			} else {
				assert.NoError(t, err, "Got unexpected error")
				assert.NotNil(t, records, "Records should not be nil")
				assert.Greater(t, len(records), 0, "Records should not be empty")
				t.Logf("✓ Successfully loaded %d records", len(records))
			}
		})
	}
}

func TestLoadDataFileNotFound(t *testing.T) {
	t.Log("Testing file not found scenario...")

	_, err := cmd.LoadData("data/test/nonexistent.csv", "tax")
	assert.Error(t, err, "Expected error for nonexistent file")
	assert.Contains(t, err.Error(), "error opening file",
		"Error message should indicate file not found")
	t.Log("✓ Successfully detected nonexistent file")
}

func TestLoadDataWithSpecialCharacters(t *testing.T) {
	t.Log("Testing CSV with special characters...")

	// Create test data directory if it doesn't exist
	err := os.MkdirAll("data/test", 0755)
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll("data/test")

	csvContent := `location,tax_rate,source
"Travis County, TX",1.9%,"tax_policies_2023.pdf"
"São Paulo, BR","2,1%","tax_records_2023.csv"`

	testFile := filepath.Join("data", "test", "special_chars.csv")
	err = os.WriteFile(testFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove(testFile)

	records, err := cmd.LoadData(testFile, "tax")
	assert.NoError(t, err, "Failed to load CSV with special characters")
	assert.Equal(t, 2, len(records), "Should have loaded 2 records")
	assert.Equal(t, "Travis County, TX", records[0].Location,
		"First record location mismatch")
	assert.Equal(t, "São Paulo, BR", records[1].Location,
		"Second record location mismatch")
	t.Log("✓ Successfully handled special characters in CSV")
}
