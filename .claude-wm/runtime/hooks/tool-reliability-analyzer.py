#!/usr/bin/env python3

"""
Tool Reliability Analyzer - Analyzes tool usage patterns and generates adaptive recommendations
Combines data from both bash tools and MCP tools to provide comprehensive insights
"""

import json
import sys
import os
from datetime import datetime, timedelta
from pathlib import Path
from collections import defaultdict, Counter

# Configuration
HOOKS_DIR = Path("/Users/a.pezzotta/.claude/hooks")
BASH_STATS_FILE = HOOKS_DIR / "logs" / "tool-reliability-stats.json"
MCP_STATS_FILE = HOOKS_DIR / "logs" / "mcp-tool-reliability-stats.json"
COMBINED_ANALYSIS_FILE = HOOKS_DIR / "logs" / "tool-reliability-combined-analysis.json"
ADAPTIVE_RECOMMENDATIONS_FILE = HOOKS_DIR / "logs" / "tool-adaptive-recommendations.json"

def load_stats():
    """Load statistics from both bash and MCP tracking files"""
    bash_stats = {}
    mcp_stats = {}
    
    if BASH_STATS_FILE.exists():
        try:
            with open(BASH_STATS_FILE, 'r') as f:
                bash_stats = json.load(f)
        except:
            pass
    
    if MCP_STATS_FILE.exists():
        try:
            with open(MCP_STATS_FILE, 'r') as f:
                mcp_stats = json.load(f)
        except:
            pass
    
    return bash_stats, mcp_stats

def analyze_tool_ecosystem(bash_stats, mcp_stats):
    """Analyze the overall tool ecosystem for patterns and insights"""
    
    analysis = {
        'ecosystem_health': {},
        'tool_categories': {},
        'reliability_tiers': {},
        'usage_patterns': {},
        'performance_insights': {},
        'error_correlations': {},
        'recommendations': {},
        'mcp_adoption_metrics': {},
        'temporal_patterns': {},
        'security_insights': {},
        'generated_at': datetime.now().isoformat()
    }
    
    # Combine all tools for analysis
    all_tools = {}
    
    # Process bash tools
    for tool, stats in bash_stats.items():
        all_tools[tool] = {
            'type': 'bash',
            'stats': stats,
            'category': categorize_bash_tool(tool)
        }
    
    # Process MCP tools
    for tool, stats in mcp_stats.items():
        all_tools[tool] = {
            'type': 'mcp',
            'stats': stats,
            'category': categorize_mcp_tool(tool)
        }
    
    # Analyze ecosystem health
    analysis['ecosystem_health'] = analyze_ecosystem_health(all_tools)
    
    # Categorize tools
    analysis['tool_categories'] = categorize_tools(all_tools)
    
    # Create reliability tiers
    analysis['reliability_tiers'] = create_reliability_tiers(all_tools)
    
    # Analyze usage patterns
    analysis['usage_patterns'] = analyze_usage_patterns(all_tools)
    
    # Performance insights
    analysis['performance_insights'] = analyze_performance(all_tools)
    
    # Error correlations
    analysis['error_correlations'] = analyze_error_correlations(all_tools)
    
    # Generate adaptive recommendations
    analysis['recommendations'] = generate_adaptive_recommendations(all_tools, analysis)
    
    # MCP adoption metrics
    analysis['mcp_adoption_metrics'] = analyze_mcp_adoption(all_tools)
    
    # Temporal patterns
    analysis['temporal_patterns'] = analyze_temporal_patterns(all_tools)
    
    # Security insights
    analysis['security_insights'] = analyze_security_patterns(all_tools)
    
    return analysis

def categorize_bash_tool(tool):
    """Categorize bash tools by function"""
    categories = {
        'file_operations': ['ls', 'find', 'cat', 'grep', 'rg', 'fd', 'bat', 'exa'],
        'git_operations': ['git'],
        'package_management': ['npm', 'pip', 'brew', 'apt'],
        'system_operations': ['chmod', 'mkdir', 'rm', 'mv', 'cp', 'pwd'],
        'process_management': ['ps', 'kill', 'top', 'htop'],
        'network_operations': ['curl', 'wget', 'ping', 'ssh'],
        'text_processing': ['sed', 'awk', 'sort', 'uniq', 'wc'],
        'development': ['node', 'python', 'java', 'make', 'docker']
    }
    
    for category, tools in categories.items():
        if tool in tools:
            return category
    
    return 'other'

def categorize_mcp_tool(tool):
    """Categorize MCP tools by function"""
    if tool.startswith('mcp__github__'):
        return 'source_control'
    elif tool.startswith('mcp__aws__'):
        return 'cloud_services'
    elif tool.startswith('mcp__playwright__'):
        return 'browser_automation'
    elif tool.startswith('mcp_mem0__'):
        return 'data_storage'
    elif tool.startswith('mcp__time__'):
        return 'utilities'
    elif tool.startswith('mcp__consult7__'):
        return 'ai_services'
    elif tool.startswith('mcp__sequential-thinking__'):
        return 'ai_services'
    else:
        return 'other'

def analyze_ecosystem_health(all_tools):
    """Analyze overall ecosystem health"""
    total_tools = len(all_tools)
    reliable_tools = sum(1 for tool_data in all_tools.values() 
                        if tool_data['stats'].get('success_rate', 0) > 0.8)
    unreliable_tools = sum(1 for tool_data in all_tools.values() 
                          if tool_data['stats'].get('success_rate', 0) < 0.5)
    
    health_score = (reliable_tools / total_tools) * 100 if total_tools > 0 else 0
    
    return {
        'total_tools': total_tools,
        'reliable_tools': reliable_tools,
        'unreliable_tools': unreliable_tools,
        'health_score': health_score,
        'health_category': get_health_category(health_score)
    }

def get_health_category(score):
    """Get health category based on score"""
    if score >= 90:
        return 'excellent'
    elif score >= 80:
        return 'good'
    elif score >= 70:
        return 'fair'
    elif score >= 60:
        return 'poor'
    else:
        return 'critical'

def create_reliability_tiers(all_tools):
    """Create reliability tiers for tools"""
    tiers = {
        'tier_1_excellent': [],  # 95%+ success rate
        'tier_2_good': [],       # 85-95% success rate
        'tier_3_fair': [],       # 70-85% success rate
        'tier_4_poor': [],       # 50-70% success rate
        'tier_5_critical': []    # <50% success rate
    }
    
    for tool_name, tool_data in all_tools.items():
        success_rate = tool_data['stats'].get('success_rate', 0)
        
        if success_rate >= 0.95:
            tiers['tier_1_excellent'].append({
                'tool': tool_name,
                'success_rate': success_rate,
                'type': tool_data['type']
            })
        elif success_rate >= 0.85:
            tiers['tier_2_good'].append({
                'tool': tool_name,
                'success_rate': success_rate,
                'type': tool_data['type']
            })
        elif success_rate >= 0.70:
            tiers['tier_3_fair'].append({
                'tool': tool_name,
                'success_rate': success_rate,
                'type': tool_data['type']
            })
        elif success_rate >= 0.50:
            tiers['tier_4_poor'].append({
                'tool': tool_name,
                'success_rate': success_rate,
                'type': tool_data['type']
            })
        else:
            tiers['tier_5_critical'].append({
                'tool': tool_name,
                'success_rate': success_rate,
                'type': tool_data['type']
            })
    
    return tiers

def categorize_tools(all_tools):
    """Categorize tools by function"""
    categories = defaultdict(list)
    
    for tool_name, tool_data in all_tools.items():
        category = tool_data['category']
        categories[category].append({
            'tool': tool_name,
            'type': tool_data['type'],
            'success_rate': tool_data['stats'].get('success_rate', 0),
            'avg_time': tool_data['stats'].get('average_response_time', 0) if tool_data['type'] == 'mcp' 
                       else tool_data['stats'].get('average_execution_time', 0)
        })
    
    return dict(categories)

def analyze_usage_patterns(all_tools):
    """Analyze tool usage patterns"""
    patterns = {
        'most_used_tools': [],
        'least_used_tools': [],
        'bash_vs_mcp_usage': {},
        'category_usage': defaultdict(int),
        'time_patterns': {}
    }
    
    # Sort by usage frequency
    sorted_tools = sorted(all_tools.items(), 
                         key=lambda x: x[1]['stats'].get('total_executions', 0) + 
                                      x[1]['stats'].get('total_calls', 0), 
                         reverse=True)
    
    patterns['most_used_tools'] = [
        {
            'tool': tool_name,
            'usage_count': tool_data['stats'].get('total_executions', 0) + 
                          tool_data['stats'].get('total_calls', 0),
            'type': tool_data['type']
        } for tool_name, tool_data in sorted_tools[:10]
    ]
    
    patterns['least_used_tools'] = [
        {
            'tool': tool_name,
            'usage_count': tool_data['stats'].get('total_executions', 0) + 
                          tool_data['stats'].get('total_calls', 0),
            'type': tool_data['type']
        } for tool_name, tool_data in sorted_tools[-10:]
    ]
    
    # Bash vs MCP usage
    bash_usage = sum(tool_data['stats'].get('total_executions', 0) 
                     for tool_data in all_tools.values() 
                     if tool_data['type'] == 'bash')
    mcp_usage = sum(tool_data['stats'].get('total_calls', 0) 
                    for tool_data in all_tools.values() 
                    if tool_data['type'] == 'mcp')
    
    patterns['bash_vs_mcp_usage'] = {
        'bash_usage': bash_usage,
        'mcp_usage': mcp_usage,
        'mcp_adoption_rate': mcp_usage / (bash_usage + mcp_usage) if (bash_usage + mcp_usage) > 0 else 0
    }
    
    # Category usage
    for tool_data in all_tools.values():
        usage_count = tool_data['stats'].get('total_executions', 0) + tool_data['stats'].get('total_calls', 0)
        patterns['category_usage'][tool_data['category']] += usage_count
    
    return patterns

def analyze_performance(all_tools):
    """Analyze performance characteristics"""
    performance = {
        'fastest_tools': [],
        'slowest_tools': [],
        'performance_by_category': {},
        'performance_recommendations': []
    }
    
    # Get tools with timing data
    timed_tools = []
    for tool_name, tool_data in all_tools.items():
        avg_time = (tool_data['stats'].get('average_response_time', 0) if tool_data['type'] == 'mcp' 
                   else tool_data['stats'].get('average_execution_time', 0))
        if avg_time > 0:
            timed_tools.append((tool_name, avg_time, tool_data['type'], tool_data['category']))
    
    # Sort by performance
    timed_tools.sort(key=lambda x: x[1])
    
    performance['fastest_tools'] = [
        {'tool': tool, 'avg_time': time, 'type': tool_type, 'category': category}
        for tool, time, tool_type, category in timed_tools[:10]
    ]
    
    performance['slowest_tools'] = [
        {'tool': tool, 'avg_time': time, 'type': tool_type, 'category': category}
        for tool, time, tool_type, category in timed_tools[-10:]
    ]
    
    # Performance by category
    category_performance = defaultdict(list)
    for tool, time, tool_type, category in timed_tools:
        category_performance[category].append(time)
    
    for category, times in category_performance.items():
        performance['performance_by_category'][category] = {
            'avg_time': sum(times) / len(times),
            'min_time': min(times),
            'max_time': max(times),
            'tool_count': len(times)
        }
    
    return performance

