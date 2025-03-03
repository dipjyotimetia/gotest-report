# Go Test Report Generator

A command-line tool that generates beautiful Markdown reports from Go test output in JSON format. Designed to be easily integrated with GitHub Actions and other CI systems.

## Features

- Transforms Go test JSON output into readable Markdown reports
- Provides summary statistics:
  - Total, passed, failed, and skipped test counts
  - Success rate percentage
  - Total test duration
- Visual status indicators:
  - Status badges (PASSED/FAILED/SKIPPED)
  - Emoji indicators for test results (✅ PASS, ❌ FAIL, ⏭️ SKIP)
- Detailed test results:
  - Hierarchical display of tests and subtests
  - Test durations with visual bar charts for the longest tests
  - Detailed failure information for debugging
- CI/CD integration:
  - GitHub Actions workflow included for automatic PR comments

## Installation

```sh
go install github.com/dipjyotimetia/gotest-report@latest
```

## Usage

### Command Line

Run your Go tests with JSON output and pipe to the tool:

```sh
# Option 1: Pipe directly
go test ./... -json | gotest-report

# Option 2: Save JSON and process
go test ./... -json > test-output.json
gotest-report -input test-output.json -output test-report.md
```

### Command Line Options

```
  -input string
        go test -json output file (default is stdin)
  -output string
        Output markdown file (default "test-report.md")
```

### GitHub Actions Integration

Add the following workflow to your repository (`.github/workflows/test.yml`):

```yml
name: Go Test Report

on:
  pull_request:
    branches: [ main ]

permissions:
  pull-requests: write
  contents: read

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go: [ '1.12', '1.23', '1.24' ]
    name: Go ${{ matrix.go }} sample
    steps:
    - uses: actions/checkout@v4

    - name: Setup go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ matrix.go }}
        
    - name: Run tests with JSON output
      run: |
        go test ./... -json > test-output.json  

    - name: Generate test report
      run: |
        go run main.go -input test-output.json -output test-report.md

    - uses: mshick/add-pr-comment@v2
      with:
        message-id: test-report
        message-path: test_report.md
```

## Output Format

The generated Markdown report includes:

1. **Summary Section** - Overall test statistics
2. **Test Status** - Visual badge indicator of overall test status
3. **Test Results** - Table of all tests with status and duration
4. **Failed Tests Details** - Detailed output for failed tests (if any)
5. **Test Durations** - Bar chart visualization of the longest-running tests
6. **Timestamp** - When the report was generated

## Example Output

# Go Test Results

## Summary

- **Total Tests:** 3
- **Passed:** 3 (100.0%)
- **Failed:** 0
- **Skipped:** 0
- **Total Duration:** 1.23s

## Test Status

![Status](https://img.shields.io/badge/Status-PASSED-brightgreen)

## Test Results

| Test | Status | Duration |
| ---- | ------ | -------- |
| **TestOne** | ✅ PASS | 0.500s |
| **TestTwo** | ✅ PASS | 0.400s |
|    ↳ SubTest1 | ✅ PASS | 0.200s |
|    ↳ SubTest2 | ✅ PASS | 0.200s |
| **TestThree** | ✅ PASS | 0.334s |

## Test Durations

| Test | Duration |
| ---- | -------- |
| TestOne | 0.500s █████████████████████ |
| TestTwo | 0.400s ████████████████ |
| TestThree | 0.334s █████████████ |

---

Report generated at: 2024-03-20T15:30:00Z

## License

MIT License