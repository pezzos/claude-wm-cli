# Claude WM CLI - Technical Architecture

## Overview

Claude WM CLI is an intelligent command-line interface built in **Go** that acts as a wrapper around `claude -p "/command"` execution. The primary objective is to orchestrate Claude Code commands intelligently by providing context-aware suggestions and guided interaction. The application provides both interactive (guided terminal interface) and headless (VSCode extension integration) operational modes, targeting solo developers who need streamlined project workflow management.

## Core Architecture

### Application Entry Points

- **Interactive Mode**: Guided terminal interface where users are presented with contextual options based on project state. Users never need to memorize complex command names. Progressive guidance with next-step suggestions.
- **Headless Mode**: JSON API mode designed for VSCode extension integration, providing structured input/output for programmatic access. Never runs concurrently with interactive mode.

### State Management

#### Simplified JSON-Based State
- **Design Philosophy**: Simple JSON files for fast parsing and solo-developer usage (sequential workflow, no parallel updates)
- **Recovery Strategy**: Git versioning for state files - rollback to previous version if corruption occurs
- **Core Workflow Files**:
  - `docs/1-project/epics.json` - All epics list with statuses
  - `docs/2-current-epic/current-epic.json` - Currently selected epic
  - `docs/2-current-epic/stories.json` - All stories for current epic + embedded tasks
  - `docs/2-current-epic/current-story.json` - Currently selected story  
  - `docs/2-current-epic/tickets.json` - Interruptions and tickets
  - `docs/3-current-task/current-task.json` - Currently selected task

#### Data Flow Pattern
Each level follows a **List + Current** pattern:
- **List files**: Contain all items at that level (epics.json, stories.json)
- **Current files**: Track the currently selected item (current-epic.json, current-story.json, current-task.json)
- **Tasks**: Embedded within stories.json (no separate todo.json files)

#### State Structure Example
```json
{
  "isInitialized": true,
  "commandHistory": [],
  "projectMetadata": {
    "name": "Claude WM CLI",
    "version": "1.0.0",
    "description": "Claude Workflow Manager Project"
  },
  "settings": {
    "autoSave": true,
    "logRetentionDays": 30
  },
  "hasExecutedImportFeedback": false,
  "hasExecutedPlanEpics": false,
  "lastUpdated": 1753429853403
}
```

### Command Structure & Naming

#### Path-Based Command System
- **Source Location**: `.claude/commands/` directory
- **Naming Convention**: `/{category}/{subcategory}/{command-name}` → `/{category}:{subcategory}:{command-name}`
- **Example**: `/1-project/2-update/1-Import-feedback` → `/1-project:2-update:1-Import-feedback`

#### Hierarchical Command Categories

1. **1-PROJECT**: Project-level initialization and management
2. **2-EPIC**: Epic-level planning and execution
3. **3-STORY**: Story-level task breakdown
4. **4-task**: Individual ticket implementation
5. **Support Tools**: DEBUG, ENRICH, METRICS, LEARNING, VALIDATION

### File Organization System

#### Documentation Structure
```
docs/
├── 1-project/          # Project vision and architecture
├── 2-current-epic/     # Active epic documentation  
├── 3-current-task/     # Current ticket implementation
└── archive/            # Completed work history
```

#### Context-Aware File Management
- **Project Context**: Global vision, roadmap, epics planning
- **Epic Context**: Current epic PRD, stories, and progress tracking
- **Task Context**: Individual ticket implementation details
- **Archive**: Historical record of completed work

## Workflow Engine

### Context Detection Algorithm

#### State Analysis Flow
```
Detection → Analysis → Contextual Suggestion → Next Step Prediction → Execution → State Update
```

#### Decision Tree Logic
1. **Project State Detection**: Check for `.claude-wm/state.json`
2. **Context Analysis**: Analyze existing documentation structure
3. **Command Suggestion**: Provide contextually appropriate next action
4. **Flow Prediction**: Anticipate subsequent workflow steps

### Intelligent Command Routing

#### Mode Detection
- **PROJECT Mode**: Project updates and improvements
- **EPIC Mode**: Epic management and planning
- **STORY Mode**: Story breakdown and task extraction
- **TICKET Mode**: Implementation and execution
- **COMPLETION Mode**: Finalization and archival

#### Context-Specific Suggestions
- File existence analysis determines available actions
- Dependency verification ensures proper workflow sequence
- State tracking prevents inconsistent operations

## Integration Architecture

### External System Interfaces

#### Git Integration
- **Branch Management**: One branch per User Story - all tickets within a story use the same branch
- **Commit Strategy**: Structured commits with epic/story/task traceability
- **Repository Health**: State validation and consistency checking
- **Interruption Handling**: Emergency fixes and GitHub issues become tickets within the current story branch

#### GitHub Integration
- **Issue Processing**: gh CLI integration for issue-driven development
- **Pull Request Workflow**: Automated PR creation and management
- **Project Synchronization**: Issue-to-ticket conversion

