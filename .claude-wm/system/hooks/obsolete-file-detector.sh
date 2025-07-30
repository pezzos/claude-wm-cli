#!/bin/bash

# Obsolete File Detector Hook
# Implements "Flag obsolete files for removal" from CLAUDE.md Core Coding Principles
# Automatically detects and flags files that may be obsolete

# Function to check if file is potentially obsolete
is_potentially_obsolete() {
    local file_path="$1"
    local file_name=$(basename "$file_path")
    local file_dir=$(dirname "$file_path")
    
    # Skip if file doesn't exist
    [[ ! -f "$file_path" ]] && return 1
    
    # Common obsolete file patterns
    local obsolete_patterns=(
        "*.bak"
        "*.old"
        "*.orig"
        "*.tmp"
        "*~"
        "*.backup"
        "*.save"
        "*-old.*"
        "*_old.*"
        "*-backup.*"
        "*_backup.*"
        "*.deprecated"
        "*-deprecated.*"
        "*_deprecated.*"
        "*.legacy"
        "*-legacy.*"
        "*_legacy.*"
        "copy-*"
        "*-copy.*"
        "*_copy.*"
        "test-*"
        "temp-*"
        "draft-*"
        "unused-*"
    )
    
    # Check against patterns
    for pattern in "${obsolete_patterns[@]}"; do
        if [[ "$file_name" == $pattern ]]; then
            return 0
        fi
    done
    
    # Check for duplicate files with similar names
    local base_name="${file_name%.*}"
    local extension="${file_name##*.}"
    
    # Look for files with similar names
    local similar_files=$(find "$file_dir" -maxdepth 1 -name "${base_name}*" 2>/dev/null | wc -l)
    if [[ $similar_files -gt 1 ]]; then
        # Check if this one looks like a backup/old version
        if [[ "$file_name" =~ (old|backup|bak|orig|copy|deprecated|legacy)$ ]]; then
            return 0
        fi
    fi
    
    return 1
}

# Function to check file usage
check_file_usage() {
    local file_path="$1"
    local file_name=$(basename "$file_path")
    local base_name="${file_name%.*}"
    
    # Skip binary files and common non-source files
    if file "$file_path" | grep -qi "binary\|executable"; then
        return 1
    fi
    
    # For source files, check if they're referenced anywhere
    local references=0
    
    # Common source file extensions to check
    if [[ "$file_name" =~ \.(js|ts|jsx|tsx|py|java|go|rs|cpp|c|h|php|rb|cs)$ ]]; then
        # Search for imports/includes of this file
        references=$(grep -r --include="*.js" --include="*.ts" --include="*.jsx" --include="*.tsx" --include="*.py" --include="*.java" --include="*.go" --include="*.rs" --include="*.cpp" --include="*.c" --include="*.h" --include="*.php" --include="*.rb" --include="*.cs" -l "$base_name" . 2>/dev/null | grep -v "$file_path" | wc -l)
    fi
    
    return $references
}

# Function to get file age and last modification
get_file_age_info() {
    local file_path="$1"
    
    # Get last modification time (days ago)
    local mod_time=$(stat -f %m "$file_path" 2>/dev/null || stat -c %Y "$file_path" 2>/dev/null)
    local current_time=$(date +%s)
    local days_ago=$(( (current_time - mod_time) / 86400 ))
    
    echo "$days_ago"
}

# Function to suggest file removal
suggest_removal() {
    local file_path="$1"
    local reason="$2"
    local age="$3"
    
    echo "🗑️  OBSOLETE FILE DETECTED: $file_path"
    echo "   📅 Last modified: $age days ago"
    echo "   💡 Reason: $reason"
    echo "   🔧 Suggested action: rm '$file_path'"
    echo ""
    
    # Log the suggestion
    local log_file="$HOME/.claude/hooks/logs/obsolete-files.log"
    mkdir -p "$(dirname "$log_file")"
    echo "$(date '+%Y-%m-%d %H:%M:%S') - OBSOLETE: $file_path ($reason, $age days old)" >> "$log_file"
}

# Function to scan for obsolete files
scan_for_obsolete_files() {
    local files_found=false
    
    # Find all files in current directory and subdirectories (excluding .git, node_modules, etc.)
    while IFS= read -r -d '' file; do
        # Skip hidden directories and common ignore patterns
        if [[ "$file" =~ /(\.git|node_modules|__pycache__|\.venv|venv|env|dist|build|target)/ ]]; then
            continue
        fi
        
        local age=$(get_file_age_info "$file")
        
        # Check if file is potentially obsolete
        if is_potentially_obsolete "$file"; then
            suggest_removal "$file" "Matches obsolete file pattern" "$age"
            files_found=true
        elif [[ $age -gt 180 ]]; then  # Files older than 6 months
            # Check if file is unused
            check_file_usage "$file"
            local usage_count=$?
            
            if [[ $usage_count -eq 0 && $age -gt 180 ]]; then
                suggest_removal "$file" "No references found, very old" "$age"
                files_found=true
            fi
        fi
        
    done < <(find . -type f ! -path "./.git/*" ! -path "./node_modules/*" ! -path "./__pycache__/*" ! -path "./.venv/*" ! -path "./venv/*" ! -path "./env/*" ! -path "./dist/*" ! -path "./build/*" ! -path "./target/*" -print0 2>/dev/null)
    
    if [[ "$files_found" == true ]]; then
        echo "💡 Obsolete File Cleanup Tips:"
        echo "   • Review suggested files before removal"
        echo "   • Move to .archive/ folder instead of deleting"
        echo "   • Update .gitignore to prevent similar files"
        echo "   • Consider setting up pre-commit hooks to catch these"
        echo ""
    else
        echo "✅ No obsolete files detected in current project"
    fi
}

# Main execution
main() {
    # Create logs directory
    mkdir -p "$HOME/.claude/hooks/logs"
    
    # Only run in directories that look like projects
    if [[ -f "package.json" || -f "requirements.txt" || -f "go.mod" || -f "Cargo.toml" || -f ".git/config" ]]; then
        echo "🔍 Scanning for obsolete files..."
        scan_for_obsolete_files
    fi
    
    exit 0
}

# Run the scanner
main