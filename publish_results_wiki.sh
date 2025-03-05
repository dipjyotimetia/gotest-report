#!/bin/bash

# Script to publish test results to GitHub Wiki
# Usage: ./publish_results_wiki.sh <test-report-file> <wiki-page-name>

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

# Check for GitHub token
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Warning: GITHUB_TOKEN environment variable not set."
    echo "You may encounter authentication issues when pushing to the wiki."
    echo "Set it with: export GITHUB_TOKEN=your_personal_access_token"
fi

# Get repository information from git
REPO_URL=$(git config --get remote.origin.url)
if [ -z "$REPO_URL" ]; then
    echo "Error: Could not determine git repository URL"
    exit 1
fi

# Parse repository owner and name
if [[ "$REPO_URL" == *"github.com"* ]]; then
    # Handle SSH URL
    if [[ "$REPO_URL" == git@* ]]; then
        REPO_PATH=$(echo "$REPO_URL" | sed -E 's|git@github.com:||' | sed -E 's|\.git$||')
    # Handle HTTPS URL
    else
        REPO_PATH=$(echo "$REPO_URL" | sed -E 's|https://github.com/||' | sed -E 's|\.git$||')
    fi
    
    REPO_OWNER=$(echo "$REPO_PATH" | cut -d '/' -f 1)
    REPO_NAME=$(echo "$REPO_PATH" | cut -d '/' -f 2)
    
    echo "Repository: $REPO_OWNER/$REPO_NAME"
else
    echo "Error: Not a GitHub repository URL: $REPO_URL"
    exit 1
fi

# Construct the wiki URL with token for authentication
if [ -n "$GITHUB_TOKEN" ]; then
    REPO_WIKI_URL="https://${GITHUB_TOKEN}@github.com/${REPO_OWNER}/${REPO_NAME}.wiki.git"
else
    REPO_WIKI_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}.wiki.git"
fi

echo "Repository Wiki URL: https://github.com/${REPO_OWNER}/${REPO_NAME}/wiki"

# Remove temp directory if it exists
rm -rf $TEMP_DIR

# Create temporary directory for wiki clone
mkdir -p $TEMP_DIR
cd $TEMP_DIR

# Clone the wiki repository
echo "Cloning wiki repository..."
git clone "$REPO_WIKI_URL" . 2>&1 || { 
    echo "Error: Failed to clone wiki repository."
    echo "Make sure the wiki is enabled and you have proper permissions."
    echo "If using GitHub Actions, ensure GITHUB_TOKEN has wiki permissions."
    cd ..
    rm -rf $TEMP_DIR
    exit 1
}

# Prepare the wiki page content
echo "Preparing wiki page content..."

# Add header with timestamp
CURRENT_TIME=$(date -u +"%Y-%m-%d %H:%M:%S UTC")
HEADER="# Test Results (Updated: $CURRENT_TIME)\n\n"

# Add the test report content
cat ../$TEST_REPORT_FILE > "$WIKI_PAGE_NAME.md"
sed -i.bak "1s/^/$HEADER/" "$WIKI_PAGE_NAME.md" && rm -f "$WIKI_PAGE_NAME.md.bak"

# Add a note about automatic updates
echo -e "\n\n---\n\n_This page is automatically updated by CI/CD pipeline._" >> "$WIKI_PAGE_NAME.md"

# Commit changes to wiki
echo "Committing changes to wiki..."
git add "$WIKI_PAGE_NAME.md"
git config user.email "github-actions@github.com"
git config user.name "GitHub Actions"
git commit -m "Update test results on $(date -u +"%Y-%m-%d")"

# Push changes (with credential handling)
echo "Pushing changes to wiki..."
if [ -n "$GITHUB_TOKEN" ]; then
    # Token is already in the URL
    git push
else
    # No token provided, try normal push
    git push || {
        echo "Error: Failed to push changes to wiki."
        echo "You may need to provide a GITHUB_TOKEN environment variable."
        cd ..
        rm -rf $TEMP_DIR
        exit 1
    }
fi

# Clean up
cd ..
rm -rf $TEMP_DIR

echo "Successfully published test results to wiki page: $WIKI_PAGE_NAME"