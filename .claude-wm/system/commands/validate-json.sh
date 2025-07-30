#!/bin/bash

# Master JSON Schema Validation System
# Unified interface for all validation operations

set -eo pipefail

TOOLS_DIR="$(dirname "$0")/tools" 
SIMPLE_VALIDATOR="$TOOLS_DIR/simple-validator.sh"
SCHEMA_ENFORCER="$TOOLS_DIR/schema-enforcer.sh"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Main validation function
validate_all() {
    echo -e "${BLUE}=== Master JSON Schema Validation ===${NC}"
    
    if "$SIMPLE_VALIDATOR" validate; then
        echo -e "${GREEN}✓ All JSON files are schema-compliant${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ Some files have validation issues${NC}"
        echo -e "${BLUE}Run individual file validation for details${NC}"
        return 1
    fi
}

# Show all schema requirements
show_schemas() {
    echo -e "${BLUE}=== All Schema Requirements ===${NC}"
    
    local schemas=("current-task" "current-story" "current-epic" "stories" "epics" "iterations" "metrics")
    
    for schema in "${schemas[@]}"; do
        echo ""
        "$SIMPLE_VALIDATOR" show-schema "$schema" 2>/dev/null || echo -e "${YELLOW}Schema not available: $schema${NC}"
    done
}

# Validate specific file
validate_file() {
    local file="$1"
    if [[ -z "$file" ]]; then
        echo "Usage: validate-file <json_file>"
        return 1
    fi
    
    "$SIMPLE_VALIDATOR" validate-file "$file"
}

# Show usage help
show_help() {
    cat << EOF
Master JSON Schema Validation System for Claude WM CLI

Usage: $0 [command] [options]

Commands:
  validate              Validate all JSON files (default)
  validate-file <file>  Validate specific JSON file
  show-schemas          Display all schema requirements
  enforce <type>        Show proactive schema enforcement for type
  status               Show validation system status
  help                 Show this help

Examples:
  $0                                    # Validate all files
  $0 validate-file current-task.json   # Validate specific file
  $0 show-schemas                      # Show all schema requirements
  $0 enforce current-task              # Show enforcement guidance

Schema Types:
  current-task, current-story, current-epic, stories, epics, iterations, metrics

The validation system:
1. Validates JSON syntax using jq
2. Checks required fields from schema
3. Validates enum values against allowed lists
4. Provides clear error messages for fixing issues
5. Integrated into all command templates for automatic validation

For proactive schema compliance, use the schema-enforcer.sh tool to generate
schema-aware prompts before creating or updating JSON files.
EOF
}

# Show enforcement guidance
show_enforcement() {
    local json_type="$1"
    if [[ -z "$json_type" ]]; then
        echo "Usage: enforce <json_type>"
        return 1
    fi
    
    echo -e "${BLUE}=== Proactive Schema Enforcement for $json_type ===${NC}"
    echo ""
    echo -e "${YELLOW}Before generating/updating JSON:${NC}"
    echo "1. Show requirements: $SCHEMA_ENFORCER show-requirements $json_type"
    echo "2. Generate template: $SCHEMA_ENFORCER template $json_type"
    echo ""
    echo -e "${YELLOW}During Claude interaction:${NC}"
    echo "Include this in your prompt:"
    echo ""
    echo "**CRITICAL: SCHEMA COMPLIANCE REQUIRED**"
    echo "The JSON must strictly follow the schema requirements."
    echo "Show requirements: $SCHEMA_ENFORCER show-requirements $json_type"
    echo ""
    echo -e "${YELLOW}After generation:${NC}"
    echo "Validate: $0 validate-file <generated_file>"
    echo ""
    
    # Show actual requirements
    "$SIMPLE_VALIDATOR" show-schema "$json_type"
}

# Show system status
show_status() {
    echo -e "${BLUE}=== JSON Schema Validation System Status ===${NC}"
    echo ""
    
    # Check if tools exist
    local tools_status=0
    
    if [[ -f "$SIMPLE_VALIDATOR" ]]; then
        echo -e "${GREEN}✓ Simple validator available${NC}"
    else
        echo -e "${RED}✗ Simple validator missing${NC}"
        ((tools_status++))
    fi
    
    if [[ -f "$SCHEMA_ENFORCER" ]]; then
        echo -e "${GREEN}✓ Schema enforcer available${NC}"
    else
        echo -e "${RED}✗ Schema enforcer missing${NC}"
        ((tools_status++))
    fi
    
    # Check schemas
    echo ""
    echo -e "${BLUE}Available Schemas:${NC}"
    local schema_dir=".claude-wm/.claude/commands/templates/schemas"
    if [[ -d "$schema_dir" ]]; then
        local schema_count=0
        for schema_file in "$schema_dir"/*.schema.json; do
            if [[ -f "$schema_file" ]]; then
                local schema_name
                schema_name=$(basename "$schema_file" .schema.json)
                echo -e "${GREEN}✓ $schema_name${NC}"
                ((schema_count++))
            fi
        done
        echo "Total schemas: $schema_count"
    else
        echo -e "${RED}✗ Schema directory not found${NC}"
    fi
    
    # Check command integration
    echo ""
    echo -e "${BLUE}Command Integration Status:${NC}"
    ./.claude-wm/.claude/commands/tools/integrate-validation.sh status
    
    echo ""
    if [[ $tools_status -eq 0 ]]; then
        echo -e "${GREEN}✓ Validation system is ready${NC}"
    else
        echo -e "${YELLOW}⚠ Some components are missing${NC}"
    fi
}

# Main function
main() {
    case "${1:-validate}" in
        "validate")
            validate_all
            ;;
        "validate-file")
            validate_file "$2"
            ;;
        "show-schemas")
            show_schemas
            ;;
        "enforce")
            show_enforcement "$2"
            ;;
        "status")
            show_status
            ;;
        "help")
            show_help
            ;;
        *)
            echo -e "${RED}Unknown command: $1${NC}"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

main "$@"