# File Ownership and Write Boundaries

## Write Permission Matrix

| Path | Config Install | Config Update | Config Sync | Dev Sandbox | Sandbox Diff --apply | Migrate Legacy |
|------|----------------|---------------|-------------|-------------|---------------------|----------------|
| `.claude/` | ✅ Create | ✅ Update | ✅ Update | ❌ | ❌ | ❌ |
| `.wm/baseline/` | ✅ Create | ❌ Read-only | ❌ Read-only | ❌ Read-only | ❌ Read-only | ❌ |
| `.wm/backups/` | ❌ | ✅ Create | ❌ | ❌ | ❌ | ❌ |
| `.wm/sandbox/` | ❌ | ❌ | ❌ | ✅ Create/Update | ❌ Read-only | ❌ |
| `.wm/meta.json` | ✅ Create | ❌ Read-only | ❌ Read-only | ❌ Read-only | ❌ Read-only | ✅ Create |
| `internal/config/system/` | ❌ Read-only | ❌ Read-only | ❌ Read-only | ❌ Read-only | ✅ Update | ❌ Read-only |

## Ownership Rules

### Protected Immutable Paths
- **`.wm/baseline/`**: Never modified after creation by `config install`
- **`internal/config/system/`**: Only modified by `dev sandbox diff --apply`

### Runtime Configuration  
- **`.claude/`**: Managed exclusively by config commands (`install`, `update`, `sync`)
- Never modified directly by users or external tools

### Development Isolation
- **`.wm/sandbox/`**: Only accessible to sandbox commands
- Completely isolated from runtime configuration

### User-Editable Areas
- **`.wm/user/`**: User overrides and customizations (future)
- **Project files**: Application code, documentation, assets

## Access Patterns

### Read-Only Access
All commands can read from any location for status, validation, and analysis.

### Write Operations
Must follow atomic patterns:
1. Write to temporary file (`path.tmp`)  
2. Validate content and permissions
3. Atomic rename (`path.tmp` → `path`)
4. Cleanup on failure

### Backup Before Write
Commands that modify critical paths must create backups:
- `.wm/backups/YYYY-MM-DD_HH-MM-SS_<operation>.zip`
- Automatic cleanup of old backups (configurable retention)

## Validation Gates

### Pre-Write Validation
- Path permission checks
- Schema validation for JSON files
- Business rule validation
- Conflict detection

### Post-Write Verification  
- File integrity checks
- Cross-reference validation
- Runtime configuration testing

## Error Recovery

### Rollback Scenarios
- Restore from backup on validation failure
- Cleanup partial writes on error
- Reset to last known good state

### Manual Intervention Required
- Merge conflicts in `config update`
- Permission denied errors
- Corrupted baseline or metadata files

## Integration with External Tools

### Git Integration
- `.wm/` directory should be gitignored (workspace metadata)
- `.claude/` directory committed (runtime configuration)
- Sandbox changes never committed directly

### IDE Integration  
- File watchers should ignore `.wm/sandbox/` changes
- Linting/formatting applied only to user-editable areas
- Auto-completion based on current `.claude/` state

### CI/CD Integration
- Use `--dry-run` flags for validation in pipelines
- Automated `config sync` in deployment scripts  
- Validation hooks for configuration changes