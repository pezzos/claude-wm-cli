#!/bin/bash
# git-status-check.sh - Ensure clean git state before starting new work

set -e

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "Warning: Not a git repository. Skipping git status check."
    exit 0
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD -- 2>/dev/null; then
    echo "Error: You have uncommitted changes. Please commit or stash them first."
    echo ""
    echo "Modified files:"
    git status --porcelain | grep -E '^(M| M)' | sed 's/^.../  /'
    echo ""
    echo "To stash changes: git stash"
    echo "To commit changes: git add . && git commit -m 'Your message'"
    exit 1
fi

# Check for untracked files
UNTRACKED=$(git ls-files --others --exclude-standard)
if [ -n "$UNTRACKED" ]; then
    echo "Warning: You have untracked files:"
    echo "$UNTRACKED" | sed 's/^/  /'
    echo ""
    echo "Consider adding them to git or .gitignore"
fi

echo "Git status: Clean working directory ✓"