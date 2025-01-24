package main

import (
	"strings"
	"testing"
)

func TestGenerateMarkdownReport(t *testing.T) {
	tests := []struct {
		name     string
		input    []TestResult
		contains []string
	}{
		{
			name:     "Empty input",
			input:    []TestResult{},
			contains: []string{"# Go Test Report", "## Test Summary", "## 📊 Overall Summary"},
		},
		{
			name: "Single package with passed test",
			input: []TestResult{
				{
					Time:    "2023-01-01T10:00:00Z",
					Action:  "run",
					Package: "package/foo",
					Test:    "TestFoo",
				},
				{
					Time:    "2023-01-01T10:00:01Z",
					Action:  "pass",
					Package: "package/foo",
					Test:    "TestFoo",
					Elapsed: 1.5,
					Output:  "test output",
				},
			},
			contains: []string{
				"📦 package/foo",
				"100.0% Success",
				"#### ✅ Passed Tests",
				"TestFoo",
			},
		},
		{
			name: "Single package with failed test",
			input: []TestResult{
				{
					Time:    "2023-01-01T10:00:00Z",
					Action:  "run",
					Package: "package/bar",
					Test:    "TestBar",
				},
				{
					Time:    "2023-01-01T10:00:01Z",
					Action:  "fail",
					Package: "package/bar",
					Test:    "TestBar",
					Elapsed: 0.5,
					Output:  "failure output",
				},
			},
			contains: []string{
				"📦 package/bar",
				"0.0% Success",
				"#### ❌ Failed Tests",
				"TestBar",
				"failure output",
			},
		},
		{
			name: "Multiple packages with mixed results",
			input: []TestResult{
				{
					Time:    "2023-01-01T10:00:00Z",
					Action:  "run",
					Package: "package/one",
					Test:    "TestOne",
				},
				{
					Time:    "2023-01-01T10:00:01Z",
					Action:  "pass",
					Package: "package/one",
					Test:    "TestOne",
				},
				{
					Time:    "2023-01-01T10:00:02Z",
					Action:  "run",
					Package: "package/two",
					Test:    "TestTwo",
				},
				{
					Time:    "2023-01-01T10:00:03Z",
					Action:  "skip",
					Package: "package/two",
					Test:    "TestTwo",
					Output:  "skip reason",
				},
			},
			contains: []string{
				"📦 package/one",
				"📦 package/two",
				"#### ✅ Passed Tests",
				"#### ⏭️ Skipped Tests",
				"Total Packages | 2",
			},
		},
		{
			name: "Package with all test types",
			input: []TestResult{
				{
					Time:    "2023-01-01T10:00:00Z",
					Action:  "run",
					Package: "package/mixed",
					Test:    "TestPass",
				},
				{
					Time:    "2023-01-01T10:00:01Z",
					Action:  "pass",
					Package: "package/mixed",
					Test:    "TestPass",
				},
				{
					Time:    "2023-01-01T10:00:02Z",
					Action:  "run",
					Package: "package/mixed",
					Test:    "TestFail",
				},
				{
					Time:    "2023-01-01T10:00:03Z",
					Action:  "fail",
					Package: "package/mixed",
					Test:    "TestFail",
				},
				{
					Time:    "2023-01-01T10:00:04Z",
					Action:  "run",
					Package: "package/mixed",
					Test:    "TestSkip",
				},
				{
					Time:    "2023-01-01T10:00:05Z",
					Action:  "skip",
					Package: "package/mixed",
					Test:    "TestSkip",
				},
			},
			contains: []string{
				"📦 package/mixed",
				"33.3% Success",
				"#### ✅ Passed Tests",
				"TestPass",
				"#### ❌ Failed Tests",
				"TestFail",
				"#### ⏭️ Skipped Tests",
				"TestSkip",
			},
		},
		{
			name: "Package with long test names",
			input: []TestResult{
				{
					Time:    "2023-01-01T10:00:00Z",
					Action:  "run",
					Package: "package/long",
					Test:    "TestVeryLongTestNameWithMultipleWordsAndNumbers123",
				},
				{
					Time:    "2023-01-01T10:00:01Z",
					Action:  "pass",
					Package: "package/long",
					Test:    "TestVeryLongTestNameWithMultipleWordsAndNumbers123",
					Output:  "test output with long\nmultiline\noutput\n",
				},
			},
			contains: []string{
				"TestVeryLongTestNameWithMultipleWordsAndNumbers123",
				"test output with long\nmultiline\noutput",
			},
		},
		{
			name: "Package with special characters",
			input: []TestResult{
				{
					Time:    "2023-01-01T10:00:00Z",
					Action:  "run",
					Package: "package/special-chars!@#$%^&*()",
					Test:    "Test_Special_Chars!@#$%^&*()",
				},
				{
					Time:    "2023-01-01T10:00:01Z",
					Action:  "pass",
					Package: "package/special-chars!@#$%^&*()",
					Test:    "Test_Special_Chars!@#$%^&*()",
				},
			},
			contains: []string{
				"package/special-chars!@#$%^&*()",
				"Test_Special_Chars!@#$%^&*()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateMarkdownReport(tt.input)

			for _, expected := range tt.contains {
				if !strings.Contains(got, expected) {
					t.Errorf("generateMarkdownReport() output does not contain expected string: %q", expected)
				}
			}
		})
	}
}

