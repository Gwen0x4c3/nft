#!/bin/bash

# Optimize imports across the Go codebase
# This script consolidates imports and removes unused imports

set -e

echo "Optimizing Go imports..."

# Find all Go files (excluding vendor and test files temporarily)
find . -name "*.go" -not -path "./vendor/*" -not -path "./tests/*" | while read file; do
    echo "Processing $file..."

    # Use goimports to format and optimize imports
    if command -v goimports &> /dev/null; then
        goimports -w "$file"
    else
        echo "goimports not found, skipping import optimization for $file"
    fi

    # Use go fmt to ensure proper formatting
    go fmt "$file"
done

echo "Running go mod tidy..."
go mod tidy

echo "Checking for unused imports..."
go mod download
go build ./...

echo "Import optimization completed!"

# Optional: Check for common import patterns that can be consolidated
echo "Checking for common import consolidation opportunities..."

# Find files with multiple error handling patterns
echo "Files with multiple error handling imports:"
grep -r "fmt.Errorf" --include="*.go" . | cut -d: -f1 | sort -u | head -5

# Find files with multiple JSON imports
echo "Files with JSON imports:"
grep -r "encoding/json" --include="*.go" . | cut -d: -f1 | sort -u | head -5

echo "Done checking import patterns."