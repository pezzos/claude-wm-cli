# Implementation Architecture - Claude WM CLI

## Vue d'Ensemble Technique

Claude WM CLI est un outil CLI développé en Go qui implémente un workflow complet de gestion de projets agiles. L'architecture privilégie la robustesse, l'atomicité des opérations, et une expérience utilisateur fluide.

**État Actuel**: Implémentation largement fonctionnelle avec 75+ fichiers Go, test coverage élevée, et patterns production-ready. Les fonctionnalités core sont opérationnelles et testées.

## Architecture Decisions

### 1. Programming Language & Framework Choice

**Go + Cobra Framework**
- **Why Go**: Fast compilation, static binaries, excellent concurrency support, cross-platform compatibility
- **Why Cobra**: Industry standard for CLI tools, built-in help generation, subcommand organization
- **Benefits**: Single binary deployment, no runtime dependencies, consistent performance

**Alternative Considered**: Python with Click/Typer was considered but rejected due to deployment complexity and dependency management.

### 2. State Management Strategy

**JSON File-Based State with Atomic Operations**
```go
// All state operations use atomic write pattern
type AtomicWriter struct {
    targetPath string
    tempPath   string
}
```

**Key Design Decisions**:
- **Atomic File Operations**: Temp file + rename pattern prevents corruption
- **Hierarchical State Structure**: `docs/1-project/`, `docs/2-current-epic/`, `docs/3-current-task/`
- **JSON Schema Validation**: All state files validate against expected schema
- **Backup Before Write**: Automatic backup creation with retention policies

**Why Not Database**: 
- Solo developer focus = no concurrent users
- Files integrate naturally with Git versioning
- Human-readable state aids debugging
- No deployment/maintenance overhead

### 3. Concurrency & File Safety

**Multi-Layer Protection**:
```go
// File locking implementation
type FileLock struct {
    file     *os.File
    locked   bool
    platform string // "windows" or "unix"  
}
```

- **File Locking**: Cross-platform (Windows/Unix) exclusive locks
- **Lock Status Tracking**: Active lock monitoring and cleanup
- **Atomic State Updates**: Prevents partial writes during interruption
- **Corruption Detection**: Checksum validation and recovery suggestions

### 4. Command Structure & Navigation

**Three-Tier Command Architecture**:
1. **Direct Commands**: `claude-wm-cli epic create "Name"`
2. **Interactive Navigation**: `claude-wm-cli interactive` with context-aware menus  
3. **Context Suggestions**: Intelligent next-action recommendations

**Navigation Implementation**:
```go
type NavigationContext struct {
    CurrentPhase    WorkflowPhase
    AvailableActions []Action
    Suggestions     []Suggestion
    Dependencies    []Dependency
}
```

### 5. Workflow State Machine

**Hierarchical Workflow Enforcement**:
- **Project → Epic → Story → Ticket** progression
- **Dependency Validation**: Cannot skip workflow levels
- **State Transitions**: Controlled state changes with validation
- **Interruption Support**: Context preservation and restoration

**Implementation Pattern**:
```go
type WorkflowValidator struct {
    currentState  *State
    dependencies  map[string][]string
    transitions   map[string][]string
}
```

### 6. External Integrations

**Git Integration**:
- **Auto-Commit**: All state changes automatically versioned
- **Recovery System**: Rollback to any previous state
- **Backup Strategy**: Multiple backup levels (local, git, external)

**GitHub Integration**:
- **OAuth Authentication**: Secure token-based access
- **Issue Synchronization**: Bi-directional sync with rate limiting
- **Webhook Support**: Real-time issue updates (planned)

**Integration Architecture**:
```go
type Integration interface {
    Connect() error
    Sync() error
    GetStatus() IntegrationStatus
    TestConnection() error
}
```

## Core Components

### 1. State Management (`internal/state/`)

**AtomicWriter**: Ensures corruption-free writes
**CorruptionDetector**: Validates JSON integrity and suggests fixes
**BackupManager**: Automated backup with rotation and cleanup
**StateLoader**: Optimized loading for large JSON files

### 2. Workflow Engine (`internal/workflow/`)

**PhaseAnalyzer**: Determines current workflow position
**DependencyValidator**: Enforces workflow dependencies  
**TransitionManager**: Controls state transitions
**ContextPreserver**: Saves/restores workflow context during interruptions

### 3. Navigation System (`internal/navigation/`)

**ContextDetector**: Analyzes project state to determine current phase
**SuggestionEngine**: Provides prioritized next actions with reasoning
**MenuSystem**: Interactive command selection with keyboard shortcuts
**ActionValidator**: Ensures actions are valid for current context

### 4. Integration Layer (`internal/git/`, `internal/github/`)

**GitManager**: State versioning, backup, and recovery operations
**GitHubClient**: Issue synchronization with rate limiting and error handling
**AuthManager**: Token management and OAuth flows
**SyncEngine**: Bi-directional data synchronization

## Performance Characteristics

### Benchmarks & Targets

**State Operations**:
- Small files (<1MB): <100ms read/write
- Medium files (1-10MB): <500ms read/write  
- Large files (10-100MB): <5s read/write
- Memory usage: <50MB baseline, <200MB peak

**Command Execution**:
- Simple commands: <200ms response time
- Complex operations: <2s response time
- Interactive navigation: <100ms menu rendering

