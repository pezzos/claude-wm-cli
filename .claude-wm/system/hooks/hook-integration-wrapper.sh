#!/bin/bash

# Hook Integration Wrapper - Integrates tool reliability tracking into existing hook system
# This script modifies the existing hook execution to include tool tracking

HOOKS_DIR="/Users/a.pezzotta/.claude/hooks"
# RELIABILITY_TRACKER merged into this file
MCP_TRACKER="$HOOKS_DIR/mcp-tool-tracker.py"
ANALYZER="$HOOKS_DIR/tool-reliability-analyzer.py"

# Function to wrap bash command execution with tracking
wrap_bash_command() {
    local command="$1"
    local session_id="${CLAUDE_SESSION_ID:-$(date +%s)}"
    local start_time=$(date +%s)
    local exit_code=0
    local error_output=""
    
    # Extract main command from complex command lines
    local main_command=$(echo "$command" | awk '{print $1}')
    
    # Execute the original command and capture output
    local temp_error=$(mktemp)
    eval "$command" 2>"$temp_error"
    exit_code=$?
    
    # Calculate execution time in milliseconds
    local end_time=$(date +%s)
    local execution_time=$(((end_time - start_time) * 1000))
    
    # Read error output
    if [[ -s "$temp_error" ]]; then
        error_output=$(cat "$temp_error")
    fi
    
    # Clean up temp file
    rm -f "$temp_error"
    
    # Inline reliability tracking (merged from tool-reliability-tracker.sh)
    track_bash_execution "$main_command" "$exit_code" "$execution_time" "$error_output" "$session_id"
    
    return $exit_code
}

# Function to track bash tool execution (merged from tool-reliability-tracker.sh)
track_bash_execution() {
    local tool_name="$1"
    local exit_code="$2"
    local execution_time="$3"
    local error_output="$4"
    local session_id="$5"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    local tracker_log="$HOOKS_DIR/logs/tool-reliability.log"
    local stats_file="$HOOKS_DIR/logs/tool-reliability-stats.json"
    
    # Ensure logs directory exists
    mkdir -p "$(dirname "$tracker_log")"
    
    # Log entry
    echo "[$timestamp] [$session_id] $tool_name $exit_code ${execution_time}ms ${error_output:0:100}" >> "$tracker_log"
    
    # Update statistics
    update_bash_tool_stats "$tool_name" "$exit_code" "$execution_time" "$error_output" "$stats_file"
    
    # Alert on repeated failures
    if [[ "$exit_code" != "0" ]]; then
        local recent_failures=$(tail -5 "$tracker_log" | grep -c "$tool_name.*[1-9]")
        if [[ "$recent_failures" -ge 3 ]]; then
            echo "⚠️  WARNING: $tool_name has failed $recent_failures times recently" >&2
        fi
    fi
}

# Function to update tool statistics (merged from tool-reliability-tracker.sh)
update_bash_tool_stats() {
    local tool_name="$1"
    local exit_code="$2"
    local execution_time="$3"
    local error_output="$4"
    local stats_file="$5"
    
    # Initialize stats file if it doesn't exist
    if [[ ! -f "$stats_file" ]]; then
        echo '{}' > "$stats_file"
    fi
    
    # Use Python to update JSON stats
    python3 -c "
import json
import sys
from datetime import datetime

try:
    with open('$stats_file', 'r') as f:
        stats = json.load(f)
except:
    stats = {}

tool = '$tool_name'
exit_code = int('$exit_code')
exec_time = float('$execution_time')
error = '''$error_output'''

if tool not in stats:
    stats[tool] = {
        'total_executions': 0,
        'successful_executions': 0,
        'failed_executions': 0,
        'success_rate': 0.0,
        'average_execution_time': 0.0,
        'total_execution_time': 0.0,
        'common_errors': {},
        'last_updated': '',
        'reliability_score': 0.0,
        'trend': 'stable'
    }

tool_stats = stats[tool]
tool_stats['total_executions'] += 1
tool_stats['total_execution_time'] += exec_time

if exit_code == 0:
    tool_stats['successful_executions'] += 1
else:
    tool_stats['failed_executions'] += 1
    if error and len(error.strip()) > 0:
        error_key = error[:50].strip()
        if error_key not in tool_stats['common_errors']:
            tool_stats['common_errors'][error_key] = 0
        tool_stats['common_errors'][error_key] += 1

# Calculate metrics
tool_stats['success_rate'] = tool_stats['successful_executions'] / tool_stats['total_executions']
tool_stats['average_execution_time'] = tool_stats['total_execution_time'] / tool_stats['total_executions']

# Calculate reliability score
success_factor = tool_stats['success_rate'] * 70
perf_factor = max(0, 20 - (tool_stats['average_execution_time'] / 1000) * 2)
stability_factor = 10 if tool_stats['total_executions'] > 5 else tool_stats['total_executions'] * 2

tool_stats['reliability_score'] = success_factor + perf_factor + stability_factor
tool_stats['last_updated'] = datetime.now().isoformat()

# Simple trend analysis
if tool_stats['total_executions'] >= 10:
    recent_success_rate = tool_stats['successful_executions'] / tool_stats['total_executions']
    if recent_success_rate > 0.9:
        tool_stats['trend'] = 'improving'
    elif recent_success_rate < 0.7:
        tool_stats['trend'] = 'declining'
    else:
        tool_stats['trend'] = 'stable'

with open('$stats_file', 'w') as f:
    json.dump(stats, f, indent=2)
" 2>/dev/null || true
}

