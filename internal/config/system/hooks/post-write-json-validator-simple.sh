#!/bin/bash

# Enhanced Post-Write JSON Validator Hook with Auto-Correction
# Validates JSON files against schemas and auto-corrects using Claude Code

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Auto-correction function using Claude
auto_correct_json() {
    local file_path="$1"
    local schema_issue="$2"
    
    echo -e "${YELLOW}🔧 Attempting auto-correction with Claude Code...${NC}" >&2
    
    # Create correction prompt
    local correction_prompt="Corrige ce fichier JSON pour qu'il respecte scrupuleusement son schema JSON.

PROBLÈME DÉTECTÉ: $schema_issue

REQUIREMENTS CRITIQUES:
1. Le fichier DOIT contenir le champ '\$schema' avec le chemin relatif vers le schema approprié
2. Tous les champs requis par le schema doivent être présents
3. Les types de données doivent correspondre exactement au schema
4. Les valeurs enum doivent être parmi les valeurs autorisées
5. Les patterns regex doivent être respectés

Corrige le fichier et assure-toi qu'il est parfaitement conforme au schema."

    # Try to call Claude Code for correction
    if command -v claude >/dev/null 2>&1; then
        echo -e "${BLUE}📝 Calling Claude Code for JSON correction...${NC}" >&2
        echo "$correction_prompt" | claude --project-path="$(dirname "$file_path")" --file="$file_path"
        
        # Check if correction was successful
        if validate_json_schema "$file_path"; then
            echo -e "${GREEN}✅ Auto-correction successful!${NC}" >&2
            return 0
        else
            echo -e "${RED}❌ Auto-correction failed${NC}" >&2
            return 1
        fi
    else
        echo -e "${RED}❌ Claude Code CLI not available for auto-correction${NC}" >&2
        return 1
    fi
}

# Schema validation function
validate_json_schema() {
    local file_path="$1"
    local filename=$(basename "$file_path")
    
    # Check if file has $schema field
    if ! grep -q '"\$schema"' "$file_path"; then
        echo -e "${RED}❌ Missing required '\$schema' field${NC}" >&2
        return 1
    fi
    
    # Additional validations based on filename
    case "$filename" in
        "epics.json")
            if ! grep -q '"epics"' "$file_path"; then
                echo -e "${RED}❌ Missing required 'epics' field${NC}" >&2
                return 1
            fi
            if ! grep -q '"project_context"' "$file_path"; then
                echo -e "${RED}❌ Missing required 'project_context' field${NC}" >&2
                return 1
            fi
            if grep -q '"userStories"' "$file_path"; then
                echo -e "${RED}❌ Forbidden 'userStories' field found${NC}" >&2
                return 1
            fi
            ;;
        "stories.json")
            if ! grep -q '"stories"' "$file_path"; then
                echo -e "${RED}❌ Missing required 'stories' field${NC}" >&2
                return 1
            fi
            if ! grep -q '"epic_context"' "$file_path"; then
                echo -e "${RED}❌ Missing required 'epic_context' field${NC}" >&2
                return 1
            fi
            ;;
        "current-story.json")
            if ! grep -q '"story"' "$file_path"; then
                echo -e "${RED}❌ Missing required 'story' field${NC}" >&2
                return 1
            fi
            ;;
        "current-epic.json")
            if ! grep -q '"epic"' "$file_path"; then
                echo -e "${RED}❌ Missing required 'epic' field${NC}" >&2
                return 1
            fi
            ;;
        "current-task.json")
            required_fields=("id" "title" "description" "type" "priority" "status")
            for field in "${required_fields[@]}"; do
                if ! grep -q "\"$field\"" "$file_path"; then
                    echo -e "${RED}❌ Missing required field '$field'${NC}" >&2
                    return 1
                fi
            done
            ;;
        "iterations.json")
            required_fields=("task_context" "iterations" "final_outcome" "recommendations")
            for field in "${required_fields[@]}"; do
                if ! grep -q "\"$field\"" "$file_path"; then
                    echo -e "${RED}❌ Missing required field '$field'${NC}" >&2
                    return 1
                fi
            done
            ;;
        "metrics.json")
            required_fields=("project_overview" "current_epic" "iteration_performance" "time_analytics" "quality_metrics" "team_performance" "trend_indicators" "last_updated")
            for field in "${required_fields[@]}"; do
                if ! grep -q "\"$field\"" "$file_path"; then
                    echo -e "${RED}❌ Missing required field '$field'${NC}" >&2
                    return 1
                fi
            done
            ;;
    esac
    
    return 0
}

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

# Basic JSON syntax validation
if ! python3 -m json.tool "$WRITTEN_FILE" >/dev/null 2>&1; then
    echo -e "${RED}❌ Invalid JSON syntax in $WRITTEN_FILE${NC}" >&2
    if ! auto_correct_json "$WRITTEN_FILE" "Invalid JSON syntax"; then
        exit 1
    fi
fi

# Enhanced schema validation with auto-correction
if ! validate_json_schema "$WRITTEN_FILE"; then
    echo -e "${RED}❌ SCHEMA VALIDATION FAILED${NC}" >&2
    
    if ! auto_correct_json "$WRITTEN_FILE" "Schema validation failed"; then
        exit 1
    fi
fi

echo -e "${GREEN}✅ JSON validation completed successfully for $WRITTEN_FILE${NC}"
exit 0