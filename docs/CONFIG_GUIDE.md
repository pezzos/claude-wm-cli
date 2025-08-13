# Configuration Management Guide

This guide provides comprehensive documentation for the configuration management system in Claude WM CLI.

## Overview

The configuration system uses a **package manager approach** with three-way merge capabilities, baseline tracking, and atomic operations. It manages configuration across multiple spaces with strong consistency guarantees.

## Configuration Spaces

### Upstream (Embedded)
- **Source**: `internal/config/system/` (embedded in binary)
- **Purpose**: System defaults and templates
- **Access**: Read-only
- **Updates**: Only through binary updates

### Baseline (`.wm/baseline/`)
- **Purpose**: Installation snapshot for diff calculations
- **Immutability**: Never modified after `config install`
- **Use**: Reference point for 3-way merges

### Local (`.claude/`)
- **Purpose**: Runtime configuration for Claude Code
- **Management**: Auto-generated from templates and user overrides
- **Updates**: Via `config update` and `config sync`

### Sandbox (`.wm/sandbox/claude/`)
- **Purpose**: Isolated development environment
- **Use**: Safe experimentation without affecting production

## Commands Reference

### `config install`

Installs initial system configuration from embedded templates.

**Syntax:**
```bash
claude-wm-cli config install
```

**Behavior:**
- Copies embedded system templates to `.claude/` and `.wm/baseline/`
- Creates `.wm/meta.json` with installation metadata
- Fails if configuration already exists

**Side Effects:**
| Directory | Action |
|-----------|---------|
| `.claude/` | ✅ Create from templates |
| `.wm/baseline/` | ✅ Create from templates |
| `.wm/meta.json` | ✅ Create metadata |

**Example:**
```bash
$ claude-wm-cli config install
📦 Installing system configuration...
   → Copying to /project/.claude
   → Copying to /project/.wm/baseline
   → Creating /project/.wm/meta.json
✅ System configuration installed successfully!
```

---

### `config status`

Shows differences between upstream, baseline, and local configurations.

**Syntax:**
```bash
claude-wm-cli config status
```

**Output Format:**
```
📊 Configuration Status
======================

🔄 Upstream vs Baseline (changes since installation):
   + new-file.json
   M modified-file.json
   - deleted-file.json

📝 Baseline vs Local (your modifications):
   M settings.json
   + custom-config.json
```

**Change Symbols:**
- `+` : New file
- `M` : Modified file
- `-` : Deleted file

**Side Effects:**
- Read-only operation
- No file system changes

---

### `config update`

Performs 3-way merge update with conflict detection.

**Syntax:**
```bash
claude-wm-cli config update [flags]
```

**Flags:**
- `--dry-run` : Show planned changes without applying
- `--no-backup` : Skip backup creation (not recommended)

**3-Way Merge Logic:**

| Upstream | Baseline | Local | Action |
|----------|----------|--------|---------|
| Changed | Unchanged | Unchanged | Apply upstream change |
| Unchanged | Unchanged | Changed | Preserve local change |
| Changed | Unchanged | Changed | **CONFLICT** - Manual resolution required |
| Unchanged | Unchanged | Unchanged | No action |

**Examples:**

**Dry Run:**
```bash
$ claude-wm-cli config update --dry-run
📋 Update Plan (dry-run)
========================
{
  "merge": [
    {
      "path": "settings.json",
      "action": "apply",
      "reason": "upstream_change_no_local_conflict"
    }
  ]
}
💡 Run without --dry-run to apply 1 changes
```

**Apply Changes:**
```bash
$ claude-wm-cli config update
🔄 Calculating 3-way merge plan...
📦 Creating backup...
   ✓ Backup created: .wm/backups/2024-01-15_14-30-25.zip
🔄 Applying 3 changes...
🎉 Update completed successfully!
```

