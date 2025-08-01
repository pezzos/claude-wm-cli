# /1-Complete-Epic
Archive completed epic, update metrics

## Steps
1. Verify all stories in docs/2-current-epic/stories.json are ✅ completed
2. Archive docs/2-current-epic/ to docs/archive/{epic-name}-{date}/
3. Update epics.json status to "✅ Completed" with metrics
4. Enrich METRICS.md with the stats of the epic

## Important
Validate epic success criteria before completing. Update METRICS.md with epic performance data.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for docs/2-current-epic/current-epic.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements current-epic
```

### Schema-Aware Generation
When updating docs/2-current-epic/current-epic.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements current-epic
```

All required fields must be present with correct types and values.

### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude/commands/tools/simple-validator.sh validate-file docs/2-current-epic/current-epic.json; then
    echo "⚠ JSON validation failed - attempting auto-correction"
    .claude/commands/tools/json-validator.sh auto-correct docs/2-current-epic/current-epic.json
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
