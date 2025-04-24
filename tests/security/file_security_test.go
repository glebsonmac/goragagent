package security

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"goragagent/cmd"

	"github.com/stretchr/testify/assert"
)

func TestMaliciousFileInput(t *testing.T) {
	t.Log("Running critical security tests for file input validation...")

	// Create test data directory if it doesn't exist
	err := os.MkdirAll("data/test", 0755)
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll("data/test")

	testCases := []struct {
		name          string
		content       string
		expectedError string
	}{
		{
			name: "Path Traversal Attempt",
			content: `location,tax_rate,source
../../../etc/passwd,1.0%,malicious.txt`,
			expectedError: "invalid data at row 2: invalid character in path",
		},
		{
			name: "Command Injection Attempt",
			content: `location,tax_rate,source
$(rm -rf /),1.0%,malicious.txt`,
			expectedError: "invalid data at row 2: invalid character in field",
		},
		{
			name:          "Null Byte Injection",
			content:       "location,tax_rate,source\nTravis\x00County,1.0%,malicious.txt",
			expectedError: "invalid data at row 2: invalid character in field",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing %s...", tc.name)

			// Create test file in data/test directory
			testFile := filepath.Join("data", "test", fmt.Sprintf("malicious_%s.csv", tc.name))
			err := os.WriteFile(testFile, []byte(tc.content), 0644)
			assert.NoError(t, err, "Failed to create test file")
			defer os.Remove(testFile)

			// Test the file
			_, err = cmd.LoadData(testFile, "test")

			assert.Error(t, err, "Expected an error but got none")
			if err != nil {
				assert.Contains(t, err.Error(), tc.expectedError, "Error message doesn't match expected")
				t.Logf("✓ Successfully detected malicious input: %v", err)
			}
		})
	}
}

func TestPathValidation(t *testing.T) {
	t.Log("Running critical path validation security tests...")

	testCases := []struct {
		name          string
		path          string
		expectedError string
	}{
		{
			name:          "Parent Directory Traversal",
			path:          "../../../etc/shadow.csv",
			expectedError: "invalid path: directory traversal not allowed",
		},
		{
			name:          "Command Injection in Path",
			path:          "$(rm -rf /).csv",
			expectedError: "invalid character in field",
		},
		{
			name:          "Null Byte in Path",
			path:          "data/test/test\x00.csv",
			expectedError: "error opening file: open data/test/test\x00.csv: invalid argument",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing path: %s", tc.path)
			_, err := cmd.LoadData(tc.path, "test")
			assert.Error(t, err, "Expected an error but got none")
			if err != nil {
				assert.Contains(t, err.Error(), tc.expectedError, "Error message doesn't match expected")
				t.Logf("✓ Successfully blocked malicious path: %v", err)
			}
		})
	}
}

func generateLargeCSV() string {
	// Generate a CSV file larger than 10MB
	var content string
	content = "location,tax_rate,source\n"
	for i := 0; i < 1000000; i++ {
		content += fmt.Sprintf("Location%d,%.2f%%,source%d.txt\n", i, float64(i)/100, i)
	}
	return content
}
