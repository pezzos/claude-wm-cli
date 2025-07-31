# /1-Select-Stories
Choose next epic from epics.json and initialize workspace.

## Steps
1. Parse epics.json for highest priority unstarted epic (P0 > P1 > P2 > P3)
2. Copy epic info to docs/2-current-epic/current-epic.json
3. Create docs/2-current-epic/CLAUDE.md with epic context
4. Mark epic as "🚧 In Progress"

## Important
Verify dependencies are met before selecting. Update epic status in epics.json with start date.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed