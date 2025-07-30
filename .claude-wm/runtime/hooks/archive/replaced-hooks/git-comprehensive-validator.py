#!/usr/bin/env python3
"""
Git Comprehensive Validator - Unified git validation system
Combines safety checks, commit validation, and gitignore enforcement
"""

import json
import os
import re
import subprocess
import sys
from pathlib import Path


REQUIRED_GITIGNORE_PATTERNS = {
    'environment': ['.env', '.env.*', '!.env.example', '!.env.sample', '!.env.template'],
    'secrets': ['*.pem', '*.key', '*.cert', '*.p12', '*.pfx', 'secrets.yml', 'secrets.json', 'credentials.json', 'service-account*.json'],
    'ide': ['.vscode/', '.idea/', '*.swp', '*.swo', '*~', '.DS_Store', 'Thumbs.db'],
    'dependencies': ['node_modules/', 'vendor/', 'venv/', 'env/', '__pycache__/', '*.pyc', '.pytest_cache/'],
    'build': ['dist/', 'build/', 'out/', '*.log', '*.tmp', '.next/', '.nuxt/', '.cache/'],
    'test': ['coverage/', '.nyc_output/', '*.test.log', 'test-results/', 'playwright-report/', 'test-artifacts/'],
    'database': ['*.db', '*.sqlite', '*.sqlite3', 'database.yml']
}

FORBIDDEN_FILES = {
    'private_keys': [r'.*\.pem$', r'.*\.key$', r'.*private.*key.*', r'id_rsa.*', r'id_dsa.*', r'id_ecdsa.*', r'id_ed25519.*'],
    'env_files': [r'^\.env$', r'^\.env\.[^.]+$', r'.*\.env\.(?!example|sample|template).*$'],
    'credentials': [r'.*credentials.*\.(json|yml|yaml)$', r'.*service[-_]?account.*\.json$', r'.*secrets?\.(json|yml|yaml|txt)$', r'.*password.*\.(txt|json|yml|yaml)$'],
    'test_scripts': [r'.*test[-_]?script.*\.(sh|bash|py|js)$', r'.*scratch.*\.(py|js|ts|sh)$', r'.*temp[-_]?test.*', r'.*debug[-_]?script.*'],
    'backups': [r'.*\.backup$', r'.*\.bak$', r'.*\.old$', r'.*~$', r'.*\.(orig|save)$'],
    'archives': [r'.*\.(zip|tar|tar\.gz|tgz|rar|7z)$'],
    'large_files': [r'.*\.(mp4|avi|mov|mkv|wmv)$', r'.*\.(psd|ai|sketch|fig)$', r'.*\.(exe|dmg|pkg|deb|rpm)$']
}

WARNING_FILES = {
    'configs': [r'config\.(json|yml|yaml)$', r'settings\.(json|yml|yaml)$'],
    'data': [r'.*\.(csv|xlsx|xls)$', r'.*\.sql$', r'.*dump.*'],
    'logs': [r'.*\.log$', r'debug\.txt$', r'error\.txt$']
}


