# /1-project:3-epics:1-Plan-Epics
Archive previous epics.json and create new epic planning based on project vision in JSON format.

## Steps
1. Archive existing epics.json to docs/archive/epics-archive-{date}/ (separate DONE/TODO)
2. Read IMPLEMENTATION.md to understand what is already working
3. Read ARCHITECTURE.md and README.md to compare and understand what is currently missing
4. Search project context **with mem0** for successful epic patterns
5. Create new epics.json in docs/1-project/ following the schema requirements exactly
6. Define 3-5 epics with story themes (not detailed stories), dependencies, success criteria in structured JSON format

## JSON Structure Required
- Use the schema from .claude-wm/.claude/commands/templates/schemas/epics.schema.json 
- Each epic must include: id, title, description, status, priority, business_value, target_users, success_criteria, dependencies, blockers, story_themes
- DO NOT include userStories - they belong in stories.json and are linked via epic_id
- Include project_context section with current_epic, total_epics, completed_epics, project_phase

## Important
Base epics on user value delivery. Sequence by dependencies and risk. Each epic should be 2-4 weeks scope.
**CRITICAL: Generate docs/1-project/epics.json (JSON format) instead of epics.json (markdown)**

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for epics.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude-wm/.claude/commands/tools/schema-enforcer.sh show-requirements epics
```

### Schema-Aware Generation
When updating docs/1-project/epics.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude-wm/.claude/commands/tools/schema-enforcer.sh show-requirements epics
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value ".claude/commands/templates/schemas/epics.schema.json"
2. All required fields must be present with correct types and values
3. No forbidden fields (like userStories) should be included

### Post-Generation Validation
After completing the main task, validate the generated JSON:

```bash
# Validate with auto-correction
if ! .claude-wm/.claude/commands/tools/simple-validator.sh validate-file docs/1-project/epics.json; then
    echo "⚠ JSON validation failed - attempting auto-correction"
    .claude-wm/.claude/commands/tools/json-validator.sh auto-correct docs/1-project/epics.json
    exit 1  # Needs iteration
fi
```

### Exit Code Integration
The command should exit with code 1 if validation fails, triggering iteration.

<!-- /JSON_SCHEMA_VALIDATION -->
