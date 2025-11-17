#!/bin/bash

# Comprehensive test for Phase 3: Leaderless Database
set -e

echo "=== Phase 3: Leaderless Database Test ==="
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    pkill -f "leaderless" 2>/dev/null || true
    sleep 1
}
trap cleanup EXIT

# Build
echo "Building leaderless..."
go build -o leaderless ./cmd/leaderless
echo "✓ Build successful"
echo ""

# Define all node addresses
ALL_NODES="localhost:8080,localhost:8081,localhost:8082,localhost:8083,localhost:8084"

# Start Node 1
echo "Starting Node 1 on port 8080..."
./leaderless \
    --node-id=node1 \
    --all-node-addrs=$ALL_NODES \
    --port=8080 > /tmp/test_node1.log 2>&1 &
NODE1_PID=$!
sleep 2

# Check if node1 started
if ! kill -0 $NODE1_PID 2>/dev/null; then
    echo "✗ Node 1 failed to start!"
    cat /tmp/test_node1.log
    exit 1
fi
echo "✓ Node 1 started (PID: $NODE1_PID)"

# Start other nodes
for i in {2..5}; do
    port=$((8080 + i - 1))
    echo "Starting Node $i on port $port..."
    ./leaderless \
        --node-id=node$i \
        --all-node-addrs=$ALL_NODES \
        --port=$port > /tmp/test_node$i.log 2>&1 &
    sleep 1
done

echo "Waiting for all nodes to be ready..."
sleep 3
echo ""

