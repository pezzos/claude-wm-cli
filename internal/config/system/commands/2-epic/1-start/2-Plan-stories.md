# /2-epic:1-start:2-Plan-stories
Extract user stories with intelligent decomposition and data-driven complexity analysis.

## Intelligent Story Decomposition (MANDATORY)
1. **Epic Analysis**: Use `mcp__consult7__consultation` to understand current codebase complexity and dependencies
2. **Story Breakdown**: Use `mcp__sequential-thinking__` for complex feature decomposition into manageable stories
3. **Historical Patterns**: Use `mcp__mem0__search_coding_preferences` to find similar story patterns and outcomes
4. **Framework Guidance**: Use `mcp__context7__` for current best practices in agile story writing

## Enhanced Planning Process
1. **Epic Context Loading**: Read docs/2-current-epic/current-epic.json for epic user stories
2. **Codebase Impact Analysis**: Analyze how each story affects existing code structure
3. **Dependency Mapping**: Identify technical dependencies between stories using consult7 insights
4. **Smart Story Creation**: Break down stories with data-driven acceptance criteria
5. **Complexity Scoring**: Use codebase analysis for accurate story point estimation
6. **Intelligent Sequencing**: Order stories by technical dependencies and development logic

## MCP-Enhanced Story Development
- **Acceptance Criteria Generation**: Create criteria based on actual component analysis
- **Technical Task Identification**: Use codebase analysis to identify implementation requirements
- **Edge Case Discovery**: Use sequential-thinking to identify potential edge cases
- **Test Scenario Planning**: Include automated testing requirements from UI analysis
- **Performance Considerations**: Include performance acceptance criteria based on current baselines

## Smart Story Features
- **Auto-Generated Templates**: Stories include relevant technical context from codebase
- **Risk Assessment**: Each story includes complexity and risk indicators
- **Testing Strategy**: Stories include automated testing approach (including MCP UI testing)
- **Documentation Requirements**: Stories specify documentation updates needed

## Deliverables
Create new docs/2-current-epic/stories.json with:
- **Data-Driven Stories**: Enhanced with codebase analysis insights
- **Smart Prioritization**: P0-P3 based on technical dependencies and business value
- **Accurate Estimation**: Complexity points informed by actual code analysis
- **Implementation Guidance**: Technical approach suggestions from mem0 patterns

## Quality Standards
- Each story should be 1-3 days scope with intelligent complexity assessment
- Include edge cases discovered through sequential-thinking analysis
- Use enhanced 1,2,3,5,8 point complexity scale with technical justification
- Include regression testing requirements for each story
- Document cross-story dependencies and integration points

## Learning Integration
- **Store Story Patterns**: Use `mcp__mem0__add_coding_preference` to capture successful story structures
- **Complexity Insights**: Save accurate estimation patterns for future epics
- **Dependency Learnings**: Document recurring technical dependency patterns

## Important
Use MCP tools for data-driven story creation backed by actual codebase analysis. Focus on realistic estimation based on technical complexity, not just feature scope.

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
1. **$schema field**: The JSON file MUST contain a "$schema" field with the value "internal/config/system/commands/templates/schemas/stories.schema.json"
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
