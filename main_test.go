package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/stretchr/testify/assert"
)

func TestProcessTestResults(t *testing.T) {
	testCases := []struct {
		name           string
		inputResults   []TestResult
		expectedOutput []PackageTestSummary
	}{
		{
			name: "Mixed Test Results",
			inputResults: []TestResult{
				{Action: "run", Package: "pkg1", Test: "Test1"},
				{Action: "pass", Package: "pkg1", Test: "Test1", Elapsed: 0.1},
				{Action: "run", Package: "pkg1", Test: "Test2"},
				{Action: "fail", Package: "pkg1", Test: "Test2", Output: "failed test"},
				{Action: "run", Package: "pkg1", Test: "Test3"},
				{Action: "skip", Package: "pkg1", Test: "Test3"},
			},
			expectedOutput: []PackageTestSummary{
				{
					Package:      "pkg1",
					TotalTests:   3,
					PassedTests:  1,
					FailedTests:  1,
					SkippedTests: 1,
					Duration:     time.Duration(100 * time.Millisecond),
					TestDetails: map[string]TestDetailInfo{
						"Test1": {Result: "PASS", Output: ""},
						"Test2": {Result: "FAIL", Output: "failed test"},
						"Test3": {Result: "SKIP", Output: ""},
					},
				},
			},
		},
		{
			name:           "Empty Input",
			inputResults:   []TestResult{},
			expectedOutput: []PackageTestSummary{},
		},
		{
			name: "Multiple Packages",
			inputResults: []TestResult{
				{Action: "run", Package: "pkg1", Test: "Test1"},
				{Action: "pass", Package: "pkg1", Test: "Test1"},
				{Action: "run", Package: "pkg2", Test: "Test2"},
				{Action: "fail", Package: "pkg2", Test: "Test2"},
			},
			expectedOutput: []PackageTestSummary{
				{
					Package:      "pkg1",
					TotalTests:   1,
					PassedTests:  1,
					FailedTests:  0,
					SkippedTests: 0,
					TestDetails: map[string]TestDetailInfo{
						"Test1": {Result: "PASS", Output: ""},
					},
				},
				{
					Package:      "pkg2",
					TotalTests:   1,
					PassedTests:  0,
					FailedTests:  1,
					SkippedTests: 0,
					TestDetails: map[string]TestDetailInfo{
						"Test2": {Result: "FAIL", Output: ""},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := processTestResults(tc.inputResults)
			assert.Equal(t, tc.expectedOutput, result)
		})
	}
}

func TestGenerateEnhancedReport(t *testing.T) {
	testCases := []struct {
		name           string
		inputResults   []TestResult
		expectedOutput func(string) bool
	}{
		{
			name: "Basic Report Generation",
			inputResults: []TestResult{
				{Action: "run", Package: "pkg1", Test: "Test1"},
				{Action: "pass", Package: "pkg1", Test: "Test1"},
			},
			expectedOutput: func(report string) bool {
				return strings.Contains(report, "Go Test Report") &&
					strings.Contains(report, "Total Tests") &&
					strings.Contains(report, "Passed Tests")
			},
		},
		{
			name: "Multiple Package Report",
			inputResults: []TestResult{
				{Action: "run", Package: "pkg1", Test: "Test1"},
				{Action: "pass", Package: "pkg1", Test: "Test1"},
				{Action: "run", Package: "pkg2", Test: "Test2"},
				{Action: "fail", Package: "pkg2", Test: "Test2"},
			},
			expectedOutput: func(report string) bool {
				return strings.Contains(report, "pkg1") &&
					strings.Contains(report, "pkg2") &&
					strings.Contains(report, "Passed Tests") &&
					strings.Contains(report, "Failed Tests")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			report := generateEnhancedReport(tc.inputResults)
			assert.True(t, tc.expectedOutput(report))
		})
	}
}

func TestGenerateOverallSummaryTable(t *testing.T) {
	testCases := []struct {
		name       string
		summaries  []PackageTestSummary
		assertions func(table.Writer) bool
	}{
		{
			name: "Single Package Summary",
			summaries: []PackageTestSummary{
				{
					Package:      "pkg1",
					TotalTests:   10,
					PassedTests:  8,
					FailedTests:  1,
					SkippedTests: 1,
					Duration:     time.Second,
				},
			},
			assertions: func(w table.Writer) bool {
				rendered := w.Render()
				return strings.Contains(rendered, "Total Packages") &&
					strings.Contains(rendered, "Total Tests") &&
					strings.Contains(rendered, "Passed Tests") &&
					strings.Contains(rendered, "80.0%")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			summaryTable := generateOverallSummaryTable(tc.summaries)
			assert.True(t, tc.assertions(summaryTable))
		})
	}
}

// Performance and Edge Case Test
func BenchmarkProcessTestResults(b *testing.B) {
	largeTestResults := generateLargeTestResultSet(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processTestResults(largeTestResults)
	}
}

func generateLargeTestResultSet(count int) []TestResult {
	results := make([]TestResult, 0, count)
	packages := []string{"pkg1", "pkg2", "pkg3"}
	actions := []string{"run", "pass", "fail", "skip"}

	for i := 0; i < count; i++ {
		result := TestResult{
			Package: packages[i%len(packages)],
			Test:    fmt.Sprintf("Test%d", i),
			Action:  actions[i%len(actions)],
			Elapsed: float64(i) * 0.01,
		}
		results = append(results, result)
	}

	return results
}