**Side Effects:**
| Directory | Action |
|-----------|---------|
| `.claude/` | ✅ Update with merged changes |
| `.wm/backups/` | ✅ Create backup (unless `--no-backup`) |

---

### `config sync`

Regenerates runtime configuration from templates and user overrides.

**Syntax:**
```bash
claude-wm-cli config sync
```

**Process:**
1. Merge system templates with user overrides
2. Generate runtime configuration in `.claude/`
3. Apply path corrections and validations

**Use Cases:**
- After manual template modifications
- Resolving configuration inconsistencies
- Regenerating after corruption

**Side Effects:**
| Directory | Action |
|-----------|---------|
| `.claude/` | ✅ Regenerate all configuration |

---

### `config upgrade`

Updates system templates while preserving user customizations.

**Syntax:**
```bash
claude-wm-cli config upgrade
```

**Process:**
1. Reinstall system templates
2. Regenerate runtime configuration
3. Preserve user customizations in overlay

**Side Effects:**
| Directory | Action |
|-----------|---------|
| `.claude/` | ✅ Update system parts, preserve user parts |

---

### `config show`

Displays effective runtime configuration.

**Syntax:**
```bash
claude-wm-cli config show [file]
```

**Arguments:**
- `file` (optional): Specific configuration file to display

**Examples:**

**Show Overview:**
```bash
$ claude-wm-cli config show
📋 Configuration Overview:

   System: ✅ 45 items
   User: ✅ 12 items  
   Runtime: ✅ 57 items
```

**Show Specific File:**
```bash
$ claude-wm-cli config show settings.json
📄 settings.json (runtime):
{
  "version": "1.0.0",
  "hooks": {
    "PreToolUse": [],
    "PostToolUse": ["post-write-json-validator-simple.sh"]
  }
}
```

---

### `config migrate-legacy`

Migrates from legacy `.claude-wm/` structure to new `.wm/` structure.

**Syntax:**
```bash
claude-wm-cli config migrate-legacy [flags]
```

**Flags:**
- `--dry-run` : Show migration plan without applying
- `--archive` : Rename `.claude-wm` to `.claude-wm.bak` after migration

**Migration Strategy:**

| Legacy Path | Category | New Path | Action |
|-------------|----------|----------|--------|
| `system/` | System | `baseline/` | Migrate |
| `user/` | User | `user/` | Migrate |
| `runtime/` | Runtime | — | Ignore (regenerated) |
| `meta.json` | Meta | `meta.json` | Convert |
| `cache/` | Cache | — | Ignore |
| `backup/` | Backup | — | Ignore |

**Examples:**

**Analysis:**
```bash
$ claude-wm-cli config migrate-legacy
🔍 Legacy Configuration Migration Analysis
═══════════════════════════════════════════

📁 Legacy directory detected: .claude-wm

📊 Migration Analysis Report
═══════════════════════════

📋 Summary:
   • Files analyzed: 25
   • Files to migrate: 15
   • Files to ignore: 10
   • Estimated size: 2.3 MB

💡 Use --dry-run to preview without applying
```

**Dry Run:**
```bash
$ claude-wm-cli config migrate-legacy --dry-run
🔍 Dry Run: What Would Be Applied
═══════════════════════════════════

   📄 CREATE system/settings.json.template → baseline/settings.json.template
   📄 CREATE user/custom.json → user/custom.json
   🔄 CONVERT meta.json → meta.json
   ⏭️ IGNORE runtime/settings.json (runtime files regenerated)

📊 Summary: 12 actions would be applied
💡 Remove --dry-run to actually apply these changes.
```

**Apply Migration:**
```bash
$ claude-wm-cli config migrate-legacy --archive
🚀 Applying Changes: .claude-wm → .wm
═══════════════════════════════════════

📄 Copying system/settings.json.template
📄 Copying user/custom.json
🔄 Converting meta.json

📦 Archiving legacy directory...
   ✓ Archived to: .claude-wm.bak

🎉 Migration Completed Successfully!

📋 Migration Summary:
   • 12 files migrated
   • 8 files ignored
   • Legacy structure: archived to .claude-wm.bak
   • New structure: .wm
```

