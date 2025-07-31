#!/bin/bash
# Smart notification system for Claude WM CLI
# Usage: smart-notify.sh <event_type>

EVENT_TYPE="$1"

case "$EVENT_TYPE" in
    "question-waiting")
        echo "🤔 Claude is waiting for your input..." >&2
        ;;
    "task-completed")
        echo "✅ Task completed successfully!" >&2
        ;;
    "error-occurred")
        echo "❌ An error occurred during execution" >&2
        ;;
    *)
        echo "📋 Claude WM CLI notification: $EVENT_TYPE" >&2
        ;;
esac

# Optional: Send system notification if available
if command -v osascript >/dev/null 2>&1; then
    # macOS notification
    osascript -e "display notification \"$EVENT_TYPE\" with title \"Claude WM CLI\"" 2>/dev/null || true
elif command -v notify-send >/dev/null 2>&1; then
    # Linux notification
    notify-send "Claude WM CLI" "$EVENT_TYPE" 2>/dev/null || true
fi

exit 0
