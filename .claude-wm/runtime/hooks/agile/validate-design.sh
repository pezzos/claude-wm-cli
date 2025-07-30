#!/bin/bash
# validate-design.sh - Validate design consistency

set -e

echo "=== Design Validation Hook ==="

# 1. Check if DESIGN.md was created
if [ ! -f "DESIGN.md" ]; then
    echo "Error: DESIGN.md not found!"
    exit 1
fi

# 2. Check if ARCHITECTURE.md was created
if [ ! -f "ARCHITECTURE.md" ]; then
    echo "Error: ARCHITECTURE.md not found!"
    exit 1
fi

# 3. Validate design contains required sections
echo "Validating DESIGN.md structure..."

REQUIRED_SECTIONS=(
    "What We're Building"
    "Technical Approach"
    "Files to Modify"
    "Implementation Steps"
    "Testing Plan"
)

for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -q "$section" DESIGN.md; then
        echo "Warning: Missing section '$section' in DESIGN.md"
    else
        echo "  ✓ Found section: $section"
    fi
done

# 4. Check for buzzwords (warning only)
BUZZWORDS=(
    "synergy"
    "leverage"
    "paradigm"
    "holistic"
    "disruptive"
    "game-changing"
    "revolutionary"
    "cutting-edge"
    "next-generation"
    "world-class"
)

echo ""
echo "Checking for buzzwords..."
BUZZWORD_COUNT=0
for word in "${BUZZWORDS[@]}"; do
    if grep -qi "$word" DESIGN.md ARCHITECTURE.md; then
        echo "  Warning: Found buzzword '$word' - consider using simpler language"
        ((BUZZWORD_COUNT++))
    fi
done

if [ $BUZZWORD_COUNT -eq 0 ]; then
    echo "  ✓ No buzzwords detected - good job keeping it simple!"
fi

# 5. Cross-reference with PRD
if [ -f "PRD.md" ]; then
    echo ""
    echo "Cross-referencing with PRD..."
    
    # Extract problem statement from PRD (simple check)
    if grep -q "Problem Statement" PRD.md && grep -q "Solution" PRD.md; then
        echo "  ✓ PRD contains problem and solution"
        echo "  ✓ Design should address the stated problem"
    fi
fi

echo ""
echo "Design validation completed ✓"