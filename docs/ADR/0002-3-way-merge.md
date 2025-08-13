# ADR-0002: 3-Way Merge Strategy for Configuration Updates

**Date:** 2025-08-13  
**Status:** Accepted  
**Deciders:** Architecture Team  
**Technical Story:** Safe configuration updates with conflict resolution

## Context

Configuration updates must safely merge system changes with user customizations. Traditional approaches either:
- Overwrite user changes (data loss)
- Skip all updates (stale system templates)  
- Use 2-way diff (cannot distinguish user vs. system changes)

With the 4-space model (ADR-0001), we have three states to reconcile:
- **Upstream (A):** New system templates from binary
- **Baseline (B):** Last known system state (immutable reference)
- **Local (L):** User workspace with customizations

The challenge is determining intent: Did the user modify a file, or is it a system change from a previous update?

## Decision

Implement **3-way merge algorithm** using content-based comparison with the following resolution strategies:

### Merge Algorithm

```
Given three versions A (upstream), B (baseline), L (local):

1. Compare A ↔ B (system changes since last install/update)
2. Compare B ↔ L (user changes since last install/update)  
3. Apply merge rules based on change patterns
4. Update both L (user workspace) and B (new baseline)
```

### Resolution Rules

| Upstream vs Baseline | Baseline vs Local | Merge Action | Rationale |
|---------------------|------------------|--------------|-----------|
| **No change** | No change | `preserve` | File unchanged, no action needed |
| **No change** | Modified | `preserve` | User modified, no system changes → keep user version |
| **Modified** | No change | `apply` | System updated, user didn't modify → apply system changes |
| **Modified** | Modified | `conflict` | Both system and user modified → manual resolution required |
| **New file** | N/A | `apply` | System added new file → add to user workspace |
| **Deleted** | No change | `delete` | System removed file, user didn't modify → remove from user workspace |
| **Deleted** | Modified | `conflict` | System removed file, user modified → manual decision required |

### File Comparison Method

**Content-based comparison (not timestamp-based):**
- Use SHA-256 file hashes for quick equality checks
- Line-by-line diff for detailed conflict analysis
- Binary file support with byte-level comparison
- Ignore-pattern support for generated files

## Implementation Details

### Merge Operations

#### preserve
- **Action:** No changes applied to local file
- **Update baseline:** No (baseline reflects last system state)
- **User impact:** None (file remains as user last saved it)

```bash
# Example: User modified .claude/settings.json, no system changes
Status: preserve .claude/settings.json (user modified, system unchanged)
```

#### apply  
- **Action:** Overwrite local file with upstream content
- **Update baseline:** Yes (baseline updated to match upstream)
- **User impact:** Local changes lost (backup created first)

```bash
# Example: System updated .claude/agents/planner.md, user didn't modify
Status: apply .claude/agents/planner.md (system updated, user unchanged)
```

#### conflict
- **Action:** Operation pauses, manual resolution required
- **Update baseline:** No (operation incomplete)
- **User impact:** Must manually edit file and retry operation

```bash
# Example: Both system and user modified .claude/commands/agile/start.md
❌ Merge conflict in .claude/commands/agile/start.md
   - System changed: Added new --template flag
   - User changed: Modified description text
💡 Edit file manually, then retry: claudewm config update
```

#### delete
- **Action:** Remove file from local workspace
- **Update baseline:** Yes (baseline reflects upstream deletion)  
- **User impact:** File removed (backup created, --allow-delete flag required)

```bash
# Example: System removed deprecated .claude/legacy/old-command.md
Status: delete .claude/legacy/old-command.md (removed from system)
Require: --allow-delete flag to proceed
```

### Safety Mechanisms

#### Backup Protection
```bash
# Create timestamped backup before merge
backup_dir=".wm/backups/$(date +%Y-%m-%d_%H-%M-%S)_config_update"
cp -r .claude/ "$backup_dir/claude_before/"
cp -r .wm/baseline/ "$backup_dir/baseline_before/"
```

#### Dry-Run Mode
```bash
# Preview all changes without applying
claudewm config update --dry-run

# Output shows planned operations:
# preserve: .claude/settings.json (user: modified, system: unchanged)  
# apply: .claude/agents/planner.md (user: unchanged, system: modified)
# conflict: .claude/commands/start.md (user: modified, system: modified)
# delete: .claude/legacy/deprecated.md (removed from system)
```

