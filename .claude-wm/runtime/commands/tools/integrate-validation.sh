#!/bin/bash

# Simple integration script for JSON schema validation
set -eo pipefail

COMMANDS_DIR=".claude-wm/.claude/commands"
TOOLS_DIR="$COMMANDS_DIR/tools"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Commands that work with current-task.json
TASK_COMMANDS=(
    "4-task/1-start/1-From-story.md"
    "4-task/1-start/2-From-issue.md"
    "4-task/1-start/3-From-input.md"
    "4-task/2-execute/1-Plan-Task.md"
    "4-task/2-execute/2-Test-design.md"
    "4-task/2-execute/3-Implement.md"
    "4-task/2-execute/4-Validate-Task.md"
    "4-task/2-execute/5-Review-Task.md"
    "4-task/3-complete/1-Archive-Task.md"
    "4-task/3-complete/2-Status-Task.md"
)

# Add validation section to a command file
add_validation_to_command() {
    local command_file="$1"
    local json_type="$2"
    
    if [[ ! -f "$command_file" ]]; then
        echo -e "${YELLOW}Command file not found: $command_file${NC}"
        return 1
    fi
    
    # Check if already integrated
    if grep -q "JSON_SCHEMA_VALIDATION" "$command_file"; then
        echo -e "${YELLOW}Already integrated: $command_file${NC}"
        return 0
    fi
    
    echo -e "${BLUE}Integrating: $command_file${NC}"
    
    # Add validation section
    cat >> "$command_file" << EOF

## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for $json_type.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

\`\`\`bash
# Show schema requirements
.claude-wm/.claude/commands/tools/schema-enforcer.sh show-requirements $json_type
\`\`\`

### Schema-Aware Generation
When updating docs/3-current-task/$json_type.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
\`\`\`bash
.claude-wm/.claude/commands/tools/schema-enforcer.sh show-requirements $json_type
\`\`\`

All required fields must be present with correct types and values.

### Post-Generation Validation
After completing the main task, validate the generated JSON:

\`\`\`bash
# Validate with auto-correction
if ! .claude-wm/.claude/commands/tools/json-validator.sh validate; then
    echo "⚠ JSON validation failed - files auto-corrected"
    exit 1  # Needs iteration
fi
\`\`\`

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
EOF

    echo -e "${GREEN}✓ Validation added to: $command_file${NC}"
    return 0
}

# Integrate validation into task commands
integrate_task_commands() {
    echo -e "${BLUE}=== Integrating Validation into Task Commands ===${NC}"
    
    local success_count=0
    local total_count=${#TASK_COMMANDS[@]}
    
    for command_path in "${TASK_COMMANDS[@]}"; do
        local full_path="$COMMANDS_DIR/$command_path"
        if add_validation_to_command "$full_path" "current-task"; then
            ((success_count++))
        fi
    done
    
    echo ""
    echo -e "${GREEN}Integration Summary: ${success_count}/${total_count} commands integrated${NC}"
}

# Create the master validation wrapper
create_master_wrapper() {
    local wrapper_file="$COMMANDS_DIR/validate-json.sh"
    
    cat > "$wrapper_file" << 'EOF'
#!/bin/bash

# Master JSON Validation Wrapper
set -eo pipefail

TOOLS_DIR="$(dirname "$0")/tools"

# Run complete validation
validate_all() {
    if "$TOOLS_DIR/json-validator.sh" validate; then
        echo "✓ All JSON files validated successfully"
        return 0
    else
        echo "⚠ Validation completed with corrections"
        return 1
    fi
}

# Show schema requirements
show_schemas() {
    echo "=== Schema Requirements ==="
    "$TOOLS_DIR/schema-enforcer.sh" show-requirements current-task
}

case "${1:-validate}" in
    "validate") validate_all ;;
    "schemas") show_schemas ;;
    "install") "$TOOLS_DIR/json-validator.sh" install ;;
    *) echo "Usage: $0 [validate|schemas|install]" ;;
esac
EOF

    chmod +x "$wrapper_file"
    echo -e "${GREEN}✓ Master wrapper created: $wrapper_file${NC}"
}

# Test the validation system
test_validation() {
    echo -e "${BLUE}=== Testing Validation System ===${NC}"
    
    # Check if current-task.json exists
    if [[ -f "docs/3-current-task/current-task.json" ]]; then
        echo -e "${BLUE}Testing with existing current-task.json${NC}"
        ./.claude-wm/.claude/commands/tools/json-validator.sh validate docs/3-current-task/current-task.json
    else
        echo -e "${YELLOW}No current-task.json found for testing${NC}"
    fi
    
    # Show schema requirements
    echo -e "${BLUE}Schema requirements for current-task:${NC}"
    ./.claude-wm/.claude/commands/tools/schema-enforcer.sh show-requirements current-task
}

# Main function
main() {
    case "${1:-integrate}" in
        "integrate")
            integrate_task_commands
            create_master_wrapper
            ;;
        "test")
            test_validation
            ;;
        "status")
            echo -e "${BLUE}=== Integration Status ===${NC}"
            local integrated=0
            for cmd in "${TASK_COMMANDS[@]}"; do
                if [[ -f "$COMMANDS_DIR/$cmd" ]] && grep -q "JSON_SCHEMA_VALIDATION" "$COMMANDS_DIR/$cmd"; then
                    echo -e "${GREEN}✓ $cmd${NC}"
                    ((integrated++))
                else
                    echo -e "${RED}✗ $cmd${NC}"
                fi
            done
            echo -e "${BLUE}Total: ${integrated}/${#TASK_COMMANDS[@]} integrated${NC}"
            ;;
        "help")
            echo "Usage: $0 [integrate|test|status|help]"
            echo "  integrate - Add validation to all task commands"
            echo "  test      - Test the validation system"
            echo "  status    - Show integration status"
            echo "  help      - Show this help"
            ;;
        *)
            echo "Unknown command: $1"
            echo "Use '$0 help' for usage"
            exit 1
            ;;
    esac
}

main "$@"