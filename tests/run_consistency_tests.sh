#!/bin/bash

# Script to run consistency tests
# This script starts the necessary nodes and runs the consistency tests

set -e

echo "=== Consistency Tests Runner ==="
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    pkill -f "leader-follower" 2>/dev/null || true
    pkill -f "leaderless" 2>/dev/null || true
    sleep 1
}
trap cleanup EXIT

# Build binaries
echo "Building binaries..."
go build -o leader-follower ./cmd/leader-follower
go build -o leaderless ./cmd/leaderless
echo "✓ Build successful"
echo ""

# Function to start Leader-Follower cluster
start_leader_follower() {
    echo "Starting Leader-Follower cluster..."
    
    # Start Leader
    ./leader-follower \
        --node-id=leader1 \
        --role=leader \
        --leader-addr=localhost:8080 \
        --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
        --port=8080 > /tmp/lf_leader.log 2>&1 &
    sleep 2

    # Start Followers
    for i in {1..4}; do
        port=$((8080 + i))
        ./leader-follower \
            --node-id=follower$i \
            --role=follower \
            --leader-addr=localhost:8080 \
            --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
            --port=$port > /tmp/lf_follower$i.log 2>&1 &
        sleep 1
    done

    echo "Waiting for Leader-Follower cluster to be ready..."
    sleep 3

    # Set W=5, R=1
    curl -s -X POST http://localhost:8080/config \
        -H "Content-Type: application/json" \
        -d '{"r":1,"w":5}' > /dev/null

    echo "✓ Leader-Follower cluster ready"
}

# Function to start Leaderless cluster
start_leaderless() {
    echo "Starting Leaderless cluster..."
    
    ALL_NODES="localhost:8080,localhost:8081,localhost:8082,localhost:8083,localhost:8084"

    # Start all nodes
    for i in {1..5}; do
        port=$((8080 + i - 1))
        ./leaderless \
            --node-id=node$i \
            --all-node-addrs=$ALL_NODES \
            --port=$port > /tmp/ll_node$i.log 2>&1 &
        sleep 1
    done

    echo "Waiting for Leaderless cluster to be ready..."
    sleep 3

    echo "✓ Leaderless cluster ready"
}

# Check which tests to run
TEST_TYPE=${1:-"all"}

if [ "$TEST_TYPE" = "leader-follower" ] || [ "$TEST_TYPE" = "all" ]; then
    echo ""
    echo "=== Running Leader-Follower Consistency Tests ==="
    
    # Clean up any existing processes
    pkill -f "leader-follower" 2>/dev/null || true
    pkill -f "leaderless" 2>/dev/null || true
    sleep 1
    
    start_leader_follower
    
    echo ""
    echo "Running tests..."
    go test -v ./tests -run TestLeaderFollowerConsistency
    
    echo ""
    echo "Running high load tests..."
    go test -v ./tests -run TestLeaderFollowerConsistencyHighLoad
    
    # Cleanup
    pkill -f "leader-follower" 2>/dev/null || true
    sleep 2
fi

if [ "$TEST_TYPE" = "leaderless" ] || [ "$TEST_TYPE" = "all" ]; then
    echo ""
    echo "=== Running Leaderless Consistency Tests ==="
    
    # Clean up any existing processes
    pkill -f "leader-follower" 2>/dev/null || true
    pkill -f "leaderless" 2>/dev/null || true
    sleep 1
    
    start_leaderless
    
    echo ""
    echo "Running tests..."
    go test -v ./tests -run TestLeaderlessConsistency
    
    echo ""
    echo "Running high load tests..."
    go test -v ./tests -run TestLeaderlessConsistencyHighLoad
    
    # Cleanup
    pkill -f "leaderless" 2>/dev/null || true
    sleep 2
fi

echo ""
echo "=== All Consistency Tests Complete ==="





