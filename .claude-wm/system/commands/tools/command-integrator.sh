#!/bin/bash

# Command Integration System for JSON Schema Validation
# Automatically integrates validation into existing command templates

set -eo pipefail

COMMANDS_DIR=".claude-wm/.claude/commands"
TOOLS_DIR="$COMMANDS_DIR/tools"
VALIDATOR="$TOOLS_DIR/json-validator.sh"
ENFORCER="$TOOLS_DIR/schema-enforcer.sh"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# JSON files mapped to their command contexts
declare -A JSON_FILE_MAPPING=(
    ["current-task.json"]="4-task"
    ["current-story.json"]="3-story" 
    ["current-epic.json"]="2-epic"
    ["stories.json"]="2-epic,3-story"
    ["epics.json"]="1-project"
    ["iterations.json"]="4-task"
    ["metrics.json"]="1-project"
)

# Commands that generate/update JSON files
declare -A COMMAND_JSON_MAPPING=(
    ["4-task/1-start/1-From-story.md"]="current-task"
    ["4-task/1-start/2-From-issue.md"]="current-task"
    ["4-task/1-start/3-From-input.md"]="current-task"
    ["4-task/2-execute/1-Plan-Task.md"]="current-task"
    ["4-task/2-execute/2-Test-design.md"]="current-task"
    ["4-task/2-execute/3-Implement.md"]="current-task"
    ["4-task/2-execute/4-Validate-Task.md"]="current-task"
    ["4-task/2-execute/5-Review-Task.md"]="current-task"
    ["4-task/3-complete/1-Archive-Task.md"]="current-task"
    ["4-task/3-complete/2-Status-Task.md"]="current-task"
    ["3-story/1-start/1-From-epic.md"]="current-story"
    ["3-story/2-execute/1-Plan-Story.md"]="current-story"
    ["3-story/3-complete/1-Archive-Story.md"]="current-story"
    ["2-epic/1-start/1-From-backlog.md"]="current-epic"
    ["2-epic/2-execute/1-Plan-Epic.md"]="current-epic"
    ["2-epic/3-complete/1-Archive-Epic.md"]="current-epic"
)

