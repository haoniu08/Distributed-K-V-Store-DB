#!/bin/bash

# Test script for Phase 2: Leader-Follower Database
# This script tests the basic functionality of the Leader-Follower implementation

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Phase 2: Leader-Follower Database Test ===${NC}\n"

# Build the binary
echo "Building leader-follower binary..."
go build -o leader-follower ./cmd/leader-follower
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful!${NC}\n"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    pkill -f "leader-follower" || true
    sleep 1
}

trap cleanup EXIT

# Start Leader
echo "Starting Leader node on port 8080..."
./leader-follower \
    --node-id=leader1 \
    --role=leader \
    --leader-addr=localhost:8080 \
    --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
    --port=8080 > /tmp/leader.log 2>&1 &
LEADER_PID=$!
sleep 2

# Check if leader started
if ! kill -0 $LEADER_PID 2>/dev/null; then
    echo -e "${RED}Leader failed to start!${NC}"
    cat /tmp/leader.log
    exit 1
fi
echo -e "${GREEN}Leader started (PID: $LEADER_PID)${NC}"

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

# Wait for all nodes to be ready
echo -e "\n${YELLOW}Waiting for nodes to be ready...${NC}"
sleep 3

# Test 1: Health checks
echo -e "\n${YELLOW}Test 1: Health Checks${NC}"
echo "Checking Leader health..."
if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo -e "${GREEN}✓ Leader is healthy${NC}"
else
    echo -e "${RED}✗ Leader health check failed${NC}"
    exit 1
fi

echo "Checking Follower 1 health..."
if curl -s http://localhost:8081/health | grep -q "healthy"; then
    echo -e "${GREEN}✓ Follower 1 is healthy${NC}"
else
    echo -e "${RED}✗ Follower 1 health check failed${NC}"
    exit 1
fi

