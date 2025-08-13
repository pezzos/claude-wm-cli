# File Ownership and Write Boundaries

Comprehensive documentation of file ownership rules, write boundaries, and access patterns for claude-wm-cli.

## Configuration Space Ownership

### Upstream Space (A) - `internal/config/system/`
**Owner:** System developers  
**Access:** Read-only (embedded in binary)  
**Write authority:** Only `dev sandbox diff --apply` (for development integration)  
**Purpose:** Canonical system templates and schemas

### Baseline Space (B) - `.wm/baseline/`
**Owner:** Package manager  
**Access:** Written only during `config install` and `config update`  
**Write authority:** Configuration management commands only  
**Purpose:** Reference state for 3-way merge operations

### Local Space (L) - `.claude/`
**Owner:** End user  
**Access:** Written by configuration commands, manually editable by users  
**Write authority:** `config install`, `config update`, user manual edits  
**Purpose:** Active workspace with user customizations

### Sandbox Space (S) - `.wm/sandbox/claude/`
**Owner:** Developer/tester  
**Access:** Isolated experimental environment  
**Write authority:** `dev sandbox` commands only  
**Purpose:** Safe testing without affecting production configuration

## Write Permission Matrix

| Path | config install | config update | config migrate-legacy | dev sandbox | dev sandbox diff --apply | guard install-hook |
|------|----------------|---------------|------------------------|-------------|-----------------------|-------------------|
| `.claude/` | ✅ Create | ✅ Update | ✅ Create | ❌ No access | ❌ No access | ❌ No access |
| `.wm/baseline/` | ✅ Create | ✅ Update | ✅ Create | ❌ Read-only | ❌ Read-only | ❌ Read-only |
| `.wm/sandbox/claude/` | ❌ No access | ❌ No access | ❌ No access | ✅ Full access | ❌ Read-only | ❌ No access |
| `.wm/backups/` | ✅ Create | ✅ Create | ✅ Create | ❌ No access | ✅ Create | ❌ No access |
| `internal/config/system/` | ❌ Read-only | ❌ Read-only | ❌ Read-only | ❌ Read-only | ✅ Update | ❌ Read-only |
| `.git/hooks/` | ❌ No access | ❌ No access | ❌ No access | ❌ No access | ❌ No access | ✅ Create |

**Legend:**
- ✅ **Full access:** Can create, read, update, delete
- ✅ **Create:** Can create new files/directories  
- ✅ **Update:** Can modify existing files
- ❌ **Read-only:** Can read but not modify
- ❌ **No access:** Cannot read or write

## File System Invariants

### Atomic Write Operations
All write operations must be atomic to prevent corruption:

```bash
# Standard atomic write pattern
temp_file=$(mktemp "${target_path}.tmp.XXXXXX")
trap "rm -f '${temp_file}'" EXIT

# Write content to temporary file
echo "$content" > "$temp_file"

# Validate content
validate_file_content "$temp_file"

# Atomic move (succeeds or fails completely)
mv "$temp_file" "$target_path"
```

### Backup Creation
Critical operations must create timestamped backups:

```bash
# Backup pattern for destructive operations
backup_dir=".wm/backups/$(date +%Y-%m-%d_%H-%M-%S)_${operation_name}"
mkdir -p "$backup_dir"

# Copy files before modification
cp -r .claude/ "$backup_dir/claude_before"
cp -r .wm/baseline/ "$backup_dir/baseline_before"
```

### Content Validation
All structured files must pass validation before write:

- **JSON files:** Schema validation against predefined schemas
- **YAML files:** Syntax and structure validation
- **Template files:** Template engine parsing validation
- **Configuration files:** Business rule validation

## Access Control Enforcement

### Guard System Rules
The guard system enforces ownership boundaries:

1. **Upstream Protection:** No command can write to `internal/config/system/` except `dev sandbox diff --apply`
2. **Baseline Integrity:** Only package manager commands can modify `.wm/baseline/`
3. **Sandbox Isolation:** Sandbox files never directly affect production spaces
4. **Hook Integration:** Git hooks prevent commits violating boundaries

### Validation Sequence
```bash
# Pre-write validation
guard_check_write_permission "$target_path" "$command_name"
schema_validate_file_content "$temp_file"
business_rule_validate "$temp_file" "$target_path"

# Write operation
atomic_write "$temp_file" "$target_path"

# Post-write verification
verify_file_integrity "$target_path"
cross_reference_validate "$target_path"
```

## Space Interaction Rules

