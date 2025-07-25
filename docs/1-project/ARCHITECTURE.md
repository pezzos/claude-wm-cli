# Claude WM CLI - Technical Architecture

## Overview

Claude WM CLI is an intelligent command-line interface that acts as a wrapper for Claude commands located in `$HOME/.claude/commands`. The application provides both interactive (manual terminal) and headless (VSCode extension integration) operational modes.

## Core Architecture

### Application Entry Points

- **Interactive Mode**: Manual terminal usage for direct user interaction
- **Headless Mode**: Designed for VSCode extension integration and automated workflows

### State Management

#### Project State File
- **Location**: `${PROJECT}/.claude-wm/state.json`
- **Purpose**: Tracks project initialization, command history, metadata, and execution state
- **Structure**:
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
4. **4-TICKET**: Individual ticket implementation
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
- **Branch Management**: Automatic feature/{epic-name} and story/{epic-name}/{story-id} branches
- **Commit Strategy**: Structured commits with epic/story/task traceability
- **Repository Health**: State validation and consistency checking

#### GitHub Integration
- **Issue Processing**: gh CLI integration for issue-driven development
- **Pull Request Workflow**: Automated PR creation and management
- **Project Synchronization**: Issue-to-ticket conversion

#### MCP (Model Context Protocol) Integration
- **consult7**: Full codebase analysis and pattern recognition
- **sequential-thinking**: Complex feature decomposition
- **mem0**: Historical context and pattern storage
- **context7**: Documentation retrieval and best practices
- **IDE diagnostics**: Real-time validation and feedback

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
- **Command Isolation**: Each command operates independently
- **Context Boundaries**: Clear separation between workflow levels
- **State Encapsulation**: Isolated state management per project

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
- **State Persistence**: JSON-based state files for simplicity and portability
- **Documentation Format**: Markdown for human-readable project artifacts
- **Version Control**: Git-centric workflow with branch-per-feature strategy
- **AI Integration**: MCP protocol for standardized AI tool interaction

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
- **Lazy Loading**: On-demand documentation and context loading
- **Caching Strategy**: State and context caching for rapid access
- **Parallel Processing**: Concurrent MCP tool execution where possible
- **Resource Management**: Cleanup and optimization of temporary artifacts

### Monitoring and Metrics
- **Velocity Tracking**: Development speed and throughput analysis
- **Quality Metrics**: Success rates and error pattern analysis
- **Resource Usage**: System resource consumption monitoring
- **User Experience**: Command execution time and responsiveness tracking