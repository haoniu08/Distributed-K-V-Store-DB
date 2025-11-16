#!/bin/bash

# Comprehensive test for Phase 2
set -e

echo "=== Phase 2 Comprehensive Test ==="
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    pkill -f "leader-follower" 2>/dev/null || true
    sleep 1
}
trap cleanup EXIT

# Build
echo "Building leader-follower..."
go build -o leader-follower ./cmd/leader-follower
echo "✓ Build successful"
echo ""

# Start Leader
echo "Starting Leader on port 8080..."
./leader-follower \
    --node-id=leader1 \
    --role=leader \
    --leader-addr=localhost:8080 \
    --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
    --port=8080 > /tmp/test_leader.log 2>&1 &
LEADER_PID=$!
sleep 3

# Check if leader started
if ! kill -0 $LEADER_PID 2>/dev/null; then
    echo "✗ Leader failed to start!"
    cat /tmp/test_leader.log
    exit 1
fi
echo "✓ Leader started (PID: $LEADER_PID)"

# Start Followers
for i in {1..4}; do
    port=$((8080 + i))
    echo "Starting Follower $i on port $port..."
    ./leader-follower \
        --node-id=follower$i \
        --role=follower \
        --leader-addr=localhost:8080 \
        --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
        --port=$port > /tmp/test_follower$i.log 2>&1 &
    sleep 1
done

echo "Waiting for all nodes to be ready..."
sleep 3
echo ""

