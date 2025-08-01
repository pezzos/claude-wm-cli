# /2-Status
Analyze task progress and provide comprehensive status report.

## Steps
1. Review task documentation (current-task.json, docs/3-current-task/TEST.md, docs/3-current-task/iterations.json) for progress metrics (pre-analyzed by preprocessing)
2. Assess completion readiness against success criteria using preprocessing analysis
3. Calculate effort vs estimates and quality metrics from JSON data
4. Enhance status report with intelligent analysis and specific next action recommendations

## Important
Show iteration history and lessons learned. Provide clear completion percentage and blockers.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for docs/3-current-task/current-task.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements current-task
```

### Schema-Aware Generation
When updating docs/3-current-task/current-task.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements current-task
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value "internal/config/system/commands/templates/schemas/current-task.schema.json"
2. All required fields must be present with correct types and values
3. All nested objects must have their required fields
### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude/commands/tools/json-validator.sh validate; then
    echo "⚠ JSON validation failed - files auto-corrected"
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
