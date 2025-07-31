#!/bin/bash
# backup-state.sh - Create timestamped backup of current project state

set -e

BACKUP_DIR=".claude-code/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="backup_${TIMESTAMP}"

echo "Creating backup: $BACKUP_NAME"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Create backup using git stash or tar
if [ -d ".git" ]; then
    # Git repository - use git stash
    git stash push -m "Claude Code backup: $BACKUP_NAME" --include-untracked
    echo "Git stash created: $BACKUP_NAME"
    
    # Also save stash info for recovery
    git stash list | head -1 > "$BACKUP_DIR/${BACKUP_NAME}.info"
else
    # Non-git project - use tar
    tar -czf "$BACKUP_DIR/${BACKUP_NAME}.tar.gz" \
        --exclude=".claude-code" \
        --exclude="node_modules" \
        --exclude="__pycache__" \
        --exclude=".env" \
        --exclude="*.pyc" \
        .
    echo "Backup archive created: $BACKUP_DIR/${BACKUP_NAME}.tar.gz"
fi

# Store backup reference
echo "$BACKUP_NAME" > "$BACKUP_DIR/latest"
echo "Backup completed successfully"