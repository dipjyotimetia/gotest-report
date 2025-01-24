package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type TestResult struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

type PackageTestSummary struct {
	Package      string
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	Duration     time.Duration
	TestDetails  map[string]TestDetailInfo
}

type TestDetailInfo struct {
	Result string
	Output string
}

func generateEnhancedReport(testResults []TestResult) string {
	packageSummaries := processTestResults(testResults)
	var reportBuilder strings.Builder

	// Overall Summary Section
	reportBuilder.WriteString("# 🧪 Go Test Report\n\n")
	reportBuilder.WriteString(fmt.Sprintf("📅 Generated: %s\n\n", time.Now().Format(time.RFC1123)))

	// Overall Summary Table
	overallSummaryTable := generateOverallSummaryTable(packageSummaries)
	reportBuilder.WriteString("## 📊 Overall Test Summary\n\n")
	reportBuilder.WriteString(fmt.Sprintf("```\n%s\n```\n\n", overallSummaryTable.Render()))

	// Detailed Package Report
	reportBuilder.WriteString("## 📦 Package Test Details\n\n")
	for _, summary := range packageSummaries {
		packageDetailTable := generatePackageDetailTable(summary)
		reportBuilder.WriteString(fmt.Sprintf("### Package: `%s`\n\n", summary.Package))
		reportBuilder.WriteString(fmt.Sprintf("```\n%s\n```\n\n", packageDetailTable.Render()))

		// Test Details Section
		if len(summary.TestDetails) > 0 {
			reportBuilder.WriteString("#### Test Case Details\n\n")
			testDetailsTable := generateTestDetailsTable(summary.TestDetails)
			reportBuilder.WriteString(fmt.Sprintf("```\n%s\n```\n\n", testDetailsTable.Render()))
		}
	}

	return reportBuilder.String()
}

func processTestResults(testResults []TestResult) []PackageTestSummary {
	packageMap := make(map[string]*PackageTestSummary)

	for _, result := range testResults {
		if result.Package == "" {
			continue
		}

		if _, exists := packageMap[result.Package]; !exists {
			packageMap[result.Package] = &PackageTestSummary{
				Package:     result.Package,
				TestDetails: make(map[string]TestDetailInfo),
			}
		}

		pkg := packageMap[result.Package]

		switch result.Action {
		case "run":
			if result.Test != "" {
				pkg.TotalTests++
			}
		case "pass":
			if result.Test != "" {
				pkg.PassedTests++
				pkg.TestDetails[result.Test] = TestDetailInfo{Result: "PASS", Output: result.Output}
			}
		case "fail":
			if result.Test != "" {
				pkg.FailedTests++
				pkg.TestDetails[result.Test] = TestDetailInfo{Result: "FAIL", Output: result.Output}
			}
		case "skip":
			if result.Test != "" {
				pkg.SkippedTests++
				pkg.TestDetails[result.Test] = TestDetailInfo{Result: "SKIP", Output: result.Output}
			}
		}

		if result.Elapsed > 0 {
			pkg.Duration += time.Duration(result.Elapsed * float64(time.Second))
		}
	}

	// Convert map to sorted slice
	summaries := make([]PackageTestSummary, 0, len(packageMap))
	for _, summary := range packageMap {
		summaries = append(summaries, *summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Package < summaries[j].Package
	})

	return summaries
}

func generateOverallSummaryTable(summaries []PackageTestSummary) table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleColoredBright)
	t.Style().Options.SeparateColumns = true
	t.Style().Options.DrawBorder = true

	var totalTests, totalPassed, totalFailed, totalSkipped int
	var totalDuration time.Duration

	for _, summary := range summaries {
		totalTests += summary.TotalTests
		totalPassed += summary.PassedTests
		totalFailed += summary.FailedTests
		totalSkipped += summary.SkippedTests
		totalDuration += summary.Duration
	}

	t.AppendHeader(table.Row{"Metric", "Count", "Percentage"})
	t.AppendRows([]table.Row{
		{"Total Packages", len(summaries), ""},
		{"Total Tests", totalTests, "100%"},
		{"Passed Tests", totalPassed, fmt.Sprintf("%.1f%%", float64(totalPassed)/float64(totalTests)*100)},
		{"Failed Tests", totalFailed, fmt.Sprintf("%.1f%%", float64(totalFailed)/float64(totalTests)*100)},
		{"Skipped Tests", totalSkipped, fmt.Sprintf("%.1f%%", float64(totalSkipped)/float64(totalTests)*100)},
		{"Total Duration", totalDuration.Round(time.Millisecond), ""},
	})

	return t
}

func generatePackageDetailTable(summary PackageTestSummary) table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleColoredBright)
	t.Style().Options.SeparateColumns = true
	t.Style().Options.DrawBorder = true

	t.AppendHeader(table.Row{"Metric", "Count", "Percentage"})
	t.AppendRows([]table.Row{
		{"Total Tests", summary.TotalTests, "100%"},
		{"Passed Tests", summary.PassedTests, fmt.Sprintf("%.1f%%", float64(summary.PassedTests)/float64(summary.TotalTests)*100)},
		{"Failed Tests", summary.FailedTests, fmt.Sprintf("%.1f%%", float64(summary.FailedTests)/float64(summary.TotalTests)*100)},
		{"Skipped Tests", summary.SkippedTests, fmt.Sprintf("%.1f%%", float64(summary.SkippedTests)/float64(summary.TotalTests)*100)},
		{"Duration", summary.Duration.Round(time.Millisecond), ""},
	})

	return t
}

func generateTestDetailsTable(details map[string]TestDetailInfo) table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleColoredBright)
	t.Style().Options.SeparateColumns = true
	t.Style().Options.DrawBorder = true

	t.AppendHeader(table.Row{"Test Name", "Result", "Output"})

	for testName, detail := range details {
		var color text.Color
		switch detail.Result {
		case "PASS":
			color = text.FgGreen
		case "FAIL":
			color = text.FgRed
		case "SKIP":
			color = text.FgYellow
		}

		trimmedOutput := detail.Output
		if len(trimmedOutput) > 50 {
			trimmedOutput = trimmedOutput[:50] + "..."
		}

		t.AppendRow(table.Row{
			testName,
			text.Colors{color}.Sprint(detail.Result),
			trimmedOutput,
		})
	}

	return t
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-report-generator.go test.json")
		os.Exit(1)
	}

	filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

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

	markdownReport := generateEnhancedReport(testResults)

	outputFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + "_report.md"
	err = os.WriteFile(outputFilename, []byte(markdownReport), 0o644)
	if err != nil {
		fmt.Printf("Error writing markdown report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Enhanced markdown report generated: %s\n", outputFilename)
}
