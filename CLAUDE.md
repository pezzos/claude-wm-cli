# Claude Code Project Configuration

## Project: Claude WM CLI

### Project Type
Window Management CLI Tool with AI Integration

### Development Approach
- **Framework**: CLI-based application
- **Language**: [Auto-detected based on first implementation]
- **Architecture**: Command-line interface with AI integration
- **Testing**: Unit tests for core functionality

### Key Commands
- **Lint**: `[to be configured]`
- **Test**: `[to be configured]`
- **Build**: `[to be configured]`
- **Deploy**: `[to be configured]`

### Development Workflow
This project uses the Claude Code agile workflow system:

1. **Epic Planning**: Use `/project:agile:start` to begin new features
2. **Technical Design**: Use `/project:agile:design` for architecture
3. **Implementation Planning**: Use `/project:agile:plan` for task breakdown
4. **Development**: Use `/project:agile:iterate` for implementation cycles
5. **Delivery**: Use `/project:agile:ship` for completion and deployment

### Memory Context
- **Project Focus**: Window management CLI with Claude AI integration
- **Target Platform**: Command-line interface
- **Integration Points**: AI services, window management APIs
- **Architecture Style**: Modular CLI design
- The .claude and .claude-wm folders are used to manage the project but are not part of the core of the project, just like .git and .vscode, which are also folders used to manage the project. You should therefore not modify any files in there unless explicitly requested.

### MCP Tools Configuration
- **Consult7 Optimization**: Always use exclude patterns to reduce token consumption and costs
  - **Standard Exclusions**: `".*logs/.*|.*metrics/.*|.*claude(-wm)?/.*|.*backup/.*|.*archive/.*|.*serena/.*"`
  - **Usage**: Add `exclude_pattern` parameter to all `mcp__consult7__consultation` calls
  - **Benefits**: ~35% cost reduction and faster processing

### Quality Standards
- Follow established CLI best practices
- Implement comprehensive error handling
- Maintain clear documentation
- Ensure cross-platform compatibility

##  Using Serena MCP
- Project uses Serena MCP for semantic code operations
- Use Serena's `find_symbol()` and `get_symbols_overview()` instead of reading entire files to reduce token consumption.
- Always run `/mcp__serena__initial_instructions` at session start.

---

*Auto-generated during project initialization*
*Update this file as the project evolves*