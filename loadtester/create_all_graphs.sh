#!/bin/bash

# Script to create graphs for all existing load test results
# This script finds all results directories and generates graphs for them

set -e

echo "=== Creating Graphs for All Load Test Results ==="
echo ""

# Find all directories containing results.csv
RESULTS_DIRS=$(find results -name "results.csv" -type f 2>/dev/null | sed 's|/results.csv||' | sort)

if [ -z "$RESULTS_DIRS" ]; then
    echo "No load test results found!"
    echo ""
    echo "To generate graphs, you need to run load tests first:"
    echo "  1. Start your database nodes"
    echo "  2. Run: ./loadtester-bin --config=loadtester/configs/lf_w5_r1_01_99.json --duration=30s --output=results/test"
    echo "  3. Then run this script again"
    exit 1
fi

echo "Found $(echo "$RESULTS_DIRS" | wc -l | tr -d ' ') result directories"
echo ""

# Check if Python dependencies are available
if ! python3 -c "import matplotlib; import numpy" 2>/dev/null; then
    echo "Error: Python dependencies not found"
    echo "Please install: pip3 install matplotlib numpy"
    exit 1
fi

# Create graphs for each results directory
COUNT=0
for results_dir in $RESULTS_DIRS; do
    COUNT=$((COUNT + 1))
    echo "[$COUNT] Processing: $results_dir"
    
    if python3 loadtester/visualize.py "$results_dir" 2>&1; then
        echo "   ✓ Graphs created"
    else
        echo "   ✗ Failed to create graphs"
    fi
    echo ""
done

echo "=== Complete ==="
echo "Created graphs for $COUNT result directories"
echo ""
echo "Graphs are located in:"
for results_dir in $RESULTS_DIRS; do
    if [ -f "$results_dir/latency_distribution.png" ]; then
        echo "  - $results_dir/latency_distribution.png"
        echo "  - $results_dir/time_intervals.png"
    fi
done





