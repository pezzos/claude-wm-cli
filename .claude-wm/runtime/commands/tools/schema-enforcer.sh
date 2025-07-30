#!/bin/bash

# Schema Enforcement System for Claude WM CLI
# Provides proactive schema guidance to prevent JSON validation errors

set -euo pipefail

# Configuration
SCHEMA_DIR=".claude-wm/.claude/commands/templates/schemas"
CLAUDE_CLI="claude-wm-cli"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get schema content for a JSON type
get_schema_content() {
    local json_type="$1"
    local schema_file="$SCHEMA_DIR/${json_type}.schema.json"
    
    if [[ -f "$schema_file" ]]; then
        cat "$schema_file"
    else
        echo "Error: Schema not found for $json_type" >&2
        return 1
    fi
}

# Generate schema-aware prompt for Claude
generate_schema_prompt() {
    local json_type="$1"
    local task_description="$2"
    local existing_file="${3:-}"
    
    local schema_content
    if ! schema_content=$(get_schema_content "$json_type"); then
        return 1
    fi
    
    cat << EOF
${task_description}

MANDATORY SCHEMA COMPLIANCE:
You MUST generate/update JSON that strictly follows this schema:

\`\`\`json
${schema_content}
\`\`\`

CRITICAL REQUIREMENTS:
1. ALL required fields must be present
2. Data types must match exactly (string, number, boolean, array, object)
3. String patterns (regex) must be followed precisely
4. Array items must conform to itemSchema
5. Enum values must be from the allowed list only
6. String length constraints (minLength, maxLength) must be respected
7. No additional properties beyond those defined in schema

VALIDATION RULES:
- IDs must follow exact patterns (e.g., "TASK-001", "STORY-001")
- Enums are case-sensitive and limited to defined values
- Arrays must have minimum required items where specified
- Objects must include all required properties
- String fields cannot be empty unless explicitly allowed

EOF

    if [[ -n "$existing_file" && -f "$existing_file" ]]; then
        echo "EXISTING FILE TO UPDATE:"
        echo "\`\`\`json"
        cat "$existing_file"
        echo "\`\`\`"
        echo ""
        echo "Update this file while maintaining schema compliance."
    else
        echo "Generate a new JSON file that perfectly matches the schema requirements."
    fi
    
    cat << EOF

OUTPUT FORMAT:
Provide ONLY the valid JSON content. No explanations, no markdown formatting.
The output will be directly written to the file and must pass schema validation.
EOF
}

# Create schema-guided JSON generation
generate_with_schema() {
    local json_type="$1"
    local output_file="$2"
    local task_description="$3"
    local existing_file="${4:-}"
    
    echo -e "${BLUE}Generating schema-compliant JSON for: $json_type${NC}"
    
    local schema_prompt
    if ! schema_prompt=$(generate_schema_prompt "$json_type" "$task_description" "$existing_file"); then
        echo -e "${RED}Failed to generate schema prompt${NC}"
        return 1
    fi
    
    # Create temporary file for output
    local temp_output=$(mktemp)
    
    # Call Claude with schema-aware prompt
    if echo "$schema_prompt" | timeout 60 "$CLAUDE_CLI" > "$temp_output" 2>/dev/null; then
        # Validate the generated JSON
        if jq empty "$temp_output" 2>/dev/null; then
            # Move to final location
            mv "$temp_output" "$output_file"
            echo -e "${GREEN}✓ Schema-compliant JSON generated: $output_file${NC}"
            return 0
        else
            echo -e "${RED}✗ Generated JSON is invalid${NC}"
            rm -f "$temp_output"
            return 1
        fi
    else
        echo -e "${RED}✗ Failed to generate JSON with Claude${NC}"
        rm -f "$temp_output"
        return 1
    fi
}

# Validate field before adding to JSON
validate_field() {
    local field_name="$1"
    local field_value="$2"
    local json_type="$3"
    
    # Get field definition from schema
    local schema_content
    if ! schema_content=$(get_schema_content "$json_type"); then
        return 1
    fi
    
    # Extract field definition (simplified check)
    if echo "$schema_content" | jq -e ".properties.\"$field_name\"" >/dev/null 2>&1; then
        local field_def
        field_def=$(echo "$schema_content" | jq ".properties.\"$field_name\"")
        
        # Basic type checking
        local expected_type
        expected_type=$(echo "$field_def" | jq -r '.type // "any"')
        
        case "$expected_type" in
            "string")
                if [[ -z "$field_value" ]]; then
                    echo -e "${YELLOW}Warning: Empty string for field '$field_name'${NC}"
                fi
                ;;
            "number")
                if ! [[ "$field_value" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
                    echo -e "${RED}Error: '$field_value' is not a valid number for field '$field_name'${NC}"
                    return 1
                fi
                ;;
            "boolean")
                if [[ "$field_value" != "true" && "$field_value" != "false" ]]; then
                    echo -e "${RED}Error: '$field_value' is not a valid boolean for field '$field_name'${NC}"
                    return 1
                fi
                ;;
        esac
        
        # Check enum constraints
        if echo "$field_def" | jq -e '.enum' >/dev/null 2>&1; then
            local allowed_values
            allowed_values=$(echo "$field_def" | jq -r '.enum[]' | tr '\n' '|' | sed 's/|$//')
            if ! echo "$field_value" | grep -qE "^($allowed_values)$"; then
                echo -e "${RED}Error: '$field_value' is not in allowed values for field '$field_name'${NC}"
                echo -e "${YELLOW}Allowed values: $(echo "$field_def" | jq -r '.enum | join(", ")')${NC}"
                return 1
            fi
        fi
        
        echo -e "${GREEN}✓ Field '$field_name' is valid${NC}"
        return 0
    else
        echo -e "${YELLOW}Warning: Field '$field_name' not defined in schema${NC}"
        return 0
    fi
}

# Show schema requirements for a JSON type
show_schema_requirements() {
    local json_type="$1"
    
    local schema_content
    if ! schema_content=$(get_schema_content "$json_type"); then
        return 1
    fi
    
    echo -e "${BLUE}=== Schema Requirements for $json_type ===${NC}"
    echo ""
    
    # Show required fields
    echo -e "${YELLOW}Required Fields:${NC}"
    echo "$schema_content" | jq -r '.required[]?' | while read -r field; do
        local field_def
        field_def=$(echo "$schema_content" | jq ".properties.\"$field\"")
        local field_type
        field_type=$(echo "$field_def" | jq -r '.type // "any"')
        local description
        description=$(echo "$field_def" | jq -r '.description // ""')
        
        echo "  • $field ($field_type): $description"
        
        # Show enum values if present
        if echo "$field_def" | jq -e '.enum' >/dev/null 2>&1; then
            local enum_values
            enum_values=$(echo "$field_def" | jq -r '.enum | join(", ")')
            echo "    Allowed values: $enum_values"
        fi
        
        # Show pattern if present
        if echo "$field_def" | jq -e '.pattern' >/dev/null 2>&1; then
            local pattern
            pattern=$(echo "$field_def" | jq -r '.pattern')
            echo "    Pattern: $pattern"
        fi
    done
    
    echo ""
    
    # Show optional fields
    echo -e "${YELLOW}Optional Fields:${NC}"
    local all_properties
    all_properties=$(echo "$schema_content" | jq -r '.properties | keys[]')
    local required_fields
    required_fields=$(echo "$schema_content" | jq -r '.required[]?' | tr '\n' '|' | sed 's/|$//')
    
    echo "$all_properties" | while read -r field; do
        if ! echo "$field" | grep -qE "^($required_fields)$"; then
            local field_def
            field_def=$(echo "$schema_content" | jq ".properties.\"$field\"")
            local field_type
            field_type=$(echo "$field_def" | jq -r '.type // "any"')
            local description
            description=$(echo "$field_def" | jq -r '.description // ""')
            
            echo "  • $field ($field_type): $description"
        fi
    done
    
    echo ""
}

# Generate template JSON with correct structure
generate_template() {
    local json_type="$1"
    local output_file="${2:-}"
    
    local schema_content
    if ! schema_content=$(get_schema_content "$json_type"); then
        return 1
    fi
    
    echo -e "${BLUE}Generating template for: $json_type${NC}"
    
    # Create basic template with required fields
    local template="{}"
    
    # Add required fields with default values
    local required_fields
    required_fields=$(echo "$schema_content" | jq -r '.required[]?' 2>/dev/null || echo "")
    
    for field in $required_fields; do
        local field_def
        field_def=$(echo "$schema_content" | jq ".properties.\"$field\"")
        local field_type
        field_type=$(echo "$field_def" | jq -r '.type // "string"')
        
        case "$field_type" in
            "string")
                if echo "$field_def" | jq -e '.enum' >/dev/null 2>&1; then
                    local first_enum
                    first_enum=$(echo "$field_def" | jq -r '.enum[0]')
                    template=$(echo "$template" | jq --arg field "$field" --arg value "$first_enum" '. + {($field): $value}')
                elif echo "$field_def" | jq -e '.pattern' >/dev/null 2>&1; then
                    # For pattern fields, provide example
                    case "$field" in
                        "id")
                            if [[ "$json_type" == "current-task" ]]; then
                                template=$(echo "$template" | jq --arg field "$field" '. + {($field): "TASK-001"}')
                            elif [[ "$json_type" == "current-story" ]]; then
                                template=$(echo "$template" | jq --arg field "$field" '. + {($field): "STORY-001"}')
                            else
                                template=$(echo "$template" | jq --arg field "$field" '. + {($field): "EXAMPLE-001"}')
                            fi
                            ;;
                        *)
                            template=$(echo "$template" | jq --arg field "$field" '. + {($field): ""}')
                            ;;
                    esac
                else
                    template=$(echo "$template" | jq --arg field "$field" '. + {($field): ""}')
                fi
                ;;
            "number")
                template=$(echo "$template" | jq --arg field "$field" '. + {($field): 0}')
                ;;
            "boolean")
                template=$(echo "$template" | jq --arg field "$field" '. + {($field): false}')
                ;;
            "array")
                template=$(echo "$template" | jq --arg field "$field" '. + {($field): []}')
                ;;
            "object")
                template=$(echo "$template" | jq --arg field "$field" '. + {($field): {}}')
                ;;
        esac
    done
    
    # Pretty print the template
    local formatted_template
    formatted_template=$(echo "$template" | jq '.')
    
    if [[ -n "$output_file" ]]; then
        echo "$formatted_template" > "$output_file"
        echo -e "${GREEN}✓ Template generated: $output_file${NC}"
    else
        echo "$formatted_template"
    fi
}

# Main function
main() {
    local command="${1:-help}"
    
    case "$command" in
        "generate")
            if [[ $# -lt 4 ]]; then
                echo "Usage: $0 generate <json_type> <output_file> <task_description> [existing_file]"
                exit 1
            fi
            generate_with_schema "$2" "$3" "$4" "${5:-}"
            ;;
        "validate-field")
            if [[ $# -lt 4 ]]; then
                echo "Usage: $0 validate-field <field_name> <field_value> <json_type>"
                exit 1
            fi
            validate_field "$2" "$3" "$4"
            ;;
        "show-requirements")
            if [[ $# -lt 2 ]]; then
                echo "Usage: $0 show-requirements <json_type>"
                exit 1
            fi
            show_schema_requirements "$2"
            ;;
        "template")
            if [[ $# -lt 2 ]]; then
                echo "Usage: $0 template <json_type> [output_file]"
                exit 1
            fi
            generate_template "$2" "${3:-}"
            ;;
        "help")
            cat << EOF
Schema Enforcement System for Claude WM CLI

Usage: $0 <command> [options]

Commands:
  generate <type> <output> <description> [existing]
                           Generate schema-compliant JSON with Claude
  validate-field <name> <value> <type>
                           Validate a field value against schema
  show-requirements <type>
                           Show schema requirements for JSON type
  template <type> [output]
                           Generate empty template following schema
  help                     Show this help

JSON Types:
  current-task, current-story, current-epic, stories, epics, iterations, metrics

Examples:
  $0 show-requirements current-task
  $0 template current-task task-template.json
  $0 generate current-task output.json "Create user login feature"
  $0 validate-field status "in_progress" current-task

This tool helps prevent JSON validation errors by providing schema-aware
generation and validation before files are created or updated.
EOF
            ;;
        *)
            echo -e "${RED}Unknown command: $command${NC}"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

main "$@"