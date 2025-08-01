# /3-Post-task
Extract lessons learned from completed task and enrich appropriate CLAUDE.md files.

## Steps
1. Analyze completed task archive for reusable patterns and insights
2. Categorize learnings (global vs epic-specific) and determine target CLAUDE.md
3. Store key patterns **with mem0** and update relevant CLAUDE.md files
4. Focus on actionable patterns, debugging techniques, and process improvements

## Important
Capture both successful patterns and lessons from failures. Include specific code examples and context.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed