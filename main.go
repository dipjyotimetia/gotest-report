package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var version = "dev"

// Color constants for HTML styling
const (
	htmlPassColor    = "#2cbe4e"
	htmlFailColor    = "#cb2431"
	htmlSkipColor    = "#eea236"
	htmlNeutralColor = "#6a737d"
)

// TestEvent represents a single event from go test -json output
type TestEvent struct {
	Time    time.Time // Time when the event occurred
	Action  string    // Action: "run", "pause", "cont", "pass", "bench", "fail", "skip", "output"
	Test    string    // Test name
	Package string    // Package being tested
	Output  string    // Output text (for "output" action)
	Elapsed float64   // Elapsed time in seconds for "pass" or "fail" events
}

// TestResult holds the aggregated result for a single test
type TestResult struct {
	Name       string
	Package    string
	Status     string // "PASS", "FAIL", "SKIP"
	Duration   float64
	Output     []string
	ParentTest string // For subtests
	SubTests   []string
	IsSubTest  bool
}

// ReportData contains all data needed for the report
type ReportData struct {
	TotalTests      int
	PassedTests     int
	FailedTests     int
	SkippedTests    int
	TotalDuration   float64
	Results         map[string]*TestResult
	SortedTestNames []string
	PackageGroups   map[string][]string
}

