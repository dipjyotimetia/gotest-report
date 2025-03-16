package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetDurationColor(t *testing.T) {
	tests := []struct {
		name        string
		duration    float64
		maxDuration float64
		want        string
	}{
		{
			name:        "zero duration",
			duration:    0,
			maxDuration: 10.0,
			want:        "#00ff00",
		},
		{
			name:        "half of max duration",
			duration:    5.0,
			maxDuration: 10.0,
			want:        "#80ff00",
		},
		{
			name:        "equal to max duration",
			duration:    10.0,
			maxDuration: 10.0,
			want:        "#ff0000",
		},
		{
			name:        "greater than max duration",
			duration:    15.0,
			maxDuration: 10.0,
			want:        "#ff0000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDurationColor(tt.duration, tt.maxDuration)
			if got != tt.want {
				t.Errorf("getDurationColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessTestEvents_BasicFlow(t *testing.T) {
	// Create test input
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []TestEvent{
		{
			Time:    testTime,
			Action:  "run",
			Test:    "TestExample",
			Package: "example/pkg",
		},
		{
			Time:    testTime.Add(2 * time.Second),
			Action:  "pass",
			Test:    "TestExample",
			Package: "example/pkg",
			Elapsed: 2.0,
		},
	}

	// Convert events to JSON and create a reader
	var buf bytes.Buffer
	for _, ev := range events {
		jsonData, _ := json.Marshal(ev)
		buf.Write(jsonData)
		buf.WriteByte('\n')
	}
	reader := bytes.NewReader(buf.Bytes())

	// Call the function
	data, err := processTestEvents(reader)
	if err != nil {
		t.Fatalf("processTestEvents() error = %v", err)
	}

	// Verify results
	if data.TotalTests != 1 {
		t.Errorf("Expected 1 test, got %d", data.TotalTests)
	}
	if data.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", data.PassedTests)
	}
	if data.FailedTests != 0 {
		t.Errorf("Expected 0 failed tests, got %d", data.FailedTests)
	}
	if data.TotalDuration != 2.0 {
		t.Errorf("Expected total duration 2.0, got %.2f", data.TotalDuration)
	}
}

func TestProcessTestEvents_WithSubtests(t *testing.T) {
	// Create test input with parent test and subtests
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []TestEvent{
		// Parent test
		{
			Time:    testTime,
			Action:  "run",
			Test:    "TestParent",
			Package: "example/pkg",
		},
		// First subtest
		{
			Time:    testTime.Add(time.Millisecond),
			Action:  "run",
			Test:    "TestParent/SubTest1",
			Package: "example/pkg",
		},
		// Second subtest
		{
			Time:    testTime.Add(2 * time.Millisecond),
			Action:  "run",
			Test:    "TestParent/SubTest2",
			Package: "example/pkg",
		},
		// Complete subtests
		{
			Time:    testTime.Add(100 * time.Millisecond),
			Action:  "pass",
			Test:    "TestParent/SubTest1",
			Package: "example/pkg",
			Elapsed: 0.1,
		},
		{
			Time:    testTime.Add(200 * time.Millisecond),
			Action:  "fail",
			Test:    "TestParent/SubTest2",
			Package: "example/pkg",
			Elapsed: 0.2,
		},
		// Complete parent test
		{
			Time:    testTime.Add(300 * time.Millisecond),
			Action:  "pass",
			Test:    "TestParent",
			Package: "example/pkg",
			Elapsed: 0.3,
		},
	}

	// Convert events to JSON and create a reader
	var buf bytes.Buffer
	for _, ev := range events {
		jsonData, _ := json.Marshal(ev)
		buf.Write(jsonData)
		buf.WriteByte('\n')
	}
	reader := bytes.NewReader(buf.Bytes())

	// Call the function
	data, err := processTestEvents(reader)
	if err != nil {
		t.Fatalf("processTestEvents() error = %v", err)
	}

	// Verify results
	if data.TotalTests != 1 {
		t.Errorf("Expected 1 root test, got %d", data.TotalTests)
	}

	parentTest := data.Results["TestParent"]
	if parentTest == nil {
		t.Fatal("Parent test not found in results")
	}

	if len(parentTest.SubTests) != 2 {
		t.Errorf("Expected 2 subtests, got %d", len(parentTest.SubTests))
	}

	subTest1 := data.Results["TestParent/SubTest1"]
	if subTest1 == nil {
		t.Fatal("SubTest1 not found in results")
	}
	if subTest1.Status != "PASS" {
		t.Errorf("Expected SubTest1 status PASS, got %s", subTest1.Status)
	}

	subTest2 := data.Results["TestParent/SubTest2"]
	if subTest2 == nil {
		t.Fatal("SubTest2 not found in results")
	}
	if subTest2.Status != "FAIL" {
		t.Errorf("Expected SubTest2 status FAIL, got %s", subTest2.Status)
	}
}

func TestProcessTestEvents_WithSkippedTests(t *testing.T) {
	// Create test input with skipped tests
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []TestEvent{
		{
			Time:    testTime,
			Action:  "run",
			Test:    "TestSkipped",
			Package: "example/pkg",
		},
		{
			Time:    testTime.Add(time.Millisecond),
			Action:  "skip",
			Test:    "TestSkipped",
			Package: "example/pkg",
		},
		{
			Time:    testTime.Add(2 * time.Millisecond),
			Action:  "output",
			Test:    "TestSkipped",
			Package: "example/pkg",
			Output:  "--- SKIP: TestSkipped (0.00s)\n",
		},
	}

	// Convert events to JSON and create a reader
	var buf bytes.Buffer
	for _, ev := range events {
		jsonData, _ := json.Marshal(ev)
		buf.Write(jsonData)
		buf.WriteByte('\n')
	}
	reader := bytes.NewReader(buf.Bytes())

	// Call the function
	data, err := processTestEvents(reader)
	if err != nil {
		t.Fatalf("processTestEvents() error = %v", err)
	}

	// Verify results
	if data.SkippedTests != 1 {
		t.Errorf("Expected 1 skipped test, got %d", data.SkippedTests)
	}

	skippedTest := data.Results["TestSkipped"]
	if skippedTest == nil {
		t.Fatal("Skipped test not found in results")
	}
	if skippedTest.Status != "SKIP" {
		t.Errorf("Expected test status SKIP, got %s", skippedTest.Status)
	}
}

func TestProcessTestEvents_WithTestOutput(t *testing.T) {
	// Create test input with test output
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []TestEvent{
		{
			Time:    testTime,
			Action:  "run",
			Test:    "TestWithOutput",
			Package: "example/pkg",
		},
		{
			Time:    testTime.Add(time.Millisecond),
			Action:  "output",
			Test:    "TestWithOutput",
			Package: "example/pkg",
			Output:  "This is line 1\n",
		},
		{
			Time:    testTime.Add(2 * time.Millisecond),
			Action:  "output",
			Test:    "TestWithOutput",
			Package: "example/pkg",
			Output:  "This is line 2\n",
		},
		{
			Time:    testTime.Add(3 * time.Millisecond),
			Action:  "fail",
			Test:    "TestWithOutput",
			Package: "example/pkg",
		},
	}

	// Convert events to JSON and create a reader
	var buf bytes.Buffer
	for _, ev := range events {
		jsonData, _ := json.Marshal(ev)
		buf.Write(jsonData)
		buf.WriteByte('\n')
	}
	reader := bytes.NewReader(buf.Bytes())

	// Call the function
	data, err := processTestEvents(reader)
	if err != nil {
		t.Fatalf("processTestEvents() error = %v", err)
	}

	// Verify results
	test := data.Results["TestWithOutput"]
	if test == nil {
		t.Fatal("Test with output not found in results")
	}

	expectedOutput := []string{"This is line 1", "This is line 2"}
	if !reflect.DeepEqual(test.Output, expectedOutput) {
		t.Errorf("Expected output %v, got %v", expectedOutput, test.Output)
	}
}

func TestProcessTestEvents_InvalidJSON(t *testing.T) {
	// Create reader with invalid JSON
	reader := strings.NewReader("This is not valid JSON\n")

	// Call the function
	_, err := processTestEvents(reader)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGenerateMarkdownReport_BasicReport(t *testing.T) {
	// Create test data
	data := &ReportData{
		TotalTests:      2,
		PassedTests:     1,
		FailedTests:     1,
		SkippedTests:    0,
		TotalDuration:   3.5,
		Results:         make(map[string]*TestResult),
		SortedTestNames: []string{"TestPass", "TestFail"},
		PackageGroups:   make(map[string][]string),
	}

	// Add test results
	data.Results["TestPass"] = &TestResult{
		Name:      "TestPass",
		Package:   "pkg/pass",
		Status:    "PASS",
		Duration:  1.5,
		Output:    []string{"--- PASS: TestPass (1.50s)"},
		IsSubTest: false,
	}

	data.Results["TestFail"] = &TestResult{
		Name:      "TestFail",
		Package:   "pkg/fail",
		Status:    "FAIL",
		Duration:  2.0,
		Output:    []string{"--- FAIL: TestFail (2.00s)", "Error: something went wrong"},
		IsSubTest: false,
	}

	// Set up package groups
	data.PackageGroups["pkg/pass"] = []string{"TestPass"}
	data.PackageGroups["pkg/fail"] = []string{"TestFail"}

	// Generate report
	report := generateMarkdownReport(data)

	// Basic validation
	if !strings.Contains(report, "# Test Summary Report") {
		t.Error("Report missing title")
	}

	if !strings.Contains(report, "Total Tests: 2") {
		t.Error("Report missing total test count")
	}

	if !strings.Contains(report, "Passed: 1") {
		t.Error("Report missing passed test count")
	}

	if !strings.Contains(report, "Failed: 1") {
		t.Error("Report missing failed test count")
	}

	if !strings.Contains(report, "50.0%") {
		t.Error("Report missing correct pass percentage")
	}

	if !strings.Contains(report, "Status-FAILED-red") {
		t.Error("Failed status badge missing")
	}
}

func TestGenerateMarkdownReport_WithSubtests(t *testing.T) {
	// Create test data with subtests
	data := &ReportData{
		TotalTests:      1,
		PassedTests:     1,
		FailedTests:     0,
		SkippedTests:    0,
		TotalDuration:   1.0,
		Results:         make(map[string]*TestResult),
		SortedTestNames: []string{"TestParent"},
		PackageGroups:   make(map[string][]string),
	}

	// Add test results with subtests
	data.Results["TestParent"] = &TestResult{
		Name:      "TestParent",
		Package:   "pkg/parent",
		Status:    "PASS",
		Duration:  1.0,
		SubTests:  []string{"TestParent/SubTest1", "TestParent/SubTest2"},
		IsSubTest: false,
	}

	data.Results["TestParent/SubTest1"] = &TestResult{
		Name:       "TestParent/SubTest1",
		Package:    "pkg/parent",
		Status:     "PASS",
		Duration:   0.5,
		ParentTest: "TestParent",
		IsSubTest:  true,
	}

	data.Results["TestParent/SubTest2"] = &TestResult{
		Name:       "TestParent/SubTest2",
		Package:    "pkg/parent",
		Status:     "PASS",
		Duration:   0.5,
		ParentTest: "TestParent",
		IsSubTest:  true,
	}

	// Set up package groups
	data.PackageGroups["pkg/parent"] = []string{"TestParent"}

	// Generate report
	report := generateMarkdownReport(data)

	// Check for subtest-specific content
	if !strings.Contains(report, "2 subtests") {
		t.Error("Report missing subtest count")
	}

	if !strings.Contains(report, "SubTest1") {
		t.Error("Report missing first subtest")
	}

	if !strings.Contains(report, "SubTest2") {
		t.Error("Report missing second subtest")
	}
}

func TestMain_Integration(t *testing.T) {
	// Skip in regular test runs - this is more of an integration test
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary files for input and output
	tmpDir, err := os.MkdirTemp("", "gotest-report-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input.json")
	outputPath := filepath.Join(tmpDir, "output.md")

	// Create sample test input
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []TestEvent{
		{
			Time:    testTime,
			Action:  "run",
			Test:    "TestExample",
			Package: "example/pkg",
		},
		{
			Time:    testTime.Add(time.Second),
			Action:  "pass",
			Test:    "TestExample",
			Package: "example/pkg",
			Elapsed: 1.0,
		},
	}

	var buf bytes.Buffer
	for _, ev := range events {
		jsonData, _ := json.Marshal(ev)
		buf.Write(jsonData)
		buf.WriteByte('\n')
	}

	err = os.WriteFile(inputPath, buf.Bytes(), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Save original args and os.Args for restoration
	oldArgs := os.Args
	oldStdout := os.Stdout

	// Redirect stdout to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set up args for the test
	os.Args = []string{"gotest-report", "-input", inputPath, "-output", outputPath}

	// Run main (in a separate function to ensure deferred cleanup)
	func() {
		defer func() {
			// Restore original state
			os.Args = oldArgs
			os.Stdout = oldStdout
		}()

		// Call main
		main()

		// Close writer to get output
		w.Close()
		var out bytes.Buffer
		io.Copy(&out, r)

		// Verify output contains success message
		if !strings.Contains(out.String(), "Report generated successfully") {
			t.Errorf("Expected success message, got: %s", out.String())
		}

		// Verify file was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output file was not created")
		}

		// Read output file to verify content
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(content), "# Test Summary Report") {
			t.Error("Output report does not contain expected title")
		}
	}()
}