### Permitted Data Flows
```
┌─────────────┐    install    ┌─────────────┐    install    ┌─────────────┐
│ Upstream(A) │ ─────────────▶ │ Baseline(B) │ ─────────────▶ │  Local(L)   │
└─────────────┘               └─────────────┘               └─────────────┘
       │                             │                             ▲
       │ sandbox                     │ 3-way merge                │ update
       ▼                             ▼                             │
┌─────────────┐    diff --apply ┌─────────────┐ ◀─────────────────┘
│ Sandbox(S)  │ ─────────────▶ │   System    │
└─────────────┘               └─────────────┘
```

### Forbidden Data Flows
- ❌ Direct Local → Upstream writes
- ❌ Sandbox → Local without baseline integration
- ❌ External tools → Baseline writes
- ❌ User edits → Upstream space

## Error Scenarios and Recovery

### Boundary Violations
**Scenario:** Command attempts to write outside authorized space

```bash
# Detection
❌ Error: Command 'config update' attempted to write to 'internal/config/system/commands/test.md'
   This violates file ownership boundaries.

# Recovery
💡 Use the appropriate command:
   - For system changes: dev sandbox diff --apply
   - For local changes: Manual editing in .claude/
```

### Permission Denied
**Scenario:** Insufficient filesystem permissions

```bash
# Detection  
❌ Error: Permission denied writing to '.claude/settings.json'
   Current user lacks write access to target directory.

# Recovery
💡 Fix permissions:
   chmod u+w .claude/
   chown $(whoami) .claude/
```

### Corrupted Baseline
**Scenario:** Baseline space contains invalid data

```bash
# Detection
❌ Error: Baseline corrupted - invalid JSON in '.wm/baseline/manifest.json'
   Cannot proceed with 3-way merge.

# Recovery  
💡 Restore from backup or reinstall:
   # Option 1: Restore from backup
   cp -r .wm/backups/2025-08-13_10-30-15/baseline_before .wm/baseline
   
   # Option 2: Clean reinstall
   rm -rf .wm/baseline .claude
   claudewm config install
```

### Merge Conflicts
**Scenario:** 3-way merge cannot resolve differences automatically

```bash
# Detection
❌ Merge conflict in '.claude/agents/planner.md'
   Upstream, baseline, and local versions all differ.

# Resolution process
💡 Manual conflict resolution required:
   1. Edit .claude/agents/planner.md to resolve conflicts
   2. Run 'claudewm config update' to retry merge
   3. Verify with 'claudewm config status'
```

## Integration with External Systems

### Git Integration
```gitignore
# .gitignore recommendations
.wm/sandbox/     # Never commit sandbox experiments
.wm/backups/     # Backup artifacts are local only  
.wm/cache/       # Cache files regenerated on demand
```

**Commit guidelines:**
- ✅ Commit `.claude/` changes (user configuration)
- ✅ Commit application code and documentation
- ❌ Never commit `.wm/baseline/` (managed by package manager)
- ❌ Never commit sandbox or backup artifacts

### IDE Integration
**File watcher exclusions:**
```json
{
  "files.watcherExclude": {
    "**/.wm/sandbox/**": true,
    "**/.wm/backups/**": true,
    "**/.wm/cache/**": true
  }
}
```

**Linting scope:**
- ✅ Apply to user project files
- ✅ Apply to `.claude/` configuration
- ❌ Skip `.wm/baseline/` (system managed)
- ❌ Skip sandbox experiments

### CI/CD Integration
**Pipeline validation:**
```yaml
# Safe validation in CI
- name: Validate Configuration
  run: |
    claudewm config install --dry-run
    claudewm config status
    claudewm guard check --strict

# Avoid destructive operations
- name: Update Configuration  
  run: |
    claudewm config update --dry-run  # Preview only
    # Never run 'config update' without --dry-run in CI
```

**Deployment automation:**
```yaml
# Production deployment
- name: Deploy Configuration
  run: |
    claudewm config install         # Safe: creates from upstream
    claudewm config status         # Verify clean state
    claudewm guard install-hook    # Install protection
```

## Performance Considerations

### Write Operation Costs
- **Atomic writes:** ~10ms overhead for safety
- **Backup creation:** Linear with content size
- **Validation:** ~50ms for typical configurations
- **Cross-space operations:** Higher latency due to multiple I/O

### Optimization Strategies
- **Selective updates:** Use `--only` patterns to limit scope
- **Skip backups:** Use `--no-backup` for non-critical changes
- **Parallel validation:** Schema checks run concurrently where possible
- **Incremental operations:** Only process changed files

### Monitoring and Alerting
- **Validation failures:** Log schema violations for analysis
- **Boundary violations:** Alert on unauthorized write attempts  
- **Performance regression:** Track operation durations
- **Backup growth:** Monitor backup directory size