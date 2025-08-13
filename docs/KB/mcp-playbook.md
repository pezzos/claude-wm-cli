# MCP Tools Playbook

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