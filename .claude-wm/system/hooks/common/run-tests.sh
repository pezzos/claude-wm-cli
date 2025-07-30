#!/bin/bash
# run-tests.sh - Run tests based on project type

set -e

echo "Running tests..."

# Track if any tests were run
TESTS_RUN=false

# JavaScript/TypeScript projects
if [ -f "package.json" ]; then
    # Check test script in package.json
    if grep -q '"test"' package.json; then
        echo "Running npm test..."
        npm test || { echo "Tests failed!"; exit 1; }
        TESTS_RUN=true
    fi
fi

# Python projects
if [ -f "requirements.txt" ] || [ -f "pyproject.toml" ] || [ -f "setup.py" ]; then
    # Check for pytest
    if command -v pytest &> /dev/null || [ -f "pytest.ini" ] || [ -f "setup.cfg" ]; then
        echo "Running pytest..."
        pytest || { echo "Tests failed!"; exit 1; }
        TESTS_RUN=true
    elif [ -d "tests" ] && command -v python &> /dev/null; then
        echo "Running Python unittest..."
        python -m unittest discover tests || { echo "Tests failed!"; exit 1; }
        TESTS_RUN=true
    fi
fi

# Go projects
if [ -f "go.mod" ]; then
    echo "Running go test..."
    go test ./... || { echo "Tests failed!"; exit 1; }
    TESTS_RUN=true
fi

# Rust projects
if [ -f "Cargo.toml" ]; then
    echo "Running cargo test..."
    cargo test || { echo "Tests failed!"; exit 1; }
    TESTS_RUN=true
fi

# PHP projects
if [ -f "composer.json" ]; then
    if [ -f "phpunit.xml" ] || [ -f "phpunit.xml.dist" ]; then
        echo "Running PHPUnit..."
        ./vendor/bin/phpunit || { echo "Tests failed!"; exit 1; }
        TESTS_RUN=true
    fi
fi

if [ "$TESTS_RUN" = true ]; then
    echo "All tests passed ✓"
else
    echo "Warning: No test suite found for this project"
    echo "Consider adding tests to ensure code quality"
fi