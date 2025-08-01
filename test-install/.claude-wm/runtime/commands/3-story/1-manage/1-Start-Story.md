# /1-Start-Story
Choose next priority story and create docs/2-current-epic/current-story.json.

## Steps
1. Read docs/2-current-epic/stories.json and identify highest priority unstarted story (P0 > P1 > P2 > P3)
2. Verify all story dependencies are marked complete
3. Create docs/2-current-epic/current-story.json with selected story details
4. Update docs/2-current-epic/stories.json: mark story as "🚧 In Progress - {date}"
5. Extract technical tasks from story and update the tasks field in docs/2-current-epic/stories.json

## Important
Validate story dependencies are complete. Tasks are stored within the story in docs/2-current-epic/stories.json, not in a separate todo.json file.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for docs/2-current-epic/current-story.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements current-story
```

### Schema-Aware Generation
When updating docs/2-current-epic/current-story.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements current-story
```

All required fields must be present with correct types and values.

### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude/commands/tools/simple-validator.sh validate-file docs/2-current-epic/current-story.json; then
    echo "⚠ JSON validation failed - attempting auto-correction"
    .claude/commands/tools/json-validator.sh auto-correct docs/2-current-epic/current-story.json
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
