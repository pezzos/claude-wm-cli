# MCP Tools Playbook

## MCP Pipeline Protocol Integration

**Always follow the MCP Pipeline Protocol for consistent tool usage:**

1. **Context7**: Load KB/ADR + authorized paths  
2. **Sequential Thinking**: Detail plan before implementation
3. **Serena**: Reuse existing patterns and documentation
4. **Zen**: Clean, noise-free outputs
5. **Chain of Verification**: Quality gate before delivery

> 📖 **Full Protocol**: See `docs/KB/mcp-pipeline-protocol.md` for complete decision trees and patterns

### Quick Reference
- **Context7** when integrating libraries (30-60s overhead)
- **Sequential Thinking** for complex tasks >3 steps (2-5min)
- **Serena** before creating new code/docs (10-30s)
- **Consult7** for architecture decisions only (60-120s)
- **Chain of Verification** always before delivery (30s)

## Serena Documentation Indexing

### Incremental Indexing Protocol
**When**: Before any task involving documentation or knowledge base queries  
**Frequency**: Automatic (GitHub Actions) + manual (`make serena-index`)  
**Purpose**: Keep Serena index synchronized with latest documentation changes

#### Automatic Triggers
- **GitHub Actions**: On push to `docs/**` paths
- **Pre-task**: Before using Serena for documentation queries
- **Development**: After creating/editing documentation files

#### Index Structure
```json
{
  "version": "1.0.0",
  "timestamp": "2024-01-15T14:30:25Z",
  "docs": [
    {
      "path": "docs/KB/glossary.md",
      "title": "Glossary",
      "category": "KB",
      "tags": ["terminology", "reference"],
      "sha": "abc123...",
      "indexed_at": "2024-01-15T14:30:25Z"
    }
  ]
}
```

#### Usage Commands
```bash
# Manual incremental indexing
make serena-index

# Check current index status
jq '.docs | length' .serena/manifest.json

# View categories breakdown
jq -r '.docs | group_by(.category) | .[] | "\(.length) \(.[0].category) documents"' .serena/manifest.json
```

### Serena Query Patterns

#### Documentation Queries (Always Index First)
```bash
# 1. Update index before querying
make serena-index

# 2. Query with specific globs for documentation
mcp__serena__search_for_pattern --relative_path "docs/KB" --substring_pattern "mcp tools"
mcp__serena__search_for_pattern --relative_path "docs/ADR" --substring_pattern "architecture decision"
```

#### Recommended Glob Patterns
- **Knowledge Base**: `docs/KB/**` (glossary, commands, file-ownership, mcp-playbook)
- **Architecture Decisions**: `docs/ADR/**` (decision records)
- **Guides**: `docs/*.md` (top-level guides like ARCHITECTURE, CONFIG_GUIDE, TESTING)
- **All Documentation**: `docs/**` (comprehensive search)

#### Pre-Task Serena Checklist
1. ✅ Run `make serena-index` (incremental, only changed files)
2. ✅ Use appropriate glob patterns for search scope
3. ✅ Query KB before ADR before guides (specificity order)
4. ✅ Combine results with code analysis via `mcp__serena__find_symbol`

## Tool Selection Matrix

### Code Analysis & Understanding

#### **mcp__consult7__consultation** 
**When**: Large codebase analysis, understanding complex systems  
**Pattern**: Always include exclude patterns: `.*logs/.*|.*metrics/.*|.*claude(-wm)?/.*|.*backup/.*|.*archive/.*|.*serena/.*`  
**Benefits**: ~35% cost reduction, faster processing  
**Example**: Understanding Go module structure, API patterns

#### **mcp__serena__** (Semantic Code Operations)
**When**: Targeted code exploration, symbol analysis  
**Advantages**: Token-efficient, precise symbol navigation  
**Key Tools**: 
- `find_symbol()` - Locate specific functions/types
- `get_symbols_overview()` - File structure analysis  
- `search_for_pattern()` - Regex search across codebase
**Usage Rule**: Always prefer over reading entire files

#### **mcp__context7__**  
**When**: Library-specific documentation needed  
**Process**: `resolve_library_id()` → `get_library_docs()`  
**Critical**: Verify version compatibility with project requirements

### Task Planning & Execution  

#### **mcp__sequential-thinking__**
**When**: Complex multi-step tasks, architectural decisions  
**Triggers**: >5 subtasks expected, unclear requirements  
**Process**: Think → decompose → verify → iterate  

#### **TodoWrite + Task**  
**When**: All implementation tasks  
**Pattern**: Plan → Execute → Mark Complete → Verify

### Memory & Learning

#### **mcp__mem0__**
**When**: Every session for context continuity  
**Pattern**: Search existing → Apply learnings → Store new insights  
**Tools**:
- `search_coding_preferences()` - Find relevant patterns
- `add_coding_preference()` - Store complete implementations  
- `get_all_coding_preferences()` - Full context review

### Development & Testing

#### **mcp__github__**
**When**: Repository operations, PR management  
**Security**: Always follow git best practices  
**Key Operations**: Push files, create PRs, manage issues

#### **mcp__playwright__** / **mcp__puppeteer__**  
**When**: Web interface testing, automation  
**Choice**: Playwright for modern apps, Puppeteer for legacy

## Activation Strategies

### Automatic Triggers
- File search → Task tool (context reduction)
- Complex analysis → Consult7 with exclusions  
- Symbol lookup → Serena before file reads
- Library integration → Context7 for docs
- Multi-step tasks → Sequential-thinking + TodoWrite

### Session Initialization Protocol
1. **Memory Load**: `mcp__mem0__search_coding_preferences()` 
2. **Context Understanding**: `mcp__consult7__consultation` with exclusions
3. **Codebase Structure**: `mcp__serena__get_symbols_overview()`
4. **Planning Setup**: `TodoWrite` for task tracking

### Cost Optimization Patterns

#### Consult7 Exclusions (Standard)
```
"exclude_pattern": ".*logs/.*|.*metrics/.*|.*claude(-wm)?/.*|.*backup/.*|.*archive/.*|.*serena/.*"
```

#### Serena First Rule
Before reading entire files:
1. Try `get_symbols_overview()`
2. Use `find_symbol()` for specific targets
3. File read only as last resort

#### Context7 Version Matching
Always verify library versions against:
- `go.mod` / `package.json` / `requirements.txt`
- Use exact version in Context7 queries

### Tool Combination Patterns

#### **Analysis → Planning → Execution**
1. `mcp__consult7__` or `mcp__serena__` (understand)
2. `mcp__sequential-thinking__` (plan)  
3. `TodoWrite` (track) + implementation tools

#### **Research → Documentation → Memory**
1. `mcp__context7__` (current docs)
2. Implementation with best practices
3. `mcp__mem0__add_coding_preference()` (store learnings)

#### **Sandbox Development**
1. `mcp__serena__` (locate target code)
2. Sandbox creation + testing
3. `mcp__github__` (PR creation)

## Performance Guidelines

### Token Efficiency
- Consult7: Always use exclusions (~35% savings)
- Serena: Symbol analysis vs full file reads
- Memory: Search before comprehensive retrieval

### Execution Speed  
- Parallel tool calls when independent
- Batch similar operations
- Cache analysis results in memory

### Quality Assurance
- Sequential-thinking for complex decisions
- TodoWrite for accountability
- Memory storage for future reference

## Error Recovery Patterns

### Tool Failures
- Consult7 timeout → Serena fallback
- Context7 unavailable → Memory search + web docs
- Serena errors → Direct file operations

### Context Loss
- Memory search for previous approaches  
- Symbol overview for code structure
- Progressive analysis rather than full reload

## Integration with Project Workflow

### Agile Command Integration
Each agile command automatically activates relevant MCP tools:
- `/project:agile:start` → Context7 + Memory search
- `/project:agile:design` → Consult7 + Sequential-thinking  
- `/project:agile:plan` → Serena + TodoWrite
- `/project:agile:iterate` → All tools as needed
- `/project:agile:ship` → GitHub integration + Memory update