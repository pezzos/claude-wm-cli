#!/bin/bash
# Backup project state
echo "💾 Backing up project state..."
if [[ -d ".claude-wm" ]]; then
    echo "📁 Found .claude-wm directory"
fi
