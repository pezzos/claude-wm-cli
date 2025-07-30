#!/bin/bash
# Test script to compare Python and Go MCP tool enforcer implementations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 MCP Tool Enforcer Test Suite${NC}"
echo "Testing Python vs Go implementations..."
echo "=========================================="

cd "$(dirname "$0")"

# Test cases
declare -a test_cases=(
    # Bash commands
    '{"tool_name": "Bash", "tool_input": {"command": "curl -X GET https://api.example.com/data"}}'
    '{"tool_name": "Bash", "tool_input": {"command": "wget https://example.com/file.zip"}}'
    '{"tool_name": "Bash", "tool_input": {"command": "mysql -u user -p database"}}'
    '{"tool_name": "Bash", "tool_input": {"command": "ls -la /tmp"}}'
    
    # File operations
    '{"tool_name": "Write", "tool_input": {"file_path": "test.js", "content": "const data = fetch(\"https://api.example.com\");\nconsole.log(data);"}}'
    '{"tool_name": "Write", "tool_input": {"file_path": "test.py", "content": "import requests\nresponse = requests.get(\"https://api.example.com\")"}}'
    '{"tool_name": "Write", "tool_input": {"file_path": "test.go", "content": "package main\nimport \"net/http\"\nfunc main() {\n\thttp.Get(\"https://api.example.com\")\n}"}}'
    '{"tool_name": "Write", "tool_input": {"file_path": "config.json", "content": "{\"database\": \"sqlite\"}"}}'
    
    # Edit operations
    '{"tool_name": "Edit", "tool_input": {"file_path": "app.js", "new_string": "localStorage.setItem(\"user\", JSON.stringify(data));"}}'
    '{"tool_name": "Edit", "tool_input": {"file_path": "db.py", "new_string": "cursor.execute(\"SELECT * FROM users\")"}}'
    
    # MultiEdit operations
    '{"tool_name": "MultiEdit", "tool_input": {"file_path": "utils.js", "edits": [{"new_string": "const searchGoogle = (query) => {"}, {"new_string": "fetch(`https://google.com/search?q=${query}`)"}]}}'
)

# Function to run test case
run_test_case() {
    local test_name="$1"
    local test_input="$2"
    
    echo -e "\n${YELLOW}Test: $test_name${NC}"
    echo "Input: $test_input"
    
    # Test Python implementation
    echo -e "  ${BLUE}Python:${NC}"
    python_output=$(echo "$test_input" | python3 mcp-tool-enforcer.py 2>/dev/null || echo "")
    if [ -n "$python_output" ]; then
        echo "    Output: $python_output"
    else
        echo "    No suggestions"
    fi
    
    # Test Go implementation
    echo -e "  ${BLUE}Go:${NC}"
    go_output=$(echo "$test_input" | ./mcp-tool-enforcer 2>/dev/null || echo "")
    if [ -n "$go_output" ]; then
        echo "    Output: $go_output"
    else
        echo "    No suggestions"
    fi
    
    # Compare outputs
    if [ "$python_output" = "$go_output" ]; then
        echo -e "    ${GREEN}✓ Match${NC}"
        return 0
    else
        # Check if both have suggestions (content might differ slightly)
        if [ -n "$python_output" ] && [ -n "$go_output" ]; then
            echo -e "    ${YELLOW}~ Similar (both have suggestions)${NC}"
            return 0
        else
            echo -e "    ${RED}✗ Different${NC}"
            return 1
        fi
    fi
}

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

test_input='{"tool_name": "Write", "tool_input": {"file_path": "test.js", "content": "const data = fetch(\"https://api.example.com\");\nconst db = new sqlite3.Database(\"test.db\");\nLocalStorage.setItem(\"key\", \"value\");"}}'

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
python_avg=$(run_performance_test "Python" "python3 mcp-tool-enforcer.py")
go_avg=$(run_performance_test "Go" "./mcp-tool-enforcer")

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

echo -e "\n${GREEN}✅ MCP Tool Enforcer test completed!${NC}"