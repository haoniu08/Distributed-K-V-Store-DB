#!/bin/bash

# Manual consistency test script
# This script demonstrates consistency tests using curl

set -e

echo "=== Manual Consistency Tests ==="
echo ""
echo "This script demonstrates consistency testing using curl commands."
echo "Make sure the nodes are running before executing these tests."
echo ""

LEADER="localhost:8080"
FOLLOWER1="localhost:8081"
FOLLOWER2="localhost:8082"
NODE1="localhost:8080"
NODE2="localhost:8081"
NODE3="localhost:8082"

echo "=== Leader-Follower Consistency Tests ==="
echo ""

# Test 1: Write to Leader, read from Leader
echo "Test 1: Write to Leader, read from Leader (should be consistent)"
echo "Writing key 'consist_test1'..."
WRITE_RESP=$(curl -s -X POST http://$LEADER/set \
    -H "Content-Type: application/json" \
    -d '{"key":"consist_test1","value":"value1"}')
echo "Write response: $WRITE_RESP"
echo ""

echo "Reading from Leader..."
READ_RESP=$(curl -s "http://$LEADER/get?key=consist_test1")
echo "Read response: $READ_RESP"
if echo "$READ_RESP" | grep -q "value1"; then
    echo "✓ Consistent"
else
    echo "✗ Inconsistent"
fi
echo ""

# Test 2: Write to Leader, read from Follower (after ack)
echo "Test 2: Write to Leader, read from Follower (after ack, should be consistent)"
echo "Writing key 'consist_test2'..."
curl -s -X POST http://$LEADER/set \
    -H "Content-Type: application/json" \
    -d '{"key":"consist_test2","value":"value2"}' > /dev/null
echo "Waiting for replication..."
sleep 3

echo "Reading from Follower 1..."
READ_RESP=$(curl -s "http://$FOLLOWER1/get?key=consist_test2")
echo "Read response: $READ_RESP"
if echo "$READ_RESP" | grep -q "value2"; then
    echo "✓ Consistent"
else
    echo "✗ Inconsistent"
fi
echo ""

# Test 3: Write to Leader, immediately local_read from Followers
echo "Test 3: Write to Leader, immediately local_read from Followers (may show inconsistency)"
echo "Writing key 'consist_test3'..."
curl -s -X POST http://$LEADER/set \
    -H "Content-Type: application/json" \
    -d '{"key":"consist_test3","value":"value3"}' > /dev/null

echo "Immediately reading locally from Follower 1..."
LOCAL_RESP=$(curl -s "http://$FOLLOWER1/local_read?key=consist_test3")
echo "Local read response: $LOCAL_RESP"

if echo "$LOCAL_RESP" | grep -q "key not found"; then
    echo "✓ Inconsistency detected (key not found during replication)"
elif echo "$LOCAL_RESP" | grep -q "value3"; then
    echo "  Note: Replication completed quickly, no inconsistency observed"
else
    echo "  Note: Different value observed (inconsistency)"
fi
echo ""

echo "=== Leaderless Consistency Tests ==="
echo ""

# Test 4: Write to random node, read from other nodes (should show inconsistency)
echo "Test 4: Write to Node 1, immediately read from Node 2 (should show inconsistency)"
echo "Writing key 'consist_test4' to Node 1..."
curl -s -X POST http://$NODE1/set \
    -H "Content-Type: application/json" \
    -d '{"key":"consist_test4","value":"value4"}' > /dev/null

echo "Immediately reading from Node 2..."
READ_RESP=$(curl -s "http://$NODE2/get?key=consist_test4")
echo "Read response: $READ_RESP"

if echo "$READ_RESP" | grep -q "key not found"; then
    echo "✓ Inconsistency detected (key not found during replication)"
elif echo "$READ_RESP" | grep -q "value4"; then
    echo "  Note: Replication completed quickly, no inconsistency observed"
else
    echo "  Note: Different value observed (inconsistency)"
fi
echo ""

# Test 5: After Coordinator ack, read from Coordinator
echo "Test 5: After Coordinator ack, read from Coordinator (should be consistent)"
echo "Writing key 'consist_test5' to Node 1..."
WRITE_RESP=$(curl -s -X POST http://$NODE1/set \
    -H "Content-Type: application/json" \
    -d '{"key":"consist_test5","value":"value5"}')
echo "Write response: $WRITE_RESP"

echo "Reading from Coordinator (Node 1)..."
READ_RESP=$(curl -s "http://$NODE1/get?key=consist_test5")
echo "Read response: $READ_RESP"
if echo "$READ_RESP" | grep -q "value5"; then
    echo "✓ Consistent"
else
    echo "✗ Inconsistent"
fi
echo ""

# Test 6: After Coordinator ack, read from another node
echo "Test 6: After Coordinator ack, read from another node (should be consistent)"
echo "Writing key 'consist_test6' to Node 1..."
curl -s -X POST http://$NODE1/set \
    -H "Content-Type: application/json" \
    -d '{"key":"consist_test6","value":"value6"}' > /dev/null
echo "Waiting for replication..."
sleep 4

echo "Reading from Node 3..."
READ_RESP=$(curl -s "http://$NODE3/get?key=consist_test6")
echo "Read response: $READ_RESP"
if echo "$READ_RESP" | grep -q "value6"; then
    echo "✓ Consistent"
else
    echo "✗ Inconsistent"
fi
echo ""

echo "=== Consistency Tests Complete ==="





