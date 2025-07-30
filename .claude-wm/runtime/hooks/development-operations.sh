#!/bin/bash

# Development Operations Hook
# Implements "Development Operations" section from CLAUDE.md
# Automatically runs linting, builds, and tests after code changes

# Read input data
INPUT=$(cat 2>/dev/null || echo '{}')

# Function to detect project type and available commands
detect_project_commands() {
    local commands=()
    
    # Node.js/JavaScript projects
    if [[ -f "package.json" ]]; then
        local package_content=$(cat package.json)
        
        # Check for available scripts
        if echo "$package_content" | jq -e '.scripts.lint' >/dev/null 2>&1; then
            commands+=("npm run lint")
        elif echo "$package_content" | jq -e '.scripts.eslint' >/dev/null 2>&1; then
            commands+=("npm run eslint")
        elif command -v eslint >/dev/null; then
            commands+=("npx eslint .")
        fi
        
        if echo "$package_content" | jq -e '.scripts.test' >/dev/null 2>&1; then
            commands+=("npm test")
        elif echo "$package_content" | jq -e '.scripts."test:unit"' >/dev/null 2>&1; then
            commands+=("npm run test:unit")
        fi
        
        if echo "$package_content" | jq -e '.scripts.build' >/dev/null 2>&1; then
            commands+=("npm run build")
        elif echo "$package_content" | jq -e '.scripts.compile' >/dev/null 2>&1; then
            commands+=("npm run compile")
        fi
        
        if echo "$package_content" | jq -e '.scripts.typecheck' >/dev/null 2>&1; then
            commands+=("npm run typecheck")
        elif command -v tsc >/dev/null && [[ -f "tsconfig.json" ]]; then
            commands+=("npx tsc --noEmit")
        fi
    fi
    
    # Python projects
    if [[ -f "requirements.txt" || -f "pyproject.toml" || -f "setup.py" ]]; then
        if command -v flake8 >/dev/null; then
            commands+=("flake8 .")
        elif command -v pylint >/dev/null; then
            commands+=("pylint .")
        elif command -v ruff >/dev/null; then
            commands+=("ruff check .")
        fi
        
        if command -v pytest >/dev/null; then
            commands+=("pytest")
        elif command -v python >/dev/null; then
            commands+=("python -m unittest discover")
        fi
        
        if command -v black >/dev/null; then
            commands+=("black --check .")
        fi
    fi
    
    # Go projects
    if [[ -f "go.mod" ]]; then
        commands+=("go fmt ./...")
        commands+=("go vet ./...")
        commands+=("go test ./...")
        commands+=("go build ./...")
    fi
    
    # Rust projects
    if [[ -f "Cargo.toml" ]]; then
        commands+=("cargo fmt --check")
        commands+=("cargo clippy")
        commands+=("cargo test")
        commands+=("cargo build")
    fi
    
    # Java projects
    if [[ -f "pom.xml" ]]; then
        commands+=("mvn compile")
        commands+=("mvn test")
    elif [[ -f "build.gradle" ]]; then
        commands+=("gradle build")
        commands+=("gradle test")
    fi
    
    printf '%s\n' "${commands[@]}"
}

# Function to run a command and capture result
run_command() {
    local cmd="$1"
    local description="$2"
    
    echo "🔧 Running $description..."
    echo "   Command: $cmd"
    
    # Create log file
    local log_file="$HOME/.claude/hooks/logs/development-operations.log"
    mkdir -p "$(dirname "$log_file")"
    
    # Log the command
    echo "$(date '+%Y-%m-%d %H:%M:%S') - RUNNING: $cmd" >> "$log_file"
    
    # Run the command with timeout
    local output
    local exit_code
    
    if timeout 300s bash -c "$cmd" > /tmp/dev_ops_output 2>&1; then
        exit_code=0
        output=$(cat /tmp/dev_ops_output)
        echo "   ✅ $description passed"
        echo "$(date '+%Y-%m-%d %H:%M:%S') - SUCCESS: $cmd" >> "$log_file"
    else
        exit_code=$?
        output=$(cat /tmp/dev_ops_output)
        echo "   ❌ $description failed (exit code: $exit_code)"
        echo "$(date '+%Y-%m-%d %H:%M:%S') - FAILED: $cmd (exit code: $exit_code)" >> "$log_file"
        
        # Show first few lines of error output
        echo "   Error output:"
        echo "$output" | head -10 | sed 's/^/      /'
        
        return $exit_code
    fi
    
    # Clean up temp file
    rm -f /tmp/dev_ops_output
    
    return 0
}

# Function to check if code changes warrant running operations
should_run_operations() {
    local file_path="$1"
    
    # Always run for source code files
    if [[ "$file_path" =~ \.(js|ts|jsx|tsx|py|java|go|rs|cpp|c|h|php|rb|cs)$ ]]; then
        return 0
    fi
    
    # Run for configuration files that might affect build
    if [[ "$file_path" =~ (package\.json|requirements\.txt|go\.mod|Cargo\.toml|pom\.xml|build\.gradle|tsconfig\.json|eslintrc|pyproject\.toml)$ ]]; then
        return 0
    fi
    
    return 1
}

# Main execution
main() {
    # Create logs directory
    mkdir -p "$HOME/.claude/hooks/logs"
    
    # Check if we should run operations
    local target_file=""
    local should_run=false
    
    if echo "$INPUT" | grep -q '"file_path"'; then
        target_file=$(echo "$INPUT" | grep -o '"file_path"[^"]*"[^"]*"' | cut -d'"' -f4)
        
        if [[ -n "$target_file" ]] && should_run_operations "$target_file"; then
            should_run=true
            echo "📝 Code changes detected in: $target_file"
        fi
    else
        # If no specific file, assume we should run for any project
        if [[ -f "package.json" || -f "requirements.txt" || -f "go.mod" || -f "Cargo.toml" || -f "pom.xml" || -f "build.gradle" ]]; then
            should_run=true
            echo "📝 Running development operations for project"
        fi
    fi
    
    if [[ "$should_run" == true ]]; then
        echo "🚀 Starting Development Operations..."
        echo ""
        
        # Get available commands for this project
        local commands=($(detect_project_commands))
        
        if [[ ${#commands[@]} -eq 0 ]]; then
            echo "⚠️  No development commands detected for this project type"
            exit 0
        fi
        
        local failed_commands=()
        
        # Run each command
        for cmd in "${commands[@]}"; do
            local description=""
            case "$cmd" in
                *lint*|*eslint*|*flake8*|*pylint*|*ruff*|*fmt*|*clippy*)
                    description="Linting"
                    ;;
                *test*)
                    description="Testing"
                    ;;
                *build*|*compile*)
                    description="Building"
                    ;;
                *typecheck*|*tsc*)
                    description="Type checking"
                    ;;
                *)
                    description="Code quality check"
                    ;;
            esac
            
            if ! run_command "$cmd" "$description"; then
                failed_commands+=("$cmd")
            fi
            echo ""
        done
        
        # Summary
        if [[ ${#failed_commands[@]} -eq 0 ]]; then
            echo "🎉 All development operations passed!"
            echo "✅ Code is ready for commit/deployment"
        else
            echo "❌ Some operations failed:"
            for failed_cmd in "${failed_commands[@]}"; do
                echo "   • $failed_cmd"
            done
            echo ""
            echo "💡 Fix these issues before committing:"
            echo "   • Review error output above"
            echo "   • Run commands individually for detailed debugging"
            echo "   • Consider using --fix flags where available"
        fi
        
    fi
    
    exit 0  # Don't block, just inform
}

# Run the operations
main