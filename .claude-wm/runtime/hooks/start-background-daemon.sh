#!/bin/bash

# Background Worker Daemon Startup Script
# This script starts the background worker as a daemon

HOOKS_DIR="/Users/a.pezzotta/.claude/hooks"
WORKER_BIN="$HOOKS_DIR/background-worker"
PID_FILE="$HOOKS_DIR/queue/daemon.pid"
LOG_FILE="$HOOKS_DIR/logs/background-worker.log"

# Ensure directories exist
mkdir -p "$(dirname "$PID_FILE")"
mkdir -p "$(dirname "$LOG_FILE")"

# Function to check if daemon is running
is_running() {
    if [[ -f "$PID_FILE" ]]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            return 0
        else
            rm -f "$PID_FILE"
            return 1
        fi
    fi
    return 1
}

# Function to start daemon
start_daemon() {
    if is_running; then
        echo "🔄 Background worker daemon is already running (PID: $(cat "$PID_FILE"))"
        return 0
    fi

    echo "🚀 Starting background worker daemon..."

    # Compile if needed
    if [[ ! -x "$WORKER_BIN" ]]; then
        echo "🔨 Compiling background worker..."
        cd "$HOOKS_DIR"
        if ! go build -o background-worker background-worker.go; then
            echo "❌ Failed to compile background worker"
            return 1
        fi
    fi

    # Start daemon in background
    nohup "$WORKER_BIN" start "$HOOKS_DIR" >> "$LOG_FILE" 2>&1 &
    local pid=$!
    
    # Save PID
    echo $pid > "$PID_FILE"
    
    # Verify startup
    sleep 2
    if is_running; then
        echo "✅ Background worker daemon started successfully (PID: $pid)"
        echo "📝 Logs: $LOG_FILE"
        return 0
    else
        echo "❌ Failed to start background worker daemon"
        rm -f "$PID_FILE"
        return 1
    fi
}

# Function to stop daemon
stop_daemon() {
    if ! is_running; then
        echo "⏹️  Background worker daemon is not running"
        return 0
    fi

    local pid=$(cat "$PID_FILE")
    echo "🛑 Stopping background worker daemon (PID: $pid)..."
    
    # Send SIGTERM
    kill "$pid"
    
    # Wait for graceful shutdown
    local count=0
    while kill -0 "$pid" 2>/dev/null && [[ $count -lt 10 ]]; do
        sleep 1
        ((count++))
    done
    
    # Force kill if necessary
    if kill -0 "$pid" 2>/dev/null; then
        echo "⚠️  Forcing shutdown..."
        kill -9 "$pid"
    fi
    
    rm -f "$PID_FILE"
    echo "✅ Background worker daemon stopped"
}

# Function to show status
show_status() {
    if is_running; then
        local pid=$(cat "$PID_FILE")
        echo "🟢 Background worker daemon is running (PID: $pid)"
        
        # Show recent stats if possible
        if [[ -x "$WORKER_BIN" ]]; then
            echo ""
            "$WORKER_BIN" stats 2>/dev/null || echo "📊 Stats not available"
        fi
    else
        echo "🔴 Background worker daemon is not running"
    fi
}

# Function to show logs
show_logs() {
    if [[ -f "$LOG_FILE" ]]; then
        echo "📝 Recent background worker logs:"
        tail -20 "$LOG_FILE"
    else
        echo "📝 No log file found at $LOG_FILE"
    fi
}

# Main command handling
case "${1:-start}" in
    start)
        start_daemon
        ;;
    stop)
        stop_daemon
        ;;
    restart)
        stop_daemon
        sleep 1
        start_daemon
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs}"
        echo ""
        echo "Commands:"
        echo "  start   - Start the background worker daemon"
        echo "  stop    - Stop the background worker daemon"
        echo "  restart - Restart the background worker daemon"
        echo "  status  - Show daemon status and stats"
        echo "  logs    - Show recent daemon logs"
        exit 1
        ;;
esac