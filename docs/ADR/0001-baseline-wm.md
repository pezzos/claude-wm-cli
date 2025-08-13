# ADR-0001: Baseline Window Management Architecture

## Status
Accepted

## Context
The project started as a CLI tool to manage Claude Code configurations but evolved to support window management functionality. We need to establish a clear baseline for the window management components and their relationship to the existing configuration system.

## Decision
We adopt a 3-tier architecture for window management:

1. **Upstream Embedded** - System templates and configurations embedded in binary
2. **Baseline (.wm/baseline/)** - Immutable snapshot for 3-way merges  
3. **Local (.claude/)** - Runtime configuration used by Claude Code

This maintains the existing configuration management pattern while enabling window management extensions.

## Consequences

### Positive
- Consistent with existing config system
- Enables safe experimentation via sandbox
- Supports atomic operations and rollback
- Clear separation of concerns

### Negative
- Additional complexity in directory structure
- Requires migration path for legacy installations
- Storage overhead for baseline copies

## Implementation Notes
- Window management configs follow same atomic write patterns
- Baseline remains immutable after installation
- Updates use 3-way merge: upstream ↔ baseline ↔ local