# Commands Reference

Comprehensive command reference for claude-wm-cli organized by category and usage patterns.

## Configuration Commands

### config install
**Purpose:** Initial system configuration installation  
**Pattern:** `claudewm config install [flags]`  
**Spaces:** A → B → L (Upstream to Baseline to Local)

```bash
# Standard installation
claudewm config install

# Force overwrite existing configuration
claudewm config install --force

# Preview installation without applying
claudewm config install --dry-run

# Install without creating backups
claudewm config install --no-backup
```

**Usage scenarios:**
- First-time project setup
- Recovery from corrupted configuration
- Clean reinstallation after major changes
- Testing installation procedures

### config status
**Purpose:** Show configuration differences across all spaces  
**Pattern:** `claudewm config status [flags]`  
**Spaces:** Reads A, B, L (no writes)

```bash
# Show all differences
claudewm config status

# Detailed diff with line-by-line changes
claudewm config status --verbose

# Filter by specific paths
claudewm config status --only commands/
claudewm config status --only "agents/,settings.json"
```

**Output interpretation:**
- `preserve`: No changes needed (files identical)
- `apply`: Changes available from upstream
- `conflict`: Manual resolution required
- `delete`: File removed in upstream version

### config update
**Purpose:** Apply upstream changes using 3-way merge  
**Pattern:** `claudewm config update [flags]`  
**Spaces:** Reads A, B; Writes B, L

```bash
# Safe workflow: preview then apply
claudewm config update --dry-run
claudewm config update

# Selective updates
claudewm config update --only commands/
claudewm config update --only "agents/planner.md,commands/agile/"

# Skip backup creation (use cautiously)
claudewm config update --no-backup

# Allow file deletions during update
claudewm config update --allow-delete
```

**Conflict resolution workflow:**
1. Run `config update --dry-run` to identify conflicts
2. Edit conflicted files manually in `.claude/`
3. Re-run `config update` to complete merge

### config migrate-legacy
**Purpose:** Migrate from .claude-wm to new .wm structure  
**Pattern:** `claudewm config migrate-legacy [flags]`  
**Spaces:** .claude-wm/ → B, L

```bash
# Standard migration with automatic backup
claudewm config migrate-legacy

# Preview migration steps
claudewm config migrate-legacy --dry-run

# Custom backup location for old configuration
claudewm config migrate-legacy --backup-dir ./legacy-backup
```

**Migration mapping:**
- `.claude-wm/baseline/` → `.wm/baseline/`
- `.claude-wm/local/` → `.claude/`
- Configuration format updates applied automatically

## Development Commands

### dev sandbox
**Purpose:** Create isolated testing environment  
**Pattern:** `claudewm dev sandbox [flags]`  
**Spaces:** A → S (Upstream to Sandbox)

```bash
# Create complete sandbox
claudewm dev sandbox

# Create partial sandbox for specific components
claudewm dev sandbox --only commands/
claudewm dev sandbox --only "agents/,templates/"

# Overwrite existing sandbox
claudewm dev sandbox --force
```

**Sandbox workflow:**
1. Create sandbox: `dev sandbox`
2. Make experimental changes in `.wm/sandbox/claude/`
3. Test changes without affecting production
4. Apply successful changes: `dev sandbox diff --apply`

### dev sandbox diff
**Purpose:** Compare sandbox with baseline and sync changes  
**Pattern:** `claudewm dev sandbox diff [flags]`  
**Spaces:** Reads S, B; Writes B (if --apply)

```bash
# Show sandbox vs baseline differences
claudewm dev sandbox diff

# Apply all changes to baseline
claudewm dev sandbox diff --apply

# Apply specific files only
claudewm dev sandbox diff --apply --only cmd/
claudewm dev sandbox diff --apply --only "agents/planner.md"

# Allow file deletion during sync
claudewm dev sandbox diff --apply --allow-delete
```

**Integration patterns:**
- Use for testing upstream changes before distribution
- Validate new command structures
- Experiment with agent configurations
- Test schema changes safely

## Guard Commands

### guard check
**Purpose:** Validate changes against file ownership boundaries  
**Pattern:** `claudewm guard check [flags]`  
**Spaces:** Reads Git working tree (no writes)

```bash
# Check all staged and unstaged changes
claudewm guard check

# Automatically fix detected violations where possible
claudewm guard check --fix

# Enable strict validation mode
claudewm guard check --strict
```

