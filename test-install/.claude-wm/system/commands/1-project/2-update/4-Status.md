# /4-Status
Analyze epics.json and provide project status with completion metrics and next actions.

## Steps
1. Parse epics.json for epic progress (completed/total, current epic)
2. Check docs/2-current-epic/ for active epic details
3. Review docs/archive/ for historical performance
4. Display formatted status with recommendations

## Important
Provide specific next command suggestions based on current state. Use visual progress indicators.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed