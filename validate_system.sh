#!/bin/bash

# System validation script
set -e

echo "=== System Validation ==="
echo ""

# Check 1: Build all binaries
echo "1. Checking binary builds..."
if go build -o /tmp/test-leader-follower ./cmd/leader-follower 2>/dev/null && \
   go build -o /tmp/test-leaderless ./cmd/leaderless 2>/dev/null && \
   go build -o /tmp/test-loadtester ./loadtester 2>/dev/null; then
    echo "   ✓ All binaries build successfully"
    rm -f /tmp/test-leader-follower /tmp/test-leaderless /tmp/test-loadtester
else
    echo "   ✗ Build failed"
    exit 1
fi

# Check 2: Configuration files
echo ""
echo "2. Checking configuration files..."
CONFIG_COUNT=$(ls -1 loadtester/configs/*.json 2>/dev/null | wc -l | tr -d ' ')
if [ "$CONFIG_COUNT" -eq 16 ]; then
    echo "   ✓ All 16 configuration files present"
else
    echo "   ✗ Expected 16 config files, found $CONFIG_COUNT"
    exit 1
fi

# Check 3: Test files
echo ""
echo "3. Checking test files..."
if [ -f "tests/consistency_test.go" ] && \
   [ -f "tests/leader_follower_consistency_test.go" ] && \
   [ -f "tests/leaderless_consistency_test.go" ]; then
    echo "   ✓ Consistency test files present"
else
    echo "   ✗ Missing test files"
    exit 1
fi

# Check 4: Documentation
echo ""
echo "4. Checking documentation..."
if [ -f "README.md" ] && \
   [ -f "LEADER_FOLLOWER_README.md" ] && \
   [ -f "LEADERLESS_README.md" ] && \
   [ -f "loadtester/README.md" ] && \
   [ -f "tests/README.md" ]; then
    echo "   ✓ Documentation files present"
else
    echo "   ✗ Missing documentation files"
    exit 1
fi

# Check 5: Phase summaries
echo ""
echo "5. Checking phase summaries..."
if [ -f "phase_summaries/PHASE_1_2_SUMMARY.md" ] && \
   [ -f "phase_summaries/PHASE_3_SUMMARY.md" ] && \
   [ -f "phase_summaries/PHASE_4_SUMMARY.md" ] && \
   [ -f "phase_summaries/PHASE_5_SUMMARY.md" ]; then
    echo "   ✓ All phase summaries present"
else
    echo "   ✗ Missing phase summaries"
    exit 1
fi

# Check 6: Docker files
echo ""
echo "6. Checking Docker files..."
if [ -f "Dockerfile" ] && \
   [ -f "Dockerfile.leader-follower" ] && \
   [ -f "Dockerfile.leaderless" ]; then
    echo "   ✓ Docker files present"
else
    echo "   ✗ Missing Docker files"
    exit 1
fi

# Check 7: Python visualization
echo ""
echo "7. Checking Python visualization script..."
if [ -f "loadtester/visualize.py" ]; then
    if python3 -c "import matplotlib; import json; import csv" 2>/dev/null; then
        echo "   ✓ Visualization script and dependencies available"
    else
        echo "   ⚠ Visualization script present but dependencies may be missing"
    fi
else
    echo "   ✗ Visualization script missing"
    exit 1
fi

echo ""
echo "=== Validation Complete ==="
echo "✓ All systems validated successfully"





