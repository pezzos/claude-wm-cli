#!/bin/bash

# Hook Resolver - Resolves hook names to absolute paths
# Ensures all hooks can be called from any directory

HOOKS_DIR="$HOME/.claude/hooks"
COMMON_DIR="$HOOKS_DIR/common"
AGILE_DIR="$HOOKS_DIR/agile"

# Function to resolve hook name to absolute path
resolve_hook() {
    local hook_name="$1"
    
    # Check if it's already an absolute path
    if [[ "$hook_name" == /* ]]; then
        echo "$hook_name"
        return 0
    fi
    
    # Check if it starts with ~/
    if [[ "$hook_name" == ~/* ]]; then
        echo "${hook_name/#\~/$HOME}"
        return 0
    fi
    
    # Map common hook names to their actual locations
    case "$hook_name" in
        "backup-current-state")
            echo "$COMMON_DIR/backup-state.sh"
            ;;
        "git-status-check")
            echo "$COMMON_DIR/git-status-check.sh"
            ;;
        "run-tests")
            echo "$COMMON_DIR/run-tests.sh"
            ;;
        "auto-format")
            echo "$COMMON_DIR/auto-format.sh"
            ;;
        "setup-project-template")
            echo "$HOOKS_DIR/development-operations.sh"
            ;;
        "initial-git-commit")
            echo "$HOOKS_DIR/development-operations.sh"
            ;;
        "smart-notify")
            echo "$HOOKS_DIR/smart-notify.sh"
            ;;
        "validate-design")
            echo "$AGILE_DIR/validate-design.sh"
            ;;
        "post-iterate")
            echo "$AGILE_DIR/post-iterate.sh"
            ;;
        *)
            # Try to find the hook in common locations
            if [[ -f "$HOOKS_DIR/$hook_name" ]]; then
                echo "$HOOKS_DIR/$hook_name"
            elif [[ -f "$COMMON_DIR/$hook_name" ]]; then
                echo "$COMMON_DIR/$hook_name"
            elif [[ -f "$AGILE_DIR/$hook_name" ]]; then
                echo "$AGILE_DIR/$hook_name"
            else
                echo "ERROR: Hook not found: $hook_name" >&2
                return 1
            fi
            ;;
    esac
}

# Function to execute a hook with proper error handling
execute_hook() {
    local hook_name="$1"
    shift
    local args="$@"
    
    local hook_path=$(resolve_hook "$hook_name")
    if [[ $? -ne 0 ]]; then
        echo "Failed to resolve hook: $hook_name" >&2
        return 1
    fi
    
    if [[ ! -f "$hook_path" ]]; then
        echo "Hook file not found: $hook_path" >&2
        return 1
    fi
    
    if [[ ! -x "$hook_path" ]]; then
        echo "Hook is not executable: $hook_path" >&2
        return 1
    fi
    
    # Execute the hook with proper logging
    echo "Executing hook: $hook_path $args" >&2
    "$hook_path" $args
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        echo "Hook failed with exit code $exit_code: $hook_path" >&2
    fi
    
    return $exit_code
}

# If called directly, execute the hook
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -eq 0 ]]; then
        echo "Usage: $0 <hook_name> [args...]" >&2
        exit 1
    fi
    
    execute_hook "$@"
fi