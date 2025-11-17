#!/bin/bash

# Manual test - run this and test manually

echo "=== Manual Test for Phase 2 ==="
echo ""
echo "This script will start the Leader and Followers."
echo "You can then test manually using curl commands."
echo ""
echo "Press Ctrl+C to stop all nodes."
echo ""

# Cleanup
cleanup() {
    echo ""
    echo "Stopping all nodes..."
    pkill -f "leader-follower" || true
    exit 0
}
trap cleanup INT TERM

# Build
go build -o leader-follower ./cmd/leader-follower

# Start Leader
echo "Starting Leader on port 8080..."
./leader-follower \
    --node-id=leader1 \
    --role=leader \
    --leader-addr=localhost:8080 \
    --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
    --port=8080 &

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
        --port=$port &
    sleep 1
done

echo ""
echo "All nodes started!"
echo ""
echo "Test commands:"
echo "1. Health check: curl http://localhost:8080/health"
echo "2. Get config: curl http://localhost:8080/config"
echo "3. Set W=5, R=1: curl -X POST http://localhost:8080/config -H 'Content-Type: application/json' -d '{\"r\":1,\"w\":5}'"
echo "4. Write: curl -X POST http://localhost:8080/set -H 'Content-Type: application/json' -d '{\"key\":\"test\",\"value\":\"hello\"}'"
echo "5. Read from Leader: curl 'http://localhost:8080/get?key=test'"
echo "6. Read from Follower: curl 'http://localhost:8081/get?key=test'"
echo ""
echo "Waiting... (Press Ctrl+C to stop)"

# Wait
while true; do
    sleep 1
done

