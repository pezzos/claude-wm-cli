#!/bin/bash

# Post-Write JSON Validator Hook
# Automatically validates JSON files after Write operations and applies auto-correction if needed

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get the file path from environment or command line
WRITTEN_FILE="${CLAUDE_TOOL_WRITE_FILE_PATH:-$1}"

# Check if we have a file path
if [[ -z "$WRITTEN_FILE" ]]; then
    echo -e "${YELLOW}⚠ PostToolUse JSON Validator: No file path provided${NC}" >&2
    exit 0
fi

# Check if file exists
if [[ ! -f "$WRITTEN_FILE" ]]; then
    echo -e "${YELLOW}⚠ PostToolUse JSON Validator: File does not exist: $WRITTEN_FILE${NC}" >&2
    exit 0
fi

# Check if it's a JSON file
if [[ ! "$WRITTEN_FILE" =~ \.json$ ]]; then
    # Not a JSON file, skip validation
    exit 0
fi

# Get the project root directory
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
TOOLS_DIR="$PROJECT_ROOT/.claude-wm/.claude/commands/tools"

# Check if we have the validation tools
if [[ ! -f "$TOOLS_DIR/simple-validator.sh" ]] || [[ ! -f "$TOOLS_DIR/json-validator.sh" ]]; then
    echo -e "${YELLOW}⚠ PostToolUse JSON Validator: Validation tools not found in $TOOLS_DIR${NC}" >&2
    exit 0
fi

echo -e "${BLUE}🔍 PostToolUse JSON Validator: Checking $WRITTEN_FILE${NC}"

# Detect schema type from filename
SCHEMA_TYPE=""
case "$(basename "$WRITTEN_FILE")" in
    "epics.json")
        SCHEMA_TYPE="epics"
        ;;
    "stories.json")
        SCHEMA_TYPE="stories"
        ;;
    "current-epic.json")
        SCHEMA_TYPE="current-epic"
        ;;
    "current-story.json")
        SCHEMA_TYPE="current-story"
        ;;
    "current-task.json")
        SCHEMA_TYPE="current-task"
        ;;
    *)
        echo -e "${YELLOW}⚠ No schema validation available for $(basename "$WRITTEN_FILE")${NC}"
        exit 0
        ;;
esac

# Validate the JSON file with appropriate schema
if "$TOOLS_DIR/simple-validator.sh" validate-schema "$SCHEMA_TYPE" "$WRITTEN_FILE" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ JSON validation passed for $WRITTEN_FILE${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ JSON validation failed for $WRITTEN_FILE - attempting auto-correction${NC}" >&2
    
    # Try auto-correction
    if "$TOOLS_DIR/json-validator.sh" validate "$WRITTEN_FILE"; then
        echo -e "${GREEN}✓ JSON auto-correction completed for $WRITTEN_FILE${NC}"
        # Re-validate after correction
        if "$TOOLS_DIR/simple-validator.sh" validate-schema "$SCHEMA_TYPE" "$WRITTEN_FILE" >/dev/null 2>&1; then
            echo -e "${GREEN}✓ JSON validation now passes after auto-correction${NC}"
            exit 0
        else
            echo -e "${RED}❌ JSON validation still fails after auto-correction${NC}" >&2
            exit 1
        fi
    else
        echo -e "${RED}❌ JSON auto-correction failed for $WRITTEN_FILE${NC}" >&2
        echo -e "${YELLOW}Manual review required for schema compliance${NC}" >&2
        exit 1
    fi
fi