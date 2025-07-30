#!/bin/bash
# Quick benchmark comparison of all three implementations

echo "🏃 Quick Performance Comparison"
echo "================================"

TEST_DATA='{"tool_name": "Bash", "tool_input": {"command": "git commit -m \"Test commit\""}}'
ITERATIONS=5

cd "$(dirname "$0")"

echo -e "\n🐍 Python Implementation:"
total=0
for i in $(seq 1 $ITERATIONS); do
    start=$(date +%s%N)
    echo "$TEST_DATA" | python3 git-comprehensive-validator.py > /dev/null 2>&1
    end=$(date +%s%N)
    duration=$(((end - start) / 1000000))
    echo "  Run $i: ${duration}ms"
    total=$((total + duration))
done
python_avg=$((total / ITERATIONS))
echo "  Average: ${python_avg}ms"

echo -e "\n🚀 Go Implementation (go-git):"
total=0
for i in $(seq 1 $ITERATIONS); do
    start=$(date +%s%N)
    echo "$TEST_DATA" | ./git-validator > /dev/null 2>&1
    end=$(date +%s%N)
    duration=$(((end - start) / 1000000))
    echo "  Run $i: ${duration}ms"
    total=$((total + duration))
done
go_avg=$((total / ITERATIONS))
echo "  Average: ${go_avg}ms"

echo -e "\n⚡ Go Implementation (optimized):"
total=0
for i in $(seq 1 $ITERATIONS); do
    start=$(date +%s%N)
    echo "$TEST_DATA" | ./git-validator-optimized > /dev/null 2>&1
    end=$(date +%s%N)
    duration=$(((end - start) / 1000000))
    echo "  Run $i: ${duration}ms"
    total=$((total + duration))
done
go_opt_avg=$((total / ITERATIONS))
echo "  Average: ${go_opt_avg}ms"

echo -e "\n📊 Summary:"
echo "================================"
echo "Python:           ${python_avg}ms"
echo "Go (go-git):      ${go_avg}ms"
echo "Go (optimized):   ${go_opt_avg}ms"

if [ $go_opt_avg -lt $python_avg ]; then
    improvement=$(echo "scale=1; (($python_avg - $go_opt_avg) * 100) / $python_avg" | bc -l 2>/dev/null || echo "N/A")
    speedup=$(echo "scale=1; $python_avg / $go_opt_avg" | bc -l 2>/dev/null || echo "N/A")
    echo "🎯 Optimized Go is ${improvement}% faster (${speedup}x speedup)"
else
    echo "❌ Optimized Go is slower than Python"
fi