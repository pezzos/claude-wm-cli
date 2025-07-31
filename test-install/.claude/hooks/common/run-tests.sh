#!/bin/bash
# Run project tests
echo "🧪 Running tests..."
if [[ -f "package.json" ]] && command -v npm >/dev/null 2>&1; then
    npm test
elif [[ -f "go.mod" ]] && command -v go >/dev/null 2>&1; then
    go test ./...
elif [[ -f "requirements.txt" ]] && command -v python3 >/dev/null 2>&1; then
    python3 -m pytest
else
    echo "ℹ️  No recognized test framework found"
fi
