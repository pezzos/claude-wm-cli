#!/bin/sh
# Claude WM CLI Pre-commit Hook
# 
# This hook runs 'claude-wm-cli guard check' before each commit.
# It will block commits that violate SELF mode restrictions.
#
# Installation: claude-wm-cli guard install-hook
# Manual installation: copy this file to .git/hooks/pre-commit and chmod +x

set -e

# Find git repository root
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "")
CLAUDE_WM_CLI=""

# Strategy 1: Look for claude-wm-cli in repo root
if [ -n "$REPO_ROOT" ] && [ -x "$REPO_ROOT/claude-wm-cli" ]; then
    CLAUDE_WM_CLI="$REPO_ROOT/claude-wm-cli"
    echo "Using local claude-wm-cli: $CLAUDE_WM_CLI"
elif [ -n "$REPO_ROOT" ] && [ -x "$REPO_ROOT/build/claude-wm-cli" ]; then
    CLAUDE_WM_CLI="$REPO_ROOT/build/claude-wm-cli"
    echo "Using built claude-wm-cli: $CLAUDE_WM_CLI"
# Strategy 2: Look in PATH
elif command -v claude-wm-cli >/dev/null 2>&1; then
    CLAUDE_WM_CLI="claude-wm-cli"
    echo "Using claude-wm-cli from PATH"
# Strategy 3: Try to build it
elif [ -n "$REPO_ROOT" ] && [ -f "$REPO_ROOT/Makefile" ]; then
    echo "claude-wm-cli not found, attempting to build..."
    cd "$REPO_ROOT"
    if make build >/dev/null 2>&1; then
        if [ -x "$REPO_ROOT/build/claude-wm-cli" ]; then
            CLAUDE_WM_CLI="$REPO_ROOT/build/claude-wm-cli"
            echo "Built and using: $CLAUDE_WM_CLI"
        fi
    fi
fi

# Final check
if [ -z "$CLAUDE_WM_CLI" ]; then
    echo "Error: claude-wm-cli not found" >&2
    echo "Tried:" >&2
    echo "  1. $REPO_ROOT/claude-wm-cli" >&2
    echo "  2. $REPO_ROOT/build/claude-wm-cli" >&2
    echo "  3. PATH lookup" >&2
    echo "  4. Building with make" >&2
    echo "" >&2
    echo "Please ensure claude-wm-cli is available or the build system works." >&2
    exit 1
fi

# Run guard check
echo "Running pre-commit guard check..."

# Execute guard check and capture result
if "$CLAUDE_WM_CLI" guard check; then
    echo "✅ Guard check passed - commit allowed"
    exit 0
else
    echo "" >&2
    echo "❌ Commit blocked by guard check" >&2
    echo "" >&2
    echo "The guard check detected violations that prevent this commit." >&2
    echo "Please fix the issues above and try committing again." >&2
    echo "" >&2
    echo "To bypass this check temporarily, use:" >&2
    echo "  git commit --no-verify" >&2
    echo "" >&2
    exit 1
fi