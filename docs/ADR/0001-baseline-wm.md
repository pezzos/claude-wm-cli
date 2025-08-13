# ADR-0001: 4-Space Configuration Model for Window Management

**Date:** 2025-08-13  
**Status:** Accepted  
**Deciders:** Architecture Team  
**Technical Story:** Package manager approach for configuration management

## Context

Claude WM CLI needs a robust configuration management system that supports:
- System-wide template distribution
- User customizations and overrides  
- Safe experimentation and testing
- Atomic updates with conflict resolution
- Clear separation of concerns between system and user spaces

Traditional approaches either:
1. Mix system and user files (causing upgrade conflicts)
2. Use simple 2-way sync (losing change attribution)
3. Lack safe experimentation environments

## Decision

We adopt a **4-space configuration model** based on package manager principles:

### Space Definitions

1. **Upstream Space (A)** - `internal/config/system/`
   - Embedded in binary (read-only)
   - System-provided templates and schemas
   - Updated via binary rebuild/upgrade
   - Owner: System developers

2. **Baseline Space (B)** - `.wm/baseline/`  
   - Installed reference state from Upstream
   - Package manager controlled
   - Enables 3-way merge calculations
   - Owner: Package manager

3. **Local Space (L)** - `.claude/`
   - User workspace and customizations
   - Manually editable by users
   - Target of 3-way merge operations
   - Owner: End user

4. **Sandbox Space (S)** - `.wm/sandbox/claude/`
   - Isolated testing environment
   - Safe experimentation without production impact
   - Integration back to system via controlled flow
   - Owner: Developer/tester

### Core Operations

- **Installation:** A → B → L (Upstream to Baseline to Local)
- **Update:** 3-way merge using A, B, L to update L, then update B
- **Sandbox:** A → S (for testing), then S → system (via integration)
- **Validation:** Guard system enforces write boundaries

## Consequences

### Positive

✅ **Clear ownership boundaries:** Each space has defined write permissions  
✅ **3-way merge capability:** Can distinguish user changes from system updates  
✅ **Safe experimentation:** Sandbox isolation prevents production corruption  
✅ **Atomic operations:** All-or-nothing guarantees for complex operations  
✅ **Package manager paradigm:** Familiar install → update → manage workflow  
✅ **Conflict detection:** Automatic identification of merge conflicts  
✅ **Rollback support:** Backup-protected operations with recovery options

### Negative

❌ **Storage overhead:** 4 copies of configuration data consume more disk space  
❌ **Complexity increase:** More directories and concepts for users to understand  
❌ **Migration requirement:** Legacy .claude-wm installations need conversion  
❌ **Performance impact:** Multiple I/O operations for cross-space validations  

### Neutral

🔄 **Directory structure:** Standardized layout may require documentation updates  
🔄 **Command interface:** New commands needed for sandbox and migration operations

## Implementation Details

### File System Layout
```
project/
├── .claude/                    # Local space (L) - User workspace
│   ├── agents/
│   ├── commands/
│   └── settings.json
├── .wm/                        # Package manager data
│   ├── baseline/               # Baseline space (B) - Reference state  
│   │   ├── agents/
│   │   ├── commands/
│   │   └── manifest.json
│   └── sandbox/                # Sandbox space (S) - Testing environment
│       └── claude/
│           ├── agents/
│           └── commands/
└── internal/config/system/     # Upstream space (A) - System templates
    ├── commands/
    ├── manifest.json
    └── settings.json.template
```

### Command Space Authorization
| Command | Reads From | Writes To | Purpose |
|---------|------------|-----------|---------|
| `config install` | A → B → L | B, L | Initial setup |
| `config status` | A, B, L | - | Show diffs |
| `config update` | A, B | B, L | Apply updates |
| `dev sandbox` | A | S | Create test env |
| `dev sandbox diff` | S, B | B | Sync changes |
| `guard check` | Git working tree | - | Validate changes |

### 3-Way Merge Algorithm
```
Compare A (upstream), B (baseline), L (local):

If L == B: User hasn't modified → Apply A to L
If A == B: No system changes → Preserve L  
If A != B and L != B: Changes on both sides → Conflict (manual resolution)
If file in B but not A: Delete operation → Requires --allow-delete flag
```

### Safety Mechanisms
- **Atomic operations:** Temp file → validate → atomic move pattern
- **Backup protection:** Timestamped backups before destructive operations  
- **Boundary enforcement:** Guard system prevents unauthorized writes
- **Schema validation:** JSON/YAML structure validation on all writes
- **Conflict handling:** Manual resolution required for merge conflicts

## Alternatives Considered

### Alternative 1: Simple 2-Space Model (A ↔ L)
**Pros:** Simpler, less storage overhead  
**Cons:** Cannot distinguish user changes from system updates, merge conflicts harder to resolve  
**Rejected:** Insufficient for complex configuration management needs

### Alternative 2: Git-Based Configuration Management  
**Pros:** Mature conflict resolution, version history  
**Cons:** Git complexity exposed to end users, requires Git knowledge  
**Rejected:** Too complex for typical CLI tool users

### Alternative 3: Database-Backed Configuration
**Pros:** Query capabilities, transaction support  
**Cons:** Deployment complexity, file-based tools incompatible  
**Rejected:** Overkill for file-based configuration management

### Alternative 4: Symlink-Based Spaces
**Pros:** No file duplication, instant updates  
**Cons:** Symlink support varies by OS, atomic operations harder  
**Rejected:** Portability and reliability concerns

## Related Decisions

- **ADR-0002:** 3-Way Merge Strategy implementation details
- **Future ADR:** Sandbox integration workflow optimization  
- **Future ADR:** Performance optimization for large configuration sets

## Notes

This decision establishes the foundational architecture for configuration management. Implementation will be incremental, starting with the core 4-space model and gradually adding advanced features like selective updates and automated conflict resolution.

The package manager approach draws inspiration from systems like APT, npm, and Homebrew, adapting their proven patterns to file-based configuration management.