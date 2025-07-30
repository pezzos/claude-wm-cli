#!/bin/bash

COMMAND=$1
MESSAGE=$2

# Create logs directory if it doesn't exist
mkdir -p ~/.claude/logs

# Lock system to prevent duplicates
LOCK_FILE="/tmp/claude_notify_${COMMAND}_lock"
LOG_FILE="$HOME/.claude/hooks/logs/notifications.log"

# Function to log notifications with timestamp
log_notification() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1: $2" >> "$LOG_FILE"
}

# Prevent focus stealing when user is away
if [ "$USER" != "$(stat -f%Su /dev/console 2>/dev/null)" ]; then
  log_notification "SKIP_AWAY" "$COMMAND - $MESSAGE"
  exit 0
fi

# Check if notifications are disabled
if [ "$CLAUDE_SILENT" = "true" ]; then
  log_notification "SILENT" "$COMMAND - $MESSAGE"
  exit 0
fi

# Check for recent duplicate notifications (last 10 seconds for same message type)
TYPED_LOCK_FILE="/tmp/claude_notify_${COMMAND}_lock"
if [ -f "$TYPED_LOCK_FILE" ]; then
    if command -v stat >/dev/null; then
        # macOS/Linux compatible stat command
        if [[ "$OSTYPE" == "darwin"* ]]; then
            LAST_NOTIFY=$(stat -f %m "$TYPED_LOCK_FILE" 2>/dev/null || echo 0)
        else
            LAST_NOTIFY=$(stat -c %Y "$TYPED_LOCK_FILE" 2>/dev/null || echo 0)
        fi
        CURRENT_TIME=$(date +%s)
        TIME_DIFF=$((CURRENT_TIME - LAST_NOTIFY))
        
        if [ $TIME_DIFF -lt 10 ]; then
            log_notification "SKIP_DUPLICATE" "$COMMAND - $MESSAGE (too recent: ${TIME_DIFF}s)"
            exit 0
        fi
    fi
fi

# Create both lock files (general and typed)
touch "$LOCK_FILE"
touch "$TYPED_LOCK_FILE"

# Log the notification
log_notification "SENT" "$COMMAND - $MESSAGE"

case $COMMAND in
  "task-completed"|"iterate-complete"|"design-complete"|"plan-ready"|"start-complete"|"ship-success"|"milestone"|"test-checkpoint")
    osascript -e "display notification \"Claude Code completed a task\" with title \"Claude Code\" sound name \"Glass\""
    # Clean up related locks after successful completion  
    rm -f /tmp/claude_notify_*_lock 2>/dev/null
    log_notification "CLEANUP" "Notification locks cleared after task completion"
    ;;
  "question-waiting"|"approval-needed"|"input-needed")
    osascript -e "display notification \"Claude Code is awaiting your approval\" with title \"Claude Code\" sound name \"Ping\""
    ;;
  "error")
    osascript -e "display notification \"❌ Error: $MESSAGE\" with title \"Claude Code\" sound name \"Basso\""
    ;;
  "rollback")
    osascript -e "display notification \"⚠️ Rolled back: $MESSAGE\" with title \"Claude Code\" sound name \"Basso\""
    ;;
  *)
    osascript -e "display notification \"$MESSAGE\" with title \"Claude Code\" sound name \"Glass\""
    ;;
esac

# Auto-cleanup locks after 60 seconds  
(sleep 60 && rm -f "$LOCK_FILE" "$TYPED_LOCK_FILE" 2>/dev/null) &