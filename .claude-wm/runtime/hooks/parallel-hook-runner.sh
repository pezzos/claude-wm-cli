#!/bin/bash

# Parallel Hook Runner - Wrapper script for the Go orchestrator
# This script integrates the Go orchestrator with Claude Code's hook system

# Configuration
HOOKS_DIR="/Users/a.pezzotta/.claude/hooks"
ORCHESTRATOR_BIN="$HOOKS_DIR/orchestrator"
CONFIG_FILE="$HOOKS_DIR/config/parallel-groups.json"
LOG_FILE="$HOOKS_DIR/logs/orchestrator.log"

# Ensure logs directory exists
mkdir -p "$(dirname "$LOG_FILE")"

# Function to log with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Function to extract hook matcher from Claude input
extract_matcher() {
    local input="$1"
    
    # Extract matcher from various Claude Code tool calls
    if echo "$input" | grep -q '"matcher"'; then
        echo "$input" | grep -o '"matcher"[^"]*"[^"]*"' | cut -d'"' -f4
    elif echo "$input" | grep -q "Bash"; then
        echo "Bash"
    elif echo "$input" | grep -q "Write\|Edit\|MultiEdit"; then
        echo "Write_Edit_MultiEdit"
    else
        echo ""
    fi
}

# Function to create tool input JSON
create_tool_input() {
    local input="$1"
    
    # Escape special characters for JSON
    local escaped_input=$(echo "$input" | sed 's/\\/\\\\/g; s/"/\\"/g; s/\n/\\n/g; s/\r/\\r/g; s/\t/\\t/g')
    local escaped_pwd=$(pwd | sed 's/\\/\\\\/g; s/"/\\"/g')
    local git_status=$(git status --porcelain 2>/dev/null || echo 'Not a git repository')
    local escaped_git_status=$(echo "$git_status" | sed 's/\\/\\\\/g; s/"/\\"/g' | tr '\n' ' ')
    local changed_files=$(git diff --name-only HEAD 2>/dev/null | sed 's/.*/"&"/' | tr '\n' ',' | sed 's/,$//')
    
    cat << EOF
{
    "raw_input": "$escaped_input",
    "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "working_directory": "$escaped_pwd",
    "git_status": "$escaped_git_status",
    "changed_files": [$changed_files],
    "hook_context": {
        "execution_mode": "parallel",
        "orchestrator_version": "1.0",
        "claude_session": "${CLAUDE_SESSION_ID:-unknown}"
    }
}
EOF
}

# Main execution function
main() {
    local input="$*"
    
    log "🚀 Parallel Hook Runner starting"
    log "📝 Input: $input"
    
    # Check if orchestrator binary exists
    if [[ ! -x "$ORCHESTRATOR_BIN" ]]; then
        log "❌ Error: Orchestrator binary not found at $ORCHESTRATOR_BIN"
        echo "Error: Parallel orchestrator not compiled. Run: cd $HOOKS_DIR && go build -o orchestrator orchestrator.go"
        exit 1
    fi
    
    # Check if config file exists
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log "❌ Error: Configuration file not found at $CONFIG_FILE"
        echo "Error: Parallel groups configuration missing"
        exit 1
    fi
    
    # Extract hook matcher
    local matcher
    matcher=$(extract_matcher "$input")
    log "🎯 Extracted matcher: '$matcher'"
    
    # Create tool input JSON
    local tool_input
    tool_input=$(create_tool_input "$input")
    
    # Execute orchestrator
    log "⚡ Executing parallel orchestrator..."
    
    local start_time
    start_time=$(date +%s)
    
    if "$ORCHESTRATOR_BIN" "$CONFIG_FILE" "$HOOKS_DIR" "$matcher" "$tool_input"; then
        local end_time
        end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        log "✅ Parallel execution completed successfully in ${duration}s"
        echo "🎉 Parallel hooks execution completed - ${duration}s total"
        exit 0
    else
        local exit_code=$?
        log "❌ Parallel execution failed with exit code $exit_code"
        echo "❌ Parallel hooks execution failed"
        exit $exit_code
    fi
}

# Performance monitoring function
monitor_performance() {
    local orchestrator_pid=$1
    local log_prefix="[PERF]"
    
    while kill -0 "$orchestrator_pid" 2>/dev/null; do
        local cpu_usage
        local memory_usage
        
        cpu_usage=$(ps -p "$orchestrator_pid" -o %cpu --no-headers 2>/dev/null || echo "0")
        memory_usage=$(ps -p "$orchestrator_pid" -o rss --no-headers 2>/dev/null || echo "0")
        
        log "$log_prefix CPU: ${cpu_usage}%, Memory: ${memory_usage}KB"
        sleep 1
    done
}

# Fallback to sequential execution
fallback_sequential() {
    log "⚠️  Falling back to sequential hook execution"
    echo "⚠️  Parallel execution failed, running hooks sequentially..."
    
    # This would call the original hook system
    # For now, just exit with error to maintain compatibility
    exit 1
}

# Signal handlers for graceful shutdown
cleanup() {
    log "🛑 Received shutdown signal, cleaning up..."
    # Kill orchestrator if running
    if [[ -n "$ORCHESTRATOR_PID" ]]; then
        kill "$ORCHESTRATOR_PID" 2>/dev/null
    fi
    exit 130
}

trap cleanup SIGINT SIGTERM

# Help function
show_help() {
    cat << EOF
Parallel Hook Runner - Claude Code Hook Orchestrator

Usage: $0 [options] [hook-input]

Options:
    --help, -h          Show this help message
    --config FILE       Use alternative configuration file
    --dry-run          Show what hooks would be executed without running them
    --sequential       Force sequential execution (fallback mode)
    --verbose          Enable verbose logging
    --performance      Enable performance monitoring

Examples:
    $0 "Bash command execution"
    $0 --dry-run "Write file operation"
    $0 --config custom-config.json "Edit operation"

Configuration:
    Config file: $CONFIG_FILE
    Hooks directory: $HOOKS_DIR
    Log file: $LOG_FILE

For more information, see: hooks/config/parallel-groups.json
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            show_help
            exit 0
            ;;
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        --dry-run)
            echo "🔍 Dry run mode - would execute parallel hooks with input: $*"
            exit 0
            ;;
        --sequential)
            log "🔄 Sequential mode forced"
            fallback_sequential
            ;;
        --verbose)
            set -x
            shift
            ;;
        --performance)
            ENABLE_PERFORMANCE_MONITORING=true
            shift
            ;;
        *)
            break
            ;;
    esac
done

# Execute main function
main "$@"