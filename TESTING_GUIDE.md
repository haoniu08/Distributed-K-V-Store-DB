# Phase 2 Testing Guide

This guide helps you test the Leader-Follower database implementation.

## Quick Start

1. **Build the binary:**
```bash
go build -o leader-follower ./cmd/leader-follower
```

2. **Start nodes manually** (use separate terminal windows):

**Terminal 1 - Leader:**
```bash
./leader-follower \
  --node-id=leader1 \
  --role=leader \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8080
```

**Terminal 2 - Follower 1:**
```bash
./leader-follower \
  --node-id=follower1 \
  --role=follower \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8081
```

**Terminal 3 - Follower 2:**
```bash
./leader-follower \
  --node-id=follower2 \
  --role=follower \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8082
```

**Terminal 4 - Follower 3:**
```bash
./leader-follower \
  --node-id=follower3 \
  --role=follower \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8083
```

**Terminal 5 - Follower 4:**
```bash
./leader-follower \
  --node-id=follower4 \
  --role=follower \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8084
```

## Test Cases

### Test 1: Health Checks

```bash
# Check Leader
curl http://localhost:8080/health

# Check Followers
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
```

Expected: All should return `{"status":"healthy","role":"leader"/"follower","time":"..."}`

### Test 2: Configuration

```bash
# Get current config
curl http://localhost:8080/config

# Set W=5, R=1
curl -X POST http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{"r":1,"w":5}'

# Verify config
curl http://localhost:8080/config
```

Expected: Should show `{"node_id":"leader1","role":"leader","n":5,"r":1,"w":5}`

### Test 3: Strategy 1 (W=5, R=1)

```bash
# Set strategy
curl -X POST http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{"r":1,"w":5}'

# Write to Leader
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test1","value":"value1"}'

# Wait for replication (W=5 means all 5 nodes must confirm)
# This will take time due to delays (200ms * 4 followers + 100ms per follower)
sleep 3

# Read from Leader
curl "http://localhost:8080/get?key=test1"

# Read from Follower 1
curl "http://localhost:8081/get?key=test1"

# Read from Follower 2
curl "http://localhost:8082/get?key=test1"
```

Expected: All reads should return `{"key":"test1","value":"value1","version":1}`

### Test 4: Strategy 2 (W=1, R=5)

```bash
# Set strategy
curl -X POST http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{"r":5,"w":1}'

# Write to Leader (should respond immediately)
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test2","value":"value2"}'

# Read from Leader (should read from all 5 nodes and return most recent)
curl "http://localhost:8080/get?key=test2"
```

Expected: Write should return immediately. Read should return value from all 5 nodes (may take time due to 50ms delays on followers).

### Test 5: Strategy 3 (R=3, W=3)

```bash
# Set strategy
curl -X POST http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{"r":3,"w":3}'

# Write to Leader (waits for 3 nodes total)
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test3","value":"value3"}'

# Read from Leader (reads from 3 nodes)
curl "http://localhost:8080/get?key=test3"
```

Expected: Write should complete after 3 nodes confirm. Read should return value from 3 nodes.

### Test 6: Write to Follower Should Fail

```bash
# Try to write to Follower
curl -X POST http://localhost:8081/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test4","value":"value4"}'
```

Expected: Should return `403 Forbidden` with message "only leader accepts write requests"

### Test 7: Local Read (for testing inconsistency)

```bash
# Write to Leader
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test5","value":"value5"}'

# Immediately read locally from Follower (might show inconsistency)
curl "http://localhost:8081/local_read?key=test5"
```

Expected: If read happens during replication window, might return old value or "key not found". After replication completes, should return correct value.

### Test 8: Version Tracking

```bash
# Set W=5, R=1
curl -X POST http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{"r":1,"w":5}'

# Write first version
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"version_test","value":"v1"}'
sleep 3

# Write second version
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"version_test","value":"v2"}'
sleep 3

# Read and check version
curl "http://localhost:8080/get?key=version_test"
```

Expected: Should return `{"key":"version_test","value":"v2","version":2}`

## Troubleshooting

### Issue: Port already in use
```bash
# Kill processes on ports
lsof -ti:8080 | xargs kill -9
lsof -ti:8081 | xargs kill -9
lsof -ti:8082 | xargs kill -9
lsof -ti:8083 | xargs kill -9
lsof -ti:8084 | xargs kill -9
```

### Issue: Replication not working
- Check that all nodes are running
- Check that follower addresses are correct
- Check logs in terminal windows
- Verify network connectivity between nodes

### Issue: Config endpoint returns 404
- Make sure you're hitting the Leader node (port 8080)
- Check that the node started successfully
- Verify the route is registered (check main.go)

## Expected Behavior

1. **W=5, R=1**: Writes take longer (all 5 nodes must confirm), but reads are fast (only from Leader)
2. **W=1, R=5**: Writes are fast (Leader only), but reads take longer (must read from all 5 nodes)
3. **R=3, W=3**: Balanced - both reads and writes need 3 nodes (quorum)

## Notes

- Replication delays are intentional to simulate real-world conditions
- The inconsistency window is visible with `local_read` during replication
- All writes must go to the Leader
- Reads can go to any node

