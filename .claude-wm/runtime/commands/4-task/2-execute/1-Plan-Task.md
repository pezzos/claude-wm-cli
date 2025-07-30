# /4-task:2-execute:1-Plan-Task
Generate comprehensive implementation plan with intelligent analysis.

## Context Available
- docs/3-current-task/current-task.json - Task context (pre-populated by preprocessing)
- docs/3-current-task/iterations.json - Iteration tracking (pre-populated by preprocessing)
- Template structures ready for content generation

## Focus
Claude should focus on intelligent planning:
1. Search similar solutions **with mem0** for proven patterns
2. Enhance current-task.json with comprehensive approach, file changes, and implementation steps
3. **Plan Regression Testing**: Include tests that must be added for continuous validation of non-regression
4. Document risks and assumptions clearly in current-task.json

## Regression Testing Requirements (MANDATORY)
When planning implementation, ALWAYS include:
- **Automated Tests**: Define which MCP UI tests (Playwright/Puppeteer) need to be created
- **Test Coverage**: Specify which user journeys require ongoing validation
- **Performance Baselines**: Set metrics that must be maintained (load time, responsiveness)
- **Visual Regression**: Identify UI components requiring screenshot comparison
- **Integration Points**: List APIs/services that need continuous validation
- **Accessibility Standards**: Define a11y requirements for ongoing compliance

## Test Integration Strategy
- **Unit Tests**: Traditional function-level testing
- **Integration Tests**: Component interaction validation
- **UI Automation**: MCP-powered browser testing for user interfaces
- **Performance Tests**: Automated performance monitoring
- **Security Tests**: Ongoing vulnerability scanning
- **Manual Test Scenarios**: Critical paths requiring human validation

## Important
Research existing patterns before planning new approach. Plan for 3 iteration maximum. ALWAYS include regression testing strategy to ensure long-term code quality and prevent future failures.

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