**Validation rules:**
- No writes to Upstream space (internal/config/system)
- Respect configuration boundaries
- Detect unauthorized system file modifications
- Validate JSON schema compliance

### guard install-hook
**Purpose:** Install git pre-commit validation hook  
**Pattern:** `claudewm guard install-hook [flags]`  
**Spaces:** Writes .git/hooks/

```bash
# Install pre-commit hook
claudewm guard install-hook

# Overwrite existing hook
claudewm guard install-hook --force
```

**Hook behavior:**
- Runs `guard check` on every commit
- Blocks commits with boundary violations
- Bypass with `git commit --no-verify` if needed
- Updates automatically with CLI version

## Project Commands

### init
**Purpose:** Initialize new project with proper structure  
**Pattern:** `claudewm init <project-name> [flags]`

```bash
# Create new project
claudewm init my-project

# Initialize with custom template
claudewm init my-project --template advanced

# Skip interactive prompts
claudewm init my-project --no-interactive
```

**Created structure:**
```
my-project/
├── .claude/
├── .wm/baseline/
├── .git/
├── README.md
└── project-config.yaml
```

### status
**Purpose:** Show current project state and progress  
**Pattern:** `claudewm status [flags]`

```bash
# Show project overview
claudewm status

# Include detailed configuration status
claudewm status --config

# Show recent activity
claudewm status --history
```

**Status categories:**
- Configuration state (clean/dirty/conflicts)
- Recent command history
- Active sandbox sessions
- Pending migrations

### execute
**Purpose:** Execute Claude commands with timeout protection  
**Pattern:** `claudewm execute \"<command>\" [flags]`

```bash
# Execute with default timeout (30s)
claudewm execute "analyze codebase structure"

# Custom timeout and retries
claudewm execute "complex analysis task" --timeout 120 --retries 3

# Execute with specific context
claudewm execute "implement feature" --context project-notes.md
```

**Timeout handling:**
- Automatic retry on network failures
- Exponential backoff: 1s, 2s, 4s delays
- Structured error reporting
- Resource cleanup guaranteed

## Global Flags

### Common Flags
- `--verbose` - Enable detailed output
- `--debug` - Enable debug logging
- `--config <path>` - Custom configuration file
- `--no-color` - Disable colored output

### Operation Flags
- `--dry-run` - Preview without applying changes
- `--force` - Overwrite existing files/configuration
- `--no-backup` - Skip backup creation
- `--apply` - Apply changes after preview

### Filtering Flags
- `--only <pattern>` - Process only matching files
- `--exclude <pattern>` - Skip matching files
- `--strict` - Enable strict validation mode

### Safety Flags
- `--allow-delete` - Permit file deletion
- `--no-verify` - Skip validation checks
- `--timeout <seconds>` - Operation timeout
- `--retries <count>` - Retry attempts

## Command Combinations

### Safe Update Workflow
```bash
claudewm config status           # Check current state
claudewm config update --dry-run # Preview changes
claudewm config update           # Apply updates
claudewm guard check            # Validate result
```

### Development Cycle
```bash
claudewm dev sandbox                      # Create test environment
# Edit files in .wm/sandbox/claude/
claudewm dev sandbox diff                 # Review changes
claudewm dev sandbox diff --apply         # Integrate changes
claudewm config update                    # Propagate to local
```

### Migration Process
```bash
claudewm config migrate-legacy --dry-run  # Preview migration
claudewm config migrate-legacy            # Execute migration
claudewm config status                    # Verify final state
claudewm guard install-hook               # Install protection
```

### Recovery Operations
```bash
# Restore from backup
ls .wm/backups/                          # List available backups
cp -r .wm/backups/2025-08-13_10-30-15/.claude .claude

# Clean reinstall
rm -rf .claude .wm/baseline
claudewm config install
```

## Exit Codes

- `0`: Success
- `1`: General error
- `2`: Validation error
- `3`: Conflict requiring manual resolution
- `4`: Permission denied
- `5`: File not found or inaccessible

## Performance Optimization

### Command Performance Targets
- `config status`: <200ms for 100 files
- `config update`: <500ms for typical changes
- `dev sandbox`: <1s for complete upstream
- `guard check`: <300ms for working tree scan

### Optimization Strategies
- Use `--only` flags for selective operations
- Skip backups with `--no-backup` for minor changes
- Parallel processing automatically enabled where safe
- Incremental operations preferred over full scans