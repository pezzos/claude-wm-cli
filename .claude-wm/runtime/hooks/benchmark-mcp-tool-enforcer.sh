#!/bin/bash
# Performance benchmark script for MCP tool enforcer implementations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 MCP Tool Enforcer Performance Benchmark${NC}"
echo "Testing Python vs Go implementations..."
echo "Iterations: 20"
echo "==========================================="

TEST_DATA='{
  "tool_name": "Write",
  "tool_input": {
    "file_path": "complex.js",
    "content": "const data = fetch(\"https://api.example.com\");\nconst db = new sqlite3.Database(\"test.db\");\nlocalStorage.setItem(\"key\", \"value\");\naxios.get(\"https://service.com\");\nconst search = google.search(\"query\");\nfs.readFile(\"file.txt\", callback);\nconst cache = new Map();\nfunction processData() {\n  return requests.post(\"https://api.com\", data);\n}\n"
  }
}'

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
if [ ! -f "mcp-tool-enforcer.py" ]; then
    echo -e "${RED}Error: mcp-tool-enforcer.py not found${NC}"
    exit 1
fi

if [ ! -f "mcp-tool-enforcer" ]; then
    echo -e "${RED}Error: mcp-tool-enforcer not found${NC}"
    exit 1
fi

# Run benchmarks
echo -e "\n${BLUE}🐍 Python Implementation Benchmark${NC}"
python_avg=$(run_benchmark "Python mcp-tool-enforcer" "python3 mcp-tool-enforcer.py")

echo -e "\n${BLUE}🚀 Go Implementation Benchmark${NC}"
go_avg=$(run_benchmark "Go mcp-tool-enforcer" "./mcp-tool-enforcer")

# Calculate improvement
if [ $python_avg -gt 0 ] && [ $go_avg -gt 0 ]; then
    improvement_ratio=$(echo "scale=2; $python_avg / $go_avg" | bc -l 2>/dev/null || echo "N/A")
    improvement_percent=$(echo "scale=1; (($python_avg - $go_avg) * 100) / $python_avg" | bc -l 2>/dev/null || echo "N/A")
else
    improvement_ratio="N/A"
    improvement_percent="N/A"
fi

echo -e "\n${BLUE}📊 Performance Comparison${NC}"
echo "========================================="
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

# Expected target from TODO.md is 60-70% improvement
target_improvement=65
if [ "$improvement_percent" != "N/A" ]; then
    if (( $(echo "$improvement_percent > $target_improvement" | bc -l) )); then
        echo -e "\n${GREEN}🎯 Target achieved! Expected 60-70% improvement, got ${improvement_percent}%${NC}"
    else
        echo -e "\n${YELLOW}⚠️  Target status: Expected ${target_improvement}% improvement, got ${improvement_percent}%${NC}"
    fi
fi

# Memory usage comparison
echo -e "\n${BLUE}🧠 Memory Usage Analysis${NC}"
echo "========================================"

# Python memory usage
echo -e "Testing Python memory usage..."
python_memory=$(echo "$TEST_DATA" | timeout 5s /usr/bin/time -l python3 mcp-tool-enforcer.py 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")

# Go memory usage  
echo -e "Testing Go memory usage..."
go_memory=$(echo "$TEST_DATA" | timeout 5s /usr/bin/time -l ./mcp-tool-enforcer 2>&1 | grep "maximum resident set size" | awk '{print $1}' || echo "N/A")

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

# Test with different input sizes
echo -e "\n${BLUE}📏 Scale Testing${NC}"
echo "========================================"

# Large input test
LARGE_TEST_DATA='{
  "tool_name": "Write",
  "tool_input": {
    "file_path": "large.js",
    "content": "'
for i in {1..50}; do
    LARGE_TEST_DATA+="const data${i} = fetch(\"https://api${i}.example.com\");\n"
    LARGE_TEST_DATA+="const db${i} = new sqlite3.Database(\"test${i}.db\");\n"
    LARGE_TEST_DATA+="localStorage.setItem(\"key${i}\", \"value${i}\");\n"
done
LARGE_TEST_DATA+='"
  }
}'

echo "Testing with large input (150 patterns to match)..."

# Test Python with large input
echo -e "Python large input test:"
large_python_total=0
for i in $(seq 1 5); do
    start=$(date +%s%N)
    echo "$LARGE_TEST_DATA" | python3 mcp-tool-enforcer.py > /dev/null 2>&1 || true
    end=$(date +%s%N)
    duration=$(((end - start) / 1000000))
    echo "  Run $i: ${duration}ms"
    large_python_total=$((large_python_total + duration))
done
large_python_avg=$((large_python_total / 5))

# Test Go with large input
echo -e "Go large input test:"
large_go_total=0
for i in $(seq 1 5); do
    start=$(date +%s%N)
    echo "$LARGE_TEST_DATA" | ./mcp-tool-enforcer > /dev/null 2>&1 || true
    end=$(date +%s%N)
    duration=$(((end - start) / 1000000))
    echo "  Run $i: ${duration}ms"
    large_go_total=$((large_go_total + duration))
done
large_go_avg=$((large_go_total / 5))

echo -e "\nLarge input results:"
echo "Python average: ${large_python_avg}ms"
echo "Go average: ${large_go_avg}ms"

if [ $large_python_avg -gt 0 ] && [ $large_go_avg -gt 0 ]; then
    large_improvement=$(echo "scale=1; (($large_python_avg - $large_go_avg) * 100) / $large_python_avg" | bc -l)
    echo -e "Scale improvement: ${GREEN}${large_improvement}%${NC}"
fi

echo -e "\n${GREEN}✅ Benchmark completed successfully!${NC}"

# Summary
echo -e "\n${BLUE}📋 Summary${NC}"
echo "========================================"
echo "• Python average: ${python_avg}ms"
echo "• Go average: ${go_avg}ms"
echo "• Performance improvement: ${improvement_percent}%"
echo "• Memory improvement: Available in memory analysis above"
echo "• Target (60-70%): $([ "$improvement_percent" != "N/A" ] && (( $(echo "$improvement_percent > 60" | bc -l) )) && echo "✅ ACHIEVED" || echo "⚠️ Check needed")"
echo "• Scale performance: ${large_improvement}% improvement on large inputs"