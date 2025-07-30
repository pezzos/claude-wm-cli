#!/bin/bash

# Startup Project Detection Hook
# Automatically detects project state and suggests next action
# Replaces the "Startup Project Detection Protocol" section from CLAUDE.md

# Create detection results file
DETECTION_FILE="$HOME/.claude/hooks/config/project-detection.json"

# Function to detect project state
detect_project_state() {
    local state=""
    local message=""
    local suggested_action=""
    local example=""
    local alternatives=""
    
    # Check for agile workflow files
    if [[ ! -f "PRD.md" ]]; then
        state="NEW_PROJECT"
        message="🚀 New project detected! Start by describing what you want to build."
        suggested_action="/project:agile:start [description]"
        example="/project:agile:start user authentication system"
        alternatives="/investigate (explore existing code)"
        
    elif [[ -f "PRD.md" ]] && [[ ! -f "DESIGN.md" || ! -f "ARCHITECTURE.md" ]]; then
        state="NEEDS_DESIGN"
        message="📋 PRD found! Ready to create technical design and architecture."
        suggested_action="/project:agile:design"
        example=""
        alternatives="/project:agile:start [new-feature]"
        
    elif [[ -f "DESIGN.md" && -f "ARCHITECTURE.md" ]] && [[ ! -f "PLAN.md" || ! -f "TODO.md" ]]; then
        state="NEEDS_PLANNING"
        message="🎯 Design complete! Ready to create project plan and task breakdown."
        suggested_action="/project:agile:plan"
        example=""
        alternatives="/project:agile:start [new-feature]"
        
    elif [[ -f "PLAN.md" && -f "TODO.md" ]]; then
        # Count uncompleted vs completed tasks
        local uncompleted=$(grep -c "- \[ \]" TODO.md 2>/dev/null || echo "0")
        local completed=$(grep -c "- \[x\]" TODO.md 2>/dev/null || echo "0")
        
        if [[ $uncompleted -gt 0 ]]; then
            state="ACTIVE_DEVELOPMENT"
            message="⚡ Ready to continue development! $uncompleted tasks remaining in TODO.md."
            suggested_action="/project:agile:iterate"
            example=""
            alternatives="/project:agile:start [new-feature], /project:agile:ship"
        else
            state="READY_TO_SHIP"
            message="🚢 All tasks completed! Ready to finalize and ship."
            suggested_action="/project:agile:ship"
            example=""
            alternatives="/project:agile:start [new-feature]"
        fi
    fi
    
    # Create JSON output
    cat > "$DETECTION_FILE" << EOF
{
    "state": "$state",
    "message": "$message",
    "suggested_action": "$suggested_action",
    "example": "$example",
    "alternatives": "$alternatives",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "working_directory": "$(pwd)"
}
EOF
    
    # Display formatted output
    echo "🔍 Project State: $state"
    echo "💡 $message"
    echo "✨ Suggested action: $suggested_action"
    [[ -n "$example" ]] && echo "📝 Example: $example"
    echo "🔄 Alternatives: $alternatives"
    
    # Log the detection
    echo "$(date '+%Y-%m-%d %H:%M:%S') - Project state detected: $state in $(pwd)" >> "$HOME/.claude/hooks/logs/startup-detection.log"
}

# Main execution
main() {
    # Create logs directory
    mkdir -p "$HOME/.claude/hooks/logs"
    
    # Only run if we're in a potential project directory
    if [[ -w "." ]]; then
        detect_project_state
    fi
}

# Run detection
main

exit 0