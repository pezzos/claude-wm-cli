# /3-Implement
Execute intelligent implementation with MCP-assisted workflow and incremental commits.

## Pre-Implementation Intelligence (MANDATORY)
1. **Load Previous Patterns**: Use `mcp__mem0__search_coding_preferences` to find similar implementations
2. **Get Current Documentation**: Use `mcp__context7__resolve-library-id` + `mcp__context7__get-library-docs` for up-to-date API docs
3. **Decompose Complex Tasks**: Use `mcp__sequential-thinking__` for features requiring >5 steps

## Implementation Phases
1. **Foundation**: Set up core structure with validated patterns from mem0
2. **Core Features**: Implement main functionality using current library docs from context7
3. **Integration**: Connect components with real-time validation via `mcp__ide__getDiagnostics`
4. **Polish**: Refine and optimize based on diagnostic feedback

## Continuous Workflow
- **Before each phase**: Search mem0 for relevant patterns
- **During implementation**: Real-time validation with ide diagnostics
- **After each feature**: Capture successful patterns in mem0 with `mcp__mem0__add_coding_preference`
- **For library calls**: Always verify syntax with context7 current docs

## Quality Standards
- Write tests alongside implementation
- Commit frequently with clear messages
- Follow existing code patterns enhanced by mem0 learning
- Update documentation as you implement

## Important
Test as you build. Make incremental commits. Learn and store successful patterns for future use.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed