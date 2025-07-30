#!/bin/bash
# Test script to compare Python vs Go duplicate detector implementations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Duplicate Detector Comparison Test${NC}"
echo "Testing Python vs Go implementations..."
echo "======================================"

# Test data - different types of files
TEST_CASES=(
    '{"tool_name": "Write", "tool_input": {"file_path": "app/dashboard/page.tsx"}}'
    '{"tool_name": "Write", "tool_input": {"file_path": "app/api/users/route.ts"}}'
    '{"tool_name": "Write", "tool_input": {"file_path": "components/Button.tsx"}}'
    '{"tool_name": "Write", "tool_input": {"file_path": "utils/helpers.ts"}}'
    '{"tool_name": "Write", "tool_input": {"file_path": "src/styles.css"}}'
)

TEST_DESCRIPTIONS=(
    "Next.js Page Route"
    "API Route"
    "React Component"
    "Utility File"
    "Non-relevant File"
)

cd "$(dirname "$0")"

# Ensure both implementations exist
if [ ! -f "duplicate-detector.py" ]; then
    echo -e "${RED}Error: duplicate-detector.py not found${NC}"
    exit 1
fi

if [ ! -f "duplicate-detector-go" ]; then
    echo -e "${RED}Error: duplicate-detector-go not found${NC}"
    exit 1
fi

echo -e "\n${BLUE}📊 Functional Comparison${NC}"
echo "========================================"

for i in "${!TEST_CASES[@]}"; do
    test_case="${TEST_CASES[$i]}"
    description="${TEST_DESCRIPTIONS[$i]}"
    
    echo -e "\n${YELLOW}Test $((i+1)): $description${NC}"
    echo "Input: $(echo "$test_case" | jq -r '.tool_input.file_path')"
    
    # Test Python implementation
    echo -e "  ${BLUE}Python:${NC}"
    python_start=$(date +%s%N)
    if python_output=$(echo "$test_case" | python3 duplicate-detector.py 2>&1); then
        python_success=true
        python_exit=0
    else
        python_success=false
        python_exit=$?
    fi
    python_end=$(date +%s%N)
    python_time=$(((python_end - python_start) / 1000000))
    
    echo "    Time: ${python_time}ms"
    echo "    Exit: $python_exit"
    if [ "$python_success" = false ] && [ ${#python_output} -gt 0 ]; then
        echo "    Output: $(echo "$python_output" | head -n 1)"
    fi
    
    # Test Go implementation
    echo -e "  ${GREEN}Go:${NC}"
    go_start=$(date +%s%N)
    if go_output=$(echo "$test_case" | ./duplicate-detector-go 2>&1); then
        go_success=true
        go_exit=0
    else
        go_success=false
        go_exit=$?
    fi
    go_end=$(date +%s%N)
    go_time=$(((go_end - go_start) / 1000000))
    
    echo "    Time: ${go_time}ms"
    echo "    Exit: $go_exit"
    if [ "$go_success" = false ] && [ ${#go_output} -gt 0 ]; then
        echo "    Output: $(echo "$go_output" | head -n 1)"
    fi
    
    # Compare results
    if [ $python_exit -eq $go_exit ]; then
        echo -e "    Result: ${GREEN}✓ Consistent exit codes${NC}"
    else
        echo -e "    Result: ${RED}✗ Different exit codes (Python: $python_exit, Go: $go_exit)${NC}"
    fi
    
    # Check if both found DETECTION_RESULT
    python_has_result=$(echo "$python_output" | grep -c "DETECTION_RESULT" || true)
    go_has_result=$(echo "$go_output" | grep -c "DETECTION_RESULT" || true)
    
    if [ $python_has_result -eq $go_has_result ]; then
        echo -e "    JSON Output: ${GREEN}✓ Consistent${NC}"
    else
        echo -e "    JSON Output: ${YELLOW}⚠ Different (Python: $python_has_result, Go: $go_has_result)${NC}"
    fi
done

echo -e "\n${BLUE}🏃 Performance Benchmark${NC}"
echo "========================================"

# Performance test with a realistic file
PERF_TEST='{"tool_name": "Write", "tool_input": {"file_path": "components/TestComponent.tsx"}}'
ITERATIONS=10

echo -e "\nTesting with $ITERATIONS iterations..."

# Python performance
echo -e "\n${BLUE}Python Implementation:${NC}"
python_total=0
for i in $(seq 1 $ITERATIONS); do
    start=$(date +%s%N)
    echo "$PERF_TEST" | python3 duplicate-detector.py > /dev/null 2>&1 || true
    end=$(date +%s%N)
    duration=$(((end - start) / 1000000))
    echo "  Run $i: ${duration}ms"
    python_total=$((python_total + duration))
done
python_avg=$((python_total / ITERATIONS))
echo "  Average: ${python_avg}ms"

# Go performance
echo -e "\n${GREEN}Go Implementation:${NC}"
go_total=0
for i in $(seq 1 $ITERATIONS); do
    start=$(date +%s%N)
    echo "$PERF_TEST" | ./duplicate-detector-go > /dev/null 2>&1 || true
    end=$(date +%s%N)
    duration=$(((end - start) / 1000000))
    echo "  Run $i: ${duration}ms"
    go_total=$((go_total + duration))
done
go_avg=$((go_total / ITERATIONS))
echo "  Average: ${go_avg}ms"

# Calculate improvement
if [ $go_avg -lt $python_avg ]; then
    improvement=$(echo "scale=1; (($python_avg - $go_avg) * 100) / $python_avg" | bc -l 2>/dev/null || echo "N/A")
    speedup=$(echo "scale=1; $python_avg / $go_avg" | bc -l 2>/dev/null || echo "N/A")
    echo -e "\n${GREEN}🎯 Performance Results:${NC}"
    echo "Python:        ${python_avg}ms"
    echo "Go:            ${go_avg}ms"
    echo "Improvement:   ${improvement}% faster"
    echo "Speedup:       ${speedup}x"
    
    # Check against target (70-80% improvement)
    target=75
    if [ "$improvement" != "N/A" ] && (( $(echo "$improvement > $target" | bc -l 2>/dev/null || echo 0) )); then
        echo -e "${GREEN}✅ Target exceeded! Expected 70-80%, got ${improvement}%${NC}"
    else
        echo -e "${YELLOW}⚠️ Target not met. Expected 70-80%, got ${improvement}%${NC}"
    fi
else
    echo -e "\n${RED}❌ Go implementation is slower than Python${NC}"
    echo "Python: ${python_avg}ms"
    echo "Go:     ${go_avg}ms"
fi

# Memory comparison
echo -e "\n${BLUE}💾 Memory Usage Comparison${NC}"
echo "========================================"

python_memory=$(echo "$PERF_TEST" | timeout 5s /usr/bin/time -l python3 duplicate-detector.py 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")
go_memory=$(echo "$PERF_TEST" | timeout 5s /usr/bin/time -l ./duplicate-detector-go 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")

if [ "$python_memory" != "N/A" ] && [ "$go_memory" != "N/A" ]; then
    echo "Python Memory: ${python_memory} bytes"
    echo "Go Memory:     ${go_memory} bytes"
    
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

echo -e "\n${GREEN}✅ Comparison completed!${NC}"