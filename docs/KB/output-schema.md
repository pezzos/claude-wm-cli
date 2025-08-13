# Standardized Claude Code Output Schema

## Required JSON Schema

All Claude Code task outputs must conform to this schema:

```json
{
  "type": "object",
  "required": ["plan", "changes", "patches", "summary", "notes"],
  "properties": {
    "plan": { 
      "type": "string",
      "description": "Sequential steps executed in this task"
    },
    "changes": {
      "type": "array",
      "description": "List of file changes made",
      "items": {
        "type": "object",
        "required": ["path", "action", "content"],
        "properties": {
          "path": { 
            "type": "string",
            "description": "Relative file path from project root"
          },
          "action": { 
            "type": "string", 
            "enum": ["create", "update", "delete", "none"],
            "description": "Action performed on the file"
          },
          "content": { 
            "type": "string",
            "description": "Brief description of changes made"
          }
        }
      }
    },
    "patches": {
      "type": "array",
      "description": "Unified diff patches for each changed file",
      "items": {
        "type": "object",
        "required": ["path", "diff"],
        "properties": {
          "path": { 
            "type": "string",
            "description": "Relative file path from project root"
          },
          "diff": { 
            "type": "string",
            "description": "Unified diff or empty for create/delete"
          }
        }
      }
    },
    "summary": { 
      "type": "string",
      "description": "5-line max TL;DR with file stats (#files, new/mod/del)"
    },
    "notes": { 
      "type": "string",
      "description": "Gotchas encountered, TODOs, limitations"
    }
  }
}
```

## Output Examples

### Example 1: New Feature Implementation

```json
{
  "plan": "1. Updated Serena index for latest docs\n2. Created MCP pipeline protocol documentation\n3. Implemented standardized output schema\n4. Added validation examples\n5. Updated mcp-playbook.md with pipeline integration",
  "changes": [
    {
      "path": "docs/KB/mcp-pipeline-protocol.md",
      "action": "create",
      "content": "Comprehensive MCP pipeline protocol with decision trees and patterns"
    },
    {
      "path": "docs/KB/output-schema.md", 
      "action": "create",
      "content": "JSON schema and examples for standardized Claude Code outputs"
    },
    {
      "path": "docs/KB/mcp-playbook.md",
      "action": "update", 
      "content": "Added pipeline protocol integration and usage patterns"
    }
  ],
  "patches": [
    {
      "path": "docs/KB/mcp-pipeline-protocol.md",
      "diff": "New file - no diff for create action"
    },
    {
      "path": "docs/KB/output-schema.md",
      "diff": "New file - no diff for create action"
    },
    {
      "path": "docs/KB/mcp-playbook.md",
      "diff": "@@ -1,5 +1,15 @@\n # MCP Tools Playbook\n \n+## MCP Pipeline Protocol Integration\n+\n+Always follow the MCP Pipeline Protocol for consistent tool usage:\n+1. Context7: Load KB/ADR + authorized paths\n+2. Sequential Thinking: Detail plan before implementation\n+3. Serena: Reuse existing patterns and documentation\n+4. Zen: Clean, noise-free outputs\n+5. Chain of Verification: Quality gate before delivery\n+\n ## Serena Documentation Indexing"
    }
  ],
  "summary": "Implemented MCP Pipeline Protocol and standardized output schema.\nFiles: 3 total (2 created, 1 updated)\nNew: mcp-pipeline-protocol.md, output-schema.md\nUpdated: mcp-playbook.md with pipeline integration\nAdded comprehensive protocol for Context7→Sequential→Serena→Zen→CoV workflow",
  "notes": "Pipeline protocol designed for realistic time constraints. Schema enforces consistent output format for all Claude Code tasks. Consider adding automated validation in CI/CD pipeline."
}
```

### Example 2: Bug Fix

```json
{
  "plan": "1. Identified issue in serena-index.sh mapfile compatibility\n2. Replaced mapfile with portable while loop\n3. Tested fix on both bash and sh\n4. Verified incremental indexing still works",
  "changes": [
    {
      "path": "scripts/serena-index.sh",
      "action": "update",
      "content": "Fixed mapfile compatibility issue for cross-platform support"
    }
  ],
  "patches": [
    {
      "path": "scripts/serena-index.sh", 
      "diff": "@@ -98,8 +98,10 @@ log_info \"Scanning $DOCS_DIR for documentation files...\"\n \n-# Find all documentation files\n-mapfile -t doc_files < <(find \"$DOCS_DIR\" -type f \\( -name \"*.md\" -o -name \"*.json\" \\) | sort)\n+# Find all documentation files\n+doc_files=()\n+while IFS= read -r -d '' file; do\n+    doc_files+=(\"$file\")\n+done < <(find \"$DOCS_DIR\" -type f \\( -name \"*.md\" -o -name \"*.json\" \\) -print0 | sort -z)\n \n if [[ ${#doc_files[@]} -eq 0 ]]; then"
    }
  ],
  "summary": "Fixed serena-index.sh compatibility issue with mapfile command.\nFiles: 1 updated\nFixed: Cross-platform shell compatibility\nReplaced mapfile with portable while loop for file array population",
  "notes": "mapfile is bash-specific and not available in sh. Using while loop with null-delimited input ensures compatibility across different shells while maintaining functionality."
}
```

