#!/bin/bash

# Simple JSON Schema Validation using jq
# Reliable fallback that doesn't require external dependencies

set -eo pipefail

SCHEMA_DIR=".claude-wm/.claude/commands/templates/schemas"
CLAUDE_CLI="claude-wm-cli"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Basic JSON validation using jq
validate_json_basic() {
    local json_file="$1"
    local schema_file="$2"
    
    echo -e "${BLUE}Validating: $json_file${NC}"
    
    # Check if JSON is valid
    if ! jq empty "$json_file" 2>/dev/null; then
        echo -e "${RED}✗ Invalid JSON syntax in $json_file${NC}"
        return 1
    fi
    
    # Check if schema exists
    if [[ ! -f "$schema_file" ]]; then
        echo -e "${YELLOW}⚠ No schema found, skipping validation${NC}"
        return 0
    fi
    
    # Extract required fields from schema
    local required_fields
    if ! required_fields=$(jq -r '.required[]?' "$schema_file" 2>/dev/null); then
        echo -e "${YELLOW}⚠ Could not read schema requirements${NC}"
        return 0
    fi
    
    local validation_errors=0
    
    # Check required fields
    echo "Checking required fields..."
    while IFS= read -r field; do
        if [[ -n "$field" ]]; then
            if jq -e "has(\"$field\")" "$json_file" >/dev/null 2>&1; then
                echo -e "${GREEN}✓ Required field '$field' present${NC}"
            else
                echo -e "${RED}✗ Missing required field: $field${NC}"
                ((validation_errors++))
            fi
        fi
    done <<< "$required_fields"
    
    # Check enum fields (basic check)
    local enum_errors
    enum_errors=$(check_enum_fields "$json_file" "$schema_file")
    if [[ -n "$enum_errors" ]]; then
        echo -e "${RED}✗ Enum validation errors:${NC}"
        echo "$enum_errors"
        ((validation_errors++))
    fi
    
    if [[ $validation_errors -eq 0 ]]; then
        echo -e "${GREEN}✓ $json_file validation passed${NC}"
        return 0
    else
        echo -e "${RED}✗ $json_file has $validation_errors validation errors${NC}"
        return 1
    fi
}

