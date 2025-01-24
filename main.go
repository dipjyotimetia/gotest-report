package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TestResult represents the structure of a single test result from go test -json output
type TestResult struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

// TestSummary aggregates test results by package
type TestSummary struct {
	Package      string
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	Duration     time.Duration
	TestResults  []TestResult
}

func generateMarkdownReport(testResults []TestResult) string {
	// Group test results by package
	packageSummaries := make(map[string]*TestSummary)

	for _, result := range testResults {
		if result.Package == "" {
			continue
		}

		// Initialize package summary if not exists
		if _, exists := packageSummaries[result.Package]; !exists {
			packageSummaries[result.Package] = &TestSummary{
				Package: result.Package,
			}
		}

		summary := packageSummaries[result.Package]
		summary.TestResults = append(summary.TestResults, result)

		// Count test statuses
		switch result.Action {
		case "run":
			summary.TotalTests++
		case "pass":
			summary.PassedTests++
		case "fail":
			summary.FailedTests++
		case "skip":
			summary.SkippedTests++
		}

		// Track duration
		if result.Elapsed > 0 {
			summary.Duration += time.Duration(result.Elapsed * float64(time.Second))
		}
	}

	// Generate markdown report
	var report strings.Builder
	report.WriteString("# Go Test Report\n\n")
	report.WriteString("## Test Summary\n\n")

	var totalPackages, totalTests, totalPassed, totalFailed, totalSkipped int
	totalDuration := time.Duration(0)

	for _, summary := range packageSummaries {
		totalPackages++
		totalTests += summary.TotalTests
		totalPassed += summary.PassedTests
		totalFailed += summary.FailedTests
		totalSkipped += summary.SkippedTests
		totalDuration += summary.Duration

		report.WriteString(fmt.Sprintf("### Package: `%s`\n\n", summary.Package))
		report.WriteString(fmt.Sprintf("- **Total Tests:** %d\n", summary.TotalTests))
		report.WriteString(fmt.Sprintf("- **Passed:** %d ✅\n", summary.PassedTests))
		report.WriteString(fmt.Sprintf("- **Failed:** %d ❌\n", summary.FailedTests))
		report.WriteString(fmt.Sprintf("- **Skipped:** %d ⏩\n", summary.SkippedTests))
		report.WriteString(fmt.Sprintf("- **Duration:** %s\n\n", summary.Duration.Round(time.Millisecond)))

		// Detailed test results
		if summary.FailedTests > 0 {
			report.WriteString("#### Failed Tests\n\n")
			for _, result := range summary.TestResults {
				if result.Action == "fail" {
					report.WriteString(fmt.Sprintf("- **%s**\n", result.Test))
					if result.Output != "" {
						report.WriteString(fmt.Sprintf("  ```\n%s  ```\n", result.Output))
					}
				}
			}
			report.WriteString("\n")
		}
	}

	// Overall summary
	report.WriteString("## Overall Summary\n\n")
	report.WriteString(fmt.Sprintf("- **Total Packages:** %d\n", totalPackages))
	report.WriteString(fmt.Sprintf("- **Total Tests:** %d\n", totalTests))
	report.WriteString(fmt.Sprintf("- **Total Passed:** %d ✅\n", totalPassed))
	report.WriteString(fmt.Sprintf("- **Total Failed:** %d ❌\n", totalFailed))
	report.WriteString(fmt.Sprintf("- **Total Skipped:** %d ⏩\n", totalSkipped))
	report.WriteString(fmt.Sprintf("- **Total Duration:** %s\n", totalDuration.Round(time.Millisecond)))

	return report.String()
}

func main() {
	// Check if a test.json file is provided as an argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-report-generator.go test.json")
		os.Exit(1)
	}

	// Read the test.json file
	filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON lines
	var testResults []TestResult
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	for decoder.More() {
		var result TestResult
		err := decoder.Decode(&result)
		if err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			os.Exit(1)
		}
		testResults = append(testResults, result)
	}

	// Generate markdown report
	markdownReport := generateMarkdownReport(testResults)

	// Write report to a markdown file
	outputFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + "_report.md"
	err = os.WriteFile(outputFilename, []byte(markdownReport), 0o644)
	if err != nil {
		fmt.Printf("Error writing markdown report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Markdown report generated: %s\n", outputFilename)
}
