#!/bin/bash

# Serena Incremental Indexing Script
# Calculates SHA256 for docs/ files and indexes only changed content

set -euo pipefail

# Configuration
DOCS_DIR="docs"
SERENA_DIR=".serena"
MANIFEST_FILE="$SERENA_DIR/manifest.json"
TEMP_MANIFEST="$SERENA_DIR/manifest.tmp.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# Create .serena directory if it doesn't exist
mkdir -p "$SERENA_DIR"

# Initialize manifest if it doesn't exist
if [[ ! -f "$MANIFEST_FILE" ]]; then
    log_info "Initializing manifest file..."
    cat > "$MANIFEST_FILE" << EOF
{
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "docs": []
}
EOF
fi

# Function to extract category from file path
get_category() {
    local file_path="$1"
    if [[ "$file_path" == *"/KB/"* ]]; then
        echo "KB"
    elif [[ "$file_path" == *"/ADR/"* ]]; then
        echo "ADR"
    elif [[ "$file_path" == *".md" ]]; then
        echo "guide"
    else
        echo "other"
    fi
}

# Function to extract title from markdown file
get_title() {
    local file_path="$1"
    local title
    
    # Try to get title from first # heading
    title=$(head -20 "$file_path" 2>/dev/null | grep -E '^# ' | head -1 | sed 's/^# //' || echo "")
    
    # Fallback to filename if no title found
    if [[ -z "$title" ]]; then
        title=$(basename "$file_path" .md)
    fi
    
    echo "$title"
}

# Function to generate tags from file path and content
get_tags() {
    local file_path="$1"
    local category="$2"
    local tags=()
    
    # Add category as primary tag
    tags+=("$category")
    
    # Add specific tags based on path
    case "$file_path" in
        *"/glossary.md") tags+=("terminology" "reference") ;;
        *"/commands.md") tags+=("cli" "reference" "commands") ;;
        *"/file-ownership.md") tags+=("security" "permissions" "boundaries") ;;
        *"/mcp-playbook.md") tags+=("mcp" "tools" "integration" "playbook") ;;
        *"ARCHITECTURE.md") tags+=("architecture" "system-design") ;;
        *"CONFIG_GUIDE.md") tags+=("configuration" "setup" "guide") ;;
        *"TESTING.md") tags+=("testing" "qa" "protocol") ;;
        *"/ADR/"*) tags+=("decision" "architecture") ;;
    esac
    
    # Join tags with commas
    IFS=',' ; echo "${tags[*]}"
}

log_info "Starting Serena incremental indexing..."
log_info "Scanning $DOCS_DIR for documentation files..."

# Find all documentation files
doc_files=()
while IFS= read -r -d '' file; do
    doc_files+=("$file")
done < <(find "$DOCS_DIR" -type f \( -name "*.md" -o -name "*.json" \) -print0 | sort -z)

if [[ ${#doc_files[@]} -eq 0 ]]; then
    log_warning "No documentation files found in $DOCS_DIR"
    exit 0
fi

log_info "Found ${#doc_files[@]} documentation files"

# Load existing manifest
current_manifest=$(cat "$MANIFEST_FILE")

# Initialize counters
new_files=0
modified_files=0
unchanged_files=0
files_to_index=()

# Check each file
for file in "${doc_files[@]}"; do
    if [[ ! -f "$file" ]]; then
        continue
    fi
    
    # Calculate current SHA256
    current_sha=$(shasum -a 256 "$file" | cut -d' ' -f1)
    
    # Get file info from existing manifest
    existing_sha=$(echo "$current_manifest" | jq -r --arg path "$file" '.docs[] | select(.path == $path) | .sha // empty')
    
    if [[ -z "$existing_sha" ]]; then
        log_info "New file: $file"
        files_to_index+=("$file")
        ((new_files++))
    elif [[ "$existing_sha" != "$current_sha" ]]; then
        log_info "Modified file: $file"
        files_to_index+=("$file")
        ((modified_files++))
    else
        ((unchanged_files++))
    fi
done

# Summary of changes
log_info "File analysis complete:"
echo "  📄 New files: $new_files"
echo "  🔄 Modified files: $modified_files"
echo "  ✅ Unchanged files: $unchanged_files"

if [[ ${#files_to_index[@]} -eq 0 ]]; then
    log_success "No changes detected. Index is up to date."
    exit 0
fi

log_info "Updating manifest for ${#files_to_index[@]} files..."

# Create new manifest
new_docs_array="[]"

# Keep unchanged entries from existing manifest
for file in "${doc_files[@]}"; do
    if [[ ! -f "$file" ]]; then
        continue
    fi
    
    current_sha=$(shasum -a 256 "$file" | cut -d' ' -f1)
    existing_entry=$(echo "$current_manifest" | jq --arg path "$file" '.docs[] | select(.path == $path)')
    existing_sha=$(echo "$existing_entry" | jq -r '.sha // empty')
    
    if [[ "$existing_sha" == "$current_sha" ]]; then
        # File unchanged, keep existing entry
        new_docs_array=$(echo "$new_docs_array" | jq ". + [$existing_entry]")
    else
        # File new or modified, create new entry
        category=$(get_category "$file")
        title=$(get_title "$file")
        tags=$(get_tags "$file" "$category")
        
        new_entry=$(jq -n \
            --arg path "$file" \
            --arg title "$title" \
            --arg category "$category" \
            --arg tags "$tags" \
            --arg sha "$current_sha" \
            --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
            '{
                path: $path,
                title: $title,
                category: $category,
                tags: ($tags | split(",")),
                sha: $sha,
                indexed_at: $timestamp
            }')
        
        new_docs_array=$(echo "$new_docs_array" | jq ". + [$new_entry]")
    fi
done

# Create updated manifest
updated_manifest=$(jq \
    --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
    --argjson docs "$new_docs_array" \
    '.timestamp = $timestamp | .docs = $docs' \
    <<< "$current_manifest")

# Write updated manifest atomically
echo "$updated_manifest" > "$TEMP_MANIFEST"
mv "$TEMP_MANIFEST" "$MANIFEST_FILE"

log_success "Manifest updated successfully!"

# Display indexing summary
echo ""
log_info "📊 Indexing Summary:"
echo "  📁 Total files in manifest: $(echo "$updated_manifest" | jq '.docs | length')"
echo "  🆕 New files indexed: $new_files"
echo "  🔄 Modified files updated: $modified_files"
echo "  ⏰ Last updated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"

# Optional: Display files that were indexed
if [[ ${#files_to_index[@]} -gt 0 ]]; then
    echo ""
    log_info "Files processed in this run:"
    for file in "${files_to_index[@]}"; do
        category=$(get_category "$file")
        title=$(get_title "$file")
        echo "  📝 $file [$category] \"$title\""
    done
fi

log_success "Serena indexing completed successfully!"