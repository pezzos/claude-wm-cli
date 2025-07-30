#!/bin/bash

# Project Initialization Command Handler
# Handles /project:init command with file import support

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
    osascript -e "display notification \"$message\" with title \"Claude Code - Project Init\""
    
    # Log the notification
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $type: $message" >> ~/.claude/hooks/logs/project-notifications.log
}

# Check if we're in a project init command
if echo "$INPUT" | grep -q "project:init"; then
    send_notification "🏗️ Project initialization started" "info"
    
    # Check if docs/project structure exists
    if [[ -d "docs/project" ]]; then
        send_notification "📁 Project structure already exists" "info"
    else
        send_notification "📁 Creating project structure" "info"
    fi
fi

# Check if VISION.md was created
if [[ -f "docs/project/VISION.md" ]] && echo "$INPUT" | grep -q "project:init"; then
    send_notification "👁️ Project vision document created" "success"
fi

# Check if all project documents are created
if [[ -f "docs/project/VISION.md" ]] && [[ -f "docs/project/ROADMAP.md" ]] && \
   [[ -f "docs/project/ARCHITECTURE.md" ]] && [[ -f "docs/project/EPICS.md" ]] && \
   echo "$INPUT" | grep -q "project:init"; then
    send_notification "✅ Project initialization completed - Ready to start first epic" "success"
fi

# Always exit successfully
exit 0