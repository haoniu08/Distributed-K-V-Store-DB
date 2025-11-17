#!/bin/bash

# Script to run a quick load test and generate graphs
# This demonstrates the full workflow

set -e

echo "=== Load Test and Graph Generation ==="
echo ""

# Check if nodes are running
echo "Checking if database nodes are running..."
LEADER_RUNNING=$(curl -s http://localhost:8080/health 2>/dev/null | grep -q "healthy" && echo "yes" || echo "no")

if [ "$LEADER_RUNNING" != "yes" ]; then
    echo ""
    echo "⚠️  Database nodes are not running!"
    echo ""
    echo "Please start your database nodes first:"
    echo ""
    echo "For Leader-Follower:"
    echo "  Terminal 1: ./leader-follower --node-id=leader1 --role=leader --leader-addr=localhost:8080 --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 --port=8080"
    echo "  Terminal 2-5: Start followers on ports 8081-8084"
    echo ""
    echo "For Leaderless:"
    echo "  Start 5 nodes on ports 8080-8084"
    echo ""
    echo "Then run this script again."
    exit 1
fi

echo "✓ Nodes are running"
echo ""

# Check Python dependencies
echo "Checking Python dependencies..."
if ! python3 -c "import matplotlib; import numpy" 2>/dev/null; then
    echo "⚠️  Python dependencies missing"
    echo "Installing matplotlib and numpy..."
    pip3 install matplotlib numpy 2>/dev/null || {
        echo "Failed to install. Please run: pip3 install matplotlib numpy"
        exit 1
    }
fi
echo "✓ Python dependencies available"
echo ""

# Build load tester if needed
if [ ! -f "./loadtester-bin" ]; then
    echo "Building load tester..."
    go build -o loadtester-bin ./loadtester
    echo "✓ Build complete"
    echo ""
fi

# Create results directory
mkdir -p results/demo

# Run a quick load test
echo "Running load test (30 seconds)..."
echo "Configuration: Leader-Follower W=5 R=1, 10% writes, 90% reads"
echo ""

./loadtester-bin \
    --config=loadtester/configs/lf_w5_r1_10_90.json \
    --duration=30s \
    --concurrency=5 \
    --output=results/demo 2>&1 | tee /tmp/loadtest.log

echo ""
echo "✓ Load test complete"
echo ""

# Check if results were generated
if [ ! -f "results/demo/results.csv" ]; then
    echo "✗ No results.csv generated. Check the log above for errors."
    exit 1
fi

# Generate graphs
echo "Generating graphs..."
if python3 loadtester/visualize.py results/demo 2>&1; then
    echo ""
    echo "✓ Graphs created successfully!"
    echo ""
    echo "Graphs saved to:"
    echo "  - results/demo/latency_distribution.png"
    echo "  - results/demo/time_intervals.png"
    echo ""
    
    # Try to open the graphs (macOS)
    if command -v open &> /dev/null; then
        echo "Opening graphs..."
        open results/demo/latency_distribution.png 2>/dev/null || true
        open results/demo/time_intervals.png 2>/dev/null || true
    fi
    
    echo "=== Complete ==="
else
    echo "✗ Failed to generate graphs"
    exit 1
fi





