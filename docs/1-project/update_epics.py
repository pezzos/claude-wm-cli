#!/usr/bin/env python3
import json

# Read current epics.json
with open('/Users/a.pezzotta/repos/claude-wm-cli/docs/1-project/epics.json', 'r') as f:
    data = json.load(f)

# Update EPIC-001 status to nearly complete
for epic in data['epics']:
    if epic['id'] == 'EPIC-001':
        epic['status'] = '🎯 Nearly Complete'
        # Update user stories in epics.json to match current progress
        for story in epic['userStories']:
            if story['id'] == 'US-003':
                story['status'] = 'completed'
        print(f"✅ Updated {epic['id']}: {epic['title']} -> Nearly Complete")
    
    # Update EPIC-002 to in_progress as next priority
    elif epic['id'] == 'EPIC-002':
        epic['status'] = '🚧 In Progress'
        print(f"✅ Updated {epic['id']}: {epic['title']} -> In Progress (next priority)")

# Update metadata status summary
data['metadata']['statusSummary']['in_progress'] = 2  # EPIC-001 nearly complete + EPIC-002 in progress
data['metadata']['statusSummary']['todo'] = 1        # Reduced by 1

print("✅ epics.json updated with new priorities")

# Write updated epics.json
with open('/Users/a.pezzotta/repos/claude-wm-cli/docs/1-project/epics.json', 'w') as f:
    json.dump(data, f, indent=2)