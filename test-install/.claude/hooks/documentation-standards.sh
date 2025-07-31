#!/bin/bash

# Documentation Standards Validation Hook
# Enforces documentation standards from CLAUDE.md automatically
# Validates format, structure, and quality of documentation files

# Read input data
INPUT=$(cat 2>/dev/null || echo '{}')

# Function to validate documentation file
validate_doc_file() {
    local file_path="$1"
    local issues=()
    
    if [[ ! -f "$file_path" ]]; then
        return 0
    fi
    
    # Check if it's a documentation file
    if [[ ! "$file_path" =~ \.(md|rst|txt)$ ]]; then
        return 0
    fi
    
    # Read file content
    local content=$(cat "$file_path")
    
    # Check for hierarchical headers
    if echo "$content" | grep -q "^#"; then
        # Check for proper header hierarchy
        local header_levels=$(echo "$content" | grep "^#" | sed 's/\(#*\).*/\1/' | sort -u)
        if ! echo "$header_levels" | head -1 | grep -q "^#$"; then
            if echo "$content" | grep -q "^##"; then
                issues+=("Missing top-level header (#)")
            fi
        fi
    fi
    
    # Check line length (≤ 100 chars)
    local long_lines=$(awk 'length > 100 {print NR}' "$file_path" | head -5)
    if [[ -n "$long_lines" ]]; then
        issues+=("Lines exceed 100 characters: $(echo $long_lines | tr '\n' ', ')")
    fi
    
    # Check for "Last Updated" or version info
    if [[ "$file_path" =~ README\.md$ ]] || [[ "$file_path" =~ CHANGELOG\.md$ ]]; then
        if ! echo "$content" | grep -qi "last updated\|version\|date"; then
            issues+=("Missing 'Last Updated' date or version info")
        fi
    fi
    
    # Check for code blocks with language hints
    local code_blocks=$(echo "$content" | grep -n "^```" | grep -v "^```[a-z]")
    if [[ -n "$code_blocks" ]]; then
        issues+=("Code blocks without language hints found")
    fi
    
    # Check for practical examples in documentation
    if [[ "$file_path" =~ (README|GUIDE|DOCS)\.md$ ]]; then
        if ! echo "$content" | grep -qi "example\|usage\|how to"; then
            issues+=("Missing practical examples or usage instructions")
        fi
    fi
    
    # Language check for technical documentation
    if echo "$content" | grep -qi "bonjour\|salut\|merci\|français\|voici\|voilà"; then
        issues+=("French text detected - technical documentation must be in English")
    fi
    
    # Report issues
    if [[ ${#issues[@]} -gt 0 ]]; then
        echo "📝 Documentation Issues in $file_path:"
        for issue in "${issues[@]}"; do
            echo "   ❌ $issue"
        done
        echo ""
        
        # Log the issues
        local log_file="$HOME/.claude/hooks/logs/documentation-issues.log"
        mkdir -p "$(dirname "$log_file")"
        echo "$(date '+%Y-%m-%d %H:%M:%S') - $file_path: ${issues[*]}" >> "$log_file"
        
        return 1
    fi
    
    return 0
}

# Function to check documentation standards across project
check_documentation_standards() {
    local has_issues=false
    
    # Find all documentation files
    local doc_files=$(find . -maxdepth 3 -name "*.md" -o -name "*.rst" -o -name "*.txt" | grep -v node_modules | grep -v .git)
    
    for file in $doc_files; do
        if ! validate_doc_file "$file"; then
            has_issues=true
        fi
    done
    
    # Check for missing essential documentation
    if [[ ! -f "README.md" ]] && [[ -f "package.json" || -f "requirements.txt" || -f "go.mod" || -f "Cargo.toml" ]]; then
        echo "📝 Missing README.md for project"
        has_issues=true
    fi
    
    # Suggestions for improvement
    if [[ "$has_issues" == true ]]; then
        echo ""
        echo "💡 Documentation Standards:"
        echo "   • Use clear hierarchical headers (##, ###, ####)"
        echo "   • Keep line length ≤ 100 chars for readability"
        echo "   • Include 'Last Updated' date and version"
        echo "   • Use code blocks with language hints"
        echo "   • Include practical examples"
        echo "   • Write all technical docs in English"
        echo ""
        echo "🔧 Auto-fix available: Run documentation formatter hook"
    fi
    
    return 0  # Don't block, just warn
}

# Main execution
main() {
    # Create logs directory
    mkdir -p "$HOME/.claude/hooks/logs"
    
    # Extract file path from input if it's a specific file operation
    local target_file=""
    if echo "$INPUT" | grep -q '"file_path"'; then
        target_file=$(echo "$INPUT" | grep -o '"file_path"[^"]*"[^"]*"' | cut -d'"' -f4)
        
        if [[ -n "$target_file" && "$target_file" =~ \.(md|rst|txt)$ ]]; then
            validate_doc_file "$target_file"
        fi
    else
        # Check all documentation in project
        check_documentation_standards
    fi
    
    exit 0
}

# Run the validation
main