#### MCP (Model Context Protocol) Integration
- **Design Philosophy**: Optional enhancements, not critical dependencies
- **Graceful Degradation**: CLI functions fully without MCP tools available
- **Available Integrations**:
  - **consult7**: Full codebase analysis and pattern recognition (optional)
  - **sequential-thinking**: Complex feature decomposition (optional) 
  - **mem0**: Historical context and pattern storage (optional)
  - **context7**: Documentation retrieval and best practices (optional)
- **Fallback Strategy**: Continue operation without MCP tools for core workflow functionality

## Quality Assurance Architecture

### Multi-Level Validation
- **Architectural Review**: Comprehensive system analysis
- **Dependency Analysis**: Inter-component relationship mapping
- **Performance Monitoring**: Metrics collection and trend analysis
- **Security Auditing**: Code security and vulnerability assessment

### Continuous Learning System
- **Pattern Recognition**: Success/failure pattern analysis
- **Performance Optimization**: Velocity and efficiency tracking
- **Knowledge Management**: Automated context enrichment
- **Adaptive Workflows**: Self-improving process optimization

## Scalability Considerations

### Modular Design
- **Command Isolation**: Each command operates independently via `claude -p "/command"` execution
- **Context Boundaries**: Clear separation between workflow levels (Project → Epic → Story → Ticket)
- **State Encapsulation**: Isolated JSON state files per project for solo-developer usage
- **No Concurrency Issues**: Single-user model eliminates race condition concerns
- **Performance Optimization**: Only current epic/story loaded, archived data not in memory
- **Future Migration Path**: Database migration considered for extreme scaling needs

### Extension Points
- **Custom Commands**: Plugin architecture for domain-specific workflows
- **Integration Hooks**: API endpoints for external tool integration
- **Workflow Customization**: Configurable process templates

## Technical Decisions

### Architecture Patterns
- **Command Pattern**: Encapsulated command execution with undo capability
- **State Machine**: Workflow state transitions with validation
- **Observer Pattern**: Event-driven state synchronization
- **Template Method**: Consistent command execution framework

### Technology Choices
- **Implementation Language**: **Go** for portability, performance, and single binary deployment
- **CLI Framework**: Cobra for command structure and Bubble Tea for interactive interface
- **State Persistence**: Simple JSON files for fast parsing and solo-developer usage
- **Documentation Format**: Markdown for human-readable project artifacts
- **Version Control**: Git-centric workflow with branch-per-story strategy (one branch per User Story)
- **AI Integration**: Optional MCP protocol enhancements (non-critical dependencies)
- **Error Handling**: Timeout-based command execution with graceful failure recovery
- **Concurrent Access**: File locking mechanism for multi-terminal protection (if needed)

## Security Architecture

### Access Control
- **Local File System**: Commands operate within project boundaries
- **Git Operations**: Secure repository operations with validation
- **External APIs**: Authenticated integration with GitHub and other services

### Data Protection
- **Sensitive Information**: Automatic detection and exclusion from commits
- **Audit Trail**: Complete command history and state change tracking
- **Backup Strategy**: Archive-based historical preservation

## Performance Characteristics

### Efficiency Optimizations
- **Fast JSON Parsing**: Simple state files for rapid project context loading  
- **Single Binary**: Go compilation produces portable executable with no dependencies
- **Lazy Loading**: On-demand documentation and context loading
- **Minimal Dependencies**: Core functionality works without external MCP tools
- **Resource Management**: Cleanup and optimization of temporary artifacts

### Monitoring and Metrics
- **Structured Logging**: JSON logs for debugging and analysis
- **Command Execution Tracking**: Response times and success rates
- **State Change Auditing**: All project state modifications logged
- **User Experience**: Interactive interface responsiveness optimization

## Development Philosophy

### MVP-First Approach
- **Pragmatic Development**: Optimized for rapid solo-developer iteration
- **Graceful Failure**: If implementation fails, rollback and retry with versioned state
- **Simplicity Over Complexity**: Intelligent wrapper around Claude Code, not a complex system
- **Personal Tool First**: Built for immediate personal use, then generalized

### Workflow Automation Vision
- **Progressive Guidance**: Users guided step-by-step with next-action suggestions
- **Automated Implementation Mode**: Future "implement everything" mode for full epic/story automation
- **Strict Workflow Enforcement**: No implementation without proper planning breakdown
- **Contextual Intelligence**: Always show current position and suggest next steps

## Implementation Strategy

### Phase 1: Interactive CLI Core
- Go-based CLI with Cobra/Bubble Tea interface
- JSON state parser (state.json, epics.json, stories.json, tickets.json)
- Claude Code command wrapper with timeout-based error handling
- Interactive navigation through workflow options with step-by-step guidance

### Phase 2: Headless Mode
- JSON API mode for VSCode extension integration (never runs concurrently with interactive CLI)
- Structured input/output for programmatic access with intermediate status guides
- Command execution with structured logging
- Extension-CLI separation: extension calls CLI in headless mode and displays JSON output

### Phase 3: VSCode Extension Integration
- Extension uses CLI in headless mode
- Real-time project state synchronization
- Visual workflow representation and control