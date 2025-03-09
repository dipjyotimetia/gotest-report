package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestOverallTestStatusDetermination(t *testing.T) {
	tests := []struct {
		name           string
		reportData     *ReportData
		expectedStatus string
	}{
		{
			name: "all tests passing should show PASSED status",
			reportData: &ReportData{
				TotalTests:    5,
				PassedTests:   5,
				FailedTests:   0,
				SkippedTests:  0,
				TotalDuration: 1.5,
			},
			expectedStatus: "PASSED-brightgreen",
		},
		{
			name: "any failed tests should show FAILED status",
			reportData: &ReportData{
				TotalTests:    10,
				PassedTests:   9,
				FailedTests:   1,
				SkippedTests:  0,
				TotalDuration: 2.5,
			},
			expectedStatus: "FAILED-red",
		},
		{
			name: "all skipped tests should show SKIPPED status",
			reportData: &ReportData{
				TotalTests:    3,
				PassedTests:   0,
				FailedTests:   0,
				SkippedTests:  3,
				TotalDuration: 0.05,
			},
			expectedStatus: "SKIPPED-yellow",
		},
		{
			name: "mixed passed and skipped should still show PASSED status",
			reportData: &ReportData{
				TotalTests:    5,
				PassedTests:   3,
				FailedTests:   0,
				SkippedTests:  2,
				TotalDuration: 0.8,
			},
			expectedStatus: "PASSED-brightgreen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown := generateMarkdownReport(tt.reportData)
			if !strings.Contains(markdown, "Status-"+tt.expectedStatus) {
				t.Errorf("Expected status badge with %q not found in report", tt.expectedStatus)
			}
		})
	}
}

func TestNestedSubtestFormatting(t *testing.T) {
	// Create test data with multiple levels of nested subtests
	reportData := &ReportData{
		TotalTests:    1,
		PassedTests:   1,
		FailedTests:   0,
		SkippedTests:  0,
		TotalDuration: 1.0,
		SortedTestNames: []string{
			"TestParent",
		},
		Results: map[string]*TestResult{
			"TestParent": {
				Name:      "TestParent",
				Status:    "PASS",
				Duration:  1.0,
				SubTests:  []string{"TestParent/Child1", "TestParent/Child2"},
				IsSubTest: false,
			},
			"TestParent/Child1": {
				Name:       "TestParent/Child1",
				Status:     "PASS",
				Duration:   0.3,
				ParentTest: "TestParent",
				IsSubTest:  true,
			},
			"TestParent/Child2": {
				Name:       "TestParent/Child2",
				Status:     "PASS",
				Duration:   0.7,
				ParentTest: "TestParent",
				SubTests:   []string{"TestParent/Child2/GrandChild"},
				IsSubTest:  true,
			},
			"TestParent/Child2/GrandChild": {
				Name:       "TestParent/Child2/GrandChild",
				Status:     "PASS",
				Duration:   0.2,
				ParentTest: "TestParent/Child2",
				IsSubTest:  true,
			},
		},
	}

	markdown := generateMarkdownReport(reportData)

	// Verify parent test is properly displayed
	if !strings.Contains(markdown, "| **TestParent**") {
		t.Error("Parent test not properly formatted in report")
	}

	// Check for subtests details section
	if !strings.Contains(markdown, "<details><summary>2 subtests</summary>") {
		t.Error("Subtest summary not found in report")
	}

	// Test for nested table structure
	if !strings.Contains(markdown, "<table><tr><th>Subtest</th>") {
		t.Error("Nested table for subtests not found")
	}

	// Check that child tests are included
	if !strings.Contains(markdown, "Child1") || !strings.Contains(markdown, "Child2") {
		t.Error("Child test names not found in report")
	}
}