func TestGenerateMarkdownReportStructure(t *testing.T) {
	input := []TestResult{
		{
			Time:    "2023-01-01T10:00:00Z",
			Action:  "run",
			Package: "package/test",
			Test:    "TestExample",
		},
		{
			Time:    "2023-01-01T10:00:01Z",
			Action:  "pass",
			Package: "package/test",
			Test:    "TestExample",
			Elapsed: 1.0,
			Output:  "test output",
		},
	}

	got := generateMarkdownReport(input)

	// Test for required sections
	requiredHeaders := []string{
		"# Go Test Report",
		"## Test Summary",
		"## 📊 Overall Summary",
	}

	for _, header := range requiredHeaders {
		if !strings.Contains(got, header) {
			t.Errorf("Missing required header: %s", header)
		}
	}

	// Test for table headers
	requiredTables := []string{
		"| Metric | Count | Status |",
		"| Metric | Count |",
	}

	for _, table := range requiredTables {
		if !strings.Contains(got, table) {
			t.Errorf("Missing required table header: %s", table)
		}
	}

	// Test for details tags
	if !strings.Contains(got, "<details>") || !strings.Contains(got, "</details>") {
		t.Error("Missing details tags in the report")
	}
}

func TestCalculatePercentage(t *testing.T) {
	tests := []struct {
		name     string
		part     int
		total    int
		expected float64
	}{
		{"Zero total", 5, 0, 0},
		{"Zero part", 0, 10, 0},
		{"Full percentage", 10, 10, 100},
		{"Half percentage", 5, 10, 50},
		{"Partial percentage", 3, 10, 30},
		{"Large numbers", 1000, 2000, 50},
		// {"Decimal result", 1, 3, 33.333333},
		{"Negative part", -5, 10, -50},
		{"Both negative", -5, -10, 50},
		{"Negative total", 5, -10, -50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculatePercentage(tt.part, tt.total)
			if got != tt.expected {
				t.Errorf("calculatePercentage(%d, %d) = %f; want %f",
					tt.part, tt.total, got, tt.expected)
			}
		})
	}
}

func TestGetColorForPercentage(t *testing.T) {
	tests := []struct {
		name       string
		percentage float64
		want       string
	}{
		{"Perfect score", 100.0, "brightgreen"},
		{"Very high score", 95.0, "brightgreen"},
		{"High score", 85.0, "green"},
		{"Medium score", 60.0, "yellow"},
		{"Low score", 30.0, "red"},
		{"Zero score", 0.0, "red"},
		{"Negative score", -10.0, "red"},
		{"Boundary - just brightgreen", 90.0, "brightgreen"},
		{"Boundary - just green", 75.0, "green"},
		{"Boundary - just yellow", 50.0, "yellow"},
		{"Very high decimal", 99.99, "brightgreen"},
		{"High decimal", 89.99, "green"},
		{"Medium decimal", 74.99, "yellow"},
		{"Low decimal", 49.99, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getColorForPercentage(tt.percentage)
			if got != tt.want {
				t.Errorf("getColorForPercentage(%f) = %s; want %s",
					tt.percentage, got, tt.want)
			}
		})
	}
}
