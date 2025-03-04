package main

import (
	"strings"
	"testing"
)

func TestGenerateMarkdownReport(t *testing.T) {
	tests := []struct {
		name       string
		reportData *ReportData
		checks     []func(t *testing.T, markdown string)
	}{
		{
			name: "all tests passed",
			reportData: &ReportData{
				TotalTests:    3,
				PassedTests:   3,
				FailedTests:   0,
				SkippedTests:  0,
				TotalDuration: 1.234,
				SortedTestNames: []string{
					"TestOne",
					"TestTwo",
					"TestThree",
				},
				Results: map[string]*TestResult{
					"TestOne": {
						Name:      "TestOne",
						Package:   "pkg/one",
						Status:    "PASS",
						Duration:  0.5,
						IsSubTest: false,
						Output:    []string{"=== RUN TestOne", "--- PASS: TestOne (0.50s)"},
					},
					"TestTwo": {
						Name:      "TestTwo",
						Package:   "pkg/one",
						Status:    "PASS",
						Duration:  0.4,
						IsSubTest: false,
						SubTests:  []string{"TestTwo/SubTest1", "TestTwo/SubTest2"},
						Output:    []string{"=== RUN TestTwo", "--- PASS: TestTwo (0.40s)"},
					},
					"TestTwo/SubTest1": {
						Name:       "TestTwo/SubTest1",
						Package:    "pkg/one",
						Status:     "PASS",
						Duration:   0.2,
						ParentTest: "TestTwo",
						IsSubTest:  true,
						Output:     []string{"=== RUN TestTwo/SubTest1", "--- PASS: TestTwo/SubTest1 (0.20s)"},
					},
					"TestTwo/SubTest2": {
						Name:       "TestTwo/SubTest2",
						Package:    "pkg/one",
						Status:     "PASS",
						Duration:   0.2,
						ParentTest: "TestTwo",
						IsSubTest:  true,
						Output:     []string{"=== RUN TestTwo/SubTest2", "--- PASS: TestTwo/SubTest2 (0.20s)"},
					},
					"TestThree": {
						Name:      "TestThree",
						Package:   "pkg/two",
						Status:    "PASS",
						Duration:  0.334,
						IsSubTest: false,
						Output:    []string{"=== RUN TestThree", "--- PASS: TestThree (0.33s)"},
					},
				},
			},
			checks: []func(t *testing.T, markdown string){
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "**Total Tests:** 3") {
						t.Errorf("Expected total tests count not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "**Passed:** 3 (100.0%)") {
						t.Errorf("Expected passed tests percentage not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "![Status](https://img.shields.io/badge/Status-PASSED-brightgreen)") {
						t.Errorf("Expected passed status badge not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "✅ PASS") {
						t.Errorf("Expected pass emoji not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if strings.Contains(markdown, "## Failed Tests Details") {
						t.Errorf("Failed tests section should not be present when all tests pass: %s", markdown)
					}
				},
			},
		},
		{
			name: "tests with failures",
			reportData: &ReportData{
				TotalTests:    3,
				PassedTests:   1,
				FailedTests:   2,
				SkippedTests:  0,
				TotalDuration: 1.234,
				SortedTestNames: []string{
					"TestOne",
					"TestTwo",
				},
				Results: map[string]*TestResult{
					"TestOne": {
						Name:      "TestOne",
						Package:   "pkg/one",
						Status:    "PASS",
						Duration:  0.5,
						IsSubTest: false,
						Output:    []string{"=== RUN TestOne", "--- PASS: TestOne (0.50s)"},
					},
					"TestTwo": {
						Name:      "TestTwo",
						Package:   "pkg/one",
						Status:    "FAIL",
						Duration:  0.7,
						IsSubTest: false,
						SubTests:  []string{"TestTwo/SubTest1"},
						Output:    []string{"=== RUN TestTwo", "--- FAIL: TestTwo (0.70s)", "    error_test.go:123: unexpected error"},
					},
					"TestTwo/SubTest1": {
						Name:       "TestTwo/SubTest1",
						Package:    "pkg/one",
						Status:     "FAIL",
						Duration:   0.2,
						ParentTest: "TestTwo",
						IsSubTest:  true,
						Output:     []string{"=== RUN TestTwo/SubTest1", "--- FAIL: TestTwo/SubTest1 (0.20s)", "    subtest_error.go:45: failed assertion"},
					},
				},
			},
			checks: []func(t *testing.T, markdown string){
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "**Failed:** 2") {
						t.Errorf("Expected failed tests count not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "![Status](https://img.shields.io/badge/Status-FAILED-red)") {
						t.Errorf("Expected failed status badge not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "❌ FAIL") {
						t.Errorf("Expected fail emoji not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "## Failed Tests Details") {
						t.Errorf("Failed tests section should be present when tests fail: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "unexpected error") {
						t.Errorf("Failed test error message not found in output: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "failed assertion") {
						t.Errorf("Failed subtest error message not found in output: %s", markdown)
					}
				},
			},
		},
		{
			name: "tests with skips",
			reportData: &ReportData{
				TotalTests:    3,
				PassedTests:   1,
				FailedTests:   0,
				SkippedTests:  2,
				TotalDuration: 0.8,
				SortedTestNames: []string{
					"TestPass",
					"TestSkip1",
					"TestSkip2",
				},
				Results: map[string]*TestResult{
					"TestPass": {
						Name:      "TestPass",
						Package:   "pkg/one",
						Status:    "PASS",
						Duration:  0.5,
						IsSubTest: false,
						Output:    []string{"=== RUN TestPass", "--- PASS: TestPass (0.50s)"},
					},
					"TestSkip1": {
						Name:      "TestSkip1",
						Package:   "pkg/one",
						Status:    "SKIP",
						Duration:  0.1,
						IsSubTest: false,
						Output:    []string{"=== RUN TestSkip1", "--- SKIP: TestSkip1 (0.10s)", "    skip_test.go:45: skipping for now"},
					},
					"TestSkip2": {
						Name:      "TestSkip2",
						Package:   "pkg/two",
						Status:    "SKIP",
						Duration:  0.2,
						IsSubTest: false,
						Output:    []string{"=== RUN TestSkip2", "--- SKIP: TestSkip2 (0.20s)", "    skip_test.go:55: not implemented"},
					},
				},
			},
			checks: []func(t *testing.T, markdown string){
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "**Skipped:** 2") {
						t.Errorf("Expected skipped tests count not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "⏭️ SKIP") {
						t.Errorf("Expected skip emoji not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					// When not all tests are skipped, we use the passed badge since there are passes
					if !strings.Contains(markdown, "![Status](https://img.shields.io/badge/Status-PASSED-brightgreen)") {
						t.Errorf("Expected passed status badge when some tests pass and others are skipped: %s", markdown)
					}
				},
			},
		},
		{
			name: "all tests skipped",
			reportData: &ReportData{
				TotalTests:    2,
				PassedTests:   0,
				FailedTests:   0,
				SkippedTests:  2,
				TotalDuration: 0.3,
				SortedTestNames: []string{
					"TestSkip1",
					"TestSkip2",
				},
				Results: map[string]*TestResult{
					"TestSkip1": {
						Name:      "TestSkip1",
						Package:   "pkg/one",
						Status:    "SKIP",
						Duration:  0.1,
						IsSubTest: false,
						Output:    []string{"=== RUN TestSkip1", "--- SKIP: TestSkip1 (0.10s)"},
					},
					"TestSkip2": {
						Name:      "TestSkip2",
						Package:   "pkg/two",
						Status:    "SKIP",
						Duration:  0.2,
						IsSubTest: false,
						Output:    []string{"=== RUN TestSkip2", "--- SKIP: TestSkip2 (0.20s)"},
					},
				},
			},
			checks: []func(t *testing.T, markdown string){
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "**Skipped:** 2") {
						t.Errorf("Expected skipped tests count not found in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "![Status](https://img.shields.io/badge/Status-SKIPPED-yellow)") {
						t.Errorf("Expected skipped status badge not found when all tests are skipped: %s", markdown)
					}
				},
			},
		},
		{
			name: "duration metrics formatting",
			reportData: &ReportData{
				TotalTests:    3,
				PassedTests:   3,
				FailedTests:   0,
				SkippedTests:  0,
				TotalDuration: 1.234,
				SortedTestNames: []string{
					"TestA",
					"TestB",
					"TestC",
				},
				Results: map[string]*TestResult{
					"TestA": {
						Name:      "TestA",
						Package:   "pkg/x",
						Status:    "PASS",
						Duration:  0.8,
						IsSubTest: false,
					},
					"TestB": {
						Name:      "TestB",
						Package:   "pkg/x",
						Status:    "PASS",
						Duration:  0.3,
						IsSubTest: false,
					},
					"TestC": {
						Name:      "TestC",
						Package:   "pkg/y",
						Status:    "PASS",
						Duration:  0.134,
						IsSubTest: false,
					},
				},
			},
			checks: []func(t *testing.T, markdown string){
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "## Test Durations") {
						t.Errorf("Test durations section missing in: %s", markdown)
					}
				},
				func(t *testing.T, markdown string) {
					// Check that durationBar exists and TestA has more blocks than TestB
					// We can't check for exact blocks since the implementation may change
					lines := strings.Split(markdown, "\n")
					var testALine, testBLine string

					for _, line := range lines {
						if strings.Contains(line, "TestA") && strings.Contains(line, "0.800s") {
							testALine = line
						}
						if strings.Contains(line, "TestB") && strings.Contains(line, "0.300s") {
							testBLine = line
						}
					}

					if testALine == "" || testBLine == "" {
						t.Errorf("Expected to find test duration lines for TestA and TestB")
						return
					}

					if !strings.Contains(testALine, "█") {
						t.Errorf("Expected to find block characters in the duration bar for TestA")
					}

					testABlocks := strings.Count(testALine, "█")
					testBBlocks := strings.Count(testBLine, "█")

					if testABlocks <= testBBlocks {
						t.Errorf("Expected TestA to have more duration blocks than TestB: TestA=%d blocks, TestB=%d blocks",
							testABlocks, testBBlocks)
					}
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			markdown := generateMarkdownReport(tc.reportData)

			for _, check := range tc.checks {
				check(t, markdown)
			}
		})
	}
}

func TestReportFormattingAndStructure(t *testing.T) {
	// Basic structure test
	reportData := &ReportData{
		TotalTests:      1,
		PassedTests:     1,
		FailedTests:     0,
		SkippedTests:    0,
		TotalDuration:   0.5,
		SortedTestNames: []string{"TestSimple"},
		Results: map[string]*TestResult{
			"TestSimple": {
				Name:      "TestSimple",
				Package:   "pkg",
				Status:    "PASS",
				Duration:  0.5,
				IsSubTest: false,
			},
		},
	}

	markdown := generateMarkdownReport(reportData)

	expectedSections := []string{
		"# Test Summery Report",
		"## Summary",
		"## Test Status",
		"## Test Results",
		"## Test Durations",
	}

	for _, section := range expectedSections {
		if !strings.Contains(markdown, section) {
			t.Errorf("Expected section not found: %s", section)
		}
	}

	// Check that tables have headers
	expectedTableHeaders := []string{
		"| Test | Status | Duration |",
		"| Test | Duration |",
	}

	for _, header := range expectedTableHeaders {
		if !strings.Contains(markdown, header) {
			t.Errorf("Expected table header not found: %s", header)
		}
	}

	// Check footer timestamp format - just check for "Report generated at:" text
	// instead of specific timestamp format which may change
	if !strings.Contains(markdown, "Report generated at:") {
		t.Errorf("Expected 'Report generated at:' in footer not found")
	}
}

func TestProcessTestEvents(t *testing.T) {
	jsonInput := `
{"Time":"2023-04-01T10:00:00Z","Action":"run","Test":"TestExample"}
{"Time":"2023-04-01T10:00:01Z","Action":"output","Test":"TestExample","Output":"running test\n"}
{"Time":"2023-04-01T10:00:02Z","Action":"pass","Test":"TestExample","Elapsed":1.5}
`
	reader := strings.NewReader(jsonInput)

	reportData, err := processTestEvents(reader)
	if err != nil {
		t.Fatalf("processTestEvents returned error: %v", err)
	}

	if reportData.TotalTests != 1 {
		t.Errorf("Expected 1 total test, got %d", reportData.TotalTests)
	}

	if reportData.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", reportData.PassedTests)
	}

	testResult, exists := reportData.Results["TestExample"]
	if !exists {
		t.Fatal("TestExample not found in results")
	}

	if testResult.Status != "PASS" {
		t.Errorf("Expected status PASS, got %s", testResult.Status)
	}

	if testResult.Duration != 1.5 {
		t.Errorf("Expected duration 1.5s, got %f", testResult.Duration)
	}
}

func TestEdgeCasesInReportGeneration(t *testing.T) {
	tests := []struct {
		name       string
		reportData *ReportData
		checks     []func(t *testing.T, markdown string)
	}{
		{
			name: "empty test results",
			reportData: &ReportData{
				TotalTests:      0,
				PassedTests:     0,
				FailedTests:     0,
				SkippedTests:    0,
				TotalDuration:   0,
				SortedTestNames: []string{},
				Results:         map[string]*TestResult{},
			},
			checks: []func(t *testing.T, markdown string){
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "**Total Tests:** 0") {
						t.Errorf("Expected zero total tests not found in: %s", markdown)
					}
				},
			},
		},
		{
			name: "extremely long test names",
			reportData: &ReportData{
				TotalTests:    1,
				PassedTests:   1,
				FailedTests:   0,
				SkippedTests:  0,
				TotalDuration: 0.5,
				SortedTestNames: []string{
					"TestWithExtremelyLongNameThatMightAffectTableFormattingInMarkdownOutput",
				},
				Results: map[string]*TestResult{
					"TestWithExtremelyLongNameThatMightAffectTableFormattingInMarkdownOutput": {
						Name:      "TestWithExtremelyLongNameThatMightAffectTableFormattingInMarkdownOutput",
						Package:   "pkg/long",
						Status:    "PASS",
						Duration:  0.5,
						IsSubTest: false,
					},
				},
			},
			checks: []func(t *testing.T, markdown string){
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "TestWithExtremelyLongNameThatMightAffectTableFormattingInMarkdownOutput") {
						t.Errorf("Expected long test name not found in: %s", markdown)
					}
				},
			},
		},
		{
			name: "extremely short and long durations",
			reportData: &ReportData{
				TotalTests:    2,
				PassedTests:   2,
				FailedTests:   0,
				SkippedTests:  0,
				TotalDuration: 10.5,
				SortedTestNames: []string{
					"TestVeryFast",
					"TestVerySlow",
				},
				Results: map[string]*TestResult{
					"TestVeryFast": {
						Name:      "TestVeryFast",
						Package:   "pkg/perf",
						Status:    "PASS",
						Duration:  0.001, // 1ms
						IsSubTest: false,
					},
					"TestVerySlow": {
						Name:      "TestVerySlow",
						Package:   "pkg/perf",
						Status:    "PASS",
						Duration:  10.5, // 10.5s
						IsSubTest: false,
					},
				},
			},
			checks: []func(t *testing.T, markdown string){
				func(t *testing.T, markdown string) {
					if !strings.Contains(markdown, "0.001s") {
						t.Errorf("Expected short duration not found in: %s", markdown)
					}
					if !strings.Contains(markdown, "10.500s") {
						t.Errorf("Expected long duration not found in: %s", markdown)
					}
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			markdown := generateMarkdownReport(tc.reportData)
			for _, check := range tc.checks {
				check(t, markdown)
			}
		})
	}
}

func TestNestedSubtestsHandling(t *testing.T) {
	reportData := &ReportData{
		TotalTests:    1,
		PassedTests:   1,
		FailedTests:   0,
		SkippedTests:  0,
		TotalDuration: 0.8,
		SortedTestNames: []string{
			"TestParent",
		},
		Results: map[string]*TestResult{
			"TestParent": {
				Name:      "TestParent",
				Package:   "pkg/nested",
				Status:    "PASS",
				Duration:  0.8,
				IsSubTest: false,
				SubTests:  []string{"TestParent/Child", "TestParent/Child/GrandChild"},
			},
			"TestParent/Child": {
				Name:       "TestParent/Child",
				Package:    "pkg/nested",
				Status:     "PASS",
				Duration:   0.6,
				ParentTest: "TestParent",
				IsSubTest:  true,
				SubTests:   []string{"TestParent/Child/GrandChild"},
			},
			"TestParent/Child/GrandChild": {
				Name:       "TestParent/Child/GrandChild",
				Package:    "pkg/nested",
				Status:     "PASS",
				Duration:   0.4,
				ParentTest: "TestParent/Child",
				IsSubTest:  true,
			},
		},
	}

	markdown := generateMarkdownReport(reportData)

	if !strings.Contains(markdown, "TestParent/Child/GrandChild") {
		t.Errorf("Expected deeply nested test name not found in: %s", markdown)
	}

	if !strings.Contains(markdown, "0.400s") {
		t.Errorf("Expected grandchild duration not found in: %s", markdown)
	}
}