#### Atomic Operations
- Write all changes to temporary files first
- Validate content and permissions
- Apply all changes atomically (all succeed or all fail)
- Update baseline only after successful local updates

### Conflict Resolution Workflow

1. **Detection Phase**
   ```bash
   claudewm config update --dry-run
   # Shows conflicts that require manual resolution
   ```

2. **Manual Resolution**
   ```bash
   # Edit conflicted files to resolve differences
   vim .claude/commands/start.md
   
   # User manually merges system changes with their modifications
   # Can accept system version, keep user version, or create hybrid
   ```

3. **Retry Phase**  
   ```bash
   claudewm config update
   # Re-runs merge, now with resolved conflicts
   ```

4. **Verification**
   ```bash
   claudewm config status
   # Confirms clean state (no pending conflicts)
   ```

## Performance Considerations

### Optimization Strategies
- **Incremental scanning:** Only process files modified since last operation
- **Parallel hashing:** Compute file hashes concurrently where possible
- **Change detection cache:** Store file hashes to avoid repeated computation
- **Selective updates:** `--only <pattern>` flag to limit scope

### Performance Targets
- **Status check:** <200ms for 100 files
- **Full merge:** <500ms for typical configuration (20-50 files)
- **Conflict detection:** <100ms additional overhead
- **Backup creation:** Linear with content size

## Error Handling

### Merge Failures
```bash
# Scenario: Filesystem permission denied during apply
❌ Error: Cannot write to .claude/settings.json - permission denied
💡 Fix permissions and retry:
   chmod u+w .claude/settings.json
   claudewm config update
```

### Corrupted State Recovery
```bash
# Scenario: Baseline corrupted, cannot perform 3-way merge
❌ Error: Baseline corrupted - cannot read .wm/baseline/manifest.json
💡 Recovery options:
   # Option 1: Restore from backup
   cp -r .wm/backups/2025-08-13_10-30-15/baseline_before .wm/baseline
   
   # Option 2: Reset and reinstall
   rm -rf .wm/baseline .claude
   claudewm config install
```

### Conflict Resolution Failure
```bash
# Scenario: User saves malformed resolution
❌ Error: Resolved file fails validation - invalid JSON syntax
💡 Fix syntax and retry:
   vim .claude/settings.json  # Fix JSON syntax error
   claudewm config update     # Retry merge operation
```

## Alternatives Considered

### Alternative 1: 2-Way Merge (Upstream ↔ Local)
**Pros:** Simpler implementation, faster execution  
**Cons:** Cannot distinguish user changes from old system changes, more conflicts  
**Rejected:** Loses important change attribution, more manual intervention required

### Alternative 2: Timestamp-Based Merge
**Pros:** Simpler comparison logic  
**Cons:** Unreliable (timestamps can be modified, filesystem differences)  
**Rejected:** Content-based comparison is more reliable

### Alternative 3: Interactive Merge Tool (like git mergetool)
**Pros:** Rich conflict resolution UI  
**Cons:** Additional dependencies, complexity for CLI tool  
**Deferred:** May implement as optional enhancement

### Alternative 4: Automatic Conflict Resolution
**Pros:** No manual intervention required  
**Cons:** Risk of losing user intent, unexpected changes  
**Rejected:** Configuration changes too critical for automatic resolution

## Dependencies

- **File I/O:** Efficient file reading/writing with atomic operations
- **Hashing:** SHA-256 for content comparison  
- **Backup system:** Timestamped backup creation and restoration
- **Diff engine:** Line-by-line text comparison for conflict display
- **JSON/YAML validation:** Schema validation for structured files

## Related Decisions

- **ADR-0001:** 4-Space Configuration Model (enables 3-way merge)
- **Future ADR:** Selective update patterns for large configurations
- **Future ADR:** Automated conflict resolution for specific file types

## Testing Strategy

### Unit Tests
- Individual merge rule validation
- File comparison accuracy 
- Hash computation correctness
- Backup/restore functionality

### Integration Tests
- End-to-end merge scenarios
- Conflict detection accuracy
- Atomic operation guarantees
- Error recovery workflows

### Acceptance Tests
- User workflow validation
- Performance under load
- Cross-platform compatibility
- Error message clarity

## Notes

The 3-way merge strategy balances automation with safety. Most updates can be applied automatically, while complex conflicts get explicit user review. This approach maintains user autonomy while enabling seamless system updates.

The content-based comparison ensures reliability across different filesystems and environments, avoiding timestamp-related issues common in other merge strategies.