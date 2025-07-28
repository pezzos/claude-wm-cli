#!/usr/bin/env python3
import json

# Read current todo.json
with open('/Users/a.pezzotta/repos/claude-wm-cli/docs/3-current-task/todo.json', 'r') as f:
    data = json.load(f)

# Update TASK-025 status to completed
for task in data['todos']:
    if task['id'] == 'TASK-025':
        task['status'] = 'completed'
        print(f"✅ Updated {task['id']}: {task['title']} -> completed")

# Update progress statistics
completed_count = sum(1 for task in data['todos'] if task['status'] == 'completed')
total_count = len(data['todos'])
data['meta']['completedTasks'] = completed_count
data['meta']['totalTasks'] = total_count
data['meta']['progressPercent'] = round((completed_count / total_count) * 100)

print(f"📊 Progress updated: {completed_count}/{total_count} tasks completed ({data['meta']['progressPercent']}%)")

# Write updated todo.json
with open('/Users/a.pezzotta/repos/claude-wm-cli/docs/3-current-task/todo.json', 'w') as f:
    json.dump(data, f, indent=2)

print("✅ todo.json updated successfully")