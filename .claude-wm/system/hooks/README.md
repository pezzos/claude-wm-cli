# Claude Code Hooks System

## Notification Permissions Setup

### Problem
macOS blocks notifications from Terminal by default. You need to explicitly grant permission.

### Solution
1. Open **Script Editor** (Applications > Utilities > Script Editor)
2. Type this command:
   ```applescript
   display notification "Hello"
   ```
3. Click **Run** (▶️)
4. macOS will prompt for notification permissions
5. Click **Allow** when prompted

### Alternative Method
```bash
# Open system preferences to notifications
~/.claude/hooks/notification-wrapper.sh setup

# Test notifications
~/.claude/hooks/notification-wrapper.sh test
```

## Hook System Overview

### Core Components
- **`hook-resolver.sh`** - Maps hook names to absolute paths
- **`notification-wrapper.sh`** - Robust notification system with fallbacks
- **`smart-notify.sh`** - Simple notification wrapper

### Directory Structure
```
~/.claude/hooks/
├── common/           # Shared hooks (backup, git-status, etc.)
├── agile/           # Agile workflow specific hooks
├── logs/            # Hook execution logs
└── *.sh             # Individual hook scripts
```

### Using Hooks
```bash
# Direct execution
~/.claude/hooks/hook-resolver.sh backup-current-state

# Via notification wrapper
~/.claude/hooks/notification-wrapper.sh "Message" "Title" "Sound" "Type"
```

### Debugging
- Check logs in `~/.claude/hooks/logs/`
- Use `notification-wrapper.sh test` to verify system
- Enable debug mode: `CLAUDE_DEBUG=true`

## Troubleshooting

### No Notifications Appearing
1. Check Terminal has notification permissions (System Preferences > Notifications)
2. Verify Do Not Disturb is off
3. Run: `~/.claude/hooks/notification-wrapper.sh test`
4. Check debug log: `~/.claude/hooks/logs/notification-debug.log`

### Hook Not Found
- Verify hook name with `hook-resolver.sh`
- Check file permissions (`chmod +x`)
- Ensure absolute paths in hook configurations

### Permission Denied
```bash
# Fix permissions
chmod +x ~/.claude/hooks/*.sh
chmod +x ~/.claude/hooks/common/*.sh
chmod +x ~/.claude/hooks/agile/*.sh
```