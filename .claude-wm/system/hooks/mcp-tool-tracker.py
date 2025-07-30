#!/usr/bin/env python3

"""
MCP Tool Tracker - Post-execution hook for MCP tool calls
Tracks success/failure rates, response times, and error patterns for all mcp__* tools
"""

import json
import sys
import os
from datetime import datetime
from pathlib import Path

# Configuration
HOOKS_DIR = Path("/Users/a.pezzotta/.claude/hooks")
LOGS_DIR = HOOKS_DIR / "logs"
TRACKER_LOG = LOGS_DIR / "mcp-tool-reliability.log"
STATS_FILE = HOOKS_DIR / "logs" / "mcp-tool-reliability-stats.json"
ANALYSIS_FILE = HOOKS_DIR / "logs" / "mcp-tool-reliability-analysis.json"

# Ensure logs directory exists
LOGS_DIR.mkdir(parents=True, exist_ok=True)

def log_mcp_execution(tool_name, success, response_time, error_output, session_id):
    """Log MCP tool execution to file"""
    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    status = "SUCCESS" if success else "FAILURE"
    error_snippet = error_output[:100] if error_output else ""
    
    log_entry = f"[{timestamp}] [{session_id}] {tool_name} {status} {response_time}ms {error_snippet}\n"
    
    with open(TRACKER_LOG, 'a') as f:
        f.write(log_entry)

def update_mcp_stats(tool_name, success, response_time, error_output):
    """Update MCP tool statistics"""
    
    # Initialize stats file if it doesn't exist
    if not STATS_FILE.exists():
        stats = {}
    else:
        try:
            with open(STATS_FILE, 'r') as f:
                stats = json.load(f)
        except:
            stats = {}
    
    # Initialize tool stats if not exists
    if tool_name not in stats:
        stats[tool_name] = {
            'total_calls': 0,
            'successful_calls': 0,
            'failed_calls': 0,
            'success_rate': 0.0,
            'average_response_time': 0.0,
            'total_response_time': 0.0,
            'common_errors': {},
            'last_updated': '',
            'reliability_score': 0.0,
            'trend': 'stable',
            'error_patterns': {},
            'performance_category': 'normal'
        }
    
    tool_stats = stats[tool_name]
    tool_stats['total_calls'] += 1
    tool_stats['total_response_time'] += response_time
    
    if success:
        tool_stats['successful_calls'] += 1
    else:
        tool_stats['failed_calls'] += 1
        
        # Track error patterns
        if error_output:
            error_key = error_output[:50].strip()
            if error_key not in tool_stats['common_errors']:
                tool_stats['common_errors'][error_key] = 0
            tool_stats['common_errors'][error_key] += 1
            
            # Categorize error patterns
            categorize_error(tool_stats, error_output)
    
    # Calculate metrics
    tool_stats['success_rate'] = tool_stats['successful_calls'] / tool_stats['total_calls']
    tool_stats['average_response_time'] = tool_stats['total_response_time'] / tool_stats['total_calls']
    
    # Calculate reliability score (0-100)
    success_factor = tool_stats['success_rate'] * 70
    perf_factor = calculate_performance_factor(tool_stats['average_response_time'])
    stability_factor = calculate_stability_factor(tool_stats['total_calls'])
    
    tool_stats['reliability_score'] = success_factor + perf_factor + stability_factor
    tool_stats['last_updated'] = datetime.now().isoformat()
    
    # Performance categorization
    if tool_stats['average_response_time'] < 500:
        tool_stats['performance_category'] = 'fast'
    elif tool_stats['average_response_time'] < 2000:
        tool_stats['performance_category'] = 'normal'
    elif tool_stats['average_response_time'] < 5000:
        tool_stats['performance_category'] = 'slow'
    else:
        tool_stats['performance_category'] = 'very_slow'
    
    # Trend analysis
    if tool_stats['total_calls'] >= 10:
        recent_success_rate = tool_stats['successful_calls'] / tool_stats['total_calls']
        if recent_success_rate > 0.9:
            tool_stats['trend'] = 'improving'
        elif recent_success_rate < 0.7:
            tool_stats['trend'] = 'declining'
        else:
            tool_stats['trend'] = 'stable'
    
    # Save updated stats
    with open(STATS_FILE, 'w') as f:
        json.dump(stats, f, indent=2)

def categorize_error(tool_stats, error_output):
    """Categorize error patterns for better analysis"""
    if 'error_patterns' not in tool_stats:
        tool_stats['error_patterns'] = {}
    
    patterns = tool_stats['error_patterns']
    
    # Common MCP error patterns
    error_lower = error_output.lower()
    
    if 'timeout' in error_lower:
        patterns['timeout'] = patterns.get('timeout', 0) + 1
    elif 'permission' in error_lower or 'unauthorized' in error_lower:
        patterns['permission'] = patterns.get('permission', 0) + 1
    elif 'network' in error_lower or 'connection' in error_lower:
        patterns['network'] = patterns.get('network', 0) + 1
    elif 'rate limit' in error_lower:
        patterns['rate_limit'] = patterns.get('rate_limit', 0) + 1
    elif 'invalid' in error_lower or 'malformed' in error_lower:
        patterns['invalid_input'] = patterns.get('invalid_input', 0) + 1
    elif 'not found' in error_lower:
        patterns['not_found'] = patterns.get('not_found', 0) + 1
    else:
        patterns['other'] = patterns.get('other', 0) + 1

def calculate_performance_factor(avg_time):
    """Calculate performance factor for reliability score"""
    if avg_time < 500:
        return 20  # Fast
    elif avg_time < 1000:
        return 15  # Good
    elif avg_time < 2000:
        return 10  # Acceptable
    elif avg_time < 5000:
        return 5   # Slow
    else:
        return 0   # Very slow

