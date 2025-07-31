# JSON Schema Validation System

A comprehensive validation system for Claude WM CLI that ensures all generated JSON files comply with their corresponding schemas.

## System Overview

This system provides two complementary approaches:

1. **Reactive Validation**: Post-generation validation with auto-correction
2. **Proactive Enforcement**: Pre-generation schema guidance to prevent errors

## Components

### Core Validation Tools

| Tool | Purpose | Description |
|------|---------|------------|
| `simple-validator.sh` | Reactive validation | Validates JSON files against schemas using jq |
| `schema-enforcer.sh` | Proactive enforcement | Generates schema-aware prompts for Claude |
| `integrate-validation.sh` | Command integration | Integrates validation into command templates |
| `validate-json.sh` | Master interface | Unified interface for all validation operations |

### Schema Files

Located in `.claude-wm/.claude/commands/templates/schemas/`:
- `current-task.schema.json` - Task data structure
- `current-story.schema.json` - Story data structure  
- `current-epic.schema.json` - Epic data structure
- `stories.schema.json` - Stories collection
- `epics.schema.json` - Epics collection
- `iterations.schema.json` - Iteration tracking
- `metrics.schema.json` - Project metrics

## Usage

### Master Interface

```bash
# Validate all JSON files
./validate-json.sh

# Validate specific file
./validate-json.sh validate-file docs/3-current-task/current-task.json

# Show all schema requirements
./validate-json.sh show-schemas

# Show proactive enforcement guidance
./validate-json.sh enforce current-task

# Show system status
./validate-json.sh status
```

### Individual Tools

```bash
# Simple validator
./tools/simple-validator.sh validate
./tools/simple-validator.sh validate-file current-task.json
./tools/simple-validator.sh show-schema current-task

# Schema enforcer
./tools/schema-enforcer.sh show-requirements current-task
./tools/schema-enforcer.sh template current-task output.json
./tools/schema-enforcer.sh generate current-task output.json "Task description"
```

## Workflow Integration

### Automatic Command Integration

All task commands now include validation hooks:

- **Pre-generation**: Schema requirements are shown to Claude
- **Post-generation**: Automatic validation with error reporting
- **Auto-correction**: Failed validation triggers Claude to fix issues
- **Exit codes**: Commands return appropriate codes based on validation results

### Command Template Enhancement

Each integrated command includes:

```markdown
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

<!-- /JSON_SCHEMA_VALIDATION -->
```

## Proactive Schema Enforcement

### For Developers Using Commands

When a command template instructs Claude to update JSON files, the system automatically:

1. **Shows schema requirements** before generation
2. **Provides schema-aware prompts** to guide Claude
3. **Validates output** after generation
4. **Auto-corrects errors** if validation fails

### For Claude Interactions

When Claude needs to generate/update JSON:

1. **Read schema requirements**:
   ```bash
   ./tools/schema-enforcer.sh show-requirements current-task
   ```

2. **Include in prompt**:
   ```
   **CRITICAL: SCHEMA COMPLIANCE REQUIRED**
   
   The JSON must strictly follow the schema requirements:
   - All required fields must be present
   - Data types must match exactly
   - Enum values must be from allowed lists
   - String patterns must be followed
   ```

3. **Validate after generation**:
   ```bash
   ./validate-json.sh validate-file generated-file.json
   ```

## Error Prevention Strategy

### Schema-Aware Generation

Instead of generating JSON and fixing errors afterward, the system provides:

- **Detailed schema requirements** before generation
- **Example templates** with correct structure
- **Field validation** for individual values
- **Enum value lists** for restricted fields

### Common Error Prevention

| Error Type | Prevention Method |
|------------|------------------|
| Missing required fields | Show complete required field list |
| Invalid enum values | Provide allowed value lists |
| Wrong data types | Specify exact type requirements |
| Pattern violations | Show regex patterns and examples |
| Empty required strings | Highlight minimum length requirements |

## Integration Status

### Currently Integrated Commands

✅ All 10 task commands are integrated:
- `4-task/1-start/1-From-story.md`
- `4-task/1-start/2-From-issue.md`
- `4-task/1-start/3-From-input.md`
- `4-task/2-execute/1-Plan-Task.md`
- `4-task/2-execute/2-Test-design.md`
- `4-task/2-execute/3-Implement.md`
- `4-task/2-execute/4-Validate-Task.md`
- `4-task/2-execute/5-Review-Task.md`
- `4-task/3-complete/1-Archive-Task.md`
- `4-task/3-complete/2-Status-Task.md`

### Future Integration

The system is designed to easily extend to:
- Story commands (`3-story/`)
- Epic commands (`2-epic/`)
- Project commands (`1-project/`)

## Technical Implementation

### Validation Engine

- **JSON syntax validation**: Using `jq` for reliable parsing
- **Required field checking**: Dynamic schema requirement extraction
- **Enum validation**: Automatic allowed value verification
- **Pattern matching**: Regex pattern validation for IDs and formats

### Auto-Correction System

When validation fails:
1. **Generate correction prompt** with specific error details
2. **Call Claude** with schema-aware correction instructions
3. **Re-validate** after correction
4. **Report results** with clear success/failure indicators

### Dependency Management

- **Primary dependency**: `jq` (widely available JSON processor)
- **Optional dependencies**: Node.js + AJV (for advanced validation)
- **Fallback mode**: Basic validation without external dependencies

## Best Practices

### For Command Authors

1. **Always integrate validation** into commands that generate JSON
2. **Use proactive enforcement** before calling Claude
3. **Validate after generation** and handle failures appropriately
4. **Provide clear error messages** when validation fails

### For Claude Interactions

1. **Always check schema requirements** before generating JSON
2. **Include schema compliance instructions** in prompts
3. **Validate immediately after generation**
4. **Fix errors promptly** using schema guidance

### For Schema Maintenance

1. **Keep schemas up-to-date** with data structure changes
2. **Test schemas regularly** with example data
3. **Document schema changes** and update commands accordingly
4. **Validate existing data** after schema updates

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Validation fails with "command not found" | Ensure scripts are executable: `chmod +x tools/*.sh` |
| Schema not found errors | Check schema files exist in `templates/schemas/` |
| jq not available | Install jq: `brew install jq` or `apt-get install jq` |
| Auto-correction fails | Check Claude CLI is available and configured |

### Debug Commands

```bash
# Check system status
./validate-json.sh status

# Test specific validation
./tools/simple-validator.sh validate-file your-file.json

# Show schema details
./tools/simple-validator.sh show-schema current-task

# Check command integration
./tools/integrate-validation.sh status
```

## Performance

- **Validation speed**: ~100ms per file using jq
- **Schema checking**: Real-time requirements display
- **Auto-correction**: 10-30 seconds depending on Claude response time
- **Integration overhead**: Minimal impact on command execution

## Security

- **Schema validation**: Prevents malformed data injection
- **Input sanitization**: JSON syntax validation before processing  
- **No external dependencies**: Core functionality works with standard tools
- **Safe auto-correction**: Uses controlled prompts to Claude

---

This system ensures reliable, schema-compliant JSON generation throughout the Claude WM CLI workflow, preventing data integrity issues and reducing manual correction overhead.