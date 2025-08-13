# Architecture Guide

Claude WM CLI implements a **package manager approach** for configuration management using a 4-space model with 3-way merge capabilities.

## Configuration Spaces

### Space Definitions

#### Upstream Space (A)
- **Path**: `internal/config/system/`
- **Nature**: Embedded in binary, read-only
- **Content**: System templates, commands, schemas, hooks
- **Update**: Via `go build` (new binary version)
- **Owner**: System/developers

#### Baseline Space (B)
- **Path**: `.wm/baseline/`
- **Nature**: Installed reference state
- **Content**: Last known good configuration from Upstream
- **Update**: Via `config install` or `config update`
- **Owner**: Package manager

#### Local Space (L)
- **Path**: `.claude/`
- **Nature**: User workspace
- **Content**: User customizations, overrides, local state
- **Update**: Manual editing or command execution
- **Owner**: User

#### Sandbox Space (S)
- **Path**: `.wm/sandbox/claude/`
- **Nature**: Isolated testing environment
- **Content**: Copy of Upstream for safe experimentation
- **Update**: Via `dev sandbox` commands
- **Owner**: Developer/tester

## Configuration Flows

### Installation Flow
```
[Binary] internal/config/system/
    ↓ config install
[Disk] .wm/baseline/ ← First install
    ↓ config install  
[Disk] .claude/ ← User workspace
```

### Update Flow (3-Way Merge)
```
Upstream (A) ──┐
               ├─→ 3-way merge ──→ Local (L)
Baseline (B) ──┘
                ↓
            Update Baseline (B)
```

### Sandbox Flow
```
Upstream (A) ──→ dev sandbox ──→ Sandbox (S)
                                      ↓
                               dev sandbox diff
                                      ↓
                                 Baseline (B)
```

## Core Invariants

### File Ownership Rules
1. **No writes outside designated boundaries**
   - Commands can only write to their authorized spaces
   - Violations are blocked by guard system

2. **Atomic operations**
   - All multi-file operations are atomic
   - Failure rolls back to previous state
   - Automatic backups created before changes

3. **3-Way merge consistency**
   - Content-based diff/merge (not timestamps)
   - Conflicts detected and reported
   - User resolution required for conflicts

4. **Backup protection**
   - Automatic backups before destructive operations
   - Timestamped backup files
   - Override with `--no-backup` flag

## Merge Resolution Strategies

### Plan Phase
- **preserve**: File unchanged, keep current version
- **apply**: File modified, apply new version
- **conflict**: Manual resolution required
- **delete**: File removed in update

### Conflict Resolution
```bash
# Preview conflicts
config update --dry-run

# Resolve manually then retry
vim .claude/conflicted-file
config update
```

## Directory Structure

```
project-root/
├── .claude/                    # Local space (L)
│   ├── agents/
│   ├── commands/
│   └── settings.json
├── .wm/                        # Package manager data
│   ├── baseline/               # Baseline space (B)
│   │   ├── agents/
│   │   ├── commands/
│   │   └── manifest.json
│   └── sandbox/                # Sandbox space (S)
│       └── claude/
│           ├── agents/
│           └── commands/
└── internal/config/system/     # Upstream space (A)
    ├── commands/
    ├── manifest.json
    └── settings.json.template
```

## Command Space Matrix

| Command | Reads From | Writes To | Purpose |
|---------|------------|-----------|----------|
| `config install` | A → B → L | B, L | Initial setup |
| `config status` | A, B, L | - | Show diffs |
| `config update` | A, B | B, L | Apply updates |
| `dev sandbox` | A | S | Create test env |
| `dev sandbox diff` | S, B | B | Sync changes |
| `guard check` | Git working tree | - | Validate changes |

## Performance Characteristics

- **Space scan**: <100ms for typical configurations
- **3-way merge**: <500ms for 100 files
- **Atomic updates**: Sub-second for normal operations
- **Backup creation**: Proportional to changed content size

## Error Handling

### Atomic Failure Recovery
```bash
# If update fails mid-operation
config update  # Automatically rolls back
config status  # Verify clean state
```

### Conflict Resolution
```bash
# Preview all conflicts
config update --dry-run

# Show detailed diff
config status

# Manual resolution workflow
vim .claude/problematic-file
config update  # Retry after fixes
```

## Security Model

### Write Boundaries
- **Upstream (A)**: Read-only, embedded in binary
- **Baseline (B)**: Written only by package manager
- **Local (L)**: User-writable, validated on read
- **Sandbox (S)**: Developer-writable, isolated

### Guard Protection
```bash
# Install protection hook
guard install-hook

# Validate before commit
guard check  # Runs automatically via hook
```

### Validation Chain
1. **JSON Schema** validation on all structured files
2. **File ownership** boundary checks
3. **3-way merge** conflict detection
4. **Backup verification** before destructive operations