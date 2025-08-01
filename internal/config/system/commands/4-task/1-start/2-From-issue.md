# /2-From-issue
Generate intelligent issue analysis and resolution planning.

## Context Available
- docs/3-current-task/current-task.json - Issue data (pre-populated by preprocessing)
- GitHub issue context and metadata

## Focus
Claude should focus on intelligent issue analysis:
1. Analyze issue complexity and root cause potential from docs/3-current-task/current-task.json
2. Generate comprehensive reproduction steps and debugging approach
3. Create resolution strategy and implementation plan
4. Update docs/3-current-task/current-task.json with analysis

## Important
Preprocessing has already handled GitHub issue selection, assignment, and workspace setup.
Focus on intelligent analysis and solution planning.

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