class GitValidator:
    def __init__(self):
        self.errors = []
        self.warnings = []
        self.repo_root = None
        self.current_dir = None

    def validate_repository_context(self):
        """Validate git repository context and location"""
        try:
            # Check if we're in a Git repository
            result = subprocess.run(['git', 'rev-parse', '--git-dir'], 
                                  capture_output=True, text=True)
            if result.returncode != 0:
                self.errors.append("Not in a Git repository")
                return False
            
            # Get repository root and current directory
            result = subprocess.run(['git', 'rev-parse', '--show-toplevel'], 
                                  capture_output=True, text=True)
            self.repo_root = result.stdout.strip() if result.returncode == 0 else None
            self.current_dir = os.getcwd()
            
            # Warn if not at repository root
            if self.repo_root and self.current_dir != self.repo_root:
                self.warnings.append(f"Not at repository root. Root: {self.repo_root}, Current: {self.current_dir}")
            
            return True
            
        except Exception as e:
            self.errors.append(f"Repository validation failed: {e}")
            return False

    def validate_staged_files(self):
        """Validate staged files for security and size"""
        try:
            # Get staged files
            result = subprocess.run(['git', 'diff', '--cached', '--name-only'], 
                                  capture_output=True, text=True)
            staged_files = result.stdout.strip().split('\n') if result.stdout else []
            
            if not staged_files or staged_files == ['']:
                return True
            
            # Check for forbidden files
            forbidden_patterns = [
                r'^\.git/', r'^\.claude/', r'\.log$', r'node_modules/', r'\.env$', r'\.DS_Store$'
            ]
            
            forbidden_files = []
            for file_path in staged_files:
                for pattern in forbidden_patterns:
                    if re.match(pattern, file_path):
                        forbidden_files.append(file_path)
                        break
            
            if forbidden_files:
                self.errors.append("Forbidden files detected in staging:")
                for file_path in forbidden_files:
                    self.errors.append(f"  - {file_path}")
                self.errors.append("Use 'git reset HEAD <file>' to unstage")
                return False
            
            # Check file sizes
            large_files = []
            for file_path in staged_files:
                if os.path.exists(file_path):
                    size = os.path.getsize(file_path)
                    if size > 10 * 1024 * 1024:  # 10MB
                        large_files.append((file_path, size))
            
            if large_files:
                self.warnings.append("Large files detected (>10MB):")
                for file_path, size in large_files:
                    self.warnings.append(f"  - {file_path} ({size / (1024*1024):.1f}MB)")
                self.warnings.append("Consider Git LFS for large files")
            
            # Check number of staged files
            if len(staged_files) > 50:
                self.warnings.append(f"Many staged files ({len(staged_files)}). Consider more atomic commits")
            
            return True
            
        except Exception as e:
            self.errors.append(f"Staged files validation failed: {e}")
            return False

    def validate_gitignore(self):
        """Validate .gitignore file existence and content"""
        try:
            gitignore_path = Path('.gitignore')
            
            if not gitignore_path.exists():
                self.errors.append("No .gitignore file found")
                self.errors.append("Create .gitignore with basic security patterns")
                return False
            
            # Parse existing .gitignore
            gitignore_patterns = []
            with open(gitignore_path, 'r') as f:
                for line in f:
                    line = line.strip()
                    if line and not line.startswith('#'):
                        gitignore_patterns.append(line)
            
            # Check for missing critical patterns
            critical_patterns = ['.git/', '.claude/', 'node_modules/', '.env']
            missing_patterns = []
            
            for pattern in critical_patterns:
                found = False
                for gitignore_pattern in gitignore_patterns:
                    if pattern in gitignore_pattern or gitignore_pattern.startswith(pattern.rstrip('/')):
                        found = True
                        break
                if not found:
                    missing_patterns.append(pattern)
            
            if missing_patterns:
                self.warnings.append("Missing critical .gitignore patterns:")
                for pattern in missing_patterns:
                    self.warnings.append(f"  - {pattern}")
            
            return True
            
        except Exception as e:
            self.errors.append(f"Gitignore validation failed: {e}")
            return False

    def validate_commit_message(self, message):
        """Validate commit message format and content"""
        if not message:
            return True  # Skip if no message provided
        
        # Check for Co-Authored-By (blocked per project rules)
        if "Co-Authored-By" in message or "co-authored-by" in message.lower():
            self.errors.append("Co-authored commits are not allowed per project rules")
        
        # Check for Claude signature (should be removed)
        if "🤖 Generated with [Claude Code]" in message or "🤖 Generated with Claude" in message:
            self.errors.append("Remove Claude signature from commit messages")
        
        # Extract main message (first line)
        lines = message.strip().split('\n')
        if not lines:
            self.errors.append("Empty commit message")
            return False
        
        main_message = lines[0].strip()
        
        # Check message length
        if len(main_message) > 72:
            self.warnings.append(f"First line should be ≤72 characters (current: {len(main_message)})")
        elif len(main_message) < 10:
            self.errors.append("Commit message too short (minimum 10 characters)")
        
        # Check for conventional commit format
        conventional_pattern = r'^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?: .+'
        if re.match(conventional_pattern, main_message):
            if not re.match(r'^[a-z]+(\(.+\))?: [a-z]', main_message):
                self.warnings.append("Conventional commits should start with lowercase after type")
        else:
            if main_message and not main_message[0].isupper():
                self.warnings.append("Commit message should start with capital letter")
        
        # Check for imperative mood
        past_tense_patterns = [
            r'\b(added|deleted|changed|fixed|updated|removed|created|modified)\b',
            r'\b(implemented|refactored|improved|optimized)\b'
        ]
        for pattern in past_tense_patterns:
            if re.search(pattern, main_message, re.IGNORECASE):
                self.warnings.append("Use imperative mood (e.g., 'Add' not 'Added')")
                break
        
        # Check body formatting
        if len(lines) > 1:
            if len(lines) > 1 and lines[1].strip() != "":
                self.errors.append("Add blank line after commit message summary")
            
            for i, line in enumerate(lines[2:], start=3):
                if len(line) > 72 and not line.startswith('http'):
                    self.warnings.append(f"Line {i} exceeds 72 characters")
        
        return True

    def validate_forbidden_files(self, files):
        """Check for forbidden files in commit"""
        issues = []
        
        for file_path in files:
            if not os.path.exists(file_path):
                continue
            
            # Check against forbidden patterns
            for category, patterns in FORBIDDEN_FILES.items():
                for pattern in patterns:
                    if re.match(pattern, file_path, re.IGNORECASE):
                        issues.append({
                            'file': file_path,
                            'category': category,
                            'severity': 'high'
                        })
                        break
            
            # Check against warning patterns
            for category, patterns in WARNING_FILES.items():
                for pattern in patterns:
                    if re.match(pattern, file_path, re.IGNORECASE):
                        if not any(issue['file'] == file_path for issue in issues):
                            issues.append({
                                'file': file_path,
                                'category': category,
                                'severity': 'medium'
                            })
                        break
        
        # Process issues
        high_severity = [i for i in issues if i['severity'] == 'high']
        medium_severity = [i for i in issues if i['severity'] == 'medium']
        
        if high_severity:
            self.errors.append("Forbidden files detected:")
            for issue in high_severity:
                self.errors.append(f"  - {issue['file']} ({issue['category'].replace('_', ' ')})")
        
        if medium_severity:
            self.warnings.append("Warning files detected:")
            for issue in medium_severity:
                self.warnings.append(f"  - {issue['file']} ({issue['category'].replace('_', ' ')})")
        
        return len(high_severity) == 0

    def extract_commit_message_from_command(self, command):
        """Extract commit message from git commit command"""
        patterns = [
            r'-m\s+"([^"]+)"',  # -m "message"
            r"-m\s+'([^']+)'",  # -m 'message'
            r'-m\s+([^\s]+)',   # -m message
            r'--message="([^"]+)"',  # --message="message"
            r"--message='([^']+)'",  # --message='message'
            r'-m\s+"\$\(cat\s+<<[\'"]?EOF[\'"]?\n(.*?)\nEOF\s*\)"',  # heredoc
            r"-m\s+'\$\(cat\s+<<['\"]?EOF['\"]?\n(.*?)\nEOF\s*\)'"
        ]
        
        for pattern in patterns:
            match = re.search(pattern, command, re.DOTALL)
            if match:
                return match.group(1)
        
        return None

    def run_full_validation(self, tool_name, tool_input):
        """Run comprehensive git validation"""
        # Always validate repository context
        if not self.validate_repository_context():
            return False
        
        # Validate based on tool and command
        if tool_name == 'Bash':
            command = tool_input.get('command', '')
            
            # Git commit validation
            if 'git commit' in command:
                # Skip amend with no-edit
                if '--amend' in command and '--no-edit' in command:
                    return True
                
                # Validate commit message
                commit_message = self.extract_commit_message_from_command(command)
                if commit_message:
                    self.validate_commit_message(commit_message)
                
                # Validate staged files
                self.validate_staged_files()
                
                # Get staged files for forbidden check
                try:
                    result = subprocess.run(['git', 'diff', '--cached', '--name-only'], 
                                          capture_output=True, text=True)
                    staged_files = result.stdout.strip().split('\n') if result.stdout else []
                    self.validate_forbidden_files(staged_files)
                except:
                    pass
            
            # Git add validation
            elif 'git add' in command:
                self.validate_staged_files()
                self.validate_gitignore()
        
        elif tool_name == 'Write':
            # Check if creating potentially sensitive files
            file_path = tool_input.get('file_path', '')
            if file_path:
                self.validate_forbidden_files([file_path])
        
        # Return True if no errors (warnings are okay)
        return len(self.errors) == 0

    def print_results(self):
        """Print validation results"""
        if self.errors:
            print("\n🚨 Git Validation Errors:", file=sys.stderr)
            for error in self.errors:
                print(f"❌ {error}", file=sys.stderr)
        
        if self.warnings:
            print("\n⚠️  Git Validation Warnings:", file=sys.stderr)
            for warning in self.warnings:
                print(f"⚠️  {warning}", file=sys.stderr)
        
        if self.errors:
            print("\n❌ Git operation blocked due to validation errors", file=sys.stderr)
            print("Please fix the errors above and try again.", file=sys.stderr)
        elif self.warnings:
            print("\nProceeding with warnings...", file=sys.stderr)


def main():
    """Main entry point"""
    try:
        # Read input
        input_data = json.load(sys.stdin)
        tool_name = input_data.get('tool_name', '')
        tool_input = input_data.get('tool_input', {})
        
        # Initialize validator
        validator = GitValidator()
        
        # Run validation
        success = validator.run_full_validation(tool_name, tool_input)
        
        # Print results
        validator.print_results()
        
        # Exit with appropriate code
        sys.exit(0 if success else 2)
        
    except json.JSONDecodeError:
        print("Error: Invalid JSON input", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error in git comprehensive validator: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()