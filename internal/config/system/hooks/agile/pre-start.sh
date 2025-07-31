#!/bin/bash
# pre-start.sh - Pre-hook for /project:agile:start command

set -e

echo "=== Pre-start Hook ==="

# 1. Backup current state
echo "Creating project state backup..."
bash "$HOME/.claude/hooks/common/backup-state.sh"

# 2. Check git status
echo "Checking git status..."
bash "$HOME/.claude/hooks/common/git-status-check.sh"

# 3. Check for existing PRD
if [ -f "PRD.md" ]; then
    echo "Warning: PRD.md already exists!"
    echo "Consider using a different feature name or archiving the current PRD"
    
    # Backup existing PRD
    mv PRD.md "PRD.md.backup.$(date +%Y%m%d_%H%M%S)"
    echo "Existing PRD backed up"
fi

# 4. Validate environment
echo "Validating project environment..."

# Check for common project files
if [ -f "package.json" ]; then
    echo "  ✓ Node.js project detected"
elif [ -f "requirements.txt" ] || [ -f "pyproject.toml" ]; then
    echo "  ✓ Python project detected"
elif [ -f "go.mod" ]; then
    echo "  ✓ Go project detected"
elif [ -f "Cargo.toml" ]; then
    echo "  ✓ Rust project detected"
else
    echo "  ℹ No standard project structure detected"
    echo "  This might be a new project"
fi

echo "Pre-start checks completed ✓"