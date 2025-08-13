# ADR-0002: 3-Way Merge Strategy for Configuration Updates

## Status
Accepted

## Context
When updating configurations, we need to reconcile three states:
- **Upstream**: New system templates (embedded in binary)
- **Baseline**: Last known system state (immutable snapshot)
- **Local**: User customizations and runtime state

## Decision
Implement 3-way merge algorithm with these resolution strategies:

1. **KEEP** - User changes preserved, system changes ignored
2. **APPLY** - System changes applied, user changes overwritten  
3. **PRESERVE_LOCAL** - User changes preserved, system changes applied to new files
4. **CONFLICT** - Manual resolution required

## Merge Rules
```
Upstream vs Baseline | Baseline vs Local | Action
---------------------|-------------------|--------
No change           | No change         | KEEP
No change           | Modified          | KEEP (preserve user)
Modified            | No change         | APPLY (accept system)
Modified            | Modified          | CONFLICT (manual)
New file            | N/A               | APPLY (add system)
Deleted             | No change         | APPLY (remove)
Deleted             | Modified          | CONFLICT (manual)
```

## Consequences

### Positive
- Prevents accidental loss of user customizations
- Enables automatic system updates
- Provides clear conflict resolution path
- Maintains configuration consistency

### Negative
- Complex merge logic implementation
- Requires careful conflict detection
- Manual intervention needed for conflicts
- Backup and rollback necessary

## Implementation Notes
- Always create backup before merge operations
- Use file content hashing (SHA256) for change detection
- Generate detailed merge reports for user review
- Support --dry-run mode for preview