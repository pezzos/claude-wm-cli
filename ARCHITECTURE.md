# Claude WM CLI - Architecture Documentation

## Overview

Claude WM CLI is a mature Go-based command-line tool that provides structured workflow management for solo developers. The architecture emphasizes atomic state management, robust error handling, and seamless integration with development tools while maintaining simplicity for the target audience.

## Current Implementation Status

### ✅ Completed Core Systems
- **75+ Go files** with mature modular architecture
- **Atomic state management** with corruption protection
- **Cross-platform file locking** (Unix/Windows)
- **Git/GitHub integration** with OAuth support
- **Interactive navigation** with complete logic implementation
- **Comprehensive test coverage** (unit and integration tests)
- **Performance optimizations** (streaming JSON, memory pooling)

### 🔄 Remaining Completion Tasks
- **Interactive action execution** - Navigation logic complete, some actions need wiring
- **Context restoration** - Interruption stack structure complete, file/git restoration needs completion
- **Task-level CLI interface** - CRUD operations for granular task management
- **Large-scale validation** - Testing with 1000+ epics/stories

## Architectural Design Principles

### 1. Atomic State Management
```
State Operations:
├── Atomic file writes with temp files
├── File locking for concurrent access prevention
├── Backup creation before modifications
└── Rollback capability on failure
```

**Key Features:**
- All state changes use atomic operations to prevent corruption
- Multi-platform file locking prevents concurrent access issues
- Comprehensive backup and recovery system
- Git integration for versioned state management

### 2. Modular Package Structure
```
internal/
├── epic/          # Epic lifecycle management
├── story/         # Story generation and tracking
├── task/        # Interruption and task handling
├── interactive/    # Interactive menu system
├── state/         # JSON state management core
├── backup/        # Backup and recovery
├── git/          # Git integration layer
├── github/       # GitHub API integration
├── locking/      # File locking utilities
└── workflow/     # Workflow analysis and validation
```

### 3. Error Handling and Recovery
- **Graceful degradation** when external services unavailable
- **Timeout management** for external command execution
- **Circuit breaker patterns** for integration resilience
- **Automatic recovery** from common error scenarios

## Strategic Architecture Considerations

### High-Priority Architectural Challenges

#### 1. State Machine Evolution Pressure
**Challenge**: Current linear epic → story → task progression needs to evolve to support parallel workflows and complex dependency graphs.

**Current State**: Atomic file I/O works well for sequential operations
**Risk**: Bottlenecks with parallel operations as complexity grows
**Mitigation Strategy**: 
- Design extension points for parallel workflow support
- Consider event-driven state updates for complex scenarios
- Maintain backward compatibility with current simple workflows

#### 2. Plugin Architecture Readiness
**Challenge**: Current `internal/` package structure has tight coupling that could complicate plugin development.

**Current State**: Well-structured but internally coupled packages
**Risk**: Plugin system could require breaking changes to internal APIs
**Mitigation Strategy**:
- Define stable plugin interfaces early
- Create abstraction layers for core functionality
- Maintain API compatibility contracts

#### 3. Integration Fragility Management
**Challenge**: Architecture depends on external integrations (GitHub API, Git, Claude Code) without circuit breaker patterns.

**Current State**: Basic error handling for external services
**Risk**: Single integration failure could cascade across workflow system
**Mitigation Strategy**:
- Implement circuit breaker patterns for external dependencies
- Add degraded mode operation capabilities
- Create fallback mechanisms for critical operations

### Performance and Scalability

#### Memory Management
**Current**: Good performance up to 100MB state files with lazy loading
**Considerations**: 
- Memory pooling for daemon mode (VSCode extension compatibility)
- Garbage collection optimization for long-running processes
- Streaming operations for large datasets

#### Command Execution Reliability
**Current**: Direct command execution with basic timeout handling
**Enhancements Needed**:
- Command format versioning for Claude Code evolution
- Retry logic with exponential backoff
- Output parsing with format validation

## Integration Architecture

### Claude Code Integration Strategy

The feedback analysis reveals a significant opportunity to bridge the gap between `claude-wm-cli` (robust state management) and `~/.claude/commands/` (rich templates and AI capabilities).

#### Proposed Integration Layer
```go
// internal/claude/executor.go - New package
type ClaudeExecutor struct {
    commandsPath string
    timeout      time.Duration
    cache        *PromptCache
}

func (ce *ClaudeExecutor) ExecutePrompt(path string, context map[string]interface{}) (*Response, error)
```

