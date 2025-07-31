# /3-From-input
Generate intelligent task analysis from user input and requirements.

## Context Available
- docs/3-current-task/current-task.json - User input data (pre-populated by preprocessing)
- User description and requirements context

## Focus
Claude should focus on intelligent requirement analysis:
1. Analyze user input to determine precise task scope and requirements from current-task.json
2. Generate comprehensive requirements analysis and clarification questions
3. Create detailed implementation strategy and approach
4. Update current-task.json with intelligent insights and planning

## Important
Preprocessing has already handled workspace setup and basic task initialization.
Focus on intelligent requirements analysis and strategic planning.

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
