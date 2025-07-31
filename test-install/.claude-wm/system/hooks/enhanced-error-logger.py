#!/usr/bin/env python3
"""
Enhanced Error Logger - Advanced error pattern tracking and analysis
"""
import json
import sys
import re
import hashlib
from datetime import datetime
from pathlib import Path
from collections import defaultdict

HOOKS_DIR = Path("/Users/a.pezzotta/.claude/hooks")
ERROR_PATTERNS_FILE = HOOKS_DIR / "logs" / "error-patterns.json"
ERROR_ANALYSIS_FILE = HOOKS_DIR / "logs" / "error-analysis.json"

def extract_error_patterns(error_text):
    """Extract structured error patterns from error text"""
    patterns = {
        'error_type': identify_error_type(error_text),
        'error_category': categorize_error(error_text),
        'severity': assess_severity(error_text),
        'actionable_info': extract_actionable_info(error_text),
        'stack_trace': extract_stack_trace(error_text),
        'exit_code': extract_exit_code(error_text),
        'timeout_indicator': detect_timeout(error_text),
        'permission_issue': detect_permission_issue(error_text),
        'resource_issue': detect_resource_issue(error_text),
        'network_issue': detect_network_issue(error_text)
    }
    return patterns

def identify_error_type(error_text):
    """Identify the type of error"""
    error_types = {
        'syntax_error': r'syntax error|parse error|unexpected token',
        'permission_error': r'permission denied|access denied|forbidden',
        'file_not_found': r'no such file|file not found|cannot find',
        'network_error': r'connection refused|timeout|network unreachable',
        'command_not_found': r'command not found|not recognized',
        'memory_error': r'out of memory|memory exhausted|killed',
        'disk_space': r'no space left|disk full|quota exceeded',
        'dependency_error': r'module not found|import error|missing dependency',
        'timeout_error': r'timeout|timed out|time limit exceeded',
        'authentication_error': r'authentication failed|invalid credentials|unauthorized'
    }
    
    for error_type, pattern in error_types.items():
        if re.search(pattern, error_text.lower()):
            return error_type
    
    return 'unknown'

def categorize_error(error_text):
    """Categorize error by domain"""
    categories = {
        'system': ['permission', 'disk', 'memory', 'process'],
        'network': ['connection', 'timeout', 'dns', 'ssl'],
        'development': ['syntax', 'compile', 'dependency', 'import'],
        'configuration': ['config', 'settings', 'environment'],
        'security': ['authentication', 'authorization', 'certificate']
    }
    
    error_lower = error_text.lower()
    for category, keywords in categories.items():
        if any(keyword in error_lower for keyword in keywords):
            return category
    
    return 'general'

def assess_severity(error_text):
    """Assess error severity"""
    critical_patterns = ['fatal', 'critical', 'segmentation fault', 'core dump']
    high_patterns = ['error', 'failed', 'exception', 'abort']
    medium_patterns = ['warning', 'deprecated', 'invalid']
    low_patterns = ['info', 'debug', 'notice']
    
    error_lower = error_text.lower()
    
    if any(pattern in error_lower for pattern in critical_patterns):
        return 'critical'
    elif any(pattern in error_lower for pattern in high_patterns):
        return 'high'
    elif any(pattern in error_lower for pattern in medium_patterns):
        return 'medium'
    elif any(pattern in error_lower for pattern in low_patterns):
        return 'low'
    else:
        return 'medium'

def extract_actionable_info(error_text):
    """Extract actionable information from error"""
    actionable_patterns = {
        'file_path': r'(?:/[^\s]+/[^\s]+)',
        'line_number': r'line (\d+)',
        'column_number': r'column (\d+)',
        'missing_package': r'(?:module|package) [\'"]([^\'"]+)[\'"] not found',
        'invalid_option': r'invalid option [\'"]([^\'"]+)[\'"]',
        'required_parameter': r'missing required (?:parameter|argument) [\'"]([^\'"]+)[\'"]'
    }
    
    actionable = {}
    for info_type, pattern in actionable_patterns.items():
        match = re.search(pattern, error_text)
        if match:
            actionable[info_type] = match.group(1) if match.groups() else match.group(0)
    
    return actionable

def extract_stack_trace(error_text):
    """Extract stack trace information"""
    stack_patterns = [
        r'Traceback \(most recent call last\):',
        r'at .+\(.+:\d+:\d+\)',
        r'    at .+',
        r'^\s*File ".+", line \d+, in .+$'
    ]
    
    for pattern in stack_patterns:
        if re.search(pattern, error_text, re.MULTILINE):
            return True
    return False

def extract_exit_code(error_text):
    """Extract exit code from error"""
    exit_patterns = [
        r'exit code (\d+)',
        r'returned (\d+)',
        r'status (\d+)'
    ]
    
    for pattern in exit_patterns:
        match = re.search(pattern, error_text.lower())
        if match:
            return int(match.group(1))
    
    return None

def detect_timeout(error_text):
    """Detect timeout-related errors"""
    timeout_patterns = [
        r'timeout',
        r'timed out',
        r'time limit exceeded',
        r'operation timed out'
    ]
    
    return any(re.search(pattern, error_text.lower()) for pattern in timeout_patterns)

def detect_permission_issue(error_text):
    """Detect permission-related errors"""
    permission_patterns = [
        r'permission denied',
        r'access denied',
        r'forbidden',
        r'not authorized',
        r'insufficient privileges'
    ]
    
    return any(re.search(pattern, error_text.lower()) for pattern in permission_patterns)

