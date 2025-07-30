#!/bin/bash
# Performance benchmark script for duplicate detector implementations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Duplicate Detector Performance Benchmark${NC}"
echo "Testing Python vs Go implementations..."
echo "Iterations: 20"
echo "==========================================="

TEST_DATA='{"tool_name": "Write", "tool_input": {"file_path": "components/NewComponent.tsx"}}'

cd "$(dirname "$0")"

# Function to run benchmark
run_benchmark() {
    local name="$1"
    local command="$2"
    local iterations=20
    local total_time=0
    local min_time=999999
    local max_time=0
    
    echo -e "\n${YELLOW}Testing $name...${NC}"
    
    for i in $(seq 1 $iterations); do
        start_time=$(date +%s%N)
        
        # Run the command and capture output
        if output=$(echo "$TEST_DATA" | eval "$command" 2>&1); then
            success=true
        else
            success=false
        fi
        
        end_time=$(date +%s%N)
        duration=$((($end_time - $start_time) / 1000000)) # Convert to milliseconds
        
        total_time=$((total_time + duration))
        
        if [ $duration -lt $min_time ]; then
            min_time=$duration
        fi
        
        if [ $duration -gt $max_time ]; then
            max_time=$duration
        fi
        
        echo "  Run $i: ${duration}ms $([ "$success" = true ] && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}")"
    done
    
    local avg_time=$((total_time / iterations))
    
    echo -e "  ${BLUE}Results for $name:${NC}"
    echo -e "    Average: ${avg_time}ms"
    echo -e "    Min:     ${min_time}ms"
    echo -e "    Max:     ${max_time}ms"
    echo -e "    Total:   ${total_time}ms"
    
    # Return average time for comparison
    echo $avg_time
}

# Check if both implementations exist
if [ ! -f "duplicate-detector.py" ]; then
    echo -e "${RED}Error: duplicate-detector.py not found${NC}"
    exit 1
fi

if [ ! -f "duplicate-detector-go" ]; then
    echo -e "${RED}Error: duplicate-detector-go not found${NC}"
    exit 1
fi

# Run benchmarks
echo -e "\n${BLUE}🐍 Python Implementation Benchmark${NC}"
python_avg=$(run_benchmark "Python duplicate-detector" "python3 duplicate-detector.py")

echo -e "\n${BLUE}🚀 Go Implementation Benchmark${NC}"
go_avg=$(run_benchmark "Go duplicate-detector" "./duplicate-detector-go")

# Calculate improvement
if [ $python_avg -gt 0 ]; then
    improvement_ratio=$(echo "scale=2; $python_avg / $go_avg" | bc -l 2>/dev/null || echo "N/A")
    improvement_percent=$(echo "scale=1; (($python_avg - $go_avg) * 100) / $python_avg" | bc -l 2>/dev/null || echo "N/A")
else
    improvement_ratio="N/A"
    improvement_percent="N/A"
fi

echo -e "\n${BLUE}📊 Performance Comparison${NC}"
echo "========================================"
echo -e "Python Implementation:  ${python_avg}ms"
echo -e "Go Implementation:      ${go_avg}ms"

if [ "$improvement_ratio" != "N/A" ] && [ "$improvement_percent" != "N/A" ]; then
    if (( $(echo "$go_avg < $python_avg" | bc -l) )); then
        echo -e "Performance Improvement: ${GREEN}${improvement_ratio}x faster (${improvement_percent}% improvement)${NC}"
    else
        echo -e "Performance Change: ${RED}${improvement_ratio}x (${improvement_percent}% slower)${NC}"
    fi
else
    echo -e "Performance Change: ${YELLOW}Unable to calculate${NC}"
fi

# Expected target from TODO.md is 70-80% improvement
target_improvement=75
if [ "$improvement_percent" != "N/A" ] && (( $(echo "$improvement_percent > $target_improvement" | bc -l) )); then
    echo -e "\n${GREEN}🎯 Target achieved! Expected 70-80% improvement, got ${improvement_percent}%${NC}"
else
    echo -e "\n${YELLOW}⚠️  Target status: Expected ${target_improvement}% improvement, got ${improvement_percent}%${NC}"
fi

# Memory usage comparison
echo -e "\n${BLUE}🧠 Memory Usage Analysis${NC}"
echo "========================================"

# Python memory usage
echo -e "Testing Python memory usage..."
python_memory=$(echo "$TEST_DATA" | timeout 5s /usr/bin/time -l python3 duplicate-detector.py 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")

# Go memory usage  
echo -e "Testing Go memory usage..."
go_memory=$(echo "$TEST_DATA" | timeout 5s /usr/bin/time -l ./duplicate-detector-go 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")

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
    echo -e "Memory analysis: ${YELLOW}Unable to measure (time command may not be available)${NC}"
fi

# Concurrent processing test
echo -e "\n${BLUE}⚡ Concurrent Processing Test${NC}"
echo "========================================"
echo "Testing with CACHE_ENABLED=true for Go implementation..."

CACHE_TEST='{"tool_name": "Write", "tool_input": {"file_path": "components/CachedTestComponent.tsx"}}'

# Test Go with cache enabled
echo -e "Go with cache (5 runs):"
cache_total=0
for i in $(seq 1 5); do
    start=$(date +%s%N)
    echo "$CACHE_TEST" | CACHE_ENABLED=true ./duplicate-detector-go > /dev/null 2>&1 || true
    end=$(date +%s%N)
    duration=$(((end - start) / 1000000))
    echo "  Run $i: ${duration}ms"
    cache_total=$((cache_total + duration))
done
cache_avg=$((cache_total / 5))
echo "Cache-enabled average: ${cache_avg}ms"

echo -e "\n${GREEN}✅ Benchmark completed successfully!${NC}"

# Summary
echo -e "\n${BLUE}📋 Summary${NC}"
echo "========================================"
echo "• Python average: ${python_avg}ms"
echo "• Go average: ${go_avg}ms"
echo "• Performance improvement: ${improvement_percent}%"
echo "• Memory improvement: Available in memory analysis above"
echo "• Target (70-80%): $([ "$improvement_percent" != "N/A" ] && (( $(echo "$improvement_percent > 70" | bc -l) )) && echo "✅ ACHIEVED" || echo "⚠️ Check needed")"