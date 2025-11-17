#!/bin/bash

# Script to run all load tests
# This script runs load tests for all configurations and ratios

set -e

DURATION=${1:-60s}
CONCURRENCY=${2:-10}
OUTPUT_BASE=${3:-results}

echo "=== Load Testing All Configurations ==="
echo "Duration: $DURATION"
echo "Concurrency: $CONCURRENCY"
echo "Output base: $OUTPUT_BASE"
echo ""

# Build load tester
echo "Building load tester..."
go build -o loadtester-bin ./loadtester
echo "✓ Build successful"
echo ""

# Function to run a test
run_test() {
    local config_file=$1
    local config_name=$(basename $config_file .json)
    local output_dir="${OUTPUT_BASE}/${config_name}"
    
    echo "Running: $config_name"
    echo "  Config: $config_file"
    echo "  Output: $output_dir"
    
    ./loadtester-bin \
        --config="$config_file" \
        --duration="$DURATION" \
        --concurrency=$CONCURRENCY \
        --output="$output_dir"
    
    echo "✓ Completed: $config_name"
    echo ""
}

# Leader-Follower W=5 R=1
echo "=== Leader-Follower W=5 R=1 ==="
run_test "loadtester/configs/lf_w5_r1_01_99.json"
run_test "loadtester/configs/lf_w5_r1_10_90.json"
run_test "loadtester/configs/lf_w5_r1_50_50.json"
run_test "loadtester/configs/lf_w5_r1_90_10.json"

# Leader-Follower W=1 R=5
echo "=== Leader-Follower W=1 R=5 ==="
run_test "loadtester/configs/lf_w1_r5_01_99.json"
run_test "loadtester/configs/lf_w1_r5_10_90.json"
run_test "loadtester/configs/lf_w1_r5_50_50.json"
run_test "loadtester/configs/lf_w1_r5_90_10.json"

# Leader-Follower W=3 R=3
echo "=== Leader-Follower W=3 R=3 ==="
run_test "loadtester/configs/lf_w3_r3_01_99.json"
run_test "loadtester/configs/lf_w3_r3_10_90.json"
run_test "loadtester/configs/lf_w3_r3_50_50.json"
run_test "loadtester/configs/lf_w3_r3_90_10.json"

# Leaderless
echo "=== Leaderless W=5 R=1 ==="
run_test "loadtester/configs/ll_01_99.json"
run_test "loadtester/configs/ll_10_90.json"
run_test "loadtester/configs/ll_50_50.json"
run_test "loadtester/configs/ll_90_10.json"

echo "=== All Load Tests Complete ==="
echo "Results saved to: $OUTPUT_BASE/"