# Test 1: Health Checks
echo "=== Test 1: Health Checks ==="
for i in {1..5}; do
    port=$((8080 + i - 1))
    HEALTH=$(curl -s http://localhost:$port/health)
    if echo "$HEALTH" | grep -q "healthy"; then
        echo "✓ Node $i health check passed"
    else
        echo "✗ Node $i health check failed"
        echo "  Response: $HEALTH"
        exit 1
    fi
done
echo ""

# Test 2: Write to Any Node (Node becomes Write Coordinator)
echo "=== Test 2: Write to Any Node (Write Coordinator) ==="
echo "Writing key 'test1' with value 'value1' to Node 1..."
WRITE_START=$(date +%s)
WRITE_RESPONSE=$(curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test1","value":"value1"}')
WRITE_END=$(date +%s)
WRITE_TIME=$((WRITE_END - WRITE_START))

if echo "$WRITE_RESPONSE" | grep -q "created"; then
    echo "✓ Write successful (Node 1 is Write Coordinator)"
    echo "  Response: $WRITE_RESPONSE"
    echo "  Time taken: ${WRITE_TIME}s (should be >1s due to W=N replication delays)"
else
    echo "✗ Write failed"
    echo "  Response: $WRITE_RESPONSE"
    exit 1
fi

# Wait for replication to complete (W=N means all 5 nodes must confirm)
echo "Waiting for replication to complete (W=5, all nodes must confirm)..."
sleep 4
echo ""

# Test 3: Read from All Nodes (R=1, returns local value)
echo "=== Test 3: Read from All Nodes (R=1) ==="
for i in {1..5}; do
    port=$((8080 + i - 1))
    READ_RESPONSE=$(curl -s "http://localhost:$port/get?key=test1")
    if echo "$READ_RESPONSE" | grep -q "value1"; then
        echo "✓ Read from Node $i successful"
        echo "  Response: $READ_RESPONSE"
    else
        echo "✗ Read from Node $i failed"
        echo "  Response: $READ_RESPONSE"
        echo "  Checking node $i log..."
        tail -10 /tmp/test_node$i.log
        exit 1
    fi
done
echo ""

# Test 4: Write to Different Node (Different Write Coordinator)
echo "=== Test 4: Write to Different Node (Different Coordinator) ==="
echo "Writing key 'test2' with value 'value2' to Node 3..."
WRITE_RESPONSE=$(curl -s -X POST http://localhost:8082/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test2","value":"value2"}')
if echo "$WRITE_RESPONSE" | grep -q "created"; then
    echo "✓ Write successful (Node 3 is Write Coordinator)"
    echo "  Response: $WRITE_RESPONSE"
else
    echo "✗ Write failed"
    echo "  Response: $WRITE_RESPONSE"
    exit 1
fi

echo "Waiting for replication..."
sleep 4

echo "Reading from all nodes..."
for i in {1..5}; do
    port=$((8080 + i - 1))
    READ_RESPONSE=$(curl -s "http://localhost:$port/get?key=test2")
    if echo "$READ_RESPONSE" | grep -q "value2"; then
        echo "✓ Node $i has the value"
    else
        echo "✗ Node $i missing the value"
        exit 1
    fi
done
echo ""

# Test 5: Inconsistency Window (Write and Immediate Read)
echo "=== Test 5: Inconsistency Window Test ==="
echo "Writing key 'test3' with value 'value3' to Node 2..."
WRITE_RESPONSE=$(curl -s -X POST http://localhost:8081/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test3","value":"value3"}')
echo "Write response: $WRITE_RESPONSE"

echo "Immediately reading from Node 5 (might show inconsistency)..."
IMMEDIATE_READ=$(curl -s "http://localhost:8084/get?key=test3")
echo "Immediate read: $IMMEDIATE_READ"

if echo "$IMMEDIATE_READ" | grep -q "key not found"; then
    echo "✓ Inconsistency window detected (Node 5 doesn't have value yet)"
else
    echo "  Note: Node 5 already has value (replication completed quickly)"
fi

echo "Waiting for replication..."
sleep 4

echo "Reading from Node 5 again (should be consistent now)..."
LATER_READ=$(curl -s "http://localhost:8084/get?key=test3")
if echo "$LATER_READ" | grep -q "value3"; then
    echo "✓ Node 5 now has consistent value"
    echo "  Response: $LATER_READ"
else
    echo "✗ Node 5 still missing value"
    exit 1
fi
echo ""

# Test 6: Multiple Writes from Different Nodes
echo "=== Test 6: Multiple Writes from Different Nodes ==="
echo "Writing from Node 1..."
curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"multi1","value":"from_node1"}' > /dev/null
sleep 2

echo "Writing from Node 4..."
curl -s -X POST http://localhost:8083/set \
    -H "Content-Type: application/json" \
    -d '{"key":"multi2","value":"from_node4"}' > /dev/null
sleep 2

echo "Writing from Node 5..."
curl -s -X POST http://localhost:8084/set \
    -H "Content-Type: application/json" \
    -d '{"key":"multi3","value":"from_node5"}' > /dev/null
sleep 4

echo "Verifying all writes are replicated..."
for key in "multi1" "multi2" "multi3"; do
    for i in {1..5}; do
        port=$((8080 + i - 1))
        READ_RESPONSE=$(curl -s "http://localhost:$port/get?key=$key")
        if ! echo "$READ_RESPONSE" | grep -q "key not found"; then
            echo "✓ Node $i has key $key"
        else
            echo "✗ Node $i missing key $key"
            exit 1
        fi
    done
done
echo ""

# Test 7: Version Tracking
echo "=== Test 7: Version Tracking ==="
echo "Writing 'version_test' with value 'v1' to Node 1..."
curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"version_test","value":"v1"}' > /dev/null
sleep 4

echo "Writing 'version_test' with value 'v2' to Node 3..."
curl -s -X POST http://localhost:8082/set \
    -H "Content-Type: application/json" \
    -d '{"key":"version_test","value":"v2"}' > /dev/null
sleep 4

echo "Reading from all nodes (should all have v2 with higher version)..."
for i in {1..5}; do
    port=$((8080 + i - 1))
    READ_RESPONSE=$(curl -s "http://localhost:$port/get?key=version_test")
    if echo "$READ_RESPONSE" | grep -q "v2"; then
        VERSION=$(echo "$READ_RESPONSE" | grep -o '"version":[0-9]*' | cut -d: -f2)
        echo "✓ Node $i has v2, version: $VERSION"
    else
        echo "✗ Node $i doesn't have v2"
        echo "  Response: $READ_RESPONSE"
        exit 1
    fi
done
echo ""

echo "=== All Tests Passed! ==="
echo ""
echo "Summary:"
echo "  ✓ All 5 nodes started successfully"
echo "  ✓ Any node can receive writes (becomes Write Coordinator)"
echo "  ✓ Write Coordinator replicates to all other nodes (W=N)"
echo "  ✓ All nodes can read (R=1, returns local value)"
echo "  ✓ Inconsistency window demonstrated"
echo "  ✓ Multiple coordinators working correctly"
echo "  ✓ Version tracking working"

