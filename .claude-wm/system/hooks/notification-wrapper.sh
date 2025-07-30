#!/bin/bash
# Compatibility wrapper for notify-milestone.sh
source "$(dirname "$0")/notification-wrapper.sh"
notify_milestone "$1"