**File Operations**:
- Atomic writes: <50ms overhead
- Backup creation: <500ms for typical project
- Lock acquisition: <10ms

### Optimizations Implemented

**Lazy Loading**: JSON files loaded on-demand
**Streaming Parser**: Memory-efficient parsing for large files
**Connection Pooling**: HTTP clients reuse connections
**Cache Layer**: Frequently accessed data cached in memory

## Error Handling Strategy

### Recovery Mechanisms

**State Corruption**:
1. Detect corruption via checksum validation
2. Attempt automatic repair from backup
3. Offer manual recovery options
4. Prevent further operations until resolved

**External Service Failures**:
1. Graceful degradation (continue with local state)
2. Retry with exponential backoff
3. Cache last known good state
4. User notification with manual retry options

**Concurrent Access**:
1. File locking prevents conflicts
2. Lock timeout prevents deadlocks  
3. Cleanup orphaned locks on startup
4. Clear error messages for lock conflicts

## Testing Strategy

### Test Coverage

**Unit Tests**: All core business logic (>90% coverage)
**Integration Tests**: External service interactions
**End-to-End Tests**: Complete workflow scenarios
**Performance Tests**: Benchmark critical operations

### Test Categories

```go
// Example test structure
func TestEpicCRUD(t *testing.T) {
    // Setup isolated test environment
    // Test create, read, update, delete operations
    // Verify state consistency and error handling
}
```

**Isolation**: Each test uses temporary directories
**Mocking**: External services mocked for unit tests
**Race Detection**: Tests run with `-race` flag
**Coverage**: Enforced minimum coverage thresholds

## Security Considerations

### Data Protection

**Local State**: File permissions restricted to user only
**Backup Files**: Encrypted backups for sensitive projects (planned)
**Token Storage**: OAuth tokens stored in system keychain
**Network Communications**: All external API calls use HTTPS

### Input Validation

**Command Arguments**: Strict validation and sanitization
**JSON State**: Schema validation prevents malformed data
**File Paths**: Path traversal attack prevention
**External Data**: All GitHub/Git input validated before processing

## Deployment & Distribution

### Build Process

```makefile
# Multi-platform build targets
build-all: build-linux build-windows build-darwin
build-linux:   GOOS=linux GOARCH=amd64 go build ...
build-windows: GOOS=windows GOARCH=amd64 go build ...
build-darwin:  GOOS=darwin GOARCH=amd64 go build ...
```

**Release Artifacts**:
- Single binary per platform (no dependencies)
- Installation scripts for common package managers
- Documentation and examples included

### System Requirements

**Minimum**: 
- Go 1.21+ (for building)
- 50MB disk space
- 64MB RAM

**Recommended**:
- Git installed (for versioning features)
- GitHub CLI (for enhanced GitHub integration)
- 200MB disk space (with backups)

## Analyse des Choix Techniques Réalisés

### Décisions Validées par l'Implémentation

1. **Go + Cobra**: Excellent choix - code structuré, performant, maintenable
2. **JSON + Atomic writes**: Robuste en pratique, bon équilibre simplicité/fiabilité  
3. **File locking multiplateforme**: Fonctionne bien, évite effectivement les corruptions
4. **Git integration**: Seamless versioning, recovery points efficaces
5. **Modular architecture**: `internal/` packages bien séparés, réutilisables

### Innovations Remarquables

1. **Interruption Stack**: Système unique de préservation de contexte
2. **Navigation contextuelle**: Interface intelligente adaptée à l'état
3. **Atomic state management**: Prévention réelle des corruptions  
4. **Cross-platform locking**: Solution robuste Unix/Windows
5. **Performance optimizations**: JSON streaming, memory pooling

## État de Maturité par Composant

### ✅ Production-Ready (>90% complet)
- Epic/Story/Ticket CRUD avec validation complète
- Atomic state management avec corruption protection
- File locking et concurrent access prevention
- Git integration avec auto-commit et recovery
- GitHub OAuth et issue synchronization
- Cross-platform compatibility (Unix/Windows)
- Comprehensive test suite (unit + integration)

### 🔄 Beta-Ready (70-90% complet)  
- Interactive navigation (logique complète, UX à finaliser)
- Interruption context preservation (structure OK, restoration partielle)
- Performance optimizations (implémenté mais à valider à grande échelle)
- Error handling et recovery (robuste mais UI perfectible)

### 🚧 Alpha-Level (<70% complet)
- Task-level granular management dans l'interface CLI
- Advanced workflow analytics et metrics
- Plugin/extension architecture
- Webhook support pour GitHub
- Database backend alternatif

## Monitoring & Observability

### Metrics Collection

**Performance Metrics**:
- Command execution times
- File operation latencies  
- Memory usage patterns
- Error rates by operation type

**Usage Analytics** (Planned):
- Command frequency analysis
- Workflow pattern recognition
- Performance bottleneck identification
- User experience optimization opportunities

### Debugging Support

**Verbose Logging**: Detailed operation logging with levels
**State Inspection**: Commands to examine internal state
**Performance Profiling**: Built-in profiling for optimization
**Error Reporting**: Structured error messages with context

---

*Last Updated: 2025-07-25*
*Implementation Status: Core features complete, integrations functional*