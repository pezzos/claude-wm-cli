# Configuration Guide

Complete reference for claude-wm-cli configuration management commands and flags.

## Command Reference

### config install
Install initial system configuration to .claude/ and .wm/baseline/

```bash
claudewm config install [flags]

# Examples:
claudewm config install                    # Standard install
claudewm config install --force            # Overwrite existing
claudewm config install --dry-run          # Preview install
```

**Flags:**
- `--force` - Overwrite existing configuration
- `--dry-run` - Preview changes without applying
- `--no-backup` - Skip backup creation

### config status
Show configuration differences between upstream, baseline, and local

```bash
claudewm config status [flags]

# Examples:
claudewm config status                     # Show all differences
claudewm config status --verbose           # Detailed diff output
claudewm config status --only commands/    # Filter by path pattern
```

**Flags:**
- `--only <pattern>` - Show only matching files/directories
- `--verbose` - Show detailed diff content

### config update
Update configuration with 3-way merge

```bash
claudewm config update [flags]

# Examples:
claudewm config update --dry-run           # Preview all changes
claudewm config update                      # Apply all changes
claudewm config update --only agents/      # Update only agents/
claudewm config update --no-backup         # Skip backup creation
claudewm config update --allow-delete      # Allow file deletion
```

**Flags:**
- `--dry-run` - Preview changes without applying
- `--no-backup` - Skip automatic backup creation
- `--only <pattern>` - Update only matching files/patterns
- `--allow-delete` - Allow deletion of files during update

### config migrate-legacy
Migrate from legacy .claude-wm to new .wm structure

```bash
claudewm config migrate-legacy [flags]

# Examples:
claudewm config migrate-legacy             # Standard migration
claudewm config migrate-legacy --dry-run   # Preview migration
claudewm config migrate-legacy --backup-dir ./old-config
```

**Flags:**
- `--dry-run` - Preview migration without applying
- `--backup-dir <path>` - Custom backup location for old config

## Development Commands

### dev sandbox
Create testing sandbox from upstream system files

```bash
claudewm dev sandbox [flags]

# Examples:
claudewm dev sandbox                       # Create full sandbox
claudewm dev sandbox --only commands/     # Sandbox specific path
claudewm dev sandbox --force              # Overwrite existing
```

**Flags:**
- `--force` - Overwrite existing sandbox
- `--only <pattern>` - Sandbox only matching files

### dev sandbox diff
Compare sandbox with baseline and optionally apply changes

```bash
claudewm dev sandbox diff [flags]

# Examples:
claudewm dev sandbox diff                  # Show differences
claudewm dev sandbox diff --apply          # Apply all changes
claudewm dev sandbox diff --apply --only cmd/
claudewm dev sandbox diff --allow-delete   # Allow deletions
```

**Flags:**
- `--apply` - Apply changes to baseline after showing diff
- `--only <pattern>` - Apply only matching files/patterns
- `--allow-delete` - Allow file deletion during sync

## Guard Commands

### guard check
Validate current changes against writing restrictions

```bash
claudewm guard check [flags]

# Examples:
claudewm guard check                       # Check all changes
claudewm guard check --fix                # Auto-fix violations
claudewm guard check --strict              # Strict mode
```

**Flags:**
- `--fix` - Automatically fix detected violations
- `--strict` - Enable strict validation mode

### guard install-hook
Install git pre-commit validation hook

```bash
claudewm guard install-hook [flags]

# Examples:
claudewm guard install-hook                # Install hook
claudewm guard install-hook --force        # Overwrite existing
```

**Flags:**
- `--force` - Overwrite existing pre-commit hook

## File Ownership Matrix

| Command | Reads From | Writes To | Backup Created |
|---------|------------|-----------|----------------|
| `config install` | A (internal/config/system) | B (.wm/baseline), L (.claude) | Yes |
| `config status` | A, B, L | - | No |
| `config update` | A, B | B, L | Yes (unless --no-backup) |
| `config migrate-legacy` | .claude-wm/ | B, L | Yes |
| `dev sandbox` | A | S (.wm/sandbox/claude) | No |
| `dev sandbox diff` | S, B | B (if --apply) | Yes (if --apply) |
| `guard check` | Git working tree | - | No |
| `guard install-hook` | - | .git/hooks/ | Yes |

**Legend:**
- **A**: Upstream (internal/config/system)
- **B**: Baseline (.wm/baseline)
- **L**: Local (.claude)
- **S**: Sandbox (.wm/sandbox/claude)

## Flag Usage Patterns

### Safe Preview Workflow
```bash
# Always preview first
claudewm config update --dry-run

# Review output, then apply
claudewm config update
```

### Selective Updates
```bash
# Update only commands
claudewm config update --only commands/

# Update multiple patterns
claudewm config update --only "commands/,agents/"
```

### Development Workflow
```bash
# Create sandbox for experimentation
claudewm dev sandbox

# Make changes in .wm/sandbox/claude/...
# Test changes

# Preview integration
claudewm dev sandbox diff

# Apply to baseline
claudewm dev sandbox diff --apply
```

### Backup Management
```bash
# Standard operation (creates backup)
claudewm config update

# Skip backup for minor changes
claudewm config update --no-backup

# Backups stored in:
# .wm/backups/YYYY-MM-DD_HH-MM-SS/
```

## Configuration Files

### Global Config
```yaml
# ~/.claude-wm-cli.yaml
verbose: false
debug: false

defaults:
  timeout: 30
  retries: 2
  backup: true

spaces:
  upstream: internal/config/system
  baseline: .wm/baseline
  local: .claude
  sandbox: .wm/sandbox/claude
```

### Project Config
```yaml
# ./.claude-wm-cli.yaml
project:
  name: my-project
  initialized: true
  
overrides:
  backup: false  # Skip backups for this project
  timeout: 60    # Longer timeout for heavy operations
```

## Environment Variables

- `CLAUDE_WM_VERBOSE=true` - Enable verbose output
- `CLAUDE_WM_DEBUG=true` - Enable debug mode
- `CLAUDE_WM_NO_BACKUP=true` - Skip backup creation globally
- `CLAUDE_WM_TIMEOUT=60` - Default timeout in seconds
- `CLAUDE_WM_CONFIG=/path/to/config` - Custom config file location

## Error Handling

### Common Error Scenarios

**Conflict during update:**
```bash
$ claudewm config update
❌ Merge conflict in .claude/agents/planner.md
💡 Resolve manually and retry:
   vim .claude/agents/planner.md
   claudewm config update
```

**Invalid file ownership:**
```bash
$ claudewm guard check
❌ Attempt to write outside permitted boundaries:
   - internal/config/system/commands/test.md (Upstream space)
💡 Use dev sandbox for system file modifications
```

**Missing baseline:**
```bash
$ claudewm config status
❌ Baseline not found at .wm/baseline/
💡 Run: claudewm config install
```

### Recovery Procedures

**Restore from backup:**
```bash
# List available backups
ls .wm/backups/

# Restore from specific backup
cp -r .wm/backups/2025-08-13_10-30-15/.claude .claude
```

**Clean corrupted state:**
```bash
# Remove and reinstall
rm -rf .claude .wm/baseline
claudewm config install
```

**Reset sandbox:**
```bash
# Clean sandbox
rm -rf .wm/sandbox/claude
claudewm dev sandbox
```