def calculate_stability_factor(total_calls):
    """Calculate stability factor based on number of calls"""
    if total_calls > 20:
        return 10  # High confidence
    elif total_calls > 10:
        return 7   # Medium confidence
    elif total_calls > 5:
        return 5   # Low confidence
    else:
        return total_calls  # Very low confidence

def generate_mcp_recommendations():
    """Generate recommendations for MCP tool usage"""
    if not STATS_FILE.exists():
        return
    
    try:
        with open(STATS_FILE, 'r') as f:
            stats = json.load(f)
    except:
        return
    
    recommendations = {
        'unreliable_tools': [],
        'slow_tools': [],
        'alternatives': {},
        'usage_patterns': {},
        'best_practices': {},
        'error_insights': {},
        'generated_at': datetime.now().isoformat()
    }
    
    for tool, data in stats.items():
        # Flag unreliable tools (success rate < 80%)
        if data['success_rate'] < 0.8 and data['total_calls'] > 3:
            recommendations['unreliable_tools'].append({
                'tool': tool,
                'success_rate': data['success_rate'],
                'common_errors': list(data['common_errors'].keys())[:3],
                'error_patterns': data.get('error_patterns', {}),
                'suggestion': generate_tool_suggestion(tool, data)
            })
        
        # Flag slow tools (avg > 3 seconds)
        if data['average_response_time'] > 3000:
            recommendations['slow_tools'].append({
                'tool': tool,
                'avg_time': data['average_response_time'],
                'performance_category': data.get('performance_category', 'unknown'),
                'suggestion': generate_performance_suggestion(tool, data)
            })
        
        # Generate best practices
        if data['success_rate'] > 0.95 and data['total_calls'] > 5:
            recommendations['best_practices'][tool] = {
                'success_rate': data['success_rate'],
                'avg_time': data['average_response_time'],
                'notes': f"Highly reliable tool with {data['success_rate']:.1%} success rate"
            }
        
        # Error insights
        if data['failed_calls'] > 0:
            recommendations['error_insights'][tool] = {
                'failure_rate': data['failed_calls'] / data['total_calls'],
                'most_common_errors': sorted(
                    data['common_errors'].items(),
                    key=lambda x: x[1],
                    reverse=True
                )[:3],
                'error_patterns': data.get('error_patterns', {})
            }
    
    # Save recommendations
    with open(ANALYSIS_FILE, 'w') as f:
        json.dump(recommendations, f, indent=2)

def generate_tool_suggestion(tool, data):
    """Generate specific suggestions for unreliable tools"""
    suggestions = {
        'mcp__github__': 'Check GitHub token permissions and rate limits',
        'mcp__playwright__': 'Ensure browser dependencies are installed',
        'mcp_mem0__': 'Check memory storage permissions and disk space',
        'mcp__time__': 'Check timezone data and system time settings'
    }
    
    for prefix, suggestion in suggestions.items():
        if tool.startswith(prefix):
            return suggestion
    
    return 'Review tool configuration and error patterns'

def generate_performance_suggestion(tool, data):
    """Generate performance improvement suggestions"""
    if 'github' in tool:
        return 'Consider using GraphQL API or reducing payload size'
    elif 'aws' in tool:
        return 'Check AWS region settings and service limits'
    elif 'playwright' in tool:
        return 'Consider headless mode or reduced wait times'
    elif 'memory' in tool:
        return 'Consider batch operations or data optimization'
    else:
        return 'Review parameters and consider alternative approaches'

def main():
    """Main execution logic"""
    if len(sys.argv) < 6:
        print("Usage: mcp-tool-tracker.py <tool_name> <success> <response_time> <error_output> <session_id>")
        sys.exit(1)
    
    tool_name = sys.argv[1]
    success = sys.argv[2].lower() == 'true'
    response_time = float(sys.argv[3])
    error_output = sys.argv[4] if sys.argv[4] != 'None' else ''
    session_id = sys.argv[5]
    
    # Skip if not an MCP tool
    if not tool_name.startswith('mcp__'):
        sys.exit(0)
    
    # Log the execution
    log_mcp_execution(tool_name, success, response_time, error_output, session_id)
    
    # Update statistics
    update_mcp_stats(tool_name, success, response_time, error_output)
    
    # Generate recommendations periodically
    if TRACKER_LOG.exists():
        line_count = sum(1 for _ in open(TRACKER_LOG))
        if line_count % 10 == 0:  # Every 10 MCP calls
            generate_mcp_recommendations()
    
    # Alert on repeated failures
    if not success:
        recent_log_lines = []
        if TRACKER_LOG.exists():
            with open(TRACKER_LOG, 'r') as f:
                recent_log_lines = f.readlines()[-10:]  # Last 10 lines
        
        recent_failures = sum(1 for line in recent_log_lines 
                             if tool_name in line and 'FAILURE' in line)
        
        if recent_failures >= 3:
            print(f"⚠️  WARNING: {tool_name} has failed {recent_failures} times recently", file=sys.stderr)
            
            # Check if we have analysis available
            if ANALYSIS_FILE.exists():
                try:
                    with open(ANALYSIS_FILE, 'r') as f:
                        analysis = json.load(f)
                    
                    # Find specific suggestions
                    for unreliable in analysis.get('unreliable_tools', []):
                        if unreliable['tool'] == tool_name:
                            print(f"💡 Suggestion: {unreliable['suggestion']}", file=sys.stderr)
                            break
                except:
                    pass

if __name__ == '__main__':
    main()