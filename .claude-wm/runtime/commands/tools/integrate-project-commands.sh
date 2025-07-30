#!/bin/bash

# Integration script for project/epic commands with JSON schema validation
set -eo pipefail

COMMANDS_DIR=".claude-wm/.claude/commands"
TOOLS_DIR="$COMMANDS_DIR/tools"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Commands that work with project-level JSON files
PROJECT_COMMANDS=(
    "1-project/3-epics/1-Plan-Epics.md:epics"
    "2-epic/1-start/2-Plan-stories.md:stories"
    "2-epic/2-manage/1-Complete-Epic.md:current-epic"
    "2-epic/2-manage/2-Status-Epic.md:current-epic"
    "3-story/1-manage/1-Start-Story.md:current-story"
    "3-story/1-manage/2-Complete-Story.md:current-story"
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
    
    echo -e "${BLUE}Integrating: $command_file for $json_type${NC}"
    
    # Determine the correct file path based on json_type
    local json_path=""
    case "$json_type" in
        "epics")
            json_path="docs/1-project/epics.json"
            ;;
        "stories")
            json_path="docs/2-current-epic/stories.json"
            ;;
        "current-epic")
            json_path="docs/2-current-epic/current-epic.json"
            ;;
        "current-story")
            json_path="docs/2-current-epic/current-story.json"
            ;;
        *)
            json_path="docs/3-current-task/$json_type.json"
            ;;
    esac
    
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
When updating $json_path, include this in your Claude prompt:

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
if ! .claude-wm/.claude/commands/tools/simple-validator.sh validate-file $json_path; then
    echo "⚠ JSON validation failed - attempting auto-correction"
    .claude-wm/.claude/commands/tools/json-validator.sh auto-correct $json_path
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

# Integrate validation into project commands
integrate_project_commands() {
    echo -e "${BLUE}=== Integrating Validation into Project Commands ===${NC}"
    
    local success_count=0
    local total_count=${#PROJECT_COMMANDS[@]}
    
    for command_entry in "${PROJECT_COMMANDS[@]}"; do
        local command_path="${command_entry%:*}"
        local json_type="${command_entry#*:}"
        local full_path="$COMMANDS_DIR/$command_path"
        
        if add_validation_to_command "$full_path" "$json_type"; then
            ((success_count++))
        fi
    done
    
    echo ""
    echo -e "${GREEN}Integration Summary: ${success_count}/${total_count} commands integrated${NC}"
}

# Test the validation for project commands
test_project_validation() {
    echo -e "${BLUE}=== Testing Project Validation System ===${NC}"
    
    # Test each JSON type
    for json_type in "epics" "stories" "current-epic" "current-story"; do
        echo -e "${BLUE}Testing schema requirements for: $json_type${NC}"
        if ./.claude-wm/.claude/commands/tools/schema-enforcer.sh show-requirements "$json_type"; then
            echo -e "${GREEN}✓ Schema available for $json_type${NC}"
        else
            echo -e "${RED}✗ Schema missing for $json_type${NC}"
        fi
        echo ""
    done
}

# Show integration status
show_status() {
    echo -e "${BLUE}=== Project Commands Integration Status ===${NC}"
    local integrated=0
    
    for command_entry in "${PROJECT_COMMANDS[@]}"; do
        local command_path="${command_entry%:*}"
        local json_type="${command_entry#*:}"
        local full_path="$COMMANDS_DIR/$command_path"
        
        if [[ -f "$full_path" ]] && grep -q "JSON_SCHEMA_VALIDATION" "$full_path"; then
            echo -e "${GREEN}✓ $command_path ($json_type)${NC}"
            ((integrated++))
        else
            echo -e "${RED}✗ $command_path ($json_type)${NC}"
        fi
    done
    
    echo -e "${BLUE}Total: ${integrated}/${#PROJECT_COMMANDS[@]} integrated${NC}"
}

# Main function
main() {
    case "${1:-integrate}" in
        "integrate")
            integrate_project_commands
            ;;
        "test")
            test_project_validation
            ;;
        "status")
            show_status
            ;;
        "help")
            echo "Usage: $0 [integrate|test|status|help]"
            echo "  integrate - Add validation to all project commands"
            echo "  test      - Test the validation system for project commands"
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