#!/bin/bash
# Load testing script for Leaderless mode (Docker Compose)

set -e

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Change to project root
cd "$PROJECT_ROOT"

DURATION=${1:-60s}
CONCURRENCY=${2:-10}
OUTPUT_BASE=${3:-results/ll_local}

echo "=== Leaderless Load Testing ==="
echo "Duration: $DURATION"
echo "Concurrency: $CONCURRENCY"
echo "Output base: $OUTPUT_BASE"
echo ""

# Check if leaderless nodes are running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ Leaderless nodes are not running!"
    echo "Start them with: docker compose -f docker-compose-leaderless.yml up -d"
    exit 1
fi

echo "✅ Leaderless nodes are running"
echo ""

# Test configurations
declare -a configs=(
    "ll_01_99:1% writes, 99% reads"
    "ll_10_90:10% writes, 90% reads"
    "ll_50_50:50% writes, 50% reads"
    "ll_90_10:90% writes, 10% reads"
)

# Run each test
for config_info in "${configs[@]}"; do
    IFS=':' read -r config_name config_desc <<< "$config_info"
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Running: $config_name ($config_desc)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    output_dir="${OUTPUT_BASE}_${config_name}"
    
    # Run load test
    go run ./loadtester/main.go \
        -config "loadtester/configs/${config_name}.json" \
        -duration "$DURATION" \
        -concurrency "$CONCURRENCY" \
        -output "$output_dir"
    
    if [ $? -eq 0 ]; then
        echo "✅ Completed: $config_name"
        echo "   Results: $output_dir"
    else
        echo "❌ Failed: $config_name"
    fi
    
    echo ""
    sleep 2  # Brief pause between tests
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "=== Generating Graphs and Reports ==="
echo ""

# Generate graphs and reports for each test
for config_info in "${configs[@]}"; do
    IFS=':' read -r config_name config_desc <<< "$config_info"
    output_dir="${OUTPUT_BASE}_${config_name}"
    
    if [ -d "$output_dir" ]; then
        echo "Processing: $config_name"
        echo "  Generating graphs..."
        python3 ./loadtester/visualize.py "$output_dir"
        echo "  Generating report..."
        python3 ./loadtester/generate_report.py "$output_dir"
        echo ""
    fi
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "=== Load Testing Complete ==="
echo "Results saved to: ${OUTPUT_BASE}_*"
echo "Graphs and reports generated in each results directory"

