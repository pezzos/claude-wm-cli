#!/bin/bash

# Cache Integration Script for Hooks
# Provides simple cache operations for shell-based hooks

HOOKS_DIR="/Users/a.pezzotta/.claude/hooks"
CACHE_BIN="$HOOKS_DIR/shared-cache"

# Ensure cache binary is compiled
ensure_cache_binary() {
    if [[ ! -x "$CACHE_BIN" ]]; then
        echo "🔨 Compiling shared cache..."
        cd "$HOOKS_DIR/cache"
        if ! go build -o ../shared-cache shared-cache.go; then
            echo "❌ Failed to compile shared cache"
            return 1
        fi
    fi
    return 0
}

# Get cached git info
get_git_info() {
    ensure_cache_binary || return 1
    "$CACHE_BIN" get-git-info 2>/dev/null
}

# Get cached file info
get_file_info() {
    local file_path="$1"
    if [[ -z "$file_path" ]]; then
        echo "Usage: get_file_info <file-path>"
        return 1
    fi
    
    ensure_cache_binary || return 1
    "$CACHE_BIN" get-file-info "$file_path" 2>/dev/null
}

# Invalidate git cache (call when git operations occur)
invalidate_git_cache() {
    ensure_cache_binary || return 1
    "$CACHE_BIN" invalidate-git 2>/dev/null
    echo "🗑️  Git cache invalidated"
}

# Invalidate file cache (call when files are modified)
invalidate_file_cache() {
    local file_path="$1"
    if [[ -z "$file_path" ]]; then
        echo "Usage: invalidate_file_cache <file-path>"
        return 1
    fi
    
    ensure_cache_binary || return 1
    "$CACHE_BIN" invalidate-file "$file_path" 2>/dev/null
    echo "🗑️  File cache invalidated for $file_path"
}

# Get cache statistics
get_cache_stats() {
    ensure_cache_binary || return 1
    "$CACHE_BIN" stats
}

# Clean up expired cache entries
cleanup_cache() {
    ensure_cache_binary || return 1
    "$CACHE_BIN" cleanup
}

# Helper function to check if git info is cached
is_git_info_cached() {
    local git_info
    git_info=$(get_git_info 2>/dev/null)
    if [[ -n "$git_info" ]] && echo "$git_info" | grep -q "timestamp"; then
        return 0  # Cached
    else
        return 1  # Not cached
    fi
}

# Helper function to check if file info is cached
is_file_info_cached() {
    local file_path="$1"
    local file_info
    file_info=$(get_file_info "$file_path" 2>/dev/null)
    if [[ -n "$file_info" ]] && echo "$file_info" | grep -q "mod_time"; then
        return 0  # Cached
    else
        return 1  # Not cached
    fi
}

# Smart git status function that uses cache when possible
smart_git_status() {
    if is_git_info_cached; then
        echo "🎯 Using cached git status"
        get_git_info | grep -A 20 '"status"' | grep -v '"status"' | head -1 | cut -d'"' -f4
    else
        echo "💫 Computing fresh git status"
        git status --porcelain
        # Cache will be populated by the cache system
    fi
}

# Smart file modification check
smart_file_modified() {
    local file_path="$1"
    local reference_time="$2"  # Optional: compare against specific time
    
    if [[ ! -f "$file_path" ]]; then
        return 1  # File doesn't exist
    fi
    
    if is_file_info_cached "$file_path"; then
        echo "🎯 Using cached file info for $file_path"
        local cached_modtime
        cached_modtime=$(get_file_info "$file_path" | grep '"mod_time"' | cut -d'"' -f4)
        
        if [[ -n "$reference_time" ]]; then
            # Compare against reference time
            if [[ "$cached_modtime" > "$reference_time" ]]; then
                return 0  # Modified
            else
                return 1  # Not modified
            fi
        else
            # Just return the modification time
            echo "$cached_modtime"
            return 0
        fi
    else
        echo "💫 Computing fresh file info for $file_path"
        if [[ -n "$reference_time" ]]; then
            # Use find to check modification
            if find "$file_path" -newer "$reference_time" 2>/dev/null | grep -q .; then
                return 0  # Modified
            else
                return 1  # Not modified
            fi
        else
            # Return modification time
            stat -f "%m" "$file_path" 2>/dev/null || stat -c "%Y" "$file_path" 2>/dev/null
            return 0
        fi
    fi
}

