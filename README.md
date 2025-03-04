# Go Test Report Generator

A command-line tool and GitHub Action that generates beautiful Markdown reports from Go test output in JSON format.

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
  - Collapsible sections for failed test details and duration metrics
- Multi-job support:
  - Combine reports from multiple jobs in a single PR
  - Individual job comments with detailed reports
  - Consolidated summary comment with links to all job details
- CI/CD integration:
  - GitHub Actions workflow included
  - Can be used as a GitHub Action or standalone CLI tool
  - Direct links to workflow runs from reports

## Sequence Diagram
```mermaid
sequenceDiagram
    participant D as Developer
    participant GH as GitHub
    participant GA as GitHub Actions
    participant GT as Go Test
    participant TR as Test Reporter
    participant PR as Pull Request
    
    D->>GH: Create/Update PR
    GH->>GA: Trigger Workflow with Multiple Jobs
    
    par Job 1
        GA->>GT: Run Go Tests with JSON output
        GT->>GA: Return test.json output
        GA->>TR: Process test.json
        Note over TR: Parse JSON events
        Note over TR: Generate Job 1 Report
        TR->>GA: Return Job 1 Report
    and Job 2
        GA->>GT: Run Go Tests with JSON output
        GT->>GA: Return test.json output
        GA->>TR: Process test.json
        Note over TR: Parse JSON events
        Note over TR: Generate Job 2 Report
        TR->>GA: Return Job 2 Report
    end
    
    GA->>GH: Upload reports as artifacts
    
    alt Comment PR is enabled
        GA->>PR: Add/update individual job comments
        GA->>PR: Add/update summary comment with links
    end
    
    D->>PR: View summary and job-specific reports
    D->>GH: Access workflow run via embedded links
```

## Usage

### As a GitHub Action

Add the following to your workflow file:

```yml
- name: Run Go tests with JSON output
  run: go test ./... -json > test-output.json || true

- name: Generate Test Report
  uses: dipjyotimetia/gotest-report@v1
  with:
    test-json-file: test-output.json
    output-file: test-report.md
    comment-pr: true
    fail-on-test-failure: false
```

### Action Inputs

| Input | Description | Required | Default |
| ----- | ----------- | -------- | ------- |
| test-json-file | Path to the go test -json output file | No | test-output.json |
| output-file | Path for the generated Markdown report | No | test-report.md |
| comment-pr | Whether to comment the PR with the test report | No | true |
| fail-on-test-failure | Whether to fail the GitHub Action if any tests fail | No | false |
| job-name | Name of the job running the tests (for multi-job reports) | No | '' |
| summary-only | Include only summary in the combined PR comment (for multi-job setups) | No | false |

### Multi-Job Setup Example

Here's how to use the action in a multi-job workflow:

```yml
name: Go Tests

on:
  pull_request:
    branches: [ main ]

permissions:
  pull-requests: write
  contents: read

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22'

    - name: Run unit tests with JSON output
      run: |
        go test ./... -tags=unit -json > test-output.json || true

    - name: Generate and Comment Unit Test Report
      uses: dipjyotimetia/gotest-report@v1
      with:
        test-json-file: test-output.json
        job-name: "Unit Tests"
        summary-only: true
        comment-pr: true

  integration-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22'

    - name: Run integration tests with JSON output
      run: |
        go test ./... -tags=integration -json > test-output.json || true

    - name: Generate and Comment Integration Test Report
      uses: dipjyotimetia/gotest-report@v1
      with:
        test-json-file: test-output.json
        job-name: "Integration Tests"
        summary-only: true
        comment-pr: true
```

This will create:
1. A detailed comment for each job (Unit Tests and Integration Tests)
2. A summary comment combining results from all jobs
3. Links to the workflow run in all comments

### Command Line

You can also use this tool directly as a CLI application:

```sh
# Install
go install github.com/dipjyotimetia/gotest-report@latest

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

## Output Format

The generated Markdown report includes:

1. **Summary Section** - Overall test statistics
2. **Test Status** - Visual badge indicator of overall test status
3. **Test Results** - Table of all tests with status and duration
4. **Failed Tests Details** - Collapsible section with detailed output for failed tests (if any)
5. **Test Durations** - Collapsible section with bar chart visualization of the longest-running tests
6. **Workflow Link** - Direct link to the GitHub Actions workflow run
7. **Timestamp** - When the report was generated

## Example Output for Multi-Job Reports

### Summary Comment

# Go Test Report Summary

## Summary

- **Total Tests:** 12
- **Passed:** 12 (100.0%)
- **Failed:** 0
- **Skipped:** 0
- **Total Duration:** 3.45s

## Test Status

![Status](https://img.shields.io/badge/Status-PASSED-brightgreen)

<details>
<summary>View details for all test jobs</summary>

This is a combined report summary. See individual job comments for detailed reports or check the [workflow run](https://github.com/dipjyotimetia/gotest-report/actions/runs/123456789).
</details>

---

[View Workflow Run](https://github.com/dipjyotimetia/gotest-report/actions/runs/123456789)


### Individual Job Comment

# Unit Tests Results

## Summary

- **Total Tests:** 5
- **Passed:** 5 (100.0%)
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

---

Job: **Unit Tests** | [View Workflow Run](https://github.com/dipjyotimetia/gotest-report/actions/runs/123456789)

Report generated at: 2024-03-20T15:30:00Z

## License

MIT License