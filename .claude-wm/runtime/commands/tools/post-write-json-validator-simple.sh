#!/bin/bash

# Simple Post-Write JSON Validator Hook 
# Uses basic JSON parsing and schema field validation

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

echo -e "${BLUE}🔍 PostToolUse JSON Validator: Checking $WRITTEN_FILE${NC}" >&2
echo -e "${YELLOW}DEBUG: Hook executed for file: $WRITTEN_FILE${NC}" >&2
echo -e "${YELLOW}DEBUG: Timestamp: $(date)${NC}" >&2
echo -e "${YELLOW}DEBUG: Working directory: $(pwd)${NC}" >&2

# Basic JSON syntax validation
if ! python3 -m json.tool "$WRITTEN_FILE" >/dev/null 2>&1; then
    echo -e "${RED}❌ Invalid JSON syntax in $WRITTEN_FILE${NC}" >&2
    exit 1
fi

# Detect schema type from filename and validate specific rules
FILENAME=$(basename "$WRITTEN_FILE")
case "$FILENAME" in
    "epics.json")
        echo -e "${BLUE}📋 Validating epics.json schema compliance...${NC}"
        
        # Check for forbidden userStories field
        if grep -q '"userStories"' "$WRITTEN_FILE"; then
            echo -e "${RED}❌ SCHEMA VIOLATION: 'userStories' field is not allowed in epics.json${NC}" >&2
            echo -e "${YELLOW}💡 Stories should be defined in stories.json, not embedded in epics${NC}" >&2
            echo -e "${YELLOW}💡 Please remove the 'userStories' fields from all epics${NC}" >&2
            exit 1
        fi
        
        # Check for required fields
        if ! grep -q '"epics"' "$WRITTEN_FILE"; then
            echo -e "${RED}❌ SCHEMA VIOLATION: Missing required 'epics' array${NC}" >&2
            exit 1
        fi
        
        if ! grep -q '"project_context"' "$WRITTEN_FILE"; then
            echo -e "${RED}❌ SCHEMA VIOLATION: Missing required 'project_context' object${NC}" >&2
            exit 1
        fi
        
        echo -e "${GREEN}✓ epics.json schema validation passed${NC}"
        ;;
        
    "stories.json")
        echo -e "${BLUE}📋 Validating stories.json schema compliance...${NC}"
        
        # Check for required fields
        if ! grep -q '"stories"' "$WRITTEN_FILE"; then
            echo -e "${RED}❌ SCHEMA VIOLATION: Missing required 'stories' array${NC}" >&2
            exit 1
        fi
        
        echo -e "${GREEN}✓ stories.json schema validation passed${NC}"
        ;;
        
    "current-story.json")
        echo -e "${BLUE}📋 Validating current-story.json schema compliance...${NC}"
        
        # Check for required top-level structure
        if ! grep -q '"story"' "$WRITTEN_FILE"; then
            echo -e "${RED}❌ SCHEMA VIOLATION: Missing required 'story' object${NC}" >&2
            exit 1
        fi
        
        # Check for required story fields
        required_fields=("id" "title" "description" "epic_id" "epic_title" "status" "priority")
        for field in "${required_fields[@]}"; do
            if ! grep -q "\"$field\"" "$WRITTEN_FILE"; then
                echo -e "${RED}❌ SCHEMA VIOLATION: Missing required field '$field' in story object${NC}" >&2
                exit 1
            fi
        done
        
        # Validate story ID format
        if ! grep -q '"id": "STORY-[0-9][0-9][0-9]"' "$WRITTEN_FILE"; then
            echo -e "${RED}❌ SCHEMA VIOLATION: Story ID must follow format STORY-XXX${NC}" >&2
            exit 1
        fi
        
        # Validate epic ID format
        if ! grep -q '"epic_id": "EPIC-[0-9][0-9][0-9]"' "$WRITTEN_FILE"; then
            echo -e "${RED}❌ SCHEMA VIOLATION: Epic ID must follow format EPIC-XXX${NC}" >&2
            exit 1
        fi
        
        # Validate status values
        if ! grep -qE '"status": "(todo|in_progress|done|blocked)"' "$WRITTEN_FILE"; then
            echo -e "${RED}❌ SCHEMA VIOLATION: Status must be one of: todo, in_progress, done, blocked${NC}" >&2
            exit 1
        fi
        
        # Validate priority values
        if ! grep -qE '"priority": "(low|medium|high|critical)"' "$WRITTEN_FILE"; then
            echo -e "${RED}❌ SCHEMA VIOLATION: Priority must be one of: low, medium, high, critical${NC}" >&2
            exit 1
        fi
        
        echo -e "${GREEN}✓ current-story.json schema validation passed${NC}"
        ;;
        
    "current-epic.json"|"current-task.json")
        echo -e "${BLUE}📋 Validating current-*.json schema compliance...${NC}"
        echo -e "${GREEN}✓ Basic JSON syntax validation passed${NC}"
        ;;
        
    *)
        echo -e "${YELLOW}⚠ No specific schema validation for $FILENAME${NC}"
        echo -e "${GREEN}✓ Basic JSON syntax validation passed${NC}"
        ;;
esac

echo -e "${GREEN}✅ JSON validation completed successfully for $WRITTEN_FILE${NC}"
exit 0