# Batch file operations with cache awareness
batch_file_check() {
    local files=("$@")
    local cached_count=0
    local total_count=${#files[@]}
    
    echo "📊 Checking $total_count files..."
    
    for file in "${files[@]}"; do
        if is_file_info_cached "$file"; then
            ((cached_count++))
        fi
    done
    
    local cache_ratio=$((cached_count * 100 / total_count))
    echo "🎯 Cache hit ratio: $cache_ratio% ($cached_count/$total_count)"
    
    return $cached_count
}

# Cache warming functions for common operations
warm_git_cache() {
    echo "🔥 Warming git cache..."
    get_git_info > /dev/null 2>&1
    echo "✅ Git cache warmed"
}

warm_file_cache() {
    local directory="${1:-.}"
    echo "🔥 Warming file cache for $directory..."
    
    local count=0
    while IFS= read -r -d '' file; do
        get_file_info "$file" > /dev/null 2>&1
        ((count++))
        if ((count % 50 == 0)); then
            echo "  Cached $count files..."
        fi
    done < <(find "$directory" -type f -print0 2>/dev/null)
    
    echo "✅ File cache warmed for $count files"
}

# Performance monitoring
measure_cache_performance() {
    local operation="$1"
    shift
    local args=("$@")
    
    local start_time
    start_time=$(python3 -c "import time; print(int(time.time() * 1000))")  # milliseconds
    
    case "$operation" in
        git-info)
            get_git_info > /dev/null
            ;;
        file-info)
            get_file_info "${args[0]}" > /dev/null
            ;;
        *)
            echo "Unknown operation: $operation"
            return 1
            ;;
    esac
    
    local end_time
    end_time=$(python3 -c "import time; print(int(time.time() * 1000))")
    local duration=$((end_time - start_time))
    
    echo "⏱️  $operation took ${duration}ms"
}

# Main CLI interface
case "${1:-help}" in
    git-info)
        get_git_info
        ;;
    file-info)
        get_file_info "$2"
        ;;
    invalidate-git)
        invalidate_git_cache
        ;;
    invalidate-file)
        invalidate_file_cache "$2"
        ;;
    stats)
        get_cache_stats
        ;;
    cleanup)
        cleanup_cache
        ;;
    smart-git-status)
        smart_git_status
        ;;
    smart-file-modified)
        smart_file_modified "$2" "$3"
        ;;
    batch-file-check)
        shift
        batch_file_check "$@"
        ;;
    warm-git)
        warm_git_cache
        ;;
    warm-files)
        warm_file_cache "$2"
        ;;
    measure)
        shift
        measure_cache_performance "$@"
        ;;
    help|*)
        echo "Cache Integration Script"
        echo ""
        echo "Usage: $0 <command> [args...]"
        echo ""
        echo "Commands:"
        echo "  git-info                    - Get cached git information"
        echo "  file-info <path>            - Get cached file information"
        echo "  invalidate-git              - Invalidate git cache"
        echo "  invalidate-file <path>      - Invalidate file cache"
        echo "  stats                       - Show cache statistics"
        echo "  cleanup                     - Clean expired cache entries"
        echo "  smart-git-status            - Git status with cache awareness"
        echo "  smart-file-modified <path>  - File modification check with cache"
        echo "  batch-file-check <files...> - Check cache hit ratio for files"
        echo "  warm-git                    - Pre-populate git cache"
        echo "  warm-files [directory]      - Pre-populate file cache"
        echo "  measure <operation> [args]  - Measure cache performance"
        echo ""
        echo "Cache-aware functions for hooks:"
        echo "  is_git_info_cached          - Check if git info is cached"
        echo "  is_file_info_cached <path>  - Check if file info is cached"
        echo "  smart_git_status            - Git status with cache"
        echo "  smart_file_modified <path>  - File modification with cache"
        ;;
esac