def analyze_error_correlations(all_tools):
    """Analyze error patterns and correlations"""
    correlations = {
        'common_error_patterns': {},
        'tools_with_similar_errors': {},
        'error_trends': {},
        'systemic_issues': []
    }
    
    # Collect all errors
    all_errors = []
    for tool_name, tool_data in all_tools.items():
        common_errors = tool_data['stats'].get('common_errors', {})
        error_patterns = tool_data['stats'].get('error_patterns', {})
        
        for error, count in common_errors.items():
            all_errors.append({
                'tool': tool_name,
                'error': error,
                'count': count,
                'type': tool_data['type']
            })
    
    # Find common error patterns
    error_counter = Counter(error['error'] for error in all_errors)
    correlations['common_error_patterns'] = dict(error_counter.most_common(10))
    
    # Find tools with similar errors
    error_groups = defaultdict(list)
    for error in all_errors:
        error_groups[error['error']].append(error['tool'])
    
    # Only include errors that affect multiple tools
    correlations['tools_with_similar_errors'] = {
        error: list(set(tools)) for error, tools in error_groups.items() 
        if len(set(tools)) > 1
    }
    
    # Identify systemic issues
    systemic_threshold = 3  # Issues affecting 3+ tools
    for error, tools in correlations['tools_with_similar_errors'].items():
        if len(tools) >= systemic_threshold:
            correlations['systemic_issues'].append({
                'error': error,
                'affected_tools': tools,
                'severity': 'high' if len(tools) > 5 else 'medium'
            })
    
    return correlations

def generate_adaptive_recommendations(all_tools, analysis):
    """Generate adaptive recommendations based on analysis"""
    recommendations = {
        'immediate_actions': [],
        'tool_substitutions': {},
        'optimization_opportunities': [],
        'habit_changes': [],
        'system_improvements': []
    }
    
    # Immediate actions for critical tools
    for tool_data in analysis['reliability_tiers']['tier_5_critical']:
        tool_name = tool_data['tool']
        recommendations['immediate_actions'].append({
            'action': f'Investigate {tool_name} failures',
            'priority': 'high',
            'reason': f'Success rate: {tool_data["success_rate"]:.1%}',
            'suggestion': get_tool_fix_suggestion(tool_name, all_tools[tool_name])
        })
    
    # Tool substitutions
    substitutions = {
        'find': 'fd',
        'grep': 'rg',
        'cat': 'bat',
        'ls': 'exa',
    }
    
    for old_tool, new_tool in substitutions.items():
        if old_tool in all_tools and all_tools[old_tool]['stats'].get('success_rate', 1) < 0.8:
            recommendations['tool_substitutions'][old_tool] = {
                'alternative': new_tool,
                'reason': f'Current success rate: {all_tools[old_tool]["stats"].get("success_rate", 0):.1%}'
            }
    
    # Optimization opportunities
    slow_tools = [tool for tool in analysis['performance_insights']['slowest_tools'] 
                  if tool['avg_time'] > 2000]
    
    for tool in slow_tools:
        recommendations['optimization_opportunities'].append({
            'tool': tool['tool'],
            'current_time': tool['avg_time'],
            'suggestion': get_optimization_suggestion(tool['tool'], tool['category'])
        })
    
    # Habit changes based on usage patterns
    mcp_adoption = analysis['usage_patterns']['bash_vs_mcp_usage']['mcp_adoption_rate']
    if mcp_adoption < 0.3:
        recommendations['habit_changes'].append({
            'habit': 'Increase MCP tool usage',
            'reason': f'Current MCP adoption: {mcp_adoption:.1%}',
            'suggestion': 'Consider using MCP equivalents for git, file operations, and cloud services'
        })
    
    # System improvements
    health_score = analysis['ecosystem_health']['health_score']
    if health_score < 80:
        recommendations['system_improvements'].append({
            'improvement': 'Overall tool reliability',
            'current_score': health_score,
            'suggestion': 'Focus on fixing tools in tier 4 and 5 reliability'
        })
    
    return recommendations

def get_tool_fix_suggestion(tool_name, tool_data):
    """Get specific fix suggestion for a tool"""
    if tool_data['type'] == 'bash':
        return f"Check {tool_name} installation and permissions"
    elif tool_name.startswith('mcp__github__'):
        return "Verify GitHub token and API limits"
    else:
        return "Review tool configuration and error logs"

def get_optimization_suggestion(tool_name, category):
    """Get optimization suggestion for slow tools"""
    suggestions = {
        'file_operations': 'Consider using modern alternatives like fd, rg, or bat',
        'source_control': 'Use batch operations or reduce payload size',
        'cloud_services': 'Check network connectivity and service regions',
        'browser_automation': 'Use headless mode or reduce wait times',
        'ai_services': 'Consider caching responses or reducing context size'
    }
    
    return suggestions.get(category, 'Review parameters and consider alternative approaches')