def detect_resource_issue(error_text):
    """Detect resource-related errors"""
    resource_patterns = [
        r'out of memory',
        r'no space left',
        r'disk full',
        r'quota exceeded',
        r'too many open files'
    ]
    
    return any(re.search(pattern, error_text.lower()) for pattern in resource_patterns)

def detect_network_issue(error_text):
    """Detect network-related errors"""
    network_patterns = [
        r'connection refused',
        r'network unreachable',
        r'dns resolution failed',
        r'ssl certificate',
        r'connection timeout'
    ]
    
    return any(re.search(pattern, error_text.lower()) for pattern in network_patterns)

def generate_error_signature(error_patterns):
    """Generate a unique signature for similar errors"""
    key_elements = [
        error_patterns['error_type'],
        error_patterns['error_category'],
        error_patterns['severity']
    ]
    
    # Add actionable info if available
    if error_patterns['actionable_info']:
        key_elements.extend(error_patterns['actionable_info'].values())
    
    signature_string = '|'.join(str(e) for e in key_elements)
    return hashlib.md5(signature_string.encode()).hexdigest()[:8]

def load_error_patterns():
    """Load existing error patterns"""
    if ERROR_PATTERNS_FILE.exists():
        try:
            with open(ERROR_PATTERNS_FILE, 'r') as f:
                return json.load(f)
        except:
            pass
    return {}

def save_error_patterns(patterns):
    """Save error patterns to file"""
    try:
        ERROR_PATTERNS_FILE.parent.mkdir(parents=True, exist_ok=True)
        with open(ERROR_PATTERNS_FILE, 'w') as f:
            json.dump(patterns, f, indent=2)
    except Exception as e:
        print(f"Warning: Could not save error patterns: {e}", file=sys.stderr)

def analyze_error_trends():
    """Analyze error trends and generate insights"""
    patterns = load_error_patterns()
    
    if not patterns:
        return
    
    analysis = {
        'total_errors': len(patterns),
        'error_types': defaultdict(int),
        'error_categories': defaultdict(int),
        'severity_distribution': defaultdict(int),
        'recurring_errors': [],
        'actionable_insights': [],
        'generated_at': datetime.now().isoformat()
    }
    
    # Analyze patterns
    for error_id, error_data in patterns.items():
        analysis['error_types'][error_data['error_type']] += 1
        analysis['error_categories'][error_data['error_category']] += 1
        analysis['severity_distribution'][error_data['severity']] += 1
        
        # Check for recurring errors
        if error_data['occurrence_count'] > 3:
            analysis['recurring_errors'].append({
                'signature': error_data['signature'],
                'error_type': error_data['error_type'],
                'count': error_data['occurrence_count'],
                'last_seen': error_data['last_seen']
            })
    
    # Generate actionable insights
    if analysis['error_types']['permission_error'] > 5:
        analysis['actionable_insights'].append({
            'insight': 'High number of permission errors detected',
            'recommendation': 'Review file permissions and user access rights'
        })
    
    if analysis['error_types']['timeout_error'] > 3:
        analysis['actionable_insights'].append({
            'insight': 'Multiple timeout errors detected',
            'recommendation': 'Consider increasing timeout values or optimizing operations'
        })
    
    # Save analysis
    try:
        with open(ERROR_ANALYSIS_FILE, 'w') as f:
            json.dump(analysis, f, indent=2)
    except Exception as e:
        print(f"Warning: Could not save error analysis: {e}", file=sys.stderr)

def main():
    """Main function to process error from tool output"""
    try:
        input_data = json.load(sys.stdin)
        
        # Check if this is a tool result with an error
        tool_result = input_data.get('tool_result', {})
        if not tool_result.get('error'):
            sys.exit(0)
        
        tool_name = input_data.get('tool_name', 'unknown')
        error_text = tool_result.get('error', '')
        
        # Extract error patterns
        patterns = extract_error_patterns(error_text)
        signature = generate_error_signature(patterns)
        
        # Load existing patterns
        all_patterns = load_error_patterns()
        
        # Update or create pattern entry
        if signature in all_patterns:
            all_patterns[signature]['occurrence_count'] += 1
            all_patterns[signature]['last_seen'] = datetime.now().isoformat()
            all_patterns[signature]['affected_tools'].add(tool_name)
        else:
            all_patterns[signature] = {
                'signature': signature,
                'error_type': patterns['error_type'],
                'error_category': patterns['error_category'],
                'severity': patterns['severity'],
                'patterns': patterns,
                'occurrence_count': 1,
                'first_seen': datetime.now().isoformat(),
                'last_seen': datetime.now().isoformat(),
                'affected_tools': {tool_name},
                'original_error': error_text[:500]  # Truncate for storage
            }
        
        # Convert sets to lists for JSON serialization
        for pattern_data in all_patterns.values():
            if isinstance(pattern_data.get('affected_tools'), set):
                pattern_data['affected_tools'] = list(pattern_data['affected_tools'])
        
        # Save updated patterns
        save_error_patterns(all_patterns)
        
        # Analyze trends
        analyze_error_trends()
        
        # Print summary for immediate feedback
        print(f"📊 Error pattern logged: {patterns['error_type']} ({signature})", file=sys.stderr)
        if patterns['severity'] == 'critical':
            print(f"🚨 Critical error detected in {tool_name}", file=sys.stderr)
        
        sys.exit(0)
        
    except json.JSONDecodeError:
        print("Error: Invalid JSON input", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Warning: Error logging failed: {e}", file=sys.stderr)
        sys.exit(0)

if __name__ == "__main__":
    main()