#!/bin/bash
# Pre-implementation Context7 verification hook

TASK="$1"
WORKING_DIR="$2"

echo "🔍 Context7 Pre-Implementation Check..."

# Extract potential libraries from task description
LIBRARIES=$(echo "$TASK" | grep -oiE '\b(react|next|nextauth|express|fastapi|django|flask|vue|angular|typescript|node|supabase|tailwind|prisma|mongodb|postgresql|redis|stripe|auth0|firebase|aws|vercel)\b' | sort -u)

if [ -n "$LIBRARIES" ]; then
    echo "📚 Libraries detected in task: $LIBRARIES"
    echo "✅ Context7 will provide version-specific documentation"
    
    # Check if package.json or requirements.txt exists for version verification
    if [ -f "$WORKING_DIR/package.json" ]; then
        echo "📦 Found package.json - will verify versions against Context7"
        
        # Show current versions for detected libraries
        for lib in $LIBRARIES; do
            if grep -q "\"$lib\"" "$WORKING_DIR/package.json" 2>/dev/null; then
                VERSION=$(grep "\"$lib\"" "$WORKING_DIR/package.json" | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')
                echo "  - $lib: $VERSION (current)"
            else
                echo "  - $lib: not installed (Context7 will use latest stable)"
            fi
        done
        
    elif [ -f "$WORKING_DIR/requirements.txt" ]; then
        echo "📦 Found requirements.txt - will verify versions against Context7"
        
        # Show current versions for detected libraries
        for lib in $LIBRARIES; do
            if grep -q "$lib" "$WORKING_DIR/requirements.txt" 2>/dev/null; then
                VERSION=$(grep "$lib" "$WORKING_DIR/requirements.txt" | head -1 | sed 's/.*==\([^=]*\).*/\1/')
                echo "  - $lib: $VERSION (current)"
            else
                echo "  - $lib: not specified (Context7 will use latest stable)"
            fi
        done
        
    elif [ -f "$WORKING_DIR/Cargo.toml" ]; then
        echo "📦 Found Cargo.toml - will verify versions against Context7"
        
    elif [ -f "$WORKING_DIR/pom.xml" ]; then
        echo "📦 Found pom.xml - will verify versions against Context7"
        
    elif [ -f "$WORKING_DIR/go.mod" ]; then
        echo "📦 Found go.mod - will verify versions against Context7"
        
    else
        echo "⚠️ No package file found - Context7 will use latest stable versions"
    fi
    
    # Log for tracking
    echo "$(date '+%Y-%m-%d %H:%M:%S') Context7 Pre-Check: $LIBRARIES" >> ~/.claude/hooks/logs/context7-checks.log
    
else
    echo "ℹ️ No major libraries detected - Context7 check skipped"
fi

# Check for common implementation patterns that benefit from Context7
IMPLEMENTATION_PATTERNS=$(echo "$TASK" | grep -oE '\b(authentication|routing|components|hooks|api|database|testing|deployment|styling|forms)\b' | sort -u)

if [ -n "$IMPLEMENTATION_PATTERNS" ]; then
    echo "🎯 Implementation patterns detected: $IMPLEMENTATION_PATTERNS"
    echo "📖 Context7 will provide pattern-specific documentation"
fi

# Always exit successfully - this is informational only
exit 0