# Test 1: Health Checks
echo "=== Test 1: Health Checks ==="
LEADER_HEALTH=$(curl -s http://localhost:8080/health)
if echo "$LEADER_HEALTH" | grep -q "healthy"; then
    echo "✓ Leader health check passed"
    echo "  Response: $LEADER_HEALTH"
else
    echo "✗ Leader health check failed"
    echo "  Response: $LEADER_HEALTH"
    exit 1
fi

FOLLOWER_HEALTH=$(curl -s http://localhost:8081/health)
if echo "$FOLLOWER_HEALTH" | grep -q "healthy"; then
    echo "✓ Follower 1 health check passed"
else
    echo "✗ Follower 1 health check failed"
    echo "  Response: $FOLLOWER_HEALTH"
    exit 1
fi
echo ""

# Test 2: Configuration Endpoint
echo "=== Test 2: Configuration Endpoint ==="
CONFIG_RESPONSE=$(curl -s http://localhost:8080/config)
if echo "$CONFIG_RESPONSE" | grep -q "node_id"; then
    echo "✓ Config endpoint works"
    echo "  Response: $CONFIG_RESPONSE"
else
    echo "✗ Config endpoint failed (got 404?)"
    echo "  Response: $CONFIG_RESPONSE"
    echo "  Checking leader log..."
    tail -10 /tmp/test_leader.log
    exit 1
fi
echo ""

# Test 3: Set Configuration (W=5, R=1)
echo "=== Test 3: Set Configuration (W=5, R=1) ==="
SET_CONFIG=$(curl -s -X POST http://localhost:8080/config \
    -H "Content-Type: application/json" \
    -d '{"r":1,"w":5}')
if echo "$SET_CONFIG" | grep -q "configuration updated"; then
    echo "✓ Configuration set successfully"
    echo "  Response: $SET_CONFIG"
else
    echo "✗ Failed to set configuration"
    echo "  Response: $SET_CONFIG"
    exit 1
fi

# Verify config
VERIFY_CONFIG=$(curl -s http://localhost:8080/config)
if echo "$VERIFY_CONFIG" | grep -q '"w":5'; then
    echo "✓ Configuration verified (W=5, R=1)"
else
    echo "✗ Configuration not set correctly"
    echo "  Response: $VERIFY_CONFIG"
    exit 1
fi
echo ""

# Test 4: Write Operation (W=5, R=1)
echo "=== Test 4: Write Operation (W=5, R=1) ==="
echo "Writing key 'test1' with value 'value1'..."
WRITE_START=$(date +%s)
WRITE_RESPONSE=$(curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test1","value":"value1"}')
WRITE_END=$(date +%s)
WRITE_TIME=$((WRITE_END - WRITE_START))

if echo "$WRITE_RESPONSE" | grep -q "created"; then
    echo "✓ Write successful"
    echo "  Response: $WRITE_RESPONSE"
    echo "  Time taken: ${WRITE_TIME}s (should be >1s due to replication delays)"
else
    echo "✗ Write failed"
    echo "  Response: $WRITE_RESPONSE"
    exit 1
fi

# Wait for replication to complete
echo "Waiting for replication to complete (W=5 means all 5 nodes must confirm)..."
sleep 4
echo ""

# Test 5: Read from Leader
echo "=== Test 5: Read from Leader ==="
READ_LEADER=$(curl -s "http://localhost:8080/get?key=test1")
if echo "$READ_LEADER" | grep -q "value1"; then
    echo "✓ Read from Leader successful"
    echo "  Response: $READ_LEADER"
else
    echo "✗ Read from Leader failed"
    echo "  Response: $READ_LEADER"
    exit 1
fi
echo ""

# Test 6: Read from Followers (This was failing before)
echo "=== Test 6: Read from Followers (Replication Test) ==="
for i in {1..4}; do
    port=$((8080 + i))
    READ_FOLLOWER=$(curl -s "http://localhost:$port/get?key=test1")
    if echo "$READ_FOLLOWER" | grep -q "value1"; then
        echo "✓ Read from Follower $i successful"
        echo "  Response: $READ_FOLLOWER"
    else
        echo "✗ Read from Follower $i failed"
        echo "  Response: $READ_FOLLOWER"
        echo "  Checking follower $i log..."
        tail -10 /tmp/test_follower$i.log
        exit 1
    fi
done
echo ""

# Test 7: Local Read
echo "=== Test 7: Local Read (for inconsistency testing) ==="
LOCAL_READ=$(curl -s "http://localhost:8081/local_read?key=test1")
if echo "$LOCAL_READ" | grep -q "value1"; then
    echo "✓ Local read successful"
    echo "  Response: $LOCAL_READ"
else
    echo "✗ Local read failed"
    echo "  Response: $LOCAL_READ"
    exit 1
fi
echo ""

# Test 8: Write to Follower Should Fail
echo "=== Test 8: Write to Follower Should Be Rejected ==="
WRITE_TO_FOLLOWER=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST http://localhost:8081/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test2","value":"value2"}')
HTTP_CODE=$(echo "$WRITE_TO_FOLLOWER" | grep "HTTP_CODE" | cut -d: -f2)
if [ "$HTTP_CODE" = "403" ]; then
    echo "✓ Write to Follower correctly rejected (403)"
else
    echo "✗ Write to Follower should be rejected but got: $HTTP_CODE"
    echo "  Response: $WRITE_TO_FOLLOWER"
    exit 1
fi
echo ""

# Test 9: Strategy 2 (W=1, R=5)
echo "=== Test 9: Strategy 2 (W=1, R=5) ==="
echo "Setting W=1, R=5..."
curl -s -X POST http://localhost:8080/config \
    -H "Content-Type: application/json" \
    -d '{"r":5,"w":1}' > /dev/null

echo "Writing key 'test2' with value 'value2'..."
WRITE_RESPONSE=$(curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test2","value":"value2"}')
if echo "$WRITE_RESPONSE" | grep -q "created"; then
    echo "✓ Write successful (should be fast with W=1)"
    echo "  Response: $WRITE_RESPONSE"
else
    echo "✗ Write failed"
    exit 1
fi

echo "Waiting a bit for async replication..."
sleep 2

echo "Reading from Leader (should read from all 5 nodes)..."
READ_RESPONSE=$(curl -s "http://localhost:8080/get?key=test2")
if echo "$READ_RESPONSE" | grep -q "value2"; then
    echo "✓ Read with R=5 successful"
    echo "  Response: $READ_RESPONSE"
else
    echo "✗ Read with R=5 failed"
    echo "  Response: $READ_RESPONSE"
    exit 1
fi
echo ""

# Test 10: Strategy 3 (R=3, W=3)
echo "=== Test 10: Strategy 3 (R=3, W=3) ==="
echo "Setting R=3, W=3..."
curl -s -X POST http://localhost:8080/config \
    -H "Content-Type: application/json" \
    -d '{"r":3,"w":3}' > /dev/null

echo "Writing key 'test3' with value 'value3'..."
WRITE_RESPONSE=$(curl -s -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"test3","value":"value3"}')
if echo "$WRITE_RESPONSE" | grep -q "created"; then
    echo "✓ Write successful (quorum W=3)"
    echo "  Response: $WRITE_RESPONSE"
else
    echo "✗ Write failed"
    exit 1
fi

echo "Waiting for quorum replication..."
sleep 2

echo "Reading from Leader (should read from 3 nodes)..."
READ_RESPONSE=$(curl -s "http://localhost:8080/get?key=test3")
if echo "$READ_RESPONSE" | grep -q "value3"; then
    echo "✓ Read with R=3 successful"
    echo "  Response: $READ_RESPONSE"
else
    echo "✗ Read with R=3 failed"
    echo "  Response: $READ_RESPONSE"
    exit 1
fi
echo ""

echo "=== All Tests Passed! ==="
echo ""
echo "Summary:"
echo "  ✓ Health checks working"
echo "  ✓ Configuration endpoint working"
echo "  ✓ Write operations working"
echo "  ✓ Replication working (followers have data)"
echo "  ✓ Read operations working"
echo "  ✓ All three replication strategies working"
echo "  ✓ Write to follower correctly rejected"

