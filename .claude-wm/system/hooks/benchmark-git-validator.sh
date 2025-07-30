#!/bin/bash
# Performance benchmark script for git validator implementations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ITERATIONS=10
TEST_DATA='{"tool_name": "Bash", "tool_input": {"command": "git commit -m \"Test commit message for benchmarking\""}}'

echo -e "${BLUE}🏃 Git Validator Performance Benchmark${NC}"
echo "Testing both Python and Go implementations..."
echo "Iterations: $ITERATIONS"
echo "==========================================="

# Function to run benchmark
run_benchmark() {
    local name="$1"
    local command="$2"
    local total_time=0
    local min_time=999999
    local max_time=0
    
    echo -e "\n${YELLOW}Testing $name...${NC}"
    
    for i in $(seq 1 $ITERATIONS); do
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
    
    local avg_time=$((total_time / ITERATIONS))
    
    echo -e "  ${BLUE}Results for $name:${NC}"
    echo -e "    Average: ${avg_time}ms"
    echo -e "    Min:     ${min_time}ms"
    echo -e "    Max:     ${max_time}ms"
    echo -e "    Total:   ${total_time}ms"
    
    # Return average time for comparison
    echo $avg_time
}

# Ensure we're in the right directory
cd "$(dirname "$0")"

# Check if Python validator exists
if [ ! -f "git-comprehensive-validator.py" ]; then
    echo -e "${RED}Error: git-comprehensive-validator.py not found${NC}"
    exit 1
fi

# Check if Go validator exists
if [ ! -f "git-validator" ]; then
    echo -e "${YELLOW}Building Go validator...${NC}"
    go build -o git-validator git-validator.go
fi

# Run benchmarks
echo -e "\n${BLUE}🐍 Python Implementation Benchmark${NC}"
python_avg=$(run_benchmark "Python git-comprehensive-validator" "python3 git-comprehensive-validator.py")

echo -e "\n${BLUE}🚀 Go Implementation Benchmark${NC}"
go_avg=$(run_benchmark "Go git-validator" "./git-validator")

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

# Expected target from TODO.md is 40-50% improvement
target_improvement=45
if [ "$improvement_percent" != "N/A" ] && (( $(echo "$improvement_percent > $target_improvement" | bc -l) )); then
    echo -e "\n${GREEN}🎯 Target achieved! Expected 40-50% improvement, got ${improvement_percent}%${NC}"
else
    echo -e "\n${YELLOW}⚠️  Target not met. Expected ${target_improvement}% improvement, got ${improvement_percent}%${NC}"
fi

# Memory usage comparison (if available)
echo -e "\n${BLUE}🧠 Memory Usage Analysis${NC}"
echo "========================================"

# Python memory usage
echo -e "Testing Python memory usage..."
python_memory=$(echo "$TEST_DATA" | timeout 5s /usr/bin/time -l python3 git-comprehensive-validator.py 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")

# Go memory usage  
echo -e "Testing Go memory usage..."
go_memory=$(echo "$TEST_DATA" | timeout 5s /usr/bin/time -l ./git-validator 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")

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

# File size comparison
echo -e "\n${BLUE}📦 Binary Size Comparison${NC}"
echo "========================================"
python_size=$(wc -c < git-comprehensive-validator.py)
go_size=$(wc -c < git-validator)

echo -e "Python script: ${python_size} bytes"
echo -e "Go binary:     ${go_size} bytes"

size_ratio=$(echo "scale=1; $go_size / $python_size" | bc -l)
echo -e "Size ratio: ${size_ratio}x (Go binary is $(echo "scale=1; $size_ratio" | bc -l)x the size of Python script)"

echo -e "\n${GREEN}✅ Benchmark completed successfully!${NC}"