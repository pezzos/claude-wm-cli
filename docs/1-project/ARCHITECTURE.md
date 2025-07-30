# Claude WM CLI - Technical Architecture

## Overview

**Claude WM CLI** is a production-grade Go-based command-line interface for agile project management designed specifically for solo developers. Built with enterprise-level robustness including atomic state management, JSON schema validation, cross-platform file locking, and comprehensive error recovery systems.

**Current Status**: 75+ Go files implementing a complete workflow management system with high test coverage, production-ready patterns, and proven architectural decisions.

## Core Architecture

### Architectural Principles

**Atomic Operations**: All state changes use temp-file + rename pattern to prevent corruption
**Schema-First Design**: 7 comprehensive JSON schemas enforce data structure integrity  
**Interface-Driven**: 48 internal packages with clear separation of concerns
**Cross-Platform**: Native Windows and Unix compatibility with platform-specific optimizations
**Error Recovery**: Multi-layer validation and automatic repair mechanisms

### Application Entry Points

- **Interactive Mode**: Context-aware navigation system that analyzes project state and suggests intelligent next actions
- **Direct Commands**: Full CLI command suite for programmatic access (epic, story, ticket, project, config, git, github, lock)
- **Schema Validation**: PostToolUse hooks automatically validate all JSON state changes

### State Management Architecture

#### Production-Grade JSON State System
- **Atomic Operations**: Temp-file + rename pattern eliminates corruption risks
- **Schema Validation**: 7 JSON schemas with PostToolUse hooks enforce structure integrity
- **File Locking**: Cross-platform exclusive locks prevent concurrent access
- **Backup System**: Automatic backup creation with retention policies
- **Git Integration**: All state changes automatically versioned with recovery points

#### State File Hierarchy
```
docs/
├── 1-project/
│   └── epics.json                    # Master epic list (EPIC-XXX pattern)
├── 2-current-epic/
│   ├── current-epic.json            # Selected epic context
│   ├── stories.json                 # Stories map (STORY-XXX keys)
│   └── current-story.json           # Selected story context
└── 3-current-task/
    ├── current-task.json            # Active task (TASK-XXX)
    └── iterations.json              # Task attempt tracking
```

#### Schema-Enforced Structure
- **Epic Schema**: Business value, success criteria, story themes, dependencies
- **Story Schema**: Acceptance criteria, embedded tasks, epic relationships
- **Task Schema**: 13 required sections (technical context, analysis, reproduction, investigation, implementation, resolution, interruption context)
- **Iteration Schema**: Attempt tracking with outcomes, learnings, and recommendations
- **Metrics Schema**: 8 performance dimensions with comprehensive analytics

#### Atomic State Operations
```go
// Production-ready atomic write pattern
type AtomicWriter struct {
    targetPath   string
    tempPath     string
    backupPath   string
    verification func([]byte) error
}

func (aw *AtomicWriter) WriteJSON(data interface{}) error {
    // 1. Create backup of existing file
    // 2. Marshal JSON with validation
    // 3. Write to temp file with permissions
    // 4. Verify content integrity
    // 5. Atomic rename temp -> target
    // 6. Update Git if enabled
}
```

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

### Command Architecture 

#### Cobra-Based CLI Structure
```go
// Root command with global configuration
rootCmd := &cobra.Command{
    Use:   "claude-wm-cli",
    Short: "Intelligent workflow management for solo developers",
}

// Command registration pattern
func init() {
    rootCmd.AddCommand(epicCmd)
    rootCmd.AddCommand(storyCmd)
    rootCmd.AddCommand(ticketCmd)
    // ... additional commands
}
```

#### Complete Command Hierarchy