def analyze_mcp_adoption(all_tools):
    """Analyze MCP tool adoption patterns"""
    mcp_tools = {k: v for k, v in all_tools.items() if v['type'] == 'mcp'}
    bash_tools = {k: v for k, v in all_tools.items() if v['type'] == 'bash'}
    
    mcp_categories = defaultdict(list)
    for tool_name, tool_data in mcp_tools.items():
        mcp_categories[tool_data['category']].append(tool_name)
    
    total_mcp_usage = sum(tool_data['stats'].get('total_calls', 0) for tool_data in mcp_tools.values())
    total_bash_usage = sum(tool_data['stats'].get('total_executions', 0) for tool_data in bash_tools.values())
    
    return {
        'mcp_tool_count': len(mcp_tools),
        'bash_tool_count': len(bash_tools),
        'mcp_categories': dict(mcp_categories),
        'adoption_rate': total_mcp_usage / (total_mcp_usage + total_bash_usage) if (total_mcp_usage + total_bash_usage) > 0 else 0,
        'top_mcp_tools': sorted(mcp_tools.items(), key=lambda x: x[1]['stats'].get('total_calls', 0), reverse=True)[:5],
        'mcp_success_rate': sum(tool_data['stats'].get('success_rate', 0) for tool_data in mcp_tools.values()) / len(mcp_tools) if mcp_tools else 0
    }

def analyze_temporal_patterns(all_tools):
    """Analyze temporal usage patterns"""
    return {
        'peak_usage_indicators': analyze_peak_usage(all_tools),
        'trend_analysis': analyze_usage_trends(all_tools),
        'seasonal_patterns': analyze_seasonal_patterns(all_tools)
    }

def analyze_peak_usage(all_tools):
    """Analyze peak usage patterns"""
    patterns = {}
    for tool_name, tool_data in all_tools.items():
        usage_count = tool_data['stats'].get('total_executions', 0) + tool_data['stats'].get('total_calls', 0)
        if usage_count > 0:
            patterns[tool_name] = {
                'usage_frequency': usage_count,
                'category': tool_data['category'],
                'type': tool_data['type']
            }
    return patterns

def analyze_usage_trends(all_tools):
    """Analyze usage trends over time"""
    trends = {}
    for tool_name, tool_data in all_tools.items():
        trend = tool_data['stats'].get('trend', 'stable')
        trends[tool_name] = {
            'trend': trend,
            'reliability_score': tool_data['stats'].get('reliability_score', 0),
            'last_updated': tool_data['stats'].get('last_updated', '')
        }
    return trends

def analyze_seasonal_patterns(all_tools):
    """Analyze seasonal usage patterns"""
    return {
        'development_cycles': analyze_development_cycles(all_tools),
        'maintenance_windows': analyze_maintenance_patterns(all_tools)
    }

def analyze_development_cycles(all_tools):
    """Analyze development cycle patterns"""
    dev_tools = {}
    for tool_name, tool_data in all_tools.items():
        if tool_data['category'] in ['development', 'source_control', 'file_operations']:
            dev_tools[tool_name] = {
                'usage': tool_data['stats'].get('total_executions', 0) + tool_data['stats'].get('total_calls', 0),
                'category': tool_data['category']
            }
    return dev_tools

def analyze_maintenance_patterns(all_tools):
    """Analyze maintenance patterns"""
    maintenance_tools = {}
    for tool_name, tool_data in all_tools.items():
        if tool_data['category'] in ['system_operations', 'process_management']:
            maintenance_tools[tool_name] = {
                'usage': tool_data['stats'].get('total_executions', 0) + tool_data['stats'].get('total_calls', 0),
                'success_rate': tool_data['stats'].get('success_rate', 0)
            }
    return maintenance_tools

def analyze_security_patterns(all_tools):
    """Analyze security-related patterns"""
    security_insights = {
        'privileged_operations': analyze_privileged_operations(all_tools),
        'network_operations': analyze_network_operations(all_tools),
        'file_access_patterns': analyze_file_access_patterns(all_tools),
        'security_recommendations': generate_security_recommendations(all_tools)
    }
    return security_insights

def analyze_privileged_operations(all_tools):
    """Analyze privileged operations"""
    privileged_tools = ['chmod', 'sudo', 'su', 'chown', 'systemctl']
    patterns = {}
    
    for tool_name, tool_data in all_tools.items():
        if any(priv in tool_name.lower() for priv in privileged_tools):
            patterns[tool_name] = {
                'usage_count': tool_data['stats'].get('total_executions', 0),
                'success_rate': tool_data['stats'].get('success_rate', 0),
                'risk_level': 'high' if tool_data['stats'].get('success_rate', 1) < 0.9 else 'medium'
            }
    return patterns

def analyze_network_operations(all_tools):
    """Analyze network operations"""
    network_tools = ['curl', 'wget', 'ssh', 'scp', 'rsync']
    patterns = {}
    
    for tool_name, tool_data in all_tools.items():
        if any(net in tool_name.lower() for net in network_tools):
            patterns[tool_name] = {
                'usage_count': tool_data['stats'].get('total_executions', 0),
                'success_rate': tool_data['stats'].get('success_rate', 0),
                'category': tool_data['category']
            }
    return patterns

def analyze_file_access_patterns(all_tools):
    """Analyze file access patterns"""
    file_tools = ['find', 'grep', 'cat', 'ls', 'rm', 'mv', 'cp']
    patterns = {}
    
    for tool_name, tool_data in all_tools.items():
        if any(file_tool in tool_name.lower() for file_tool in file_tools):
            patterns[tool_name] = {
                'usage_count': tool_data['stats'].get('total_executions', 0),
                'success_rate': tool_data['stats'].get('success_rate', 0),
                'performance': tool_data['stats'].get('average_execution_time', 0)
            }
    return patterns

def generate_security_recommendations(all_tools):
    """Generate security recommendations"""
    recommendations = []
    
    # Check for high-risk operations
    for tool_name, tool_data in all_tools.items():
        if tool_data['stats'].get('success_rate', 1) < 0.8 and 'rm' in tool_name.lower():
            recommendations.append({
                'type': 'high_risk_operation',
                'tool': tool_name,
                'issue': 'Low success rate for destructive operation',
                'recommendation': 'Review rm command usage and consider safer alternatives'
            })
    
    # Check for excessive network operations
    network_usage = sum(tool_data['stats'].get('total_executions', 0) 
                       for tool_name, tool_data in all_tools.items() 
                       if 'curl' in tool_name.lower() or 'wget' in tool_name.lower())
    
    if network_usage > 100:
        recommendations.append({
            'type': 'network_security',
            'issue': 'High network tool usage',
            'recommendation': 'Review network operations for potential security implications'
        })
    
    return recommendations

def main():
    """Main execution function"""
    # Load statistics
    bash_stats, mcp_stats = load_stats()
    
    # Perform comprehensive analysis
    analysis = analyze_tool_ecosystem(bash_stats, mcp_stats)
    
    # Save combined analysis
    with open(COMBINED_ANALYSIS_FILE, 'w') as f:
        json.dump(analysis, f, indent=2)
    
    # Save adaptive recommendations
    with open(ADAPTIVE_RECOMMENDATIONS_FILE, 'w') as f:
        json.dump(analysis['recommendations'], f, indent=2)
    
    # Print summary to stderr for immediate feedback
    print("🔍 Tool Reliability Analysis Complete", file=sys.stderr)
    print(f"📊 Analyzed {analysis['ecosystem_health']['total_tools']} tools", file=sys.stderr)
    print(f"🏆 Health Score: {analysis['ecosystem_health']['health_score']:.1f}%", file=sys.stderr)
    
    # Print critical issues
    critical_tools = len(analysis['reliability_tiers']['tier_5_critical'])
    if critical_tools > 0:
        print(f"⚠️  {critical_tools} tools need immediate attention", file=sys.stderr)
    
    # Print immediate actions
    immediate_actions = len(analysis['recommendations']['immediate_actions'])
    if immediate_actions > 0:
        print(f"🔧 {immediate_actions} immediate actions recommended", file=sys.stderr)

if __name__ == '__main__':
    main()