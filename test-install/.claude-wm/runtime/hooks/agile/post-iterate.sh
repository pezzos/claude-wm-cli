#!/bin/bash
# post-iterate.sh - Post-hook for /project:agile:iterate command

set -e

echo "=== Post-iterate Hook ==="

# 1. Run tests
echo "Running tests..."
bash "$HOME/.claude/hooks/common/run-tests.sh" || {
    echo "Tests failed! Please fix before continuing."
    exit 1
}

# 2. Auto-format code
echo "Formatting code..."
bash "$HOME/.claude/hooks/common/auto-format.sh"

# 3. Update CHANGELOG if it exists
if [ -f "CHANGELOG.md" ]; then
    echo "CHANGELOG.md already updated by iterate command"
else
    # Create CHANGELOG if it doesn't exist
    echo "Creating CHANGELOG.md..."
    cat > CHANGELOG.md << EOF
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]
### Added
- Initial project setup
EOF
fi

# 4. Check for TODOs in code
echo "Checking for TODO comments..."
TODO_COUNT=$(grep -r "TODO\|FIXME\|HACK" --include="*.js" --include="*.ts" --include="*.py" --include="*.go" --include="*.rs" . 2>/dev/null | wc -l || echo "0")
if [ "$TODO_COUNT" -gt 0 ]; then
    echo "  Found $TODO_COUNT TODO/FIXME/HACK comments in code"
    echo "  Consider addressing these in future iterations"
fi

# 5. Update task completion stats
if [ -f "TODO.md" ]; then
    COMPLETED=$(grep -c "\- \[x\]" TODO.md || echo "0")
    REMAINING=$(grep -c "\- \[ \]" TODO.md || echo "0")
    TOTAL=$((COMPLETED + REMAINING))
    
    if [ $TOTAL -gt 0 ]; then
        PERCENTAGE=$((COMPLETED * 100 / TOTAL))
        echo "Task Progress: $COMPLETED/$TOTAL ($PERCENTAGE% complete)"
    fi
fi

echo "Post-iterate hook completed ✓"