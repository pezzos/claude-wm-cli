# /1-From-story
Generate intelligent task analysis and planning from current story context.

## Context Available
- docs/3-current-task/current-task.json - Current task data (pre-populated by preprocessing)
- docs/2-current-epic/stories.json - Story context

## Focus
Claude should focus on intelligent analysis and content generation:
1. Analyze task complexity and requirements from current-task.json
2. Generate comprehensive task description and approach
3. Create implementation strategy based on story context
4. Update current-task.json with intelligent insights

## Important
Preprocessing has already handled file management, task selection, and status updates.
Focus on intelligent content generation and analysis.

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
