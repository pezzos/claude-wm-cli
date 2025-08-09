# Project Context for Claude

## Project Overview
{{PROJECT_NAME}} - {{PROJECT_DESCRIPTION}}

## Architecture Patterns
{{ARCHITECTURE_PATTERNS}}

## Coding Standards
{{CODING_STANDARDS}}

## Common Commands
{{COMMON_COMMANDS}}

## Lessons Learned
{{LESSONS_LEARNED}}

## Development Environment
{{DEV_ENVIRONMENT}}
### Using Serena MCP
- Project uses Serena MCP for semantic code operations
- Use Serena's `find_symbol()` and `get_symbols_overview()` instead of reading entire files to reduce token consumption.
- Always run `/mcp__serena__initial_instructions` at session start.
### MCP Tools Configuration
- **Consult7 Optimization**: Always use exclude patterns to reduce token consumption and costs
  - **Standard Exclusions**: `".*logs/.*|.*metrics/.*|.*claude(-wm)?/.*|.*backup/.*|.*archive/.*|.*serena/.*"`
  - **Usage**: Add `exclude_pattern` parameter to all `mcp__consult7__consultation` calls
  - **Benefits**: ~35% cost reduction and faster processing

## Quality Standards
{{QUALITY_STANDARDS}}