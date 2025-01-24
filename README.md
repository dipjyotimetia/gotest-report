# Go Test Report Generator

A command-line tool that generates beautiful Markdown reports from Go test output in JSON format.

## Features

- Generates detailed test reports in Markdown format
- Shows per-package test summaries with:
  - Total tests, passed, failed, and skipped counts
  - Test duration
  - Success rate percentage
  - Failed test details with full output
- Creates an overall summary with:
  - Total package and test counts
  - Success rate visualization
  - Color-coded status badges

## Installation

```sh
go install github.com/dipjyotimetia/gotest-report@latest
```

## Usage

1. Run your Go tests with JSON output:

```sh
go test ./... -json > test.json
```

2. Generate the report:

```sh
gotest-report test.json
```

This will create `test_report.md` in the same directory.

## Output Format

The generated report includes:
- Timestamp of report generation
- Collapsible sections for each package
- Test statistics with visual indicators
- Detailed failure information
- Overall test execution summary

## Example Output

# Go Test Report

Generated at: Mon, 26 Feb 2024 10:00:00 EST

## Test Summary

<details>
<summary><strong>📦 package/name</strong> (95.0% Success)</summary>

| Metric | Count | Status |
|--------|--------|--------|
| Total Tests | 20 | |
| Passed | 19 | ![](https://img.shields.io/badge/passed-19-%2373D216) |
| Failed | 1 | ![](https://img.shields.io/badge/failed-1-red) |
| Skipped | 0 | ![](https://img.shields.io/badge/skipped-0-yellow) |
| Duration | 1.234s | |
</details>


## License

MIT License
