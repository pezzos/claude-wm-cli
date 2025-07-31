# /1-Archive
Archive completed task and capture learnings.

## Steps
1. Review archived task documentation (pre-archived by preprocessing from docs/3-current-task/)
2. Update epic TECHNICAL.md with technical decisions and patterns
3. Store key learnings **with mem0** and enrich global CLAUDE.md
4. Validate task completion status (pre-updated by preprocessing in PRD.md and stories.json)

## Important
Capture reusable patterns and lessons learned. Clean up current task workspace.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for current-task.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude-wm/.claude/commands/tools/schema-enforcer.sh show-requirements current-task
```

### Schema-Aware Generation
When updating docs/3-current-task/current-task.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude-wm/.claude/commands/tools/schema-enforcer.sh show-requirements current-task
```

All required fields must be present with correct types and values.

### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude-wm/.claude/commands/tools/json-validator.sh validate; then
    echo "⚠ JSON validation failed - files auto-corrected"
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
