# MCP Tools Playbook

Comprehensive guide for using Model Context Protocol (MCP) tools effectively within claude-wm-cli development and operations.

## MCP Tools Overview

Claude WM CLI leverages multiple MCP servers to provide specialized capabilities:
- **Consult7:** Codebase analysis and consultation
- **Context7:** Library documentation and examples
- **Mem0:** Persistent memory and learning
- **Serena:** Semantic code operations
- **Sequential Thinking:** Complex problem decomposition
- **Time:** Time zone conversions and scheduling
- **GitHub:** Repository management and automation

## Token Optimization Strategies

### Consult7 Cost Optimization
**Standard exclusion pattern:**
```
.*logs/.*|.*metrics/.*|.*claude(-wm)?/.*|.*backup/.*|.*archive/.*|.*serena/.*
```

**Usage pattern:**
```bash
# Always use exclude patterns to reduce costs by ~35%
mcp__consult7__consultation \
  --path /project/root \
  --pattern ".*\\.go$" \
  --exclude_pattern ".*logs/.*|.*metrics/.*|.*claude(-wm)?/.*|.*backup/.*|.*archive/.*|.*serena/.*" \
  --query "Analyze configuration management patterns"
```

### Serena Semantic Operations
**Priority:** Use Serena over reading entire files
```bash
# Instead of reading full files:
# ❌ Read /path/to/large/file.go

# Use semantic operations:
# ✅ mcp__serena__get_symbols_overview --relative_path file.go
# ✅ mcp__serena__find_symbol --name_path ConfigManager --include_body true
```

## Development Workflows

### Code Analysis Workflow

1. **Project Overview**
   ```bash
   mcp__consult7__consultation \
     --path . \
     --pattern ".*\\.(go|md|yaml)$" \
     --exclude_pattern ".*logs/.*|.*claude(-wm)?/.*|.*backup/.*" \
     --query "Provide project architecture overview focusing on configuration management"
   ```

2. **Specific Component Analysis**
   ```bash
   mcp__serena__get_symbols_overview --relative_path internal/config/manager.go
   mcp__serena__find_symbol --name_path ConfigManager --include_body true
   ```

3. **Cross-Reference Analysis**
   ```bash
   mcp__serena__find_referencing_symbols \
     --name_path ConfigManager \
     --relative_path internal/config/manager.go
   ```

### Memory-Enhanced Development

1. **Store Learning Insights**
   ```bash
   mcp__mem0__add_coding_preference \
     --text "Configuration Management Pattern: The 4-space model (A/B/L/S) with 3-way merge provides excellent conflict resolution. Key insight: content-based comparison is more reliable than timestamp-based. Performance target: <200ms for status checks on 100 files."
   ```

2. **Retrieve Context Before Tasks**
   ```bash
   mcp__mem0__search_coding_preferences --query "configuration management 3-way merge"
   ```

3. **Get Complete Memory Context**
   ```bash
   mcp__mem0__get_all_coding_preferences
   ```

### Documentation Generation Workflow

1. **Library Documentation**
   ```bash
   # Resolve Go CLI library
   mcp__context7__resolve_library_id --libraryName "urfave/cli"
   
   # Get version-specific documentation
   mcp__context7__get_library_docs \
     --context7CompatibleLibraryID "/urfave/cli" \
     --topic "command structure and flags" \
     --tokens 5000
   ```

2. **Template Generation**
   ```bash
   # Use Task tool with claude-wm-templates subagent
   Task --subagent_type claude-wm-templates \
     --description "Generate architecture template" \
     --prompt "Create ARCHITECTURE.md template for Go CLI with 4-space configuration model"
   ```

## Problem-Solving Patterns

### Complex Analysis with Sequential Thinking

```bash
mcp__sequential_thinking__sequentialthinking \
  --thought "Analyzing the claude-wm-cli architecture to identify optimization opportunities" \
  --nextThoughtNeeded true \
  --thoughtNumber 1 \
  --totalThoughts 5
```

**Use cases:**
- Architecture decisions requiring multi-step analysis
- Performance optimization planning
- Complex debugging scenarios
- Feature design with multiple considerations

### Code Review Automation

```bash
# Use claude-wm-reviewer for comprehensive reviews
Task --subagent_type claude-wm-reviewer \
  --description "Review configuration manager" \
  --prompt "Review the ConfigManager implementation for security vulnerabilities, performance issues, and maintainability concerns"
```

### Status Dashboard Generation

```bash
# Use claude-wm-status for project insights
Task --subagent_type claude-wm-status \
  --description "Generate project dashboard" \
  --prompt "Create comprehensive status dashboard showing configuration management health, recent changes, and performance metrics"
```

## Integration Patterns

### GitHub Workflow Automation

1. **Repository Analysis**
   ```bash
   mcp__github__get_file_contents \
     --owner organization \
     --repo claude-wm-cli \
     --path internal/config/manager.go
   ```

2. **Issue Management**
   ```bash
   # Create issues for documentation updates
   mcp__github__create_issue \
     --owner organization \
     --repo claude-wm-cli \
     --title "Update configuration guide for v2.0" \
     --body "Documentation needs updating for new 4-space model"
   ```

3. **Pull Request Creation**
   ```bash
   mcp__github__create_pull_request \
     --owner organization \
     --repo claude-wm-cli \
     --title "Add 3-way merge documentation" \
     --head feature/merge-docs \
     --base main \
     --body "Comprehensive documentation for 3-way merge strategy"
   ```

