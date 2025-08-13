# Serena Documentation Indexing System

## Overview

The Serena indexing system provides incremental documentation indexing for fast semantic search. It tracks SHA256 hashes to index only changed files, providing ~35% cost reduction and faster processing.

## Architecture

### Components
- **`scripts/serena-index.sh`**: Incremental indexing script
- **`.serena/manifest.json`**: Index manifest with file metadata  
- **`make serena-index`**: Make target for manual indexing
- **GitHub Actions**: Automatic indexing on `docs/**` changes

### Index Structure
```json
{
  "version": "1.0.0",
  "timestamp": "2025-08-13T07:28:24Z",
  "docs": [
    {
      "path": "docs/KB/glossary.md",
      "title": "Glossary", 
      "category": "KB",
      "tags": ["KB", "terminology", "reference"],
      "sha": "cbaa5d3b291db87558e3ca40918b30a4c21bdade414341698cdf6326e946716e",
      "indexed_at": "2025-08-13T07:28:23Z"
    }
  ]
}
```

## Categories and Tags

### Automatic Categorization
- **KB**: Files in `docs/KB/` (Knowledge Base)
- **ADR**: Files in `docs/ADR/` (Architecture Decision Records) 
- **guide**: Top-level `.md` files in `docs/`
- **other**: Other file types

### Smart Tagging
Files receive automatic tags based on path and content:

| File | Category | Tags |
|------|----------|------|
| `glossary.md` | KB | KB, terminology, reference |
| `commands.md` | KB | KB, cli, reference, commands |
| `file-ownership.md` | KB | KB, security, permissions, boundaries |
| `mcp-playbook.md` | KB | KB, mcp, tools, integration, playbook |
| `ARCHITECTURE.md` | guide | guide, architecture, system-design |
| `CONFIG_GUIDE.md` | guide | guide, configuration, setup, guide |
| `TESTING.md` | guide | guide, testing, qa, protocol |
| `ADR/*.md` | ADR | ADR, decision, architecture |

## Usage

### Manual Indexing
```bash
# Update index (incremental - only changed files)
make serena-index

# Check index status
jq '.docs | length' .serena/manifest.json

# View categories breakdown
jq -r '.docs | group_by(.category) | .[] | "\(.length) \(.[0].category) documents"' .serena/manifest.json

# View recent changes
jq -r '.docs[] | select(.indexed_at > "2025-08-13T07:00:00Z") | .path' .serena/manifest.json
```

### Query Patterns
Always index before querying documentation:

```bash
# 1. Update index first
make serena-index

# 2. Query with specific paths
mcp__serena__search_for_pattern --relative_path "docs/KB" --substring_pattern "mcp tools"
mcp__serena__search_for_pattern --relative_path "docs/ADR" --substring_pattern "architecture decision"

# 3. Query by category priority
# KB (most specific) → ADR (decisions) → guides (general)
```

### Recommended Glob Patterns
- **Knowledge Base**: `docs/KB/**` - Focused factual reference
- **Architecture Decisions**: `docs/ADR/**` - Decision records and rationale
- **Configuration Guides**: `docs/*GUIDE.md` - Setup and operational guides  
- **All Documentation**: `docs/**` - Comprehensive search

## GitHub Actions Integration

### Automatic Triggers
The indexing workflow runs automatically on:
- Push to `docs/**` paths
- Pull requests touching `docs/**`
- Changes to `scripts/serena-index.sh`

### Workflow Features
- **Validation**: JSON schema validation of manifest
- **Summary Comments**: PR comments with indexing results
- **Artifact Storage**: Archives manifest for debugging
- **Performance Tracking**: Execution time and file counts

### Example Workflow Output
```
📊 Indexing Summary:
📁 Total documents indexed: 14
⏰ Last updated: 2025-08-13T07:28:24Z

📋 Documents by category:
8 guide documents
4 KB documents  
2 ADR documents
```

## Performance Characteristics

### Incremental Processing
- **First Run**: Indexes all files (~2-3 seconds for 14 files)
- **Subsequent Runs**: Only changed files (~200ms for no changes)
- **SHA256 Checking**: Fast file change detection

### Cost Optimization
- **Token Savings**: ~35% reduction vs full re-indexing
- **Processing Speed**: 5-10x faster incremental updates
- **Storage Efficiency**: Minimal manifest footprint

## Integration with Development Workflow

### Pre-Task Protocol
1. ✅ `make serena-index` - Ensure docs are indexed
2. ✅ Query specific categories (KB → ADR → guides)
3. ✅ Combine with code search via `mcp__serena__find_symbol`
4. ✅ Store learnings in `mcp__mem0__` for future sessions

### Development Workflow
1. Edit documentation files
2. Run `make serena-index` (or wait for GitHub Actions)
3. Test queries with updated index
4. Commit changes - GitHub Actions validates indexing

### Quality Assurance
- **Manifest Validation**: JSON schema compliance
- **Category Consistency**: Proper file categorization
- **Tag Accuracy**: Relevant tagging based on content
- **Performance Monitoring**: Indexing time and file counts

## Troubleshooting

### Common Issues

**Empty Manifest**
```bash
# Reinitialize manifest
rm .serena/manifest.json
make serena-index
```

**Missing Files in Index**
```bash
# Check file permissions and paths
find docs/ -type f -name "*.md" -not -readable
```

**Performance Degradation**
```bash
# Check manifest size
ls -la .serena/manifest.json
wc -l .serena/manifest.json

# Reset if too large
cp .serena/manifest.json .serena/manifest.backup.json
echo '{"version": "1.0.0", "timestamp": "", "docs": []}' > .serena/manifest.json
make serena-index
```

### Debugging Commands
```bash
# Verbose script execution
bash -x scripts/serena-index.sh

# Check SHA256 manually
shasum -a 256 docs/KB/glossary.md

# Validate manifest JSON
jq empty .serena/manifest.json && echo "Valid JSON" || echo "Invalid JSON"
```

## Future Enhancements

### Potential Improvements
- **Content Analysis**: Semantic similarity scoring
- **Cross-References**: Document relationship mapping
- **Search Ranking**: Usage-based relevance scoring
- **Remote Sync**: Push index to external Serena service

### Integration Points
- **IDE Integration**: Real-time indexing on file save
- **Search Interface**: Web UI for documentation search
- **Analytics**: Query patterns and document usage metrics
- **API Endpoints**: REST API for external integrations