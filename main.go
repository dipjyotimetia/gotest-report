package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TestResult struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

type TestSummary struct {
	Package        string
	TotalTests     int
	PassedTests    int
	FailedTests    int
	SkippedTests   int
	Duration       time.Duration
	TestResults    []TestResult
	PassedDetails  map[string]string
	FailedDetails  map[string]string
	SkippedDetails map[string]string
}

func calculatePercentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) * 100 / float64(total)
}

func generateMarkdownReport(testResults []TestResult) string {
	packageSummaries := make(map[string]*TestSummary)

	// Single pass processing with improved error handling
	for _, result := range testResults {
		if result.Package == "" {
			continue
		}

		if _, exists := packageSummaries[result.Package]; !exists {
			packageSummaries[result.Package] = &TestSummary{
				Package:        result.Package,
				PassedDetails:  make(map[string]string),
				FailedDetails:  make(map[string]string),
				SkippedDetails: make(map[string]string),
			}
		}

		summary := packageSummaries[result.Package]

		// Handle test actions with proper counting
		switch result.Action {
		case "run":
			if result.Test != "" { // Only count named tests
				summary.TotalTests++
			}
		case "pass":
			if result.Test != "" {
				summary.PassedTests++
				summary.PassedDetails[result.Test] = strings.TrimSpace(result.Output)
			}
		case "fail":
			if result.Test != "" {
				summary.FailedTests++
				summary.FailedDetails[result.Test] = strings.TrimSpace(result.Output)
			}
		case "skip":
			if result.Test != "" {
				summary.SkippedTests++
				summary.SkippedDetails[result.Test] = strings.TrimSpace(result.Output)
			}
		}

		// Accumulate duration properly
		if result.Elapsed > 0 {
			summary.Duration += time.Duration(result.Elapsed * float64(time.Second))
		}
	}

	var report strings.Builder
	report.WriteString("# Go Test Report\n\n")
	report.WriteString(fmt.Sprintf("Generated at: %s\n\n", time.Now().Format(time.RFC1123)))
	report.WriteString("## Test Summary\n\n")

	// Track totals with proper initialization
	totals := struct {
		packages, tests, passed, failed, skipped int
		duration                                 time.Duration
	}{}

	// Generate package summaries with improved formatting
	for _, summary := range packageSummaries {
		totals.packages++
		totals.tests += summary.TotalTests
		totals.passed += summary.PassedTests
		totals.failed += summary.FailedTests
		totals.skipped += summary.SkippedTests
		totals.duration += summary.Duration

		passRate := calculatePercentage(summary.PassedTests, summary.TotalTests)

		// Use consistent emoji and better spacing
		report.WriteString(fmt.Sprintf("<details>\n<summary><strong>📦 %s</strong> (%.1f%% Success)</summary>\n\n",
			sanitizePackageName(summary.Package), passRate))

		// Add table headers with proper alignment
		report.WriteString("| Metric | Count | Status |\n")
		report.WriteString("|:-------|:------:|:-------|\n")
		report.WriteString(fmt.Sprintf("| Total Tests | %d | |\n", summary.TotalTests))
		report.WriteString(fmt.Sprintf("| Passed | %d | ![](https://img.shields.io/badge/passed-%d-%%2373D216) |\n",
			summary.PassedTests, summary.PassedTests))
		report.WriteString(fmt.Sprintf("| Failed | %d | ![](https://img.shields.io/badge/failed-%d-red) |\n",
			summary.FailedTests, summary.FailedTests))
		report.WriteString(fmt.Sprintf("| Skipped | %d | ![](https://img.shields.io/badge/skipped-%d-yellow) |\n",
			summary.SkippedTests, summary.SkippedTests))
		report.WriteString(fmt.Sprintf("| Duration | %s | |\n\n", formatDuration(summary.Duration)))

		// Write test details sections with consistent formatting
		writeTestDetails(&report, "✅ Passed Tests", summary.PassedDetails)
		writeTestDetails(&report, "❌ Failed Tests", summary.FailedDetails)
		writeTestDetails(&report, "⏭️ Skipped Tests", summary.SkippedDetails)

		report.WriteString("</details>\n\n")
	}

	// Write overall summary with proper formatting
	writeOverallSummary(&report, totals)

	return report.String()
}

// Helper function to format test details sections
func writeTestDetails(report *strings.Builder, title string, details map[string]string) {
	if len(details) > 0 {
		report.WriteString(fmt.Sprintf("#### %s\n\n", title))
		for testName, output := range details {
			report.WriteString(fmt.Sprintf("<details>\n<summary><code>%s</code></summary>\n\n", testName))
			if output != "" {
				report.WriteString("```\n" + output + "```\n")
			}
			report.WriteString("</details>\n\n")
		}
	}
}

// Helper function to write overall summary section
func writeOverallSummary(report *strings.Builder, totals struct {
	packages, tests, passed, failed, skipped int
	duration                                 time.Duration
},
) {
	totalPassRate := calculatePercentage(totals.passed, totals.tests)
	report.WriteString("## 📊 Overall Summary\n\n")
	report.WriteString(fmt.Sprintf("![](https://img.shields.io/badge/Total%%20Success-%.1f%%25-%s)\n\n",
		totalPassRate, getColorForPercentage(totalPassRate)))

	report.WriteString("| Metric | Count |\n")
	report.WriteString("|:-------|:------|\n")
	report.WriteString(fmt.Sprintf("| Total Packages | %d |\n", totals.packages))
	report.WriteString(fmt.Sprintf("| Total Tests | %d |\n", totals.tests))
	report.WriteString(fmt.Sprintf("| Total Passed | %d |\n", totals.passed))
	report.WriteString(fmt.Sprintf("| Total Failed | %d |\n", totals.failed))
	report.WriteString(fmt.Sprintf("| Total Skipped | %d |\n", totals.skipped))
	report.WriteString(fmt.Sprintf("| Total Duration | %s |\n", formatDuration(totals.duration)))
}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	return d.Round(time.Millisecond).String()
}

// Helper function to sanitize package names
func sanitizePackageName(name string) string {
	return strings.TrimSpace(name)
}

func getColorForPercentage(percentage float64) string {
	switch {
	case percentage >= 90:
		return "brightgreen"
	case percentage >= 75:
		return "green"
	case percentage >= 50:
		return "yellow"
	default:
		return "red"
	}
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
