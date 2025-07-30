#!/bin/bash

# JSON Schema Validation System for Claude WM CLI
# Validates generated JSON files against their corresponding schemas
# and provides auto-correction through Claude

set -euo pipefail

# Configuration
SCHEMA_DIR=".claude-wm/.claude/commands/templates/schemas"
VALIDATION_LOG=".claude-wm/.claude/validation.log"
CLAUDE_CLI="claude-wm-cli"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize validation log
init_validation_log() {
    mkdir -p "$(dirname "$VALIDATION_LOG")"
    echo "$(date): JSON validation system initialized" >> "$VALIDATION_LOG"
}

# Get schema file for a JSON file
get_schema_file() {
    local json_file="$1"
    local filename=$(basename "$json_file" .json)
    local schema_file="$SCHEMA_DIR/${filename}.schema.json"
    
    if [[ -f "$schema_file" ]]; then
        echo "$schema_file"
    else
        return 1
    fi
}

# Validate JSON against schema using Node.js (more robust than bash solutions)
validate_json_schema() {
    local json_file="$1"
    local schema_file="$2"
    
    # Create temporary validation script
    local validator_script=$(mktemp)
    cat > "$validator_script" << 'EOF'
const fs = require('fs');
const Ajv = require('ajv');
const addFormats = require('ajv-formats');

const ajv = new Ajv({allErrors: true, verbose: true});
addFormats(ajv);

try {
    const jsonFile = process.argv[2];
    const schemaFile = process.argv[3];
    
    const data = JSON.parse(fs.readFileSync(jsonFile, 'utf8'));
    const schema = JSON.parse(fs.readFileSync(schemaFile, 'utf8'));
    
    const validate = ajv.compile(schema);
    const valid = validate(data);
    
    if (!valid) {
        console.log('VALIDATION_FAILED');
        console.log(JSON.stringify(validate.errors, null, 2));
        process.exit(1);
    } else {
        console.log('VALIDATION_SUCCESS');
        process.exit(0);
    }
} catch (error) {
    console.log('VALIDATION_ERROR');
    console.log(error.message);
    process.exit(2);
}
EOF

    # Check if node and ajv are available
    if ! command -v node >/dev/null 2>&1; then
        echo -e "${YELLOW}Warning: Node.js not found. Installing validation dependencies...${NC}"
        # Fallback to basic JSON validation
        if jq empty "$json_file" 2>/dev/null; then
            echo "VALIDATION_SUCCESS"
            return 0
        else
            echo "VALIDATION_FAILED"
            echo "Invalid JSON syntax"
            return 1
        fi
    fi
    
    # Run validation
    local validation_output
    if validation_output=$(node "$validator_script" "$json_file" "$schema_file" 2>&1); then
        rm -f "$validator_script"
        echo "$validation_output"
        return 0
    else
        local exit_code=$?
        rm -f "$validator_script"
        echo "$validation_output"
        return $exit_code
    fi
}

# Generate correction prompt for Claude
generate_correction_prompt() {
    local json_file="$1"
    local schema_file="$2"
    local validation_errors="$3"
    
    cat << EOF
Fix the following JSON file to comply with its schema:

JSON File: $json_file
Schema File: $schema_file

Validation Errors:
$validation_errors

IMPORTANT: 
1. Read the current JSON file and the schema file
2. Fix ALL validation errors while preserving existing valid data
3. Ensure the corrected JSON is properly formatted and valid
4. Replace the entire file content with the corrected version
5. DO NOT add explanations - just fix the file

The JSON file should be updated to pass schema validation.
EOF
}

# Auto-correct JSON file using Claude
auto_correct_json() {
    local json_file="$1"
    local schema_file="$2"
    local validation_errors="$3"
    
    echo -e "${BLUE}Auto-correcting $json_file using Claude...${NC}"
    
    local correction_prompt
    correction_prompt=$(generate_correction_prompt "$json_file" "$schema_file" "$validation_errors")
    
    # Log correction attempt
    echo "$(date): Attempting auto-correction for $json_file" >> "$VALIDATION_LOG"
    echo "Validation errors: $validation_errors" >> "$VALIDATION_LOG"
    
    # Call Claude for correction
    if echo "$correction_prompt" | timeout 60 "$CLAUDE_CLI" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Auto-correction completed for $json_file${NC}"
        echo "$(date): Auto-correction successful for $json_file" >> "$VALIDATION_LOG"
        return 0
    else
        echo -e "${RED}✗ Auto-correction failed for $json_file${NC}"
        echo "$(date): Auto-correction failed for $json_file" >> "$VALIDATION_LOG"
        return 1
    fi
}