### Time Management Integration

```bash
# Schedule maintenance windows
mcp__time__get_current_time --timezone Europe/Paris
mcp__time__convert_time \
  --source_timezone Europe/Paris \
  --time "14:00" \
  --target_timezone America/New_York
```

## Performance Optimization

### Token Usage Reduction

1. **Selective File Analysis**
   ```bash
   # Target specific file patterns
   --pattern "internal/config/.*\\.go$"
   
   # Always exclude non-essential directories
   --exclude_pattern ".*logs/.*|.*metrics/.*|.*claude(-wm)?/.*|.*backup/.*|.*archive/.*"
   ```

2. **Semantic Operations Over Full Reads**
   ```bash
   # Efficient symbol-based exploration
   mcp__serena__find_symbol --name_path Manager --substring_matching true
   mcp__serena__get_symbols_overview --relative_path internal/config/
   ```

3. **Memory-Guided Development**
   ```bash
   # Leverage stored insights to avoid repeated analysis
   mcp__mem0__search_coding_preferences --query "performance optimization patterns"
   ```

### Response Optimization

**Structured queries for focused responses:**
- Specific component analysis over broad codebase scans
- Targeted documentation requests with token limits
- Incremental exploration using follow-up queries

## Error Handling and Recovery

### Common Issues and Solutions

1. **Token Limit Exceeded**
   ```bash
   # Solution: Use more restrictive patterns
   --exclude_pattern ".*logs/.*|.*test/.*|.*vendor/.*|.*node_modules/.*"
   --pattern "internal/config/.*\\.go$"  # More specific pattern
   ```

2. **Context Loss**
   ```bash
   # Solution: Store important insights in memory
   mcp__mem0__add_coding_preference --text "Key finding from analysis: [detailed insight]"
   ```

3. **Analysis Overwhelm**
   ```bash
   # Solution: Use sequential thinking for complex problems
   mcp__sequential_thinking__sequentialthinking --thought "Breaking down the problem..."
   ```

### Recovery Patterns

**State Reconstruction:**
1. Check memory for previous insights: `mcp__mem0__get_all_coding_preferences`
2. Use Serena for quick code structure review: `mcp__serena__get_symbols_overview`
3. Consult7 for focused re-analysis: `mcp__consult7__consultation` with specific patterns

## Best Practices

### Query Design

1. **Specificity:** Use precise patterns and queries
   - ✅ "Analyze ConfigManager's 3-way merge implementation"
   - ❌ "Tell me about the code"

2. **Context:** Provide relevant background
   ```bash
   --query "Given the 4-space model (A/B/L/S), analyze how the 3-way merge algorithm handles conflicts in the ConfigManager class"
   ```

3. **Constraints:** Set appropriate limits
   ```bash
   --tokens 3000  # For focused documentation
   --tokens 8000  # For comprehensive analysis
   ```

### Memory Management

1. **Structured Storage**
   ```bash
   # Use prefixes for categorization
   mcp__mem0__add_coding_preference --text "ARCHITECTURE: 4-space model provides clear separation..."
   mcp__mem0__add_coding_preference --text "PERFORMANCE: 3-way merge averages 200ms for 100 files..."
   mcp__mem0__add_coding_preference --text "DEBUGGING: Common conflict resolution pattern..."
   ```

2. **Regular Maintenance**
   - Search existing memory before adding new entries
   - Update outdated insights with new learnings
   - Use descriptive, searchable content

### Workflow Integration

1. **Task Planning**
   ```bash
   # Always start complex tasks with planning subagents
   Task --subagent_type claude-wm-planner --description "Plan feature implementation"
   ```

2. **Code Quality**
   ```bash
   # Review before completion
   Task --subagent_type claude-wm-reviewer --description "Final code review"
   ```

3. **Documentation**
   ```bash
   # Generate comprehensive documentation
   Task --subagent_type claude-wm-templates --description "Create user guide"
   ```

## Monitoring and Metrics

### Usage Tracking

Track token consumption patterns:
- **High-cost operations:** Consult7 without exclusions
- **Medium-cost operations:** Sequential thinking sessions
- **Low-cost operations:** Serena symbol operations, memory queries

### Efficiency Metrics

- **Analysis speed:** Time to insight for common queries
- **Token efficiency:** Insights per token consumed
- **Memory hit rate:** Percentage of queries answered from memory

### Quality Indicators

- **Accuracy:** Solutions that work on first attempt
- **Completeness:** Analysis covering all relevant aspects
- **Actionability:** Insights leading to concrete improvements

## Future Optimization Opportunities

### Potential Enhancements

1. **Automated Pattern Learning**
   - Train exclusion patterns based on usage
   - Optimize query structures for common tasks
   - Develop domain-specific prompt templates

2. **Intelligent Caching**
   - Cache analysis results for unchanged code sections
   - Share insights across similar project structures
   - Implement incremental analysis updates

3. **Workflow Automation**
   - Chain MCP tools for common patterns
   - Create decision trees for tool selection
   - Implement feedback loops for continuous improvement

This playbook provides the foundation for efficient MCP tool usage within claude-wm-cli development. Regular updates based on usage patterns and new capabilities will ensure continued optimization.