# Test 2: Configuration
echo -e "\n${YELLOW}Test 2: Configuration${NC}"
echo "Getting current configuration..."
CONFIG=$(curl -s http://localhost:8080/config)
echo "$CONFIG" | jq '.' || echo "$CONFIG"

# Test 3: Strategy 1 (W=5, R=1)
echo -e "\n${YELLOW}Test 3: Strategy 1 (W=5, R=1)${NC}"
echo "Setting W=5, R=1..."
curl -s -X POST http://localhost:8080/config \
    -H "Content-Type: application/json" \
    -d '{"r":1,"w":5}' > /dev/null

echo "Writing key 'test1' with value 'value1' to Leader..."
WRITE_RESPONSE=$(curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test1","value":"value1"}')
echo "$WRITE_RESPONSE" | jq '.' || echo "$WRITE_RESPONSE"

# Wait for replication (W=5 means all nodes must be updated)
echo "Waiting for replication to complete..."
sleep 2

echo "Reading from Leader..."
READ_LEADER=$(curl -s "http://localhost:8080/get?key=test1")
echo "$READ_LEADER" | jq '.' || echo "$READ_LEADER"

if echo "$READ_LEADER" | grep -q "value1"; then
    echo -e "${GREEN}✓ Read from Leader successful${NC}"
else
    echo -e "${RED}✗ Read from Leader failed${NC}"
    exit 1
fi

echo "Reading from Follower 1..."
READ_FOLLOWER=$(curl -s "http://localhost:8081/get?key=test1")
echo "$READ_FOLLOWER" | jq '.' || echo "$READ_FOLLOWER"

if echo "$READ_FOLLOWER" | grep -q "value1"; then
    echo -e "${GREEN}✓ Read from Follower successful${NC}"
else
    echo -e "${RED}✗ Read from Follower failed${NC}"
    exit 1
fi

# Test 4: Strategy 2 (W=1, R=5)
echo -e "\n${YELLOW}Test 4: Strategy 2 (W=1, R=5)${NC}"
echo "Setting W=1, R=5..."
curl -s -X POST http://localhost:8080/config \
    -H "Content-Type: application/json" \
    -d '{"r":5,"w":1}' > /dev/null

echo "Writing key 'test2' with value 'value2' to Leader..."
WRITE_RESPONSE=$(curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test2","value":"value2"}')
echo "$WRITE_RESPONSE" | jq '.' || echo "$WRITE_RESPONSE"

# W=1 means Leader responds immediately, but replication happens async
echo "Waiting a bit for async replication..."
sleep 1

echo "Reading from Leader (should read from all 5 nodes)..."
READ_RESPONSE=$(curl -s "http://localhost:8080/get?key=test2")
echo "$READ_RESPONSE" | jq '.' || echo "$READ_RESPONSE"

if echo "$READ_RESPONSE" | grep -q "value2"; then
    echo -e "${GREEN}✓ Read with R=5 successful${NC}"
else
    echo -e "${RED}✗ Read with R=5 failed${NC}"
    exit 1
fi

# Test 5: Strategy 3 (R=3, W=3)
echo -e "\n${YELLOW}Test 5: Strategy 3 (R=3, W=3)${NC}"
echo "Setting R=3, W=3..."
curl -s -X POST http://localhost:8080/config \
    -H "Content-Type: application/json" \
    -d '{"r":3,"w":3}' > /dev/null

echo "Writing key 'test3' with value 'value3' to Leader..."
WRITE_RESPONSE=$(curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test3","value":"value3"}')
echo "$WRITE_RESPONSE" | jq '.' || echo "$WRITE_RESPONSE"

# Wait for quorum replication
echo "Waiting for quorum replication..."
sleep 1

echo "Reading from Leader (should read from 3 nodes)..."
READ_RESPONSE=$(curl -s "http://localhost:8080/get?key=test3")
echo "$READ_RESPONSE" | jq '.' || echo "$READ_RESPONSE"

if echo "$READ_RESPONSE" | grep -q "value3"; then
    echo -e "${GREEN}✓ Read with R=3 successful${NC}"
else
    echo -e "${RED}✗ Read with R=3 failed${NC}"
    exit 1
fi

# Test 6: Write to follower should fail
echo -e "\n${YELLOW}Test 6: Write to Follower should be rejected${NC}"
WRITE_TO_FOLLOWER=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8081/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test4","value":"value4"}')
HTTP_CODE=$(echo "$WRITE_TO_FOLLOWER" | tail -n1)
if [ "$HTTP_CODE" = "403" ]; then
    echo -e "${GREEN}✓ Write to Follower correctly rejected (403)${NC}"
else
    echo -e "${RED}✗ Write to Follower should be rejected but got: $HTTP_CODE${NC}"
    exit 1
fi

# Test 7: Local read for testing inconsistency
echo -e "\n${YELLOW}Test 7: Local Read (for testing inconsistency)${NC}"
echo "Reading locally from Follower 1..."
LOCAL_READ=$(curl -s "http://localhost:8081/local_read?key=test1")
echo "$LOCAL_READ" | jq '.' || echo "$LOCAL_READ"

if echo "$LOCAL_READ" | grep -q "value1"; then
    echo -e "${GREEN}✓ Local read successful${NC}"
else
    echo -e "${RED}✗ Local read failed${NC}"
    exit 1
fi

# Test 8: Multiple writes and version tracking
echo -e "\n${YELLOW}Test 8: Version Tracking${NC}"
curl -s -X POST http://localhost:8080/config \
    -H "Content-Type: application/json" \
    -d '{"r":1,"w":5}' > /dev/null

echo "Writing 'test5' with value 'v1'..."
curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test5","value":"v1"}' > /dev/null
sleep 2

echo "Writing 'test5' with value 'v2'..."
curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test5","value":"v2"}' > /dev/null
sleep 2

READ_V1=$(curl -s "http://localhost:8080/get?key=test5")
VERSION=$(echo "$READ_V1" | jq -r '.version' 2>/dev/null || echo "unknown")
echo "Current version: $VERSION"
if echo "$READ_V1" | grep -q "v2"; then
    echo -e "${GREEN}✓ Version tracking works (should be v2)${NC}"
else
    echo -e "${RED}✗ Version tracking failed${NC}"
    exit 1
fi

echo -e "\n${GREEN}=== All Tests Passed! ===${NC}"

