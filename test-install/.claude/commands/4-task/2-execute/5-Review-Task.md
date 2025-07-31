# /5-Review-Task
Execute comprehensive review with quality gates validation and plan consistency checks.

## Pre-Review Intelligence (MANDATORY)
1. **Load Review Patterns**: Use `mcp__mem0__search_coding_preferences` to find effective review approaches
2. **Analyze Complex Requirements**: Use `mcp__sequential-thinking__` for complex requirement decomposition
3. **Get Quality Documentation**: Use `mcp__context7__` for current quality standards and best practices

## Review Steps
1. **Plan-Implementation Consistency**: Compare implementation against original plan and requirements
2. **Code Quality Assessment**: Review code standards, maintainability, and architectural consistency
3. **Security & Performance Review**: Validate security requirements and performance benchmarks (pre-checked by preprocessing)
4. **Integration & Compatibility**: Check integration points and backward compatibility
5. **Documentation Completeness**: Ensure all documentation is complete and accurate
6. **Quality Gates Validation**: Verify all quality gates are satisfied before approval

## Complex Requirements Analysis
For requirements involving multiple validation points or complex interactions:
- **Use `mcp__sequential-thinking__`** to decompose requirements into checkable criteria
- **Map Quality Gates**: Identify all quality gates and validation checkpoints
- **Regression Testing**: Verify no regressions in existing functionality
- **Acceptance Criteria**: Validate all acceptance criteria are met
- **End-to-End Review**: Ensure complete solution integrity

## Iteration Management
Review iteration status from preprocessing:
- Review preprocessing results for quality checks and task status updates
- If review fails: update iterations.json with specific guidance for re-planning
- If review passes: approve for archiving and task completion
- No iteration limit for review - continue until quality standards are met

## Learning Capture
- **Store Successful Patterns**: Use `mcp__mem0__add_coding_preference` to capture effective review approaches
- **Document Quality Insights**: Save quality gate validation strategies
- **Performance Metrics**: Record quality metrics for future reference

## Important
All quality gates must pass before approval. Use sequential-thinking for complex validation scenarios. Document any iteration lessons learned and successful review patterns.

## EXIT CODE REQUIREMENTS
**CRITICAL**: You MUST exit with the appropriate code based on review results:

### Exit Code 0 (Success)
Use when ALL conditions are met:
- Implementation matches plan and requirements completely
- All code quality standards are satisfied
- Security and performance requirements are met
- Integration and backward compatibility verified
- Documentation is complete and accurate
- All acceptance criteria are validated
- No regressions or quality issues found

**Action**: Add this to your final response:
```
REVIEW_RESULT=SUCCESS
EXIT_CODE=0
```

### Exit Code 1 (Needs Iteration)
Use when review fails but issues can be addressed:
- Implementation doesn't fully match requirements
- Code quality standards not met but fixable
- Performance or security issues that can be resolved
- Documentation gaps or inconsistencies
- Integration issues that can be addressed
- Quality gates not satisfied but achievable

**Action**: Add this to your final response:
```
REVIEW_RESULT=NEEDS_ITERATION
EXIT_CODE=1
RETRY_REASON=[specific reason for retry - will trigger re-planning]
```

### Exit Code 2 (Blocked)
Use when review fails due to fundamental issues:
- Requirements are fundamentally unclear or conflicting
- Technical blockers that require external expertise
- Architectural decisions that need stakeholder input
- Issues that cannot be resolved within current context
- Quality standards that are unattainable with current approach

**Action**: Add this to your final response:
```
REVIEW_RESULT=BLOCKED
EXIT_CODE=2
BLOCK_REASON=[specific reason for blocking]
```

# Exit codes:
- 0: Success - review passed completely, ready for archiving
- 1: Needs iteration - review failed but retryable, triggers re-planning
- 2: Blocked - review failed due to fundamental issues
- 3: User input needed
## JSON Schema Validation
<!-- JSON_SCHEMA_VALIDATION -->

### MANDATORY: Schema Compliance for iterations.json

Before generating or updating JSON files, Claude MUST use schema-aware prompts:

```bash
# Show schema requirements
.claude/commands/tools/schema-enforcer.sh show-requirements current-task
```

### Schema-Aware Generation
When updating docs/3-current-task/iterations.json, include this in your Claude prompt:

**CRITICAL: SCHEMA COMPLIANCE REQUIRED**

You MUST generate JSON that strictly follows the schema. Use:
```bash
.claude/commands/tools/schema-enforcer.sh show-requirements iterations
```

All required fields must be present with correct types and values.

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