**Primary Commands (18 total)**:
- `init` - Initialize project structure
- `status` - Show current project state
- `interactive` (aliases: `nav`, `menu`) - Context-aware navigation
- `project` - Project-level workflow management
- `epic` - Epic CRUD and dashboard operations
- `story` - Story management and generation
- `ticket` - Task/interruption handling with full workflow
- `config` - Configuration management
- `git` - Git integration and versioning
- `github` - GitHub OAuth and issue synchronization
- `lock` - File locking operations
- `interrupt` - Workflow interruption management
- `help` - Enhanced help system
- `version` - Version and build information

#### Command Implementation Pattern
```go
type CommandContext struct {
    ProjectPath   string
    StateManager  *state.Manager
    LockManager   *locking.Manager
    GitManager    *git.Manager
    Validator     *workflow.Validator
}
```

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

## Workflow Engine Architecture

### Context Detection System

#### Navigation Context Analysis
```go
type ContextDetector struct {
    projectPath string
}

type ProjectContext struct {
    State            WorkflowState
    CurrentEpic      *EpicContext
    CurrentStory     *StoryContext  
    CurrentTask      *TaskContext
    AvailableActions []string
    Issues           []string
}

// Context detection flow
func (cd *ContextDetector) DetectContext() (*ProjectContext, error) {
    // 1. Validate project structure
    // 2. Load current epic/story/task state
    // 3. Analyze workflow dependencies
    // 4. Generate available actions
    // 5. Provide intelligent suggestions
}
```

#### Workflow State Machine
```go
type WorkflowState int

const (
    StateNotInitialized WorkflowState = iota
    StateProjectInitialized
    StateHasEpics
    StateEpicInProgress
    StateStoryInProgress
    StateTaskInProgress
)
```

### Intelligent Action System

#### Action Registry Pattern
```go
type Action struct {
    ID           string
    Name         string
    Description  string
    Prerequisites []string
    Command      string
    Priority     int
}

type ActionValidator struct {
    currentContext *ProjectContext
    dependencies   map[string][]string
}
```

#### Suggestion Engine
The system analyzes current state and provides prioritized recommendations:
- **Dependency Analysis**: Ensures workflow prerequisites are met
- **Context Awareness**: Suggests actions based on current position
- **Progress Tracking**: Monitors completion status and suggests next steps
- **Error Recovery**: Provides repair actions when issues detected

## Internal Package Architecture

### Core Business Logic (48 Files)

#### Entity Management Layer
```go
// internal/epic/manager.go
type Manager struct {
    stateWriter  *state.AtomicWriter
    validator    *validation.Validator
    gitManager   *git.Manager
}

// internal/story/generator.go  
type Generator struct {
    stateManager *state.Manager
    epicManager  *epic.Manager
}

// internal/ticket/manager.go
type Manager struct {
    stateManager   *state.Manager
    lockManager    *locking.Manager
    preprocessor   *preprocessing.TaskProcessor
}
```

#### State Management Layer
```go
// internal/state/atomic.go
type AtomicWriter struct {
    targetPath     string
    verification   func([]byte) error
    backupEnabled  bool
    gitEnabled     bool
}

// internal/state/corruption.go
type CorruptionDetector struct {
    checksumValidator  func([]byte) error
    schemaValidator    func([]byte) error
    recoveryStrategies []RecoveryStrategy
}

// internal/state/performance.go
type PerformanceMetrics struct {
    operationTimes map[string]time.Duration
    memoryUsage    map[string]int64
    cacheStats     map[string]CacheMetrics
}
```

#### Integration Layer
```go
// internal/git/repository.go
type Repository struct {
    workingDir     string
    stateManager   *state.Manager
    backupStrategy BackupStrategy
}

// internal/github/integration.go
type Integration struct {
    client       *github.Client
    auth         *AuthManager
    syncHistory  *SyncHistory
    rateLimiter  *RateLimiter
}

// internal/locking/filelock.go
type FileLock struct {
    file       *os.File
    platform   string  // "windows" or "unix"
    metadata   LockMetadata
    staleCheck func() bool
}
```

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
- JSON state parser (epics.json, stories.json, ...)
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