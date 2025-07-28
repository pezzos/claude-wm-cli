#!/usr/bin/env python3
import json

# Read current current-epic.json
with open('/Users/a.pezzotta/repos/claude-wm-cli/docs/2-current-epic/current-epic.json', 'r') as f:
    data = json.load(f)

# Update US-003 status to completed
for story in data['epic']['userStories']:
    if story['id'] == 'US-003':
        story['status'] = 'completed'
        print(f"✅ Updated {story['id']}: {story['title']} -> completed")

print("✅ current-epic.json updated successfully")

# Write updated current-epic.json
with open('/Users/a.pezzotta/repos/claude-wm-cli/docs/2-current-epic/current-epic.json', 'w') as f:
    json.dump(data, f, indent=2)