#!/bin/bash

# "The Protocol: Before You Say Fixed" Validation Hook
# Enforces the testing protocol from CLAUDE.md automatically
# Prevents claims of "fixed" without proper verification

# Read input data
INPUT=$(cat 2>/dev/null || echo '{}')

# Function to check if this is a "fix" claim
is_fix_claim() {
    local content="$1"
    
    # Check for common "fixed" phrases
    if echo "$content" | grep -qi "fixed\|this should work\|try it now\|i've made the necessary changes\|the logic is correct"; then
        return 0
    fi
    return 1
}

# Function to check if proper testing was done
check_testing_evidence() {
    local content="$1"
    
    # Look for evidence of testing
    if echo "$content" | grep -qi "tested\|verified\|ran\|executed\|confirmed\|checked"; then
        return 0
    fi
    return 1
}

# Function to suggest minimum viable test
suggest_minimum_test() {
    local file_type="$1"
    
    case "$file_type" in
        *".js"|*".ts"|*".jsx"|*".tsx")
            echo "• Run: npm test or yarn test"
            echo "• For UI changes: Actually click the button/form"
            echo "• For API changes: Test with curl or Postman"
            ;;
        *".py")
            echo "• Run: python -m pytest or python -m unittest"
            echo "• For Flask/Django: Test the actual endpoint"
            echo "• For data changes: Query to verify state"
            ;;
        *".go")
            echo "• Run: go test ./..."
            echo "• For API changes: Test with curl"
            ;;
        *".java")
            echo "• Run: mvn test or gradle test"
            echo "• For Spring: Test actual endpoints"
            ;;
        *)
            echo "• Build/compile the project"
            echo "• Run any available test suite"
            echo "• Manually verify the specific feature"
            ;;
    esac
}

# Function to log protocol violation
log_violation() {
    local violation_type="$1"
    local context="$2"
    
    local log_file="$HOME/.claude/hooks/logs/protocol-violations.log"
    mkdir -p "$(dirname "$log_file")"
    
    echo "$(date '+%Y-%m-%d %H:%M:%S') - PROTOCOL VIOLATION: $violation_type - $context" >> "$log_file"
}

# Main validation logic
main() {
    # Extract file paths from input
    local files_changed=""
    if echo "$INPUT" | grep -q "file_path"; then
        files_changed=$(echo "$INPUT" | grep -o '"file_path"[^"]*"[^"]*"' | cut -d'"' -f4)
    fi
    
    # Get commit message if this is a git commit
    local commit_msg=""
    if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
        commit_msg=$(git log -1 --pretty=%B 2>/dev/null || echo "")
    fi
    
    # Check for fix claims without testing evidence
    if is_fix_claim "$commit_msg$INPUT"; then
        if ! check_testing_evidence "$commit_msg$INPUT"; then
            # Protocol violation detected
            log_violation "FIX_WITHOUT_TEST" "$(pwd)"
            
            echo "🚨 PROTOCOL VIOLATION: 'Fixed' claim without testing evidence!"
            echo ""
            echo "❌ Before claiming something is 'fixed', you must:"
            echo "□ Did I run/build the code?"
            echo "□ Did I trigger the exact feature I changed?"
            echo "□ Did I see the expected result?"
            echo "□ Did I check for error messages?"
            echo "□ Would I bet \$100 this works?"
            echo ""
            echo "⚠️  Time saved by skipping tests: 30 seconds"
            echo "⏰ Time wasted when it doesn't work: 30 minutes"
            echo "💔 User trust lost: Immeasurable"
            echo ""
            echo "🔧 Minimum Viable Test for your changes:"
            
            if [[ -n "$files_changed" ]]; then
                suggest_minimum_test "$files_changed"
            else
                echo "• Test the specific scenario that uses your changes"
                echo "• Verify the application starts/builds successfully"
                echo "• Check logs for any errors"
            fi
            
            echo ""
            echo "💡 Remember: The user isn't paying for code. They're paying for solutions."
            echo "   Untested code isn't a solution—it's a guess."
            
            # Don't block, but warn strongly
            exit 0
        fi
    fi
    
    # Check for common anti-patterns
    if echo "$commit_msg$INPUT" | grep -qi "should work\|try this\|this might work"; then
        echo "⚠️  WARNING: Uncertain language detected. Consider testing before committing."
        log_violation "UNCERTAIN_LANGUAGE" "$(pwd)"
    fi
    
    exit 0
}

# Run the validation
main