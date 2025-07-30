#!/bin/bash

# Simple API for enqueuing background hooks
# Usage: ./enqueue-hook.sh <hook-name> <args> [priority]

HOOKS_DIR="/Users/a.pezzotta/.claude/hooks"
WORKER_BIN="$HOOKS_DIR/background-worker"

# Default values
HOOK_NAME="$1"
ARGS="$2"
PRIORITY="${3:-5}"

if [[ -z "$HOOK_NAME" ]]; then
    echo "Usage: $0 <hook-name> <args> [priority]"
    echo ""
    echo "Background-eligible hooks:"
    echo "  log-commands.py           - Async command logging"
    echo "  enhanced-error-logger.py  - Async error reporting"
    echo "  tool-reliability-analyzer.py - Async analytics"
    echo "  startup-detection.sh      - Async state detection"
    echo ""
    echo "Priority levels:"
    echo "  1-3: High priority (critical logging)"
    echo "  4-6: Normal priority (standard logging)"
    echo "  7-9: Low priority (analytics)"
    exit 1
fi

# Check if background worker is compiled
if [[ ! -x "$WORKER_BIN" ]]; then
    echo "⚠️  Background worker not compiled. Compiling now..."
    cd "$HOOKS_DIR"
    if ! go build -o background-worker background-worker.go; then
        echo "❌ Failed to compile background worker"
        exit 1
    fi
    echo "✅ Background worker compiled successfully"
fi

# Enqueue the job
if "$WORKER_BIN" enqueue "$HOOK_NAME" "$ARGS" "$PRIORITY"; then
    echo "🔄 Background job queued: $HOOK_NAME (priority: $PRIORITY)"
else
    echo "❌ Failed to enqueue background job"
    exit 1
fi