**Side Effects:**
| Directory | Action |
|-----------|---------|
| `.wm/` | ✅ Create with migrated content |
| `.wm/meta.json` | ✅ Create migration metadata |
| `.claude-wm` | ⚠️ Rename to `.claude-wm.bak` (if `--archive`) |

## Advanced Usage

### Pattern Filtering

Several commands support pattern-based filtering:

**Glob Patterns:**
```bash
# Apply only agent-related changes
claude-wm-cli dev sandbox diff --apply --only "agents/**"

# Multiple patterns
claude-wm-cli dev sandbox diff --apply --only "agents/**" --only "commands/**"
```

**Pattern Syntax:**
- `*` : Matches any characters within a single directory level
- `**` : Matches any characters across multiple directory levels
- `?` : Matches any single character
- `[abc]` : Matches any character in brackets

### Backup Management

**Automatic Backups:**
- Created before destructive operations
- Stored in `.wm/backups/` as timestamped ZIP files
- Include full `.claude/` directory state

**Manual Backup Management:**
```bash
# List backups
ls -la .wm/backups/

# Restore from backup (manual process)
cd .wm/backups/
unzip 2024-01-15_14-30-25.zip -d /tmp/restore
# Manually copy desired files back
```

### Conflict Resolution

When `config update` encounters conflicts:

1. **Identify Conflicts:**
   ```bash
   $ claude-wm-cli config update
   ❌ Update failed: conflicts detected
   
   Conflicts:
   - settings.json: both upstream and local modified
   ```

2. **Manual Resolution:**
   - Edit conflicted files in `.claude/`
   - Resolve conflicts by choosing desired state
   - Run `config update` again

3. **Force Strategies:**
   ```bash
   # Keep local changes (lose upstream updates)
   git checkout HEAD -- .claude/settings.json
   claude-wm-cli config sync
   
   # Accept upstream changes (lose local changes)
   claude-wm-cli config update --force-upstream
   ```

### Integration with Development Workflow

**Pre-commit Integration:**
```bash
# Install validation hook
claude-wm-cli guard install-hook

# Manual validation
claude-wm-cli guard check
```

**CI/CD Integration:**
```bash
# Validate configuration in CI
claude-wm-cli config status
claude-wm-cli config update --dry-run

# Automated updates in CD
claude-wm-cli config update --no-backup
```

### Troubleshooting

**Common Issues:**

1. **"Configuration already installed" Error:**
   ```bash
   # Remove existing installation
   rm -rf .wm/ .claude/
   claude-wm-cli config install
   ```

2. **"Baseline not found" Error:**
   ```bash
   # Reinstall baseline
   claude-wm-cli config install
   ```

3. **Merge Conflicts:**
   ```bash
   # Check status first
   claude-wm-cli config status
   
   # Resolve manually then retry
   claude-wm-cli config update
   ```

4. **Corrupted Configuration:**
   ```bash
   # Regenerate from templates
   claude-wm-cli config sync
   
   # Or restore from backup
   # (manual restoration process)
   ```

### Best Practices

**For Users (Projects):**
1. Always run `config status` before `config update`
2. Use `--dry-run` to preview changes
3. Never modify `.wm/baseline/` manually
4. Keep backups of important customizations
5. Use sandbox for experimental changes

**For Self-Mode (Repository Development):**
1. Use `dev sandbox` for experimentation
2. Apply changes selectively with `--only` patterns
3. Test changes thoroughly before upstreaming
4. Use `config migrate-legacy` when upgrading structure
5. Validate with `guard check` before commits

**For CI/CD:**
1. Always use `--dry-run` in validation phases
2. Use `--no-backup` in automated environments
3. Implement proper error handling
4. Log all configuration changes
5. Validate before and after updates