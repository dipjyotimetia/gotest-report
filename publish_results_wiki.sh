#!/bin/bash

# Script to publish test results to GitHub Wiki
# Usage: ./publish_results_to_wiki.sh <test-report-file> <wiki-page-name>

set -e

# Default parameters
TEST_REPORT_FILE=${1:-"test-report.md"}
WIKI_PAGE_NAME=${2:-"Test-Results"}
TEMP_DIR=".wiki_temp"

# Check if test report exists
if [ ! -f "$TEST_REPORT_FILE" ]; then
    echo "Error: Test report file $TEST_REPORT_FILE not found!"
    exit 1
fi

# Check if git is installed
if ! command -v git &> /dev/null; then
    echo "Error: git is not installed or not in PATH"
    exit 1
fi

# Get repository information from git
REPO_URL=$(git config --get remote.origin.url)
if [ -z "$REPO_URL" ]; then
    echo "Error: Could not determine git repository URL"
    exit 1
fi

# Convert SSH URL to HTTPS for the wiki
REPO_WIKI_URL=$(echo $REPO_URL | sed -E 's|git@github.com:|https://github.com/|' | sed -E 's|\.git$||')/wiki

echo "Repository Wiki URL: $REPO_WIKI_URL"

# Create temporary directory for wiki clone
mkdir -p $TEMP_DIR
cd $TEMP_DIR

# Clone the wiki repository
echo "Cloning wiki repository..."
git clone "$REPO_WIKI_URL" .

# Check if clone was successful
if [ $? -ne 0 ]; then
    echo "Error: Failed to clone wiki repository."
    echo "Make sure the wiki is enabled for your repository and you have proper permissions."
    cd ..
    rm -rf $TEMP_DIR
    exit 1
fi

# Prepare the wiki page content
echo "Preparing wiki page content..."

# Add header with timestamp
CURRENT_TIME=$(date -u +"%Y-%m-%d %H:%M:%S UTC")
HEADER="# Test Results (Updated: $CURRENT_TIME)\n\n"

# Add the test report content
cat ../$TEST_REPORT_FILE > "$WIKI_PAGE_NAME.md"
sed -i "1s/^/$HEADER/" "$WIKI_PAGE_NAME.md"

# Add a note about automatic updates
echo -e "\n\n---\n\n_This page is automatically updated by CI/CD pipeline._" >> "$WIKI_PAGE_NAME.md"

# Commit changes to wiki
echo "Committing changes to wiki..."
git add "$WIKI_PAGE_NAME.md"
git config user.email "github-actions@github.com"
git config user.name "GitHub Actions"
git commit -m "Update test results on $(date -u +"%Y-%m-%d")"
git push

# Clean up
cd ..
rm -rf $TEMP_DIR

echo "Successfully published test results to wiki page: $WIKI_PAGE_NAME"