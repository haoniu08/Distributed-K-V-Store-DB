#!/bin/bash

# Complete load testing workflow
# This script helps you run all required load tests and generate graphs

set -e

echo "=== Complete Load Testing Workflow ==="
echo ""
echo "This script will guide you through running all load tests required by the homework."
echo ""

# Check prerequisites
echo "Checking prerequisites..."

# Check if nodes are running
LEADER_RUNNING=$(curl -s http://localhost:8080/health 2>/dev/null | grep -q "healthy" && echo "yes" || echo "no")

if [ "$LEADER_RUNNING" != "yes" ]; then
    echo ""
    echo "⚠️  Database nodes are not running!"
    echo ""
    echo "Please start your database nodes first."
    echo ""
    echo "For Leader-Follower tests:"
    echo "  1. Start Leader on port 8080"
    echo "  2. Start 4 Followers on ports 8081-8084"
    echo "  3. Configure R/W values:"
    echo "     curl -X POST http://localhost:8080/config -H 'Content-Type: application/json' -d '{\"r\":1,\"w\":5}'"
    echo ""
    echo "For Leaderless tests:"
    echo "  1. Start 5 nodes on ports 8080-8084"
    echo ""
    echo "Then run this script again."
    exit 1
fi

echo "✓ Nodes are running"
echo ""

# Check Python dependencies
if ! python3 -c "import matplotlib; import numpy" 2>/dev/null; then
    echo "⚠️  Python dependencies missing"
    echo "Please install: pip3 install --user matplotlib numpy"
    exit 1
fi

echo "✓ Python dependencies available"
echo ""

# Build load tester
if [ ! -f "./loadtester-bin" ]; then
    echo "Building load tester..."
    go build -o loadtester-bin ./loadtester
    echo "✓ Build complete"
    echo ""
fi

# Configuration
DURATION=${1:-60s}
CONCURRENCY=${2:-10}
OUTPUT_BASE=${3:-results}

echo "Test Configuration:"
echo "  Duration: $DURATION"
echo "  Concurrency: $CONCURRENCY workers"
echo "  Output base: $OUTPUT_BASE"
echo ""

read -p "Continue with load testing? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 1
fi

echo ""
echo "=== Starting Load Tests ==="
echo ""

# Function to run a test
run_test() {
    local config_file=$1
    local config_name=$(basename $config_file .json)
    local output_dir="${OUTPUT_BASE}/${config_name}"
    
    echo "[$(date +%H:%M:%S)] Running: $config_name"
    
    ./loadtester-bin \
        --config="$config_file" \
        --duration="$DURATION" \
        --concurrency=$CONCURRENCY \
        --output="$output_dir" > /tmp/loadtest_${config_name}.log 2>&1
    
    if [ $? -eq 0 ]; then
        echo "  ✓ Completed: $config_name"
    else
        echo "  ✗ Failed: $config_name (check /tmp/loadtest_${config_name}.log)"
    fi
}

# Run all tests
echo "Running Leader-Follower W=5 R=1 tests..."
# Note: Make sure R/W is configured correctly before running
run_test "loadtester/configs/lf_w5_r1_01_99.json"
run_test "loadtester/configs/lf_w5_r1_10_90.json"
run_test "loadtester/configs/lf_w5_r1_50_50.json"
run_test "loadtester/configs/lf_w5_r1_90_10.json"

echo ""
echo "Running Leader-Follower W=1 R=5 tests..."
# Configure R=5, W=1 before running
run_test "loadtester/configs/lf_w1_r5_01_99.json"
run_test "loadtester/configs/lf_w1_r5_10_90.json"
run_test "loadtester/configs/lf_w1_r5_50_50.json"
run_test "loadtester/configs/lf_w1_r5_90_10.json"

echo ""
echo "Running Leader-Follower W=3 R=3 tests..."
# Configure R=3, W=3 before running
run_test "loadtester/configs/lf_w3_r3_01_99.json"
run_test "loadtester/configs/lf_w3_r3_10_90.json"
run_test "loadtester/configs/lf_w3_r3_50_50.json"
run_test "loadtester/configs/lf_w3_r3_90_10.json"

echo ""
echo "Running Leaderless tests..."
run_test "loadtester/configs/ll_01_99.json"
run_test "loadtester/configs/ll_10_90.json"
run_test "loadtester/configs/ll_50_50.json"
run_test "loadtester/configs/ll_90_10.json"

echo ""
echo "=== Generating Graphs ==="
echo ""

# Generate graphs for all results
for results_dir in ${OUTPUT_BASE}/*/; do
    if [ -f "${results_dir}results.csv" ]; then
        config_name=$(basename "$results_dir")
        echo "Generating graphs for: $config_name"
        python3 loadtester/visualize.py "$results_dir" 2>&1 | grep -E "(Saved|Error)" || true
    fi
done

echo ""
echo "=== Load Testing Complete ==="
echo ""
echo "Results saved to: $OUTPUT_BASE/"
echo "Graphs generated for all test runs"
echo ""
echo "Next steps:"
echo "  1. Review summary.json files in each results directory"
echo "  2. Examine graphs (latency_distribution.png, time_intervals.png)"
echo "  3. Fill in LOAD_TEST_ANALYSIS_TEMPLATE.md with your findings"
echo ""





