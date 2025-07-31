#!/bin/bash
# Detects obsolete files in the project
# Usage: obsolete-file-detector.sh

echo "🔍 Checking for obsolete files..."

# Common obsolete file patterns
OBSOLETE_PATTERNS=(
    "*.tmp"
    "*.bak"
    "*.old"
    "*~"
    ".DS_Store"
    "Thumbs.db"
    "*.log"
    "node_modules/.cache"
)

FOUND_OBSOLETE=false

for pattern in "${OBSOLETE_PATTERNS[@]}"; do
    while IFS= read -r -d '' file; do
        if [[ -f "$file" ]]; then
            echo "⚠️  Found obsolete file: $file"
            FOUND_OBSOLETE=true
        fi
    done < <(find . -name "$pattern" -type f -print0 2>/dev/null)
done

if [[ "$FOUND_OBSOLETE" == "false" ]]; then
    echo "✅ No obsolete files detected"
fi

exit 0