func main() {
	inputFile := flag.String("input", "", "go test -json output file (default is stdin)")
	outputFile := flag.String("output", "test-report.md", "Output markdown file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("gotest-report version %s\n", version)
		os.Exit(0)
	}
	flag.Parse()

	var reader io.Reader = os.Stdin
	if *inputFile != "" {
		file, err := os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	}

	reportData, err := processTestEvents(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing test events: %v\n", err)
		os.Exit(1)
	}

	markdown := generateMarkdownReport(reportData)

	if err := os.WriteFile(*outputFile, []byte(markdown), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Report generated successfully: %s\n", *outputFile)
}

func processTestEvents(reader io.Reader) (*ReportData, error) {
	scanner := bufio.NewScanner(reader)
	results := make(map[string]*TestResult)
	testOutputMap := make(map[string][]string)

	testStartTime := make(map[string]time.Time)

	for scanner.Scan() {
		line := scanner.Text()
		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
		}

		testFullName := event.Test
		if testFullName == "" {
			// Skip package-level events
			continue
		}

		if _, exists := results[testFullName]; !exists && (event.Action == "run" || event.Action == "pass" || event.Action == "fail" || event.Action == "skip") {
			results[testFullName] = &TestResult{
				Name:      testFullName,
				Package:   event.Package,
				Status:    "UNKNOWN",
				Duration:  0,
				Output:    []string{},
				IsSubTest: strings.Contains(testFullName, "/"),
			}

			if results[testFullName].IsSubTest {
				parentName := testFullName[:strings.LastIndex(testFullName, "/")]
				results[testFullName].ParentTest = parentName

				if _, exists := results[parentName]; !exists {
					results[parentName] = &TestResult{
						Name:      parentName,
						Package:   event.Package,
						Status:    "UNKNOWN",
						Duration:  0,
						Output:    []string{},
						SubTests:  []string{},
						IsSubTest: strings.Contains(parentName, "/"),
					}
				}

				results[parentName].SubTests = append(results[parentName].SubTests, testFullName)
			}
		}

		switch event.Action {
		case "run":
			testStartTime[testFullName] = event.Time

		case "pass":
			results[testFullName].Status = "PASS"
			if event.Elapsed > 0 {
				results[testFullName].Duration = event.Elapsed
			} else if !testStartTime[testFullName].IsZero() {
				results[testFullName].Duration = event.Time.Sub(testStartTime[testFullName]).Seconds()
			}

		case "fail":
			results[testFullName].Status = "FAIL"
			if event.Elapsed > 0 {
				results[testFullName].Duration = event.Elapsed
			} else if !testStartTime[testFullName].IsZero() {
				results[testFullName].Duration = event.Time.Sub(testStartTime[testFullName]).Seconds()
			}

		case "skip":
			results[testFullName].Status = "SKIP"

		case "output":
			// Collect test output lines
			if _, exists := testOutputMap[testFullName]; !exists {
				testOutputMap[testFullName] = []string{}
			}
			// Clean output (remove trailing newlines)
			output := strings.TrimSuffix(event.Output, "\n")
			if output != "" {
				testOutputMap[testFullName] = append(testOutputMap[testFullName], output)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %v", err)
	}

	// Add collected output to each test
	for testName, output := range testOutputMap {
		if result, exists := results[testName]; exists {
			result.Output = output
		}
	}

	reportData := &ReportData{
		Results: results,
	}

	// Group tests by package
	packageGroups := make(map[string][]string)
	var sortedNames []string
	for name, result := range results {
		// Only count root tests in summary (not subtests)
		if !result.IsSubTest {
			sortedNames = append(sortedNames, name)
			reportData.TotalTests++
			reportData.TotalDuration += result.Duration

			// Group by package
			pkg := result.Package
			if pkg == "" {
				pkg = "unknown"
			}
			packageGroups[pkg] = append(packageGroups[pkg], name)

			switch result.Status {
			case "PASS":
				reportData.PassedTests++
			case "FAIL":
				reportData.FailedTests++
			case "SKIP":
				reportData.SkippedTests++
			}
		}
	}

	sort.Strings(sortedNames)
	reportData.SortedTestNames = sortedNames
	reportData.PackageGroups = packageGroups

	return reportData, nil
}

// getDurationColor returns a color gradient based on duration percentage
func getDurationColor(duration, maxDuration float64) string {
	// Green to red gradient based on duration percentage
	ratio := duration / maxDuration
	if ratio > 1.0 {
		ratio = 1.0
	}

	// Blend from green (low duration) to yellow (medium) to red (high duration)
	r := int(255 * math.Min(1.0, ratio*2))
	g := int(255 * math.Min(1.0, 2-ratio*2))
	return fmt.Sprintf("#%02x%02x00", r, g)
}

func generateMarkdownReport(data *ReportData) string {
	var sb strings.Builder

	// Generate header
	sb.WriteString("# Test Summary Report\n\n")

	// Add visual summary cards using HTML
	passPercentage := 0.0
	if data.TotalTests > 0 {
		passPercentage = float64(data.PassedTests) / float64(data.TotalTests) * 100
	}
	passColor := htmlPassColor
	if passPercentage < 80 {
		passColor = htmlFailColor
	} else if passPercentage < 100 {
		passColor = htmlSkipColor
	}

	sb.WriteString("<div style=\"display: flex; gap: 20px; margin-bottom: 20px;\">\n")

	// Total Tests Card
	sb.WriteString("<div style=\"flex: 1; padding: 10px; border: 1px solid #ddd; border-radius: 5px; text-align: center;\">\n")
	sb.WriteString(fmt.Sprintf("<div style=\"font-size: 24px; font-weight: bold;\">%d</div>\n", data.TotalTests))
	sb.WriteString("<div style=\"font-size: 12px; color: #666;\">Total Tests</div>\n")
	sb.WriteString("</div>\n")

	// Success Rate Card
	sb.WriteString("<div style=\"flex: 1; padding: 10px; border: 1px solid #ddd; border-radius: 5px; text-align: center;\">\n")
	sb.WriteString(fmt.Sprintf("<div style=\"font-size: 24px; font-weight: bold; color: %s;\">%.1f%%</div>\n",
		passColor, passPercentage))
	sb.WriteString("<div style=\"font-size: 12px; color: #666;\">Success Rate</div>\n")
	sb.WriteString("</div>\n")

	// Duration Card
	sb.WriteString("<div style=\"flex: 1; padding: 10px; border: 1px solid #ddd; border-radius: 5px; text-align: center;\">\n")
	sb.WriteString(fmt.Sprintf("<div style=\"font-size: 24px; font-weight: bold;\">%.2fs</div>\n", data.TotalDuration))
	sb.WriteString("<div style=\"font-size: 12px; color: #666;\">Total Duration</div>\n")
	sb.WriteString("</div>\n")

	sb.WriteString("</div>\n\n")

	// Generate summary
	passPercentageDisplay := "N/A"
	if data.TotalTests > 0 {
		passPercentageDisplay = fmt.Sprintf("%.1f%%", passPercentage)
	}

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Tests:** %d\n", data.TotalTests))
	sb.WriteString(fmt.Sprintf("- **Passed:** %d (%s)\n", data.PassedTests, passPercentageDisplay))
	sb.WriteString(fmt.Sprintf("- **Failed:** %d\n", data.FailedTests))
	sb.WriteString(fmt.Sprintf("- **Skipped:** %d\n", data.SkippedTests))
	sb.WriteString(fmt.Sprintf("- **Total Duration:** %.2fs\n\n", data.TotalDuration))

	// Visual pass/fail indicator
	sb.WriteString("## Test Status\n\n")

	// Create status badges
	if data.FailedTests > 0 {
		sb.WriteString("![Status](https://img.shields.io/badge/Status-FAILED-red)\n\n")
	} else if data.SkippedTests == data.TotalTests {
		sb.WriteString("![Status](https://img.shields.io/badge/Status-SKIPPED-yellow)\n\n")
	} else {
		sb.WriteString("![Status](https://img.shields.io/badge/Status-PASSED-brightgreen)\n\n")
	}

	// Add Coverage Badge if available
	// Note: This is a placeholder - you would need to integrate with actual coverage data
	// coveragePercentage := 65.4 // This would come from your actual coverage data
	// sb.WriteString(fmt.Sprintf("![Coverage](https://img.shields.io/badge/Coverage-%.1f%%25-%s)\n\n",
	//     coveragePercentage, getCoverageColor(coveragePercentage)))

	// Group tests by package
	sb.WriteString("## Test Results by Package\n\n")

	var packageNames []string
	for pkg := range data.PackageGroups {
		packageNames = append(packageNames, pkg)
	}
	sort.Strings(packageNames)

	for _, pkg := range packageNames {
		testNames := data.PackageGroups[pkg]
		sb.WriteString(fmt.Sprintf("<details>\n<summary>Package: <strong>%s</strong> (%d tests)</summary>\n\n",
			pkg, len(testNames)))

		// Create a table of test results for this package
		sb.WriteString("| Test | Status | Duration | Details |\n")
		sb.WriteString("| ---- | ------ | -------- | ------- |\n")

		// Sort package tests by name
		sort.Strings(testNames)

		for _, testName := range testNames {
			result := data.Results[testName]

			// Skip subtests here - we'll show them nested
			if result.IsSubTest {
				continue
			}

			// Determine status emoji and color
			statusEmoji := "⏺️"
			statusColor := htmlNeutralColor
			switch result.Status {
			case "PASS":
				statusEmoji = "✅"
				statusColor = htmlPassColor
			case "FAIL":
				statusEmoji = "❌"
				statusColor = htmlFailColor
			case "SKIP":
				statusEmoji = "⏭️"
				statusColor = htmlSkipColor
			}

			// Format test name to be more readable (remove package prefix if present)
			displayName := result.Name
			if strings.Contains(displayName, "/") && !result.IsSubTest {
				displayName = filepath.Base(displayName)
			}

			// Prepare details column content
			detailsColumn := ""
			if len(result.SubTests) > 0 {
				detailsColumn = fmt.Sprintf("<details><summary>%d subtests</summary>", len(result.SubTests))

				// Add a nested table for subtests
				detailsColumn += "<table><tr><th>Subtest</th><th>Status</th><th>Duration</th></tr>"

				sort.Strings(result.SubTests)
				for _, subTestName := range result.SubTests {
					subTest := data.Results[subTestName]
					subTestDisplayName := subTestName[strings.LastIndex(subTestName, "/")+1:]

					subStatusEmoji := "⏺️"
					subStatusColor := htmlNeutralColor
					switch subTest.Status {
					case "PASS":
						subStatusEmoji = "✅"
						subStatusColor = htmlPassColor
					case "FAIL":
						subStatusEmoji = "❌"
						subStatusColor = htmlFailColor
					case "SKIP":
						subStatusEmoji = "⏭️"
						subStatusColor = htmlSkipColor
					}

					detailsColumn += fmt.Sprintf("<tr><td>%s</td><td><span style=\"color: %s\">%s %s</span></td><td>%.3fs</td></tr>",
						subTestDisplayName, subStatusColor, subStatusEmoji, subTest.Status, subTest.Duration)
				}

				detailsColumn += "</table></details>"
			} else {
				detailsColumn = "-"
			}

			sb.WriteString(fmt.Sprintf("| **%s** | <span style=\"color: %s\">%s %s</span> | %.3fs | %s |\n",
				displayName, statusColor, statusEmoji, result.Status, result.Duration, detailsColumn))
		}

		sb.WriteString("\n</details>\n\n")
	}

	if data.FailedTests > 0 {
		sb.WriteString("## Failed Tests Details\n\n")
		sb.WriteString("<details>\n")
		sb.WriteString("<summary>Click to expand failed test details</summary>\n\n")

		for _, testName := range data.SortedTestNames {
			result := data.Results[testName]

			// Check if this test or any of its subtests failed
			testFailed := result.Status == "FAIL"

			// Check subtests for failures
			for _, subTestName := range result.SubTests {
				if data.Results[subTestName].Status == "FAIL" {
					testFailed = true
					break
				}
			}

			if testFailed {
				displayName := testName
				if strings.Contains(displayName, "/") && !result.IsSubTest {
					displayName = filepath.Base(displayName)
				}

				sb.WriteString(fmt.Sprintf("<div style=\"margin-bottom: 20px; padding: 10px; border-left: 4px solid %s; background-color: #ffeef0\">\n", htmlFailColor))
				sb.WriteString(fmt.Sprintf("<h3>%s</h3>\n\n", displayName))

				// Output for the main test
				if result.Status == "FAIL" && len(result.Output) > 0 {
					sb.WriteString("```go\n")
					for _, line := range result.Output {
						if strings.Contains(line, "FAIL") || strings.Contains(line, "Error") ||
							strings.Contains(line, "panic:") || strings.Contains(line, "--- FAIL") {
							sb.WriteString(fmt.Sprintf("%s\n", line))
						}
					}
					sb.WriteString("```\n\n")
				}

				// Output for failed subtests
				for _, subTestName := range result.SubTests {
					subTest := data.Results[subTestName]
					if subTest.Status == "FAIL" {
						subTestDisplayName := subTestName[strings.LastIndex(subTestName, "/")+1:]
						sb.WriteString(fmt.Sprintf("<h4>%s</h4>\n\n", subTestDisplayName))

						if len(subTest.Output) > 0 {
							sb.WriteString("```go\n")
							for _, line := range subTest.Output {
								if strings.Contains(line, "FAIL") || strings.Contains(line, "Error") ||
									strings.Contains(line, "panic:") || strings.Contains(line, "--- FAIL") {
									sb.WriteString(fmt.Sprintf("%s\n", line))
								}
							}
							sb.WriteString("```\n\n")
						}
					}
				}
				sb.WriteString("</div>\n\n")
			}
		}

		// Close the details tag
		sb.WriteString("</details>\n\n")
	}

	// Add duration metrics
	sb.WriteString("## Test Durations\n\n")
	sb.WriteString("<details>\n")
	sb.WriteString("<summary>Click to expand test durations</summary>\n\n")
	sb.WriteString("| Test | Duration |\n")
	sb.WriteString("| ---- | -------- |\n")

	// Sort tests by duration (descending)
	type testDuration struct {
		name     string
		duration float64
		isRoot   bool
	}

	var durations []testDuration
	for testName, result := range data.Results {
		durations = append(durations, testDuration{
			name:     testName,
			duration: result.Duration,
			isRoot:   !result.IsSubTest,
		})
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i].duration > durations[j].duration
	})

	// Scale factor for bar chart - handle outliers better
	maxDuration := 0.0
	if len(durations) > 0 {
		maxDuration = durations[0].duration
		if len(durations) > 1 && maxDuration > durations[1].duration*3 {
			// If top test is 3x longer than second, use second test as scale to prevent skewed visualization
			maxDuration = durations[1].duration * 1.5
		}
	}

	// Take top 15 longest tests
	count := 0
	for _, d := range durations {
		if count >= 15 {
			break
		}

		// Format test name to be more readable
		displayName := d.name
		if d.isRoot {
			if strings.Contains(displayName, "/") {
				displayName = filepath.Base(displayName)
			}
		} else {
			// For subtests, show parent/child relationship
			displayName = "↳ " + d.name[strings.LastIndex(d.name, "/")+1:]
		}

		// Add bar chart using unicode block characters with color
		barColor := getDurationColor(d.duration, maxDuration)
		scaleFactor := 25.0
		barLength := max(int(d.duration*scaleFactor/maxDuration), 1)
		durationBar := strings.Repeat("█", barLength)

		sb.WriteString(fmt.Sprintf("| %s | %.3fs <span style=\"color: %s\">%s</span> |\n",
			displayName, d.duration, barColor, durationBar))
		count++
	}

	// Close the details tag
	sb.WriteString("\n</details>\n\n")

	// Add test timeline visualization
	sb.WriteString("## Test Timeline\n\n")
	sb.WriteString("<details>\n")
	sb.WriteString("<summary>Click to expand test execution timeline</summary>\n\n")

	// Create a timeline diagram using mermaid
	sb.WriteString("```mermaid\ngantt\n")
	sb.WriteString("    title Test Execution Timeline\n")
	sb.WriteString("    dateFormat X\n")
	sb.WriteString("    axisFormat %S.%L\n\n")

	// Add timeline data
	var startTime float64 = 0
	timelineTests := durations
	if len(timelineTests) > 15 {
		timelineTests = timelineTests[:15] // Top 15 tests by duration
	}

	for _, d := range timelineTests {
		testName := d.name
		if len(testName) > 30 {
			testName = "..." + testName[len(testName)-27:]
		}

		// Escape characters that might break mermaid syntax
		testName = strings.ReplaceAll(testName, ":", " -")
		testName = strings.ReplaceAll(testName, "/", "-")

		sb.WriteString(fmt.Sprintf("    %s: %f, %f\n",
			testName, startTime, startTime+d.duration))
		startTime += d.duration * 0.2 // Offset for visualization
	}

	sb.WriteString("```\n</details>\n\n")

	// Format the timestamp more elegantly
	currentTime := time.Now()
	sb.WriteString("\n---\n\n")
	sb.WriteString(fmt.Sprintf("📆 **Report Date:** %s  \n", currentTime.Format("January 2, 2006")))
	sb.WriteString(fmt.Sprintf("⏰ **Report Time:** %s  \n", currentTime.Format("15:04:05 MST")))
	sb.WriteString(fmt.Sprintf("🖥 **Generated On:** %s\n", currentTime.Format("Monday at 15:04")))

	return sb.String()
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