#### Missing Command Categories
1. **LEARNING System**: Pattern recognition and optimization
2. **METRICS System**: Enhanced analytics with AI analysis  
3. **ENRICHMENT System**: Context enhancement capabilities
4. **TEMPLATE System**: Automated document generation
5. **VALIDATION System**: Architecture review and quality assurance

### Configuration Architecture
```json
{
    "claude_commands_path": "/Users/user/.claude/commands",
    "enhanced_mode": true,
    "claude_cli_path": "claude",
    "cache_enabled": true,
    "cache_ttl": "5m",
    "timeout": "30s"
}
```

#### Mode Detection Strategy
- **Enhanced Mode**: ~/.claude/commands exists → AI-powered features
- **Basic Mode**: Fallback → Core workflow management only
- **Hybrid Mode**: Selective enhanced features based on availability

## Security Architecture

### Current Security Model
- File permissions for state protection
- OAuth token storage for GitHub integration
- Single-user trusted environment assumptions

### Enterprise Security Considerations
**Identified Gaps**:
- Token storage lacks encryption at rest
- No audit trail capabilities
- Limited multi-user access controls

**Future Enhancements**:
- Encrypted state files for sensitive projects
- Audit trail implementation
- Role-based access controls for team environments

## Performance Benchmarks and Targets

### Current Performance
- **State Operations**: Atomic writes with minimal overhead
- **Memory Usage**: Efficient with lazy loading up to 100MB
- **File Locking**: Cross-platform with minimal contention
- **JSON Processing**: Streaming operations for large datasets

### Target Metrics
- **Plugin Compatibility**: 0 breaking changes after initial plugin API design
- **Memory Stability**: <50MB growth over 24h daemon operation
- **Integration Resilience**: 99.9% uptime despite external service failures
- **User Experience**: Navigation complexity doesn't reduce daily usage frequency

## Technology Stack Validation

### Core Stack Justification
- **Go + Cobra**: Excellent choice validated by implementation success
- **Bubble Tea**: Appropriate for interactive CLI needs
- **JSON + Atomic writes**: Robust approach confirmed in practice
- **Git integration**: Seamless versioning implementation

### Stack Advantages for Target Use Case
- **Solo Developer Focus**: Simple, fast iteration without team complexity
- **Cross-platform**: Works consistently across development environments
- **Minimal Dependencies**: Reduces installation and maintenance overhead
- **Performance**: Fast startup and execution for daily development workflow

## Migration and Evolution Strategy

### Backward Compatibility Strategy
- All existing commands continue unchanged
- Enhanced features are additive, not replacements
- Configuration flags for disabling advanced features if needed

### Progressive Enhancement Phases
1. **Phase 1**: Infrastructure for prompt execution (2-4 weeks)
2. **Phase 2**: Enhanced existing commands with AI capabilities (1-2 months)
3. **Phase 3**: New command categories (learning, metrics, enrichment) (1-2 months)
4. **Phase 4**: Performance optimization and UX refinement (3-6 months)

## Risk Assessment and Mitigation

### High-Risk Areas
1. **External Dependency Changes**: Claude Code command format evolution
2. **State Corruption**: File system failures during atomic operations
3. **Memory Leaks**: Long-running daemon processes
4. **Integration Failures**: GitHub API rate limiting or authentication issues

### Mitigation Strategies
1. **Command Format Versioning**: Version detection and compatibility layers
2. **Backup and Recovery**: Comprehensive backup before all operations
3. **Memory Management**: Explicit resource cleanup and monitoring
4. **Circuit Breakers**: Graceful degradation when integrations fail

## Future Architecture Considerations

### Collaborative Development Evolution
**Current**: File-based state optimized for solo development
**Future Consideration**: Evolution path to distributed state for team collaboration
**Strategy**: Design migration paths without breaking existing single-user workflows

### Plugin Ecosystem
**Current**: Monolithic internal architecture
**Future**: Extensible plugin framework for custom integrations
**Strategy**: Define stable interfaces early to prevent breaking changes

### AI Integration Evolution
**Current**: Basic Claude Code command execution
**Future**: Deep AI integration with learning and optimization
**Strategy**: Build flexible prompt execution infrastructure that can evolve with AI capabilities

## Conclusion

The Claude WM CLI architecture is significantly more mature and robust than initial assessments suggested. The core systems are well-implemented with strong foundations for future enhancement. The primary focus should be on:

1. **Completing the remaining 10-15%** of interactive functionality
2. **Integrating with Claude Code ecosystem** for enhanced AI capabilities  
3. **Validating at scale** with large projects and complex workflows
4. **Preparing for plugin architecture** to enable future extensibility

The architecture successfully balances simplicity for solo developers with sufficient robustness for professional development workflows.