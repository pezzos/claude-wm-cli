# /4-Validate-Task
Execute comprehensive testing with intelligent user journey analysis.

## Pre-Validation Intelligence (MANDATORY)
1. **Load Test Patterns**: Use `mcp__mem0__search_coding_preferences` to find validation approaches
2. **Analyze Complex Journeys**: Use `mcp__sequential-thinking__` for complex user journeys decomposition
3. **Get Testing Documentation**: Use `mcp__context7__` for current testing best practices

## Validation Steps
1. **Review Test Results**: Analyze automated test results (pre-executed by preprocessing)
2. **MCP UI Testing**: Execute Playwright/Puppeteer automated UI tests when applicable
3. **Complex Journey Testing**: Use `mcp__sequential-thinking__` to break down and validate multi-step user workflows
4. **Manual Test Scenarios**: Execute remaining manual test scenarios from docs/3-current-task/TEST.md
5. **Performance & Security**: Review performance baselines and security requirements (pre-checked by preprocessing)
6. **Visual Regression**: Compare UI screenshots for consistency (when UI tests exist)

## Complex User Journey Analysis
For user journeys involving multiple steps or complex interactions:
- **Use `mcp__sequential-thinking__`** to decompose into testable steps
- **Map Dependencies**: Identify step-by-step dependencies and validation points  
- **Error Path Testing**: Test failure scenarios at each step
- **Recovery Testing**: Validate system recovery from failures
- **End-to-End Validation**: Ensure complete journey integrity

## Iteration Management
Review iteration status from preprocessing:
- If tests fail and iterations < 3: update docs/3-current-task/iterations.json and provide guidance for next iteration
- If failure AND iterations = 3: document the blockage in docs/3-current-task/iterations.json, stop and ask for help
- If success: validate completion status (pre-updated by preprocessing) and provide final validation report

## Learning Capture
- **Store Successful Patterns**: Use `mcp__mem0__add_coding_preference` to capture effective validation approaches
- **Document Journey Insights**: Save complex user journey testing strategies
- **Performance Baselines**: Record performance metrics for future regression testing

## Important
All tests must pass before proceeding. Use sequential-thinking for complex validation scenarios. Document any iteration lessons learned and successful patterns.

## EXIT CODE REQUIREMENTS
**CRITICAL**: You MUST exit with the appropriate code based on validation results:

### Exit Code 0 (Success)
Use when ALL conditions are met:
- All automated tests pass successfully
- Manual test scenarios validate correctly
- Performance baselines are maintained
- Security requirements are satisfied
- No blockers or critical issues found

**Action**: Add this to your final response:
```
VALIDATION_RESULT=SUCCESS
EXIT_CODE=0
```

### Exit Code 1 (Needs Iteration)
Use when validation fails but can be retried:
- Some tests fail but issues are addressable
- Performance issues that can be optimized
- Implementation doesn't fully meet requirements
- Current iteration < 3 and problems are solvable

**Action**: Add this to your final response:
```
VALIDATION_RESULT=NEEDS_ITERATION
EXIT_CODE=1
RETRY_REASON=[specific reason for retry]
```

### Exit Code 2 (Blocked)
Use when validation fails due to fundamental issues:
- Requirements are unclear or conflicting
- Technical blockers that require external help
- Maximum iterations (3) reached without success
- Issues that cannot be resolved in current context

**Action**: Add this to your final response:
```
VALIDATION_RESULT=BLOCKED
EXIT_CODE=2
BLOCK_REASON=[specific reason for blocking]
```

# Exit codes:
- 0: Success - validation passed completely
- 1: Needs iteration - validation failed but retryable
- 2: Blocked - validation failed due to fundamental issues
- 3: User input needed
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for docs/3-current-task/iterations.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements iterations
```

### Schema-Aware Generation
When updating docs/3-current-task/iterations.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements iterations
```

**MANDATORY REQUIREMENTS:**
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value ".claude/commands/templates/schemas/iterations.schema.json"
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