# Validate a single JSON file
validate_single_file() {
    local json_file="$1"
    local auto_correct="${2:-true}"
    
    if [[ ! -f "$json_file" ]]; then
        echo -e "${RED}✗ File not found: $json_file${NC}"
        return 1
    fi
    
    # Get corresponding schema
    local schema_file
    if ! schema_file=$(get_schema_file "$json_file"); then
        echo -e "${YELLOW}⚠ No schema found for $json_file, skipping validation${NC}"
        return 0
    fi
    
    echo -e "${BLUE}Validating: $json_file${NC}"
    echo -e "${BLUE}Schema: $schema_file${NC}"
    
    # Perform validation
    local validation_result
    local validation_errors=""
    
    if validation_result=$(validate_json_schema "$json_file" "$schema_file"); then
        if echo "$validation_result" | grep -q "VALIDATION_SUCCESS"; then
            echo -e "${GREEN}✓ $json_file is valid${NC}"
            echo "$(date): Validation passed for $json_file" >> "$VALIDATION_LOG"
            return 0
        fi
    fi
    
    # Extract validation errors
    validation_errors=$(echo "$validation_result" | sed '1d') # Remove first line (status)
    
    echo -e "${RED}✗ $json_file failed validation${NC}"
    echo -e "${RED}Errors:${NC}"
    echo "$validation_errors"
    
    # Attempt auto-correction if enabled
    if [[ "$auto_correct" == "true" ]]; then
        if auto_correct_json "$json_file" "$schema_file" "$validation_errors"; then
            # Re-validate after correction
            echo -e "${BLUE}Re-validating after correction...${NC}"
            validate_single_file "$json_file" "false" # Prevent infinite recursion
            return $?
        else
            echo -e "${RED}Manual intervention required for $json_file${NC}"
            return 1
        fi
    else
        return 1
    fi
}

# Validate all JSON files in current epic/story/task
validate_current_context() {
    local auto_correct="${1:-true}"
    local validation_failed=false
    
    echo -e "${BLUE}=== JSON Schema Validation ===${NC}"
    
    # Find all JSON files in docs directories
    local json_files=(
        "docs/2-current-epic/current-epic.json"
        "docs/2-current-epic/stories.json"
        "docs/3-current-task/current-task.json"
        "docs/3-current-task/iterations.json"
        "docs/1-project/epics.json"
        "docs/1-project/metrics.json"
    )
    
    for json_file in "${json_files[@]}"; do
        if [[ -f "$json_file" ]]; then
            if ! validate_single_file "$json_file" "$auto_correct"; then
                validation_failed=true
            fi
            echo ""
        fi
    done
    
    if [[ "$validation_failed" == "true" ]]; then
        echo -e "${RED}=== Validation Summary: FAILED ===${NC}"
        echo -e "${RED}Some files failed validation. Check the log: $VALIDATION_LOG${NC}"
        return 1
    else
        echo -e "${GREEN}=== Validation Summary: PASSED ===${NC}"
        echo -e "${GREEN}All JSON files are valid${NC}"
        return 0
    fi
}

# Install validation dependencies if needed
install_dependencies() {
    if ! command -v node >/dev/null 2>&1; then
        echo -e "${YELLOW}Installing Node.js for JSON validation...${NC}"
        if command -v brew >/dev/null 2>&1; then
            brew install node
        elif command -v apt-get >/dev/null 2>&1; then
            sudo apt-get update && sudo apt-get install -y nodejs npm
        elif command -v yum >/dev/null 2>&1; then
            sudo yum install -y nodejs npm
        else
            echo -e "${RED}Please install Node.js manually for full validation support${NC}"
            return 1
        fi
    fi
    
    # Install ajv if not present
    if ! node -e "require('ajv')" 2>/dev/null; then
        echo -e "${BLUE}Installing JSON schema validation library...${NC}"
        npm install -g ajv ajv-formats 2>/dev/null || {
            echo -e "${YELLOW}Could not install global packages. Using local fallback.${NC}"
            mkdir -p .claude-wm/.claude/node_modules
            cd .claude-wm/.claude
            npm init -y >/dev/null 2>&1
            npm install ajv ajv-formats >/dev/null 2>&1
            cd - >/dev/null
        }
    fi
}

# Main function
main() {
    local command="${1:-validate}"
    local auto_correct="${2:-true}"
    
    init_validation_log
    
    case "$command" in
        "install")
            install_dependencies
            ;;
        "validate")
            if [[ $# -eq 2 && -f "$2" ]]; then
                validate_single_file "$2" "$auto_correct"
            else
                validate_current_context "$auto_correct"
            fi
            ;;
        "validate-no-fix")
            validate_current_context "false"
            ;;
        "help")
            cat << EOF
JSON Schema Validation System for Claude WM CLI

Usage: $0 [command] [options]

Commands:
  install           Install validation dependencies
  validate          Validate all JSON files (default, auto-correct enabled)
  validate-no-fix   Validate without auto-correction
  validate [file]   Validate specific file
  help             Show this help

Examples:
  $0                                    # Validate all files with auto-correction
  $0 validate docs/current-task.json   # Validate specific file
  $0 validate-no-fix                   # Validate without correction
  $0 install                           # Install dependencies

Auto-correction uses Claude to fix validation errors automatically.
Validation log is stored in: $VALIDATION_LOG
EOF
            ;;
        *)
            echo -e "${RED}Unknown command: $command${NC}"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"