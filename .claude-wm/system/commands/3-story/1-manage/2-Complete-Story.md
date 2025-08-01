# /2-Complete-Story
Mark story complete and update docs/2-current-epic/stories.json status.

## Steps
1. Verify all tasks in the current story (within docs/2-current-epic/stories.json) are ✅ completed and acceptance criteria met
2. Mark story as "✅ Completed" in docs/2-current-epic/stories.json with completion metrics
3. Remove docs/2-current-epic/current-story.json to clear current selection
4. Update IMPLEMENTATION.md with the story's implementation details

## Important
Run tests and validate acceptance criteria before marking complete. Store lessons learned **with mem0**. Tasks are stored within docs/2-current-epic/stories.json, not in separate todo.json files.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for docs/2-current-epic/stories.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements stories
```

### Schema-Aware Generation
When updating docs/2-current-epic/stories.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements stories
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value ".claude/commands/templates/schemas/stories.schema.json"
2. All required fields must be present with correct types and values
3. All nested objects must have their required fields
### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude/commands/tools/simple-validator.sh validate-file docs/2-current-epic/stories.json; then
    echo "⚠ JSON validation failed - attempting auto-correction"
    .claude/commands/tools/json-validator.sh auto-correct docs/2-current-epic/stories.json
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
