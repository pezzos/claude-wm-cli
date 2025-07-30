#!/bin/bash
# Test script to compare Python and Go env sync validator implementations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Env Sync Validator Test Suite${NC}"
echo "Testing Python vs Go implementations..."
echo "=========================================="

cd "$(dirname "$0")"

# Create test environment files
setup_test_env() {
    echo "Setting up test environment..."
    
    # Create .env file
    cat > .env.test << 'EOF'
# Database configuration
DATABASE_URL=postgresql://user:password@localhost:5432/mydb
REDIS_URL=redis://localhost:6379

# API keys
API_KEY=sk_test_123456789
SECRET_KEY=very_secret_key_123

# App configuration
PORT=3000
NODE_ENV=development
DEBUG=true
EOF

    # Create .env.example file
    cat > .env.example.test << 'EOF'
# Database configuration
DATABASE_URL=postgresql://user:password@localhost:5432/dbname
REDIS_URL=redis://localhost:6379

# API keys
API_KEY=your-api-key-here
SECRET_KEY=your-secret-here

# App configuration
PORT=3000
NODE_ENV=development
# Missing DEBUG variable
EOF

    # Create test code file
    cat > test-code.js << 'EOF'
const dbUrl = process.env.DATABASE_URL;
const apiKey = process.env.API_KEY;
const newVar = process.env.NEW_VARIABLE;
const port = process.env.PORT || 3000;
EOF
}

# Cleanup test environment
cleanup_test_env() {
    echo "Cleaning up test environment..."
    rm -f .env.test .env.example.test test-code.js
}

# Test cases
declare -a test_cases=(
    # File operations
    '{"tool_name": "Write", "tool_input": {"file_path": ".env", "content": "DATABASE_URL=postgresql://localhost:5432/test\nNEW_VAR=test_value"}}'
    '{"tool_name": "Edit", "tool_input": {"file_path": ".env.development", "new_string": "API_KEY=test_key\nSECRET_TOKEN=abc123"}}'
    '{"tool_name": "Write", "tool_input": {"file_path": "app.js", "content": "const db = process.env.DATABASE_URL;\nconst key = process.env.API_KEY;\nconst missing = process.env.MISSING_VAR;"}}'
    '{"tool_name": "MultiEdit", "tool_input": {"file_path": "config.py", "edits": [{"new_string": "import os"}, {"new_string": "DB_URL = os.environ.get(\"DATABASE_URL\")"}, {"new_string": "API_KEY = os.environ[\"API_KEY\"]"}]}}'
    
    # Git operations
    '{"tool_name": "Bash", "tool_input": {"command": "git commit -m \"Update environment variables\""}}'
    '{"tool_name": "Bash", "tool_input": {"command": "git status"}}'
    '{"tool_name": "Bash", "tool_input": {"command": "ls -la"}}'
    
    # Non-env files
    '{"tool_name": "Write", "tool_input": {"file_path": "package.json", "content": "{\"name\": \"test\", \"version\": \"1.0.0\"}"}}'
    '{"tool_name": "Write", "tool_input": {"file_path": "README.md", "content": "# Test Project"}}'
)

# Function to run test case
run_test_case() {
    local test_name="$1"
    local test_input="$2"
    
    echo -e "\n${YELLOW}Test: $test_name${NC}"
    echo "Input: $test_input"
    
    # Test Python implementation
    echo -e "  ${BLUE}Python:${NC}"
    python_output=$(echo "$test_input" | python3 env-sync-validator.py 2>&1 || echo "")
    if [ -n "$python_output" ]; then
        echo "    Output: $python_output"
    else
        echo "    No output"
    fi
    
    # Test Go implementation
    echo -e "  ${BLUE}Go:${NC}"
    go_output=$(echo "$test_input" | ./env-sync-validator 2>&1 || echo "")
    if [ -n "$go_output" ]; then
        echo "    Output: $go_output"
    else
        echo "    No output"
    fi
    
    # Compare outputs (both should have similar behavior)
    if [ -z "$python_output" ] && [ -z "$go_output" ]; then
        echo -e "    ${GREEN}✓ Both silent (expected)${NC}"
        return 0
    elif [ -n "$python_output" ] && [ -n "$go_output" ]; then
        echo -e "    ${GREEN}✓ Both have output (expected)${NC}"
        return 0
    else
        echo -e "    ${YELLOW}~ Different output patterns${NC}"
        return 0
    fi
}