# Function to wrap MCP tool execution with tracking
wrap_mcp_tool() {
    local tool_name="$1"
    local session_id="${CLAUDE_SESSION_ID:-$(date +%s)}"
    local start_time=$(date +%s)
    local success="false"
    local error_output=""
    
    # This function would be called after MCP tool execution
    # Parameters: tool_name, success, error_output (passed from Claude)
    
    # For now, we'll create a placeholder that can be called manually
    # In a real implementation, this would be integrated into Claude's MCP execution
    
    shift # Remove tool_name from args
    local success_flag="$1"
    local error_msg="$2"
    
    # Calculate execution time in milliseconds
    local end_time=$(date +%s)
    local response_time=$(((end_time - start_time) * 1000))
    
    # Normalize success flag
    if [[ "$success_flag" == "true" || "$success_flag" == "1" || "$success_flag" == "success" ]]; then
        success="true"
    else
        success="false"
        error_output="$error_msg"
    fi
    
    # Call MCP tracker
    if [[ -x "$MCP_TRACKER" ]]; then
        python3 "$MCP_TRACKER" "$tool_name" "$success" "$response_time" "$error_output" "$session_id" 2>/dev/null || true
    fi
}

# Function to run periodic analysis
run_periodic_analysis() {
    local force_run="$1"
    
    # Run analysis every 50 tool executions or when forced
    local bash_log_lines=$(wc -l < "$HOOKS_DIR/logs/tool-reliability.log" 2>/dev/null || echo 0)
    local mcp_log_lines=$(wc -l < "$HOOKS_DIR/logs/mcp-tool-reliability.log" 2>/dev/null || echo 0)
    local total_lines=$((bash_log_lines + mcp_log_lines))
    
    if [[ "$force_run" == "true" ]] || (( total_lines > 0 && total_lines % 50 == 0 )); then
        if [[ -x "$ANALYZER" ]]; then
            python3 "$ANALYZER" 2>/dev/null || true
        fi
    fi
}

# Function to generate summary report
generate_summary_report() {
    local analysis_file="$HOOKS_DIR/logs/tool-reliability-combined-analysis.json"
    local recommendations_file="$HOOKS_DIR/logs/tool-adaptive-recommendations.json"
    
    if [[ -f "$analysis_file" ]]; then
        echo "📊 Tool Reliability Summary"
        echo "=========================="
        
        # Extract key metrics using Python
        python3 -c "
import json
import sys

try:
    with open('$analysis_file', 'r') as f:
        analysis = json.load(f)
    
    health = analysis['ecosystem_health']
    print('Health Score: {:.1f}% ({})'.format(health['health_score'], health['health_category']))
    print('Total Tools: {}'.format(health['total_tools']))
    print('Reliable Tools: {}'.format(health['reliable_tools']))
    print('Unreliable Tools: {}'.format(health['unreliable_tools']))
    print()
    
    # Show top 5 most reliable tools
    tier1 = analysis['reliability_tiers']['tier_1_excellent']
    if tier1:
        print('🏆 Most Reliable Tools:')
        for i, tool in enumerate(tier1[:5]):
            print('  {}. {} ({:.1%})'.format(i+1, tool['tool'], tool['success_rate']))
        print()
    
    # Show tools needing attention
    tier5 = analysis['reliability_tiers']['tier_5_critical']
    if tier5:
        print('⚠️  Tools Needing Attention:')
        for tool in tier5:
            print('  - {} ({:.1%})'.format(tool['tool'], tool['success_rate']))
        print()
    
    # Show immediate actions
    try:
        with open('$recommendations_file', 'r') as f:
            recommendations = json.load(f)
        
        immediate = recommendations.get('immediate_actions', [])
        if immediate:
            print('🔧 Immediate Actions:')
            for action in immediate[:3]:
                print('  - {} ({})'.format(action['action'], action['priority']))
            print()
    except:
        pass

except Exception as e:
    print('Error reading analysis: {}'.format(e))
" 2>/dev/null || echo "No analysis data available yet"
    else
        echo "No analysis data available. Run some tools first!"
    fi
}

# Main execution logic
main() {
    local action="$1"
    
    # If no arguments provided, run analyze by default (for hook orchestrator)
    if [[ $# -eq 0 ]]; then
        run_periodic_analysis "false"
        return 0
    fi
    
    shift
    
    case "$action" in
        "wrap_bash")
            wrap_bash_command "$@"
            ;;
        "wrap_mcp")
            wrap_mcp_tool "$@"
            ;;
        "analyze")
            run_periodic_analysis "true"
            ;;
        "summary")
            generate_summary_report
            ;;
        "install")
            install_hooks
            ;;
        *)
            echo "Usage: $0 {wrap_bash|wrap_mcp|analyze|summary|install} [args...]"
            echo ""
            echo "Commands:"
            echo "  wrap_bash 'command'     - Wrap bash command execution with tracking"
            echo "  wrap_mcp tool success error - Track MCP tool execution"
            echo "  analyze                 - Run analysis and generate recommendations"
            echo "  summary                 - Show current reliability summary"
            echo "  install                 - Install hooks into system"
            exit 1
            ;;
    esac
}

# Installation function
install_hooks() {
    echo "Installing tool reliability tracking hooks..."
    
    # Make all scripts executable
    chmod +x "$MCP_TRACKER" "$ANALYZER" 2>/dev/null || true
    
    # Create logs directory
    mkdir -p "$HOOKS_DIR/logs"
    
    # Create initial analysis
    if [[ -x "$ANALYZER" ]]; then
        python3 "$ANALYZER" 2>/dev/null || true
    fi
    
    echo "✅ Tool reliability tracking hooks installed successfully!"
    echo ""
    echo "Usage examples:"
    echo "  $0 wrap_bash 'git status'    # Track bash command"
    echo "  $0 wrap_mcp mcp__mem0__ true   # Track MCP tool (success)"
    echo "  $0 analyze                    # Run analysis"
    echo "  $0 summary                    # Show summary"
}

# Execute main function
main "$@"