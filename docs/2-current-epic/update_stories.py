#!/usr/bin/env python3
import json

# Read current stories.json
with open('/Users/a.pezzotta/repos/claude-wm-cli/docs/2-current-epic/stories.json', 'r') as f:
    data = json.load(f)

# Update STORY-004 status to completed
for story in data['stories']:
    if story['id'] == 'STORY-004':
        story['status'] = 'completed'
        print(f"✅ Updated {story['id']}: {story['title']} -> completed")

print("✅ stories.json updated successfully")

# Write updated stories.json
with open('/Users/a.pezzotta/repos/claude-wm-cli/docs/2-current-epic/stories.json', 'w') as f:
    json.dump(data, f, indent=2)