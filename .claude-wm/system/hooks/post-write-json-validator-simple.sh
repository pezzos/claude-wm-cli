#!/bin/bash
# Simple JSON validator for post-write operations
# Usage: post-write-json-validator-simple.sh <file_path>

FILE_PATH="$1"

if [[ -z "$FILE_PATH" ]]; then
    exit 0
fi

# Only validate JSON files
if [[ "$FILE_PATH" == *.json ]]; then
    if command -v jq >/dev/null 2>&1; then
        if ! jq empty "$FILE_PATH" >/dev/null 2>&1; then
            echo "⚠️  JSON validation failed for $FILE_PATH" >&2
            exit 1
        else
            echo "✅ JSON validation passed for $FILE_PATH"
        fi
    else
        # Fallback to basic Python validation if jq not available
        if command -v python3 >/dev/null 2>&1; then
            if ! python3 -m json.tool "$FILE_PATH" >/dev/null 2>&1; then
                echo "⚠️  JSON validation failed for $FILE_PATH" >&2
                exit 1
            else
                echo "✅ JSON validation passed for $FILE_PATH"
            fi
        fi
    fi
fi

exit 0
