#!/bin/bash

# Project Workflow Handler
# Handles project-level commands and epic management

# Read stdin for command data
INPUT=$(cat)

# Function to send notification
send_notification() {
    local message="$1"
    local type="$2"
    
    # Play notification sound
    if [[ "$type" == "success" ]]; then
        afplay /System/Library/Sounds/Hero.aiff 2>/dev/null &
    elif [[ "$type" == "error" ]]; then
        afplay /System/Library/Sounds/Basso.aiff 2>/dev/null &
    else
        afplay /System/Library/Sounds/Glass.aiff 2>/dev/null &
    fi
    
    # Send system notification
    osascript -e "display notification \"$message\" with title \"Claude Code - Project Workflow\""
    
    # Log the notification
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $type: $message" >> ~/.claude/hooks/logs/project-notifications.log
}

# Handle epic transitions
if echo "$INPUT" | grep -q "start.*--next-epic"; then
    send_notification "🔄 Epic transition initiated" "info"
    
    # Check if current epic exists to archive
    if [[ -d "docs/current-epic" ]]; then
        send_notification "📦 Archiving current epic" "info"
    fi
    
    # Check if next epic documentation is created
    if [[ -f "docs/current-epic/PRD.md" ]] && echo "$INPUT" | grep -q "start.*--next-epic"; then
        send_notification "📝 Next epic PRD created - Epic transition completed" "success"
    fi
fi

# Handle enhanced design command
if echo "$INPUT" | grep -q "design" && [[ -f "docs/project/ARCHITECTURE.md" ]]; then
    send_notification "🏗️ Updating global architecture" "info"
fi

# Always exit successfully
exit 0