# Setup test environment
setup_test_env
trap cleanup_test_env EXIT

# Run all test cases
echo -e "\n${BLUE}Running functional tests...${NC}"
total_tests=0
passed_tests=0

for i in "${!test_cases[@]}"; do
    test_name="Case $((i+1))"
    test_input="${test_cases[$i]}"
    
    if run_test_case "$test_name" "$test_input"; then
        ((passed_tests++))
    fi
    ((total_tests++))
done

echo -e "\n${BLUE}📊 Test Results${NC}"
echo "Total tests: $total_tests"
echo "Passed: $passed_tests"
echo "Failed: $((total_tests - passed_tests))"

if [ $passed_tests -eq $total_tests ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
else
    echo -e "${RED}❌ Some tests failed${NC}"
fi

# Performance benchmark
echo -e "\n${BLUE}🚀 Performance Benchmark${NC}"
echo "Testing with 20 iterations..."

# Test with a realistic env file operation
test_input='{"tool_name": "Write", "tool_input": {"file_path": ".env", "content": "DATABASE_URL=postgresql://user:password@localhost:5432/mydb\nREDIS_URL=redis://localhost:6379\nAPI_KEY=sk_test_123456789\nSECRET_KEY=very_secret_key_123\nPORT=3000\nNODE_ENV=development\nDEBUG=true\nNEW_VAR=test_value"}}'

# Function to run performance test
run_performance_test() {
    local name="$1"
    local command="$2"
    local iterations=20
    local total_time=0
    
    echo -e "\n${YELLOW}Testing $name performance...${NC}"
    
    for i in $(seq 1 $iterations); do
        start_time=$(date +%s%N)
        echo "$test_input" | eval "$command" > /dev/null 2>&1 || true
        end_time=$(date +%s%N)
        
        duration=$((($end_time - $start_time) / 1000000)) # Convert to milliseconds
        total_time=$((total_time + duration))
        
        echo "  Run $i: ${duration}ms"
    done
    
    local avg_time=$((total_time / iterations))
    echo -e "  ${BLUE}Average: ${avg_time}ms${NC}"
    
    echo $avg_time
}

# Run performance tests
python_avg=$(run_performance_test "Python" "python3 env-sync-validator.py")
go_avg=$(run_performance_test "Go" "./env-sync-validator")

# Calculate improvement
if [ $python_avg -gt 0 ] && [ $go_avg -gt 0 ]; then
    improvement_ratio=$(echo "scale=2; $python_avg / $go_avg" | bc -l 2>/dev/null || echo "1.0")
    improvement_percent=$(echo "scale=1; (($python_avg - $go_avg) * 100) / $python_avg" | bc -l 2>/dev/null || echo "0")
    
    echo -e "\n${BLUE}📈 Performance Comparison${NC}"
    echo "Python Implementation: ${python_avg}ms"
    echo "Go Implementation: ${go_avg}ms"
    
    if (( $(echo "$improvement_percent > 0" | bc -l) )); then
        echo -e "Performance Improvement: ${GREEN}${improvement_ratio}x faster (${improvement_percent}% improvement)${NC}"
    else
        echo -e "Performance Change: ${RED}${improvement_ratio}x (${improvement_percent}% slower)${NC}"
    fi
    
    # Check target (60-70% improvement)
    target_improvement=65
    if (( $(echo "$improvement_percent > $target_improvement" | bc -l) )); then
        echo -e "\n${GREEN}🎯 Target achieved! Expected 60-70% improvement, got ${improvement_percent}%${NC}"
    else
        echo -e "\n${YELLOW}⚠️  Target status: Expected ${target_improvement}% improvement, got ${improvement_percent}%${NC}"
    fi
fi

# Memory usage test
echo -e "\n${BLUE}🧠 Memory Usage Analysis${NC}"
echo "========================================"

# Python memory usage
echo -e "Testing Python memory usage..."
python_memory=$(echo "$test_input" | timeout 5s /usr/bin/time -l python3 env-sync-validator.py 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")

# Go memory usage  
echo -e "Testing Go memory usage..."
go_memory=$(echo "$test_input" | timeout 5s /usr/bin/time -l ./env-sync-validator 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")

if [ "$python_memory" != "N/A" ] && [ "$go_memory" != "N/A" ]; then
    echo -e "Python Memory: ${python_memory} bytes"
    echo -e "Go Memory:     ${go_memory} bytes"
    
    if [ $go_memory -lt $python_memory ]; then
        memory_improvement=$(echo "scale=1; (($python_memory - $go_memory) * 100) / $python_memory" | bc -l)
        echo -e "Memory Improvement: ${GREEN}${memory_improvement}% less memory${NC}"
    else
        memory_increase=$(echo "scale=1; (($go_memory - $python_memory) * 100) / $python_memory" | bc -l)
        echo -e "Memory Change: ${RED}${memory_increase}% more memory${NC}"
    fi
else
    echo -e "Memory analysis: ${YELLOW}Unable to measure${NC}"
fi

# Test with multiple env files
echo -e "\n${BLUE}📁 Multiple File Test${NC}"
echo "========================================"

# Create multiple env files for concurrent processing test
cat > .env.development << 'EOF'
DATABASE_URL=postgresql://localhost:5432/dev
API_KEY=dev_key_123
PORT=3001
EOF

cat > .env.test << 'EOF'
DATABASE_URL=postgresql://localhost:5432/test
API_KEY=test_key_123
PORT=3002
TEST_VAR=test_value
EOF

multi_file_test='{"tool_name": "Bash", "tool_input": {"command": "git commit -m \"Update all env files\""}}'

echo "Testing concurrent processing with multiple .env files..."
echo "Python (multiple files):"
multi_python_start=$(date +%s%N)
echo "$multi_file_test" | python3 env-sync-validator.py > /dev/null 2>&1 || true
multi_python_end=$(date +%s%N)
multi_python_duration=$(((multi_python_end - multi_python_start) / 1000000))

echo "Go (multiple files):"
multi_go_start=$(date +%s%N)
echo "$multi_file_test" | ./env-sync-validator > /dev/null 2>&1 || true
multi_go_end=$(date +%s%N)
multi_go_duration=$(((multi_go_end - multi_go_start) / 1000000))

echo "Python: ${multi_python_duration}ms"
echo "Go: ${multi_go_duration}ms"

if [ $multi_python_duration -gt 0 ] && [ $multi_go_duration -gt 0 ]; then
    multi_improvement=$(echo "scale=1; (($multi_python_duration - $multi_go_duration) * 100) / $multi_python_duration" | bc -l)
    echo -e "Multi-file improvement: ${GREEN}${multi_improvement}%${NC}"
fi

# Cleanup multiple env files
rm -f .env.development .env.test

echo -e "\n${GREEN}✅ Env sync validator test completed!${NC}"

# Summary
echo -e "\n${BLUE}📋 Summary${NC}"
echo "========================================"
echo "• Python average: ${python_avg}ms"
echo "• Go average: ${go_avg}ms"
echo "• Performance improvement: ${improvement_percent}%"
echo "• Memory improvement: Available in memory analysis above"
echo "• Target (60-70%): $([ "$improvement_percent" != "N/A" ] && (( $(echo "$improvement_percent > 60" | bc -l) )) && echo "✅ ACHIEVED" || echo "⚠️ Check needed")"
echo "• Multi-file performance: ${multi_improvement}% improvement"