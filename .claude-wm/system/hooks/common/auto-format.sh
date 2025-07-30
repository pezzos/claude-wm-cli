#!/bin/bash
# auto-format.sh - Automatically format code based on project type

set -e

echo "Running auto-format..."

# Track if any formatter was run
FORMATTED=false

# JavaScript/TypeScript projects
if [ -f "package.json" ]; then
    # Check for prettier
    if grep -q "prettier" package.json; then
        echo "Running Prettier..."
        if [ -f ".prettierrc" ] || [ -f ".prettierrc.js" ] || [ -f ".prettierrc.json" ]; then
            npx prettier --write "**/*.{js,jsx,ts,tsx,json,css,scss,md}" 2>/dev/null || true
            FORMATTED=true
        fi
    fi
    
    # Check for ESLint with --fix
    if grep -q "eslint" package.json; then
        echo "Running ESLint --fix..."
        npx eslint --fix "**/*.{js,jsx,ts,tsx}" 2>/dev/null || true
        FORMATTED=true
    fi
fi

# Python projects
if [ -f "requirements.txt" ] || [ -f "pyproject.toml" ] || [ -f "setup.py" ]; then
    # Check for black
    if command -v black &> /dev/null; then
        echo "Running Black..."
        black . 2>/dev/null || true
        FORMATTED=true
    fi
    
    # Check for isort
    if command -v isort &> /dev/null; then
        echo "Running isort..."
        isort . 2>/dev/null || true
        FORMATTED=true
    fi
    
    # Check for autopep8
    if command -v autopep8 &> /dev/null && [ "$FORMATTED" = false ]; then
        echo "Running autopep8..."
        find . -name "*.py" -not -path "./venv/*" -not -path "./.venv/*" -exec autopep8 --in-place {} \;
        FORMATTED=true
    fi
fi

# Go projects
if [ -f "go.mod" ]; then
    echo "Running go fmt..."
    go fmt ./... 2>/dev/null || true
    FORMATTED=true
fi

# Rust projects
if [ -f "Cargo.toml" ]; then
    echo "Running cargo fmt..."
    cargo fmt 2>/dev/null || true
    FORMATTED=true
fi

if [ "$FORMATTED" = true ]; then
    echo "Auto-formatting completed ✓"
else
    echo "No formatter found for this project type"
fi