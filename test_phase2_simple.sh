#!/bin/bash

# Simple test script for Phase 2
set -e

echo "=== Phase 2 Simple Test ==="

# Build
echo "Building..."
go build -o leader-follower ./cmd/leader-follower

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    pkill -f "leader-follower" || true
    sleep 1
}
trap cleanup EXIT

# Start Leader
echo "Starting Leader on port 8080..."
./leader-follower \
    --node-id=leader1 \
    --role=leader \
    --leader-addr=localhost:8080 \
    --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
    --port=8080 > /tmp/leader.log 2>&1 &
sleep 2

# Start Followers
for i in {1..4}; do
    port=$((8080 + i))
    echo "Starting Follower $i on port $port..."
    ./leader-follower \
        --node-id=follower$i \
        --role=follower \
        --leader-addr=localhost:8080 \
        --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
        --port=$port > /tmp/follower$i.log 2>&1 &
    sleep 1
done

echo "Waiting for nodes to start..."
sleep 3

# Test health
echo -e "\n=== Test 1: Health Check ==="
if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo "✓ Leader is healthy"
else
    echo "✗ Leader health check failed"
    cat /tmp/leader.log
    exit 1
fi

# Test config
echo -e "\n=== Test 2: Configuration ==="
echo "Current config:"
curl -s http://localhost:8080/config

# Test write (W=5, R=1)
echo -e "\n=== Test 3: Write (W=5, R=1) ==="
echo "Setting W=5, R=1..."
curl -s -X POST http://localhost:8080/config \
    -H "Content-Type: application/json" \
    -d '{"r":1,"w":5}' > /dev/null

echo "Writing test key..."
curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test","value":"hello"}' | head -c 200
echo ""

echo "Waiting for replication..."
sleep 3

echo "Reading from Leader:"
curl -s "http://localhost:8080/get?key=test" | head -c 200
echo ""

echo "Reading from Follower 1:"
curl -s "http://localhost:8081/get?key=test" | head -c 200
echo ""

echo -e "\n=== Test Complete ==="
echo "Check the logs in /tmp/leader.log and /tmp/follower*.log for details"

