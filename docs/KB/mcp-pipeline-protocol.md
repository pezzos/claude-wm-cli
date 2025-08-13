# MCP Pipeline Protocol

## Short Playbook (For Every Prompt)

**Always activate MCP tools in this sequence when useful:**

### 1. **Context7**: Load KB/ADR + Authorized Paths
- **When**: Before any implementation task
- **Purpose**: Get current library docs and authorized file boundaries
- **Pattern**: Resolve library ID → Get version-specific docs
- **Check**: Verify against project requirements (go.mod, package.json, etc.)

### 2. **Sequential Thinking**: Detail Plan Before Writing
- **When**: Complex tasks (>3 steps), architectural decisions, unclear requirements
- **Purpose**: Break down task, identify dependencies, verify approach
- **Pattern**: Think → Decompose → Verify → Iterate until clear
- **Output**: Clear step-by-step plan ready for execution

### 3. **Serena**: Reuse Existing Helpers/Code/Docs
- **When**: Before creating new code or documentation
- **Purpose**: Find existing patterns, utilities, documentation to reuse
- **Pattern**: `make serena-index` → Search KB/ADR → Find symbols → Reuse patterns
- **Focus**: Hash utilities, diff engines, copy patterns, FS operations

### 4. **Zen**: Reduce Noise in Outputs/Errors
- **When**: Processing large datasets, complex operations
- **Purpose**: Filter relevant information, reduce cognitive load
- **Pattern**: Focus on actionable information, hide verbose details
- **Output**: Clean, scannable results with clear status

### 5. **Chain of Verification**: Short Checklist Before Delivery
- **When**: Before completing any task
- **Purpose**: Quality gate to ensure completeness and correctness
- **Checklist**:
  - ✅ Plan executed completely
  - ✅ Files created/modified as intended
  - ✅ No breaking changes introduced
  - ✅ Documentation updated if needed
  - ✅ Follows project conventions

## Tool Selection Decision Tree

### Context7 Usage
```
Need library integration? → Yes → Context7 for docs
↓ No
Implementation only with existing libs? → Skip Context7
```

### Sequential Thinking Usage
```
Task complexity > 3 steps? → Yes → Sequential Thinking
↓ No
Architectural decision needed? → Yes → Sequential Thinking  
↓ No
Requirements unclear? → Yes → Sequential Thinking
↓ No
Direct implementation → Skip Sequential Thinking
```

### Serena Usage
```
Creating new code/docs? → Yes → Search existing patterns first
↓ No
Need project context? → Yes → Load KB/ADR via Serena
↓ No
File operations needed? → Yes → Find existing FS utilities
↓ No
Skip Serena
```

### Consult7 Usage (Architecture Only)
```
Need architecture understanding? → Yes → Consult7 with exclusions
↓ No
Complex system integration? → Yes → Consult7 for context
↓ No
Basic implementation? → Skip Consult7 (use Serena instead)
```

## Pipeline Execution Patterns

### Pattern A: New Feature Implementation
1. **Context7**: Load library docs for required integrations
2. **Sequential Thinking**: Break down feature into implementable steps
3. **Serena**: Find existing patterns for similar features
4. **Implementation**: Execute plan with verified patterns
5. **Zen**: Clean output focusing on key changes
6. **Chain of Verification**: Validate completeness

### Pattern B: Bug Fix/Maintenance
1. **Serena**: Understand existing code structure and find issue
2. **Sequential Thinking**: Plan fix approach (if complex)
3. **Implementation**: Apply fix with minimal changes
4. **Zen**: Focus output on fix verification
5. **Chain of Verification**: Ensure no regressions

### Pattern C: Documentation/Knowledge
1. **Serena**: `make serena-index` → Search existing docs
2. **Context7**: Get current library/framework docs if needed
3. **Implementation**: Create/update documentation
4. **Zen**: Clear, scannable documentation output
5. **Chain of Verification**: Ensure accuracy and completeness

### Pattern D: Architecture Decisions  
1. **Consult7**: Understand current system architecture
2. **Sequential Thinking**: Analyze options and trade-offs
3. **Serena**: Check existing ADRs and patterns
4. **Implementation**: Document decision and implement
5. **Chain of Verification**: Validate against requirements

## Realistic Time Estimates

### Tool Activation Overhead
- **Context7**: 30-60s (library resolution + doc fetch)
- **Sequential Thinking**: 2-5 minutes (complex decomposition)
- **Serena**: 10-30s (index update + search)
- **Consult7**: 60-120s (codebase analysis with exclusions)
- **Chain of Verification**: 30s (checklist validation)

### When to Skip Tools
- **Time-critical tasks**: Skip Sequential Thinking for simple fixes
- **Routine operations**: Skip Context7 for known patterns
- **Documentation-only**: Skip Consult7 for pure doc tasks
- **Minor updates**: Skip full pipeline for typo fixes

## Quality Gates

### Pre-Implementation Checkpoints
- [ ] Requirements clearly understood
- [ ] Existing patterns identified via Serena
- [ ] Library compatibility verified via Context7
- [ ] Plan decomposed if complex (Sequential Thinking)

### Post-Implementation Checkpoints  
- [ ] All planned files created/modified
- [ ] No unauthorized path modifications
- [ ] Documentation updated if needed
- [ ] Tests passing (if applicable)
- [ ] Output clean and scannable (Zen)

## Pipeline Optimization

### Parallel Tool Usage
When possible, run tools in parallel:
```bash
# Parallel execution example
Context7 library-docs & 
make serena-index &
wait  # Wait for both to complete
```

### Caching Strategy
- **Context7**: Cache library docs per session
- **Serena**: Incremental indexing (only changed files)
- **Consult7**: Use exclusion patterns to reduce processing
- **Sequential Thinking**: Reuse decomposition patterns

### Error Recovery
- **Context7 fails**: Fall back to memory + web docs
- **Serena fails**: Use direct file operations
- **Sequential Thinking timeout**: Use simpler planning
- **Consult7 timeout**: Use Serena for targeted analysis

## Integration with Project Workflow

### Agile Command Integration
Each agile command automatically follows appropriate pipeline:

- **`/project:agile:start`**: Context7 + Sequential Thinking
- **`/project:agile:design`**: Consult7 + Sequential Thinking + Serena
- **`/project:agile:plan`**: Serena + Sequential Thinking
- **`/project:agile:iterate`**: Full pipeline based on task type
- **`/project:agile:ship`**: Zen + Chain of Verification

### Development Workflow
1. **Session Start**: `make serena-index` + load mem0 context
2. **Task Analysis**: Apply appropriate pipeline pattern
3. **Implementation**: Execute with MCP tool support
4. **Quality Check**: Chain of Verification before completion
5. **Session End**: Store learnings in mem0

## Success Metrics

### Pipeline Effectiveness
- **Context Accuracy**: Fewer integration issues with Context7
- **Code Reuse**: Higher pattern reuse with Serena
- **Plan Quality**: Better decomposition with Sequential Thinking
- **Output Quality**: Cleaner results with Zen
- **Delivery Quality**: Fewer defects with Chain of Verification

### Performance Targets
- **Total Pipeline Time**: <5 minutes for complex tasks
- **Tool Success Rate**: >95% successful activations
- **Quality Score**: <5% rework needed after Chain of Verification
- **Developer Satisfaction**: Clear, actionable outputs