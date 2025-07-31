#!/bin/bash

# Enhanced Post-Write JSON Validator Hook with Schema Validation
# Validates JSON files against their schemas using proper JSON Schema validators

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Path to schemas directory
SCHEMAS_DIR="internal/config/system/commands/templates/schemas"

# Schema mapping for JSON files
declare -A SCHEMA_MAP=(
    ["epics.json"]="epics.schema.json"
    ["stories.json"]="stories.schema.json"
    ["current-story.json"]="current-story.schema.json"
    ["current-epic.json"]="current-epic.schema.json"
    ["current-task.json"]="current-task.schema.json"
    ["iterations.json"]="iterations.schema.json"
    ["metrics.json"]="metrics.schema.json"
)

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
6. La structure DOIT correspondre exactement au schema (objets vs tableaux)

Corrige le fichier et assure-toi qu'il est parfaitement conforme au schema."

    # Try to call Claude Code for correction
    if command -v claude >/dev/null 2>&1; then
        echo -e "${BLUE}📝 Calling Claude Code for JSON correction...${NC}" >&2
        echo "$correction_prompt" | claude --project-path="$(dirname "$file_path")" --file="$file_path"
        
        # Check if correction was successful
        if validate_json_with_schema "$file_path"; then
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

# Install validator if needed
ensure_validator_installed() {
    # Try Python jsonschema first (most reliable)
    if command -v python3 >/dev/null 2>&1; then
        if python3 -c "import jsonschema" 2>/dev/null; then
            echo "python-jsonschema"
            return 0
        else
            echo -e "${YELLOW}📦 Installing jsonschema Python library...${NC}" >&2
            if pip3 install jsonschema >/dev/null 2>&1; then
                echo "python-jsonschema"
                return 0
            fi
        fi
    fi
    
    # Fallback to ajv-cli (Node.js)
    if command -v npx >/dev/null 2>&1; then
        echo "ajv-cli"
        return 0
    fi
    
    # No validator available
    echo -e "${RED}❌ No JSON Schema validator available${NC}" >&2
    echo -e "${YELLOW}💡 Install jsonschema: pip3 install jsonschema${NC}" >&2
    echo -e "${YELLOW}💡 Or install ajv-cli: npm install -g ajv-cli${NC}" >&2
    return 1
}

# Validate JSON against schema using Python jsonschema
validate_with_python_jsonschema() {
    local json_file="$1"
    local schema_file="$2"
    
    python3 -c "
import json
import jsonschema
import sys

try:
    with open('$json_file', 'r') as f:
        data = json.load(f)
    
    with open('$schema_file', 'r') as f:
        schema = json.load(f)
    
    jsonschema.validate(data, schema)
    print('✅ Schema validation passed')
    sys.exit(0)
    
except jsonschema.ValidationError as e:
    print(f'❌ Schema validation failed: {e.message}', file=sys.stderr)
    if hasattr(e, 'path') and e.path:
        print(f'   Path: {\" > \".join(map(str, e.path))}', file=sys.stderr)
    if e.schema_path:
        print(f'   Schema path: {\" > \".join(map(str, e.schema_path))}', file=sys.stderr)
    sys.exit(1)
    
except json.JSONDecodeError as e:
    print(f'❌ Invalid JSON: {e}', file=sys.stderr)
    sys.exit(1)
    
except Exception as e:
    print(f'❌ Validation error: {e}', file=sys.stderr)
    sys.exit(1)
"
}

# Main schema validation function
validate_json_with_schema() {
    local file_path="$1"
    local filename=$(basename "$file_path")
    
    # Check if we have a schema for this file
    if [[ ! "${SCHEMA_MAP[$filename]+_}" ]]; then
        echo -e "${YELLOW}⚠ No schema defined for $filename, skipping schema validation${NC}" >&2
        return 0
    fi
    
    local schema_name="${SCHEMA_MAP[$filename]}"
    local schema_path="$SCHEMAS_DIR/$schema_name"
    
    # Check if schema file exists
    if [[ ! -f "$schema_path" ]]; then
        echo -e "${RED}❌ Schema file not found: $schema_path${NC}" >&2
        return 1
    fi
    
    # Determine validator to use
    local validator
    if ! validator=$(ensure_validator_installed); then
        return 1
    fi
    
    echo -e "${BLUE}🔍 Validating $filename against $schema_name${NC}" >&2
    
    # Validate using appropriate validator
    case $validator in
        "python-jsonschema")
            if validate_with_python_jsonschema "$file_path" "$schema_path"; then
                return 0
            else
                return 1
            fi
            ;;
        "ajv-cli")
            if npx ajv validate -s "$schema_path" -d "$file_path" 2>&1; then
                return 0
            else
                return 1
            fi
            ;;
        *)
            echo -e "${RED}❌ Unknown validator: $validator${NC}" >&2
            return 1
            ;;
    esac
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

echo -e "${BLUE}🔍 Enhanced JSON Validator: Checking $WRITTEN_FILE${NC}" >&2

# Basic JSON syntax validation
if ! python3 -m json.tool "$WRITTEN_FILE" >/dev/null 2>&1; then
    echo -e "${RED}❌ Invalid JSON syntax in $WRITTEN_FILE${NC}" >&2
    if ! auto_correct_json "$WRITTEN_FILE" "Invalid JSON syntax"; then
        exit 1
    fi
fi

# Enhanced schema validation with auto-correction
if ! validate_json_with_schema "$WRITTEN_FILE"; then
    echo -e "${RED}❌ SCHEMA VALIDATION FAILED${NC}" >&2
    
    if ! auto_correct_json "$WRITTEN_FILE" "Schema validation failed"; then
        exit 1
    fi
fi

echo -e "${GREEN}✅ Enhanced JSON validation completed successfully for $WRITTEN_FILE${NC}"
exit 0