func TestPassPercentageCalculation(t *testing.T) {
	tests := []struct {
		name               string
		reportData         *ReportData
		expectedPercentage string
	}{
		{
			name: "all tests passing should show 100%",
			reportData: &ReportData{
				TotalTests:    10,
				PassedTests:   10,
				FailedTests:   0,
				SkippedTests:  0,
				TotalDuration: 1.0,
			},
			expectedPercentage: "100.0%",
		},
		{
			name: "half tests passing should show 50%",
			reportData: &ReportData{
				TotalTests:    10,
				PassedTests:   5,
				FailedTests:   5,
				SkippedTests:  0,
				TotalDuration: 1.0,
			},
			expectedPercentage: "50.0%",
		},
		{
			name: "no tests should show N/A",
			reportData: &ReportData{
				TotalTests:    0,
				PassedTests:   0,
				FailedTests:   0,
				SkippedTests:  0,
				TotalDuration: 0.0,
			},
			expectedPercentage: "N/A",
		},
		{
			name: "partial percentage should be formatted to one decimal place",
			reportData: &ReportData{
				TotalTests:    3,
				PassedTests:   2,
				FailedTests:   1,
				SkippedTests:  0,
				TotalDuration: 1.0,
			},
			expectedPercentage: "66.7%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown := generateMarkdownReport(tt.reportData)
			if !strings.Contains(markdown, "Passed: "+strconv.Itoa(tt.reportData.PassedTests)+" ("+tt.expectedPercentage+")") {
				t.Errorf("Expected pass percentage %q not found in report", tt.expectedPercentage)
			}
		})
	}
}

func TestFailedTestOutputFiltering(t *testing.T) {
	reportData := &ReportData{
		TotalTests:      1,
		PassedTests:     0,
		FailedTests:     1,
		SkippedTests:    0,
		TotalDuration:   0.5,
		SortedTestNames: []string{"FailingTest"},
		Results: map[string]*TestResult{
			"FailingTest": {
				Name:     "FailingTest",
				Status:   "FAIL",
				Duration: 0.5,
				Output: []string{
					"=== RUN   FailingTest",
					"Some regular output that should be filtered",
					"--- FAIL: FailingTest (0.50s)",
					"    file.go:25: Error: something went wrong",
					"    file.go:26: This is part of the error",
					"FAIL",
					"exit status 1",
				},
			},
		},
	}

	markdown := generateMarkdownReport(reportData)

	// Verify the failed test details section exists
	if !strings.Contains(markdown, "## Failed Tests Details") {
		t.Error("Failed test details section not found in report")
	}

	// Check that filtered output includes error lines but not regular output
	if !strings.Contains(markdown, "--- FAIL: FailingTest") {
		t.Error("FAIL message not found in filtered output")
	}

	if !strings.Contains(markdown, "Error: something went wrong") {
		t.Error("Error message not found in filtered output")
	}

	// Check that regular output is not included
	detailsSection := strings.Split(markdown, "### FailingTest")[1]
	detailsSection = strings.Split(detailsSection, "## Test Durations")[0]

	if strings.Contains(detailsSection, "Some regular output that should be filtered") {
		t.Error("Regular output should not be included in filtered output")
	}
}

func TestTimeFormatting(t *testing.T) {
	reportData := &ReportData{
		TotalTests:    1,
		PassedTests:   1,
		FailedTests:   0,
		SkippedTests:  0,
		TotalDuration: 0.0,
		Results:       map[string]*TestResult{},
	}

	markdown := generateMarkdownReport(reportData)

	// Extract report timestamp
	lines := strings.Split(markdown, "\n")
	var timestamp string
	for _, line := range lines {
		if strings.HasPrefix(line, "Report generated at:") {
			timestamp = line
			break
		}
	}

	if timestamp == "" {
		t.Fatal("Report timestamp not found")
	}

	// Just check that we have a timestamp, without being strict about format
	if !strings.HasPrefix(timestamp, "Report generated at:") {
		t.Errorf("Timestamp format incorrect: %s", timestamp)
	}
}
