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

	var sortedNames []string
	for name, result := range results {
		// Only count root tests in summary (not subtests)
		if !result.IsSubTest {
			sortedNames = append(sortedNames, name)
			reportData.TotalTests++
			reportData.TotalDuration += result.Duration

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

	return reportData, nil
}

func generateMarkdownReport(data *ReportData) string {
	var sb strings.Builder

	// Generate header with better styling
	sb.WriteString("# Test Summary Report\n\n")

	// Calculate pass percentage more efficiently
	passPercentage := 0.0
	passPercentageDisplay := "N/A"
	if data.TotalTests > 0 {
		passPercentage = float64(data.PassedTests) / float64(data.TotalTests) * 100
		passPercentageDisplay = fmt.Sprintf("%.1f%%", passPercentage)
	}

	// Improved summary section with better formatting
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Tests:** %d\n", data.TotalTests))
	sb.WriteString(fmt.Sprintf("- **Passed:** %d (%s)\n", data.PassedTests, passPercentageDisplay))
	sb.WriteString(fmt.Sprintf("- **Failed:** %d\n", data.FailedTests))
	sb.WriteString(fmt.Sprintf("- **Skipped:** %d\n", data.SkippedTests))
	sb.WriteString(fmt.Sprintf("- **Total Duration:** %.2fs\n\n", data.TotalDuration))

	// Enhanced visual status indicator with more distinctive badges
	sb.WriteString("## Test Status\n\n")
	if data.FailedTests > 0 {
		sb.WriteString("![Status](https://img.shields.io/badge/Status-FAILED-red?style=for-the-badge)\n\n")
	} else if data.SkippedTests == data.TotalTests {
		sb.WriteString("![Status](https://img.shields.io/badge/Status-SKIPPED-yellow?style=for-the-badge)\n\n")
	} else {
		sb.WriteString("![Status](https://img.shields.io/badge/Status-PASSED-brightgreen?style=for-the-badge)\n\n")
	}

	// More readable test results table with better column widths
	sb.WriteString("## Test Results\n\n")
	sb.WriteString("| Test | Status | Duration |\n")
	sb.WriteString("| :--- | :----: | --------: |\n")

	// Create a map to track whether we need to add the table header after subtests
	needsTableHeader := false

	for _, testName := range data.SortedTestNames {
		result := data.Results[testName]

		if result.IsSubTest {
			continue
		}

		// More distinctive status emojis
		var statusEmoji string
		switch result.Status {
		case "PASS":
			statusEmoji = "✅"
		case "FAIL":
			statusEmoji = "❌"
		case "SKIP":
			statusEmoji = "⏭️"
		default:
			statusEmoji = "⏺️"
		}

		// Better display name formatting
		displayName := testName
		if strings.Contains(displayName, "/") && !result.IsSubTest {
			displayName = filepath.Base(displayName)
		}

		sb.WriteString(fmt.Sprintf("| **%s** | %s %s | %.3fs |\n",
			displayName, statusEmoji, result.Status, result.Duration))

		// Improved subtest rendering
		if len(result.SubTests) > 0 {
			sb.WriteString("\n")
			sb.WriteString("<details>\n")
			sb.WriteString("<summary>Show Subtests</summary>\n\n")
			sb.WriteString("| SubTest | Status | Duration |\n")
			sb.WriteString("| :------ | :----: | --------: |\n")

			sort.Strings(result.SubTests)
			for _, subTestName := range result.SubTests {
				subTest := data.Results[subTestName]
				// Extract just the subtest name without the full path
				subTestDisplayName := subTestName[strings.LastIndex(subTestName, "/")+1:]

				var statusEmoji string
				switch subTest.Status {
				case "PASS":
					statusEmoji = "✅"
				case "FAIL":
					statusEmoji = "❌"
				case "SKIP":
					statusEmoji = "⏭️"
				default:
					statusEmoji = "⏺️"
				}

				sb.WriteString(fmt.Sprintf("| &nbsp;&nbsp;↳ %s | %s %s | %.3fs |\n",
					subTestDisplayName, statusEmoji, subTest.Status, subTest.Duration))
			}
			sb.WriteString("</details>\n\n")
			needsTableHeader = true
		}

		// Add table header after subtests for better readability
		if needsTableHeader {
			sb.WriteString("| Test | Status | Duration |\n")
			sb.WriteString("| :--- | :----: | --------: |\n")
			needsTableHeader = false
		}
	}
	sb.WriteString("\n")

	// Enhanced failed tests section
	if data.FailedTests > 0 {
		sb.WriteString("## Failed Tests Details\n\n")
		sb.WriteString("<details>\n")
		sb.WriteString("<summary>Click to expand failed test details</summary>\n\n")

		failedTestCount := 0
		for _, testName := range data.SortedTestNames {
			result := data.Results[testName]

			// Check if this test or any of its subtests failed
			testFailed := result.Status == "FAIL"
			failedSubtests := []string{}

			// Check subtests for failures
			for _, subTestName := range result.SubTests {
				if data.Results[subTestName].Status == "FAIL" {
					testFailed = true
					failedSubtests = append(failedSubtests, subTestName)
				}
			}

			if testFailed {
				failedTestCount++
				displayName := testName
				if strings.Contains(displayName, "/") && !result.IsSubTest {
					displayName = filepath.Base(displayName)
				}

				// Add horizontal rule between failed tests for better separation
				if failedTestCount > 1 {
					sb.WriteString("---\n\n")
				}

				sb.WriteString(fmt.Sprintf("### ❌ %s\n\n", displayName))

				// More focused error output for the main test
				if result.Status == "FAIL" {
					const maxOutputLines = 20 // Limit output lines to prevent excessive length
					var errorLines []string

					for _, line := range result.Output {
						if strings.Contains(line, "FAIL") || strings.Contains(line, "Error") ||
							strings.Contains(line, "panic:") || strings.Contains(line, "--- FAIL") {
							errorLines = append(errorLines, line)
						}
					}

					if len(errorLines) > 0 {
						sb.WriteString("```go\n") // Add language for syntax highlighting
						if len(errorLines) > maxOutputLines {
							sb.WriteString("...(truncated)...\n")
							errorLines = errorLines[len(errorLines)-maxOutputLines:]
						}
						for _, line := range errorLines {
							sb.WriteString(fmt.Sprintf("%s\n", line))
						}
						sb.WriteString("```\n\n")
					}
				}

				// Process failed subtests with better formatting
				if len(failedSubtests) > 0 {
					for _, subTestName := range failedSubtests {
						subTest := data.Results[subTestName]
						subTestDisplayName := subTestName[strings.LastIndex(subTestName, "/")+1:]
						sb.WriteString(fmt.Sprintf("#### ↳ %s\n\n", subTestDisplayName))

						const maxOutputLines = 15 // Limit output lines for subtests
						var errorLines []string

						for _, line := range subTest.Output {
							if strings.Contains(line, "FAIL") || strings.Contains(line, "Error") ||
								strings.Contains(line, "panic:") || strings.Contains(line, "--- FAIL") {
								errorLines = append(errorLines, line)
							}
						}

						if len(errorLines) > 0 {
							sb.WriteString("```go\n") // Add language for syntax highlighting
							if len(errorLines) > maxOutputLines {
								sb.WriteString("...(truncated)...\n")
								errorLines = errorLines[len(errorLines)-maxOutputLines:]
							}
							for _, line := range errorLines {
								sb.WriteString(fmt.Sprintf("%s\n", line))
							}
							sb.WriteString("```\n\n")
						}
					}
				}
			}
		}

		// Close the details tag
		sb.WriteString("</details>\n\n")
	}

	// Improved duration metrics with better visualization
	sb.WriteString("## Test Durations\n\n")
	sb.WriteString("<details>\n")
	sb.WriteString("<summary>Click to expand test durations</summary>\n\n")
	sb.WriteString("| Test | Duration | Relative Time |\n")
	sb.WriteString("| :--- | --------: | :--- |\n")

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

	// Improved bar chart scaling to handle outliers better
	maxDuration := 0.0
	if len(durations) > 0 {
		maxDuration = durations[0].duration
		// Use 95th percentile instead of arbitrary scaling
		if len(durations) > 20 {
			percentile95Index := int(float64(len(durations)) * 0.05)
			maxDuration = durations[percentile95Index].duration
		} else if len(durations) > 1 && maxDuration > durations[1].duration*3 {
			// If top test is 3x longer than second, use second test as scale
			maxDuration = durations[1].duration * 1.5
		}
	}

	// Show top 20 longest tests with improved visualization
	count := 0
	maxBarLength := 30 // Maximum bar length
	for _, d := range durations {
		if count >= 20 {
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

		// Add improved bar chart using unicode block characters
		var barLength int
		if maxDuration > 0 {
			barLength = int(math.Ceil(float64(maxBarLength) * math.Min(d.duration/maxDuration, 1.0)))
			barLength = max(barLength, 1) // At least 1 character
		} else {
			barLength = 1
		}

		// Use gradient colors for the bar
		durationBar := ""
		for range barLength {
			durationBar += "█"
		}

		sb.WriteString(fmt.Sprintf("| %s | %.3fs | %s |\n", displayName, d.duration, durationBar))
		count++
	}

	// Add note if there are more tests not shown
	if len(durations) > 20 {
		sb.WriteString(fmt.Sprintf("\n*%d more tests not shown*\n", len(durations)-20))
	}

	// Close the details tag
	sb.WriteString("\n</details>\n\n")

	// Add visual duration chart
	if len(durations) > 0 {
		sb.WriteString("## Duration Distribution\n\n")
		sb.WriteString("<details>\n")
		sb.WriteString("<summary>Click to expand duration distribution</summary>\n\n")

		// Create a simple histogram
		const numBuckets = 10
		buckets := make([]int, numBuckets)
		maxValue := durations[0].duration

		for _, d := range durations {
			if d.duration > 0 {
				bucketIndex := int(math.Min(float64(numBuckets-1), math.Floor(float64(numBuckets)*d.duration/maxValue)))
				buckets[bucketIndex]++
			} else {
				buckets[0]++
			}
		}

		// Find the maximum bucket count for scaling
		maxBucketCount := 0
		for _, count := range buckets {
			if count > maxBucketCount {
				maxBucketCount = count
			}
		}

		// Generate the histogram
		sb.WriteString("```\n")
		sb.WriteString("Duration Distribution (number of tests):\n\n")

		for i := 0; i < numBuckets; i++ {
			lowerBound := maxValue * float64(i) / float64(numBuckets)
			upperBound := maxValue * float64(i+1) / float64(numBuckets)

			// Format the duration range
			rangeStr := fmt.Sprintf("%6.3fs - %6.3fs", lowerBound, upperBound)

			// Create the bar
			barLength := 0
			if maxBucketCount > 0 {
				barLength = (buckets[i] * 40) / maxBucketCount
			}

			bar := ""
			for j := 0; j < barLength; j++ {
				bar += "█"
			}

			sb.WriteString(fmt.Sprintf("%s | %s %d\n", rangeStr, bar, buckets[i]))
		}

		sb.WriteString("```\n")
		sb.WriteString("\n</details>\n\n")
	}

	// Add useful footer with improved timestamp formatting
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("*Report generated at: %s*\n",
		time.Now().UTC().Format("2006-01-02 15:04:05 UTC")))

	return sb.String()
}