# Add validation hooks to a command file
add_validation_hooks() {
    local command_file="$1"
    local json_type="$2"
    
    if [[ ! -f "$command_file" ]]; then
        echo -e "${RED}Command file not found: $command_file${NC}"
        return 1
    fi
    
    # Check if already integrated
    if grep -q "JSON_SCHEMA_VALIDATION" "$command_file"; then
        echo -e "${YELLOW}Validation already integrated in: $command_file${NC}"
        return 0
    fi
    
    echo -e "${BLUE}Integrating validation hooks into: $command_file${NC}"
    
    # Create backup
    cp "$command_file" "${command_file}.backup"
    
    # Add schema enforcement section
    cat >> "$command_file" << EOF

## JSON Schema Validation Integration
<!-- JSON_SCHEMA_VALIDATION -->

### Pre-Generation Schema Guidance
Before generating or updating JSON files, Claude should use schema-aware prompts to prevent validation errors:

\`\`\`bash
# Show schema requirements to Claude
$ENFORCER show-requirements $json_type

# Generate schema-compliant content
$ENFORCER generate $json_type output-file.json "Task description"
\`\`\`

### Post-Generation Validation
After generating or updating JSON files, validate against schema:

\`\`\`bash
# Validate current context
$VALIDATOR validate

# Auto-correct if validation fails
$VALIDATOR validate docs/3-current-task/current-task.json
\`\`\`

### Schema-Aware Prompt Enhancement
When updating **docs/3-current-task/current-task.json** (or related files), include this in your prompt to Claude:

**MANDATORY SCHEMA COMPLIANCE:**
You MUST ensure the JSON strictly follows the schema requirements. Use the schema enforcer tool to understand requirements:

\`\`\`bash
$ENFORCER show-requirements $json_type
\`\`\`

Generate/update JSON that passes validation. All required fields must be present with correct types and values.

### Exit Code Enhancement
After completing the main task, run validation:

\`\`\`bash
# Validate and auto-correct if needed
if ! $VALIDATOR validate; then
    echo "JSON validation failed - files have been auto-corrected"
    exit 1  # Indicate iteration needed
fi
\`\`\`

<!-- /JSON_SCHEMA_VALIDATION -->
EOF

    echo -e "${GREEN}✓ Validation hooks added to: $command_file${NC}"
}

# Remove validation hooks from a command file
remove_validation_hooks() {
    local command_file="$1"
    
    if [[ ! -f "$command_file" ]]; then
        echo -e "${RED}Command file not found: $command_file${NC}"
        return 1
    fi
    
    if ! grep -q "JSON_SCHEMA_VALIDATION" "$command_file"; then
        echo -e "${YELLOW}No validation hooks found in: $command_file${NC}"
        return 0
    fi
    
    echo -e "${BLUE}Removing validation hooks from: $command_file${NC}"
    
    # Create backup
    cp "$command_file" "${command_file}.backup"
    
    # Remove validation section
    sed -i.tmp '/<!-- JSON_SCHEMA_VALIDATION -->/,/<!-- \/JSON_SCHEMA_VALIDATION -->/d' "$command_file"
    rm -f "${command_file}.tmp"
    
    echo -e "${GREEN}✓ Validation hooks removed from: $command_file${NC}"
}

# Integrate validation into all relevant commands
integrate_all_commands() {
    echo -e "${BLUE}=== Integrating JSON Schema Validation into Commands ===${NC}"
    
    local integrated_count=0
    local failed_count=0
    
    for command_path in "${!COMMAND_JSON_MAPPING[@]}"; do
        local json_type="${COMMAND_JSON_MAPPING[$command_path]}"
        local full_path="$COMMANDS_DIR/$command_path"
        
        if add_validation_hooks "$full_path" "$json_type"; then
            ((integrated_count++))
        else
            ((failed_count++))
        fi
    done
    
    echo ""
    echo -e "${GREEN}Integration complete:${NC}"
    echo -e "  ${GREEN}✓ Integrated: $integrated_count commands${NC}"
    if [[ $failed_count -gt 0 ]]; then
        echo -e "  ${RED}✗ Failed: $failed_count commands${NC}"
    fi
}

# Remove validation from all commands
remove_all_hooks() {
    echo -e "${BLUE}=== Removing JSON Schema Validation from Commands ===${NC}"
    
    local removed_count=0
    local failed_count=0
    
    for command_path in "${!COMMAND_JSON_MAPPING[@]}"; do
        local full_path="$COMMANDS_DIR/$command_path"
        
        if remove_validation_hooks "$full_path"; then
            ((removed_count++))
        else
            ((failed_count++))
        fi
    done
    
    echo ""
    echo -e "${GREEN}Removal complete:${NC}"
    echo -e "  ${GREEN}✓ Removed: $removed_count commands${NC}"
    if [[ $failed_count -gt 0 ]]; then
        echo -e "  ${RED}✗ Failed: $failed_count commands${NC}"
    fi
}

# Create master validation wrapper
create_validation_wrapper() {
    local wrapper_file="$COMMANDS_DIR/validate-json.sh"
    
    cat > "$wrapper_file" << 'EOF'
#!/bin/bash

# Master JSON Schema Validation Wrapper
# Provides unified interface for all validation operations

set -euo pipefail

TOOLS_DIR="$(dirname "$0")/tools"
VALIDATOR="$TOOLS_DIR/json-validator.sh"
ENFORCER="$TOOLS_DIR/schema-enforcer.sh"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Master validation function
validate_and_correct() {
    echo -e "${BLUE}=== Master JSON Schema Validation ===${NC}"
    
    # Install dependencies if needed
    if ! "$VALIDATOR" install; then
        echo -e "${RED}Failed to install validation dependencies${NC}"
        return 1
    fi
    
    # Run validation with auto-correction
    if "$VALIDATOR" validate; then
        echo -e "${GREEN}✓ All JSON files are valid${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ Some files required auto-correction${NC}"
        return 1
    fi
}

# Show all schema requirements
show_all_schemas() {
    echo -e "${BLUE}=== All Schema Requirements ===${NC}"
    
    local schemas=("current-task" "current-story" "current-epic" "stories" "epics" "iterations" "metrics")
    
    for schema in "${schemas[@]}"; do
        echo ""
        "$ENFORCER" show-requirements "$schema" 2>/dev/null || echo -e "${YELLOW}Schema not found: $schema${NC}"
    done
}

# Generate templates for all JSON types
generate_all_templates() {
    local template_dir="templates/generated"
    mkdir -p "$template_dir"
    
    echo -e "${BLUE}Generating all JSON templates...${NC}"
    
    local schemas=("current-task" "current-story" "current-epic" "stories" "epics" "iterations" "metrics")
    
    for schema in "${schemas[@]}"; do
        local template_file="$template_dir/${schema}-template.json"
        if "$ENFORCER" template "$schema" "$template_file"; then
            echo -e "${GREEN}✓ Generated: $template_file${NC}"
        else
            echo -e "${RED}✗ Failed: $template_file${NC}"
        fi
    done
}

# Main function
main() {
    case "${1:-help}" in
        "validate")
            validate_and_correct
            ;;
        "show-schemas")
            show_all_schemas
            ;;
        "generate-templates")
            generate_all_templates
            ;;
        "install")
            "$VALIDATOR" install
            ;;
        "help")
            cat << 'HELP'
Master JSON Schema Validation System

Usage: ./validate-json.sh [command]

Commands:
  validate          Validate all JSON files with auto-correction
  show-schemas      Display all schema requirements
  generate-templates Generate template files for all JSON types
  install           Install validation dependencies
  help              Show this help

This wrapper provides unified access to the complete validation system.
Use this as the main entry point for all validation operations.
HELP
            ;;
        *)
            echo -e "${RED}Unknown command: ${1:-}${NC}"
            echo "Use './validate-json.sh help' for usage information"
            exit 1
            ;;
    esac
}

main "$@"
EOF

    chmod +x "$wrapper_file"
    echo -e "${GREEN}✓ Master validation wrapper created: $wrapper_file${NC}"
}

# Show integration status
show_status() {
    echo -e "${BLUE}=== JSON Schema Validation Integration Status ===${NC}"
    echo ""
    
    local integrated_count=0
    local total_count=0
    
    for command_path in "${!COMMAND_JSON_MAPPING[@]}"; do
        local json_type="${COMMAND_JSON_MAPPING[$command_path]}"
        local full_path="$COMMANDS_DIR/$command_path"
        ((total_count++))
        
        if [[ -f "$full_path" ]]; then
            if grep -q "JSON_SCHEMA_VALIDATION" "$full_path"; then
                echo -e "${GREEN}✓ $command_path (${json_type})${NC}"
                ((integrated_count++))
            else
                echo -e "${RED}✗ $command_path (${json_type})${NC}"
            fi
        else
            echo -e "${YELLOW}? $command_path (file not found)${NC}"
        fi
    done
    
    echo ""
    echo -e "${BLUE}Summary: ${integrated_count}/${total_count} commands integrated${NC}"
    
    if [[ $integrated_count -eq $total_count ]]; then
        echo -e "${GREEN}All commands have validation integration${NC}"
    else
        echo -e "${YELLOW}Run 'integrate' to add validation to remaining commands${NC}"
    fi
}

# Main function
main() {
    local command="${1:-help}"
    
    case "$command" in
        "integrate")
            integrate_all_commands
            create_validation_wrapper
            ;;
        "remove")
            remove_all_hooks
            ;;
        "status")
            show_status
            ;;
        "create-wrapper")
            create_validation_wrapper
            ;;
        "help")
            cat << EOF
Command Integration System for JSON Schema Validation

Usage: $0 <command>

Commands:
  integrate        Add validation hooks to all relevant commands
  remove           Remove validation hooks from all commands
  status           Show integration status
  create-wrapper   Create master validation wrapper script
  help             Show this help

Integration adds:
- Pre-generation schema guidance
- Post-generation validation
- Auto-correction on validation failure
- Schema-aware prompt instructions

The system ensures all JSON files generated by commands are valid
and automatically corrects them if validation fails.
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