### Example 3: Documentation Update

```json
{
  "plan": "1. Analyzed current documentation gaps\n2. Created comprehensive output schema guide\n3. Added validation examples and patterns\n4. Updated index to include new documentation",
  "changes": [
    {
      "path": "docs/KB/output-schema.md",
      "action": "create", 
      "content": "Complete guide for Claude Code JSON output format with examples"
    },
    {
      "path": ".serena/manifest.json",
      "action": "update",
      "content": "Added new documentation file to Serena index"
    }
  ],
  "patches": [
    {
      "path": "docs/KB/output-schema.md",
      "diff": "New file - no diff for create action"
    },
    {
      "path": ".serena/manifest.json",
      "diff": "@@ -172,5 +172,15 @@\n       \"indexed_at\": \"2025-08-13T07:29:35Z\"\n     }\n   ]\n+  {\n+    \"path\": \"docs/KB/output-schema.md\",\n+    \"title\": \"Standardized Claude Code Output Schema\", \n+    \"category\": \"KB\",\n+    \"tags\": [\"KB\", \"schema\", \"output\", \"validation\"],\n+    \"sha\": \"abc123def456...\",\n+    \"indexed_at\": \"2025-08-13T08:15:23Z\"\n+  }\n }"
    }
  ],
  "summary": "Created standardized output schema documentation.\nFiles: 2 total (1 created, 1 updated)\nNew: output-schema.md with JSON schema and examples\nUpdated: Serena manifest with new documentation file",
  "notes": "Schema designed to provide consistent structure for all Claude Code outputs. Enables automated validation and better human readability of task results."
}
```

## Validation Requirements

### Required Fields
- **plan**: Must describe actual steps taken (not future plans)
- **changes**: Must list all file modifications with accurate actions
- **patches**: Must provide diffs for all updated files
- **summary**: Must be ≤5 lines with file statistics
- **notes**: Must capture important context, limitations, or follow-ups

### File Statistics Format
Summary must include:
```
Files: X total (Y created, Z updated, A deleted)
New: file1.ext, file2.ext
Updated: file3.ext, file4.ext  
Deleted: file5.ext
[Brief description of key changes]
```

### Diff Format Requirements
- Use unified diff format (`diff -u`)
- Show context lines (±3 lines)
- Empty diff for create/delete actions with note
- Include file headers with `@@` line markers

## Usage in Different Contexts

### Implementation Tasks
Focus on:
- **plan**: Sequential implementation steps
- **changes**: All code/config files modified  
- **patches**: Full diffs showing exact changes
- **summary**: Feature/functionality added
- **notes**: Technical decisions, edge cases

### Documentation Tasks  
Focus on:
- **plan**: Content creation and organization steps
- **changes**: Documentation files and index updates
- **patches**: Content additions and modifications
- **summary**: Documentation scope and coverage
- **notes**: Content gaps, future improvements

### Bug Fix Tasks
Focus on:
- **plan**: Problem identification and resolution steps
- **changes**: Minimal set of files changed
- **patches**: Precise fixes with context
- **summary**: Issue resolved and verification
- **notes**: Root cause, prevention measures

### Refactoring Tasks
Focus on:
- **plan**: Refactoring approach and validation
- **changes**: All affected files with reason
- **patches**: Before/after code structure
- **summary**: Improvements achieved
- **notes**: Compatibility, testing requirements

## Quality Checklist

Before outputting JSON schema response:

- [ ] **plan** describes what was actually done (past tense)
- [ ] **changes** includes all modified files with accurate actions
- [ ] **patches** provides readable diffs for all updates
- [ ] **summary** is ≤5 lines with complete file statistics
- [ ] **notes** captures important context and limitations
- [ ] JSON is valid and conforms to schema
- [ ] File paths are relative to project root
- [ ] Actions match actual operations (create/update/delete/none)