# Check enum field values
check_enum_fields() {
    local json_file="$1"
    local schema_file="$2"
    
    # Get enum fields from schema
    local enum_fields
    enum_fields=$(jq -r '
        .properties | to_entries[] | 
        select(.value.enum) | 
        "\(.key):\(.value.enum | join(","))"
    ' "$schema_file" 2>/dev/null || echo "")
    
    if [[ -z "$enum_fields" ]]; then
        return 0
    fi
    
    local errors=""
    
    while IFS=':' read -r field_name allowed_values; do
        if [[ -n "$field_name" && -n "$allowed_values" ]]; then
            local current_value
            current_value=$(jq -r ".$field_name // \"\"" "$json_file" 2>/dev/null)
            
            if [[ -n "$current_value" && "$current_value" != "null" ]]; then
                if ! echo "$allowed_values" | grep -q "$current_value"; then
                    errors+="Field '$field_name' has invalid value '$current_value'. Allowed: $allowed_values\n"
                fi
            fi
        fi
    done <<< "$enum_fields"
    
    echo -e "$errors"
}

# Generate correction prompt for Claude
generate_correction_prompt() {
    local json_file="$1"
    local schema_file="$2"
    local validation_errors="$3"
    
    cat << EOF
Fix the JSON file to comply with its schema requirements.

JSON File: $json_file
Schema File: $schema_file

Validation Errors:
$validation_errors

INSTRUCTIONS:
1. Read the current JSON file content
2. Read the schema file to understand requirements
3. Fix ALL validation errors while preserving valid existing data
4. Ensure all required fields are present with correct types
5. Ensure enum values are from allowed lists
6. Update the JSON file with the corrected content

The corrected file must pass schema validation.
EOF
}

# Auto-correct using Claude
auto_correct_json() {
    local json_file="$1"
    local schema_file="$2"
    local validation_errors="$3"
    
    echo -e "${BLUE}Attempting auto-correction with Claude...${NC}"
    
    local prompt
    prompt=$(generate_correction_prompt "$json_file" "$schema_file" "$validation_errors")
    
    # Try to call Claude (if available)
    if command -v "$CLAUDE_CLI" >/dev/null 2>&1; then
        if echo "$prompt" | timeout 60 "$CLAUDE_CLI" >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Auto-correction completed${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠ Claude auto-correction failed${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠ Claude CLI not available for auto-correction${NC}"
        return 1
    fi
}

# Validate a single file
validate_file() {
    local json_file="$1"
    local auto_correct="${2:-true}"
    
    if [[ ! -f "$json_file" ]]; then
        echo -e "${RED}✗ File not found: $json_file${NC}"
        return 1
    fi
    
    # Get schema file
    local filename
    filename=$(basename "$json_file" .json)
    local schema_file="$SCHEMA_DIR/${filename}.schema.json"
    
    # Validate
    local validation_output
    if validation_output=$(validate_json_basic "$json_file" "$schema_file" 2>&1); then
        echo "$validation_output"
        return 0
    else
        echo "$validation_output"
        
        # Attempt auto-correction if enabled
        if [[ "$auto_correct" == "true" ]]; then
            if auto_correct_json "$json_file" "$schema_file" "$validation_output"; then
                echo -e "${BLUE}Re-validating after correction...${NC}"
                validate_file "$json_file" "false"  # Prevent recursion
                return $?
            fi
        fi
        
        return 1
    fi
}

# Validate all relevant JSON files
validate_all() {
    local auto_correct="${1:-true}"
    
    echo -e "${BLUE}=== JSON Schema Validation ===${NC}"
    
    local files=(
        "docs/3-current-task/current-task.json"
        "docs/2-current-epic/current-epic.json"
        "docs/2-current-epic/stories.json"
        "docs/1-project/epics.json"
        "docs/1-project/metrics.json"
    )
    
    local validation_failed=false
    
    for file in "${files[@]}"; do
        if [[ -f "$file" ]]; then
            echo ""
            if ! validate_file "$file" "$auto_correct"; then
                validation_failed=true
            fi
        fi
    done
    
    echo ""
    if [[ "$validation_failed" == "true" ]]; then
        echo -e "${RED}=== Validation FAILED ===${NC}"
        echo -e "${RED}Some files failed validation${NC}"
        return 1
    else
        echo -e "${GREEN}=== Validation PASSED ===${NC}"
        echo -e "${GREEN}All JSON files are valid${NC}"
        return 0
    fi
}

# Show schema requirements
show_schema() {
    local json_type="$1"
    local schema_file="$SCHEMA_DIR/${json_type}.schema.json"
    
    if [[ ! -f "$schema_file" ]]; then
        echo -e "${RED}Schema not found: $schema_file${NC}"
        return 1
    fi
    
    echo -e "${BLUE}=== Schema Requirements for $json_type ===${NC}"
    echo ""
    
    # Show required fields
    echo -e "${YELLOW}Required Fields:${NC}"
    jq -r '.required[]?' "$schema_file" 2>/dev/null | while read -r field; do
        if [[ -n "$field" ]]; then
            local field_type
            field_type=$(jq -r ".properties.\"$field\".type // \"unknown\"" "$schema_file" 2>/dev/null)
            local description
            description=$(jq -r ".properties.\"$field\".description // \"\"" "$schema_file" 2>/dev/null)
            
            echo "  • $field ($field_type): $description"
            
            # Show enum values
            local enum_values
            enum_values=$(jq -r ".properties.\"$field\".enum // empty | join(\", \")" "$schema_file" 2>/dev/null)
            if [[ -n "$enum_values" ]]; then
                echo "    Allowed values: $enum_values"
            fi
        fi
    done
    
    echo ""
}

# Main function
main() {
    case "${1:-validate}" in
        "validate")
            validate_all "${2:-true}"
            ;;
        "validate-file")
            if [[ -z "$2" ]]; then
                echo "Usage: $0 validate-file <json_file>"
                exit 1
            fi
            validate_file "$2" "${3:-true}"
            ;;
        "show-schema")
            if [[ -z "$2" ]]; then
                echo "Usage: $0 show-schema <json_type>"
                exit 1
            fi
            show_schema "$2"
            ;;
        "help")
            cat << EOF
Simple JSON Schema Validator

Usage: $0 [command] [options]

Commands:
  validate                    Validate all JSON files (default)
  validate-file <file>        Validate specific file
  show-schema <type>          Show schema requirements
  help                        Show this help

Examples:
  $0                                    # Validate all files
  $0 validate-file current-task.json   # Validate specific file
  $0 show-schema current-task          # Show schema requirements

This validator uses jq for reliable validation without external dependencies.
EOF
            ;;
        *)
            echo "Unknown command: $1"
            echo "Use '$0 help' for usage"
            exit 1
            ;;
    esac
}

main "$@"