# Testing Leaderless KV Store

## ✅ Current Status

Your leaderless KV store is **running successfully**!

- ✅ All 5 nodes started
- ✅ Configuration: W=5 (N), R=1
- ✅ Write/Read operations working

## Test Replication

Since W=5 (write to all nodes), let's verify replication is working:

### 1. Write to Node 1
```bash
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"replication-test","value":"written-to-node1"}'
```

### 2. Read from Different Nodes

```bash
# Read from Node 1 (should have the value)
curl "http://localhost:8080/get?key=replication-test"

# Read from Node 2 (should eventually have it after replication)
curl "http://localhost:8082/get?key=replication-test"

# Read from Node 3
curl "http://localhost:8083/get?key=replication-test"

# Read from Node 4
curl "http://localhost:8084/get?key=replication-test"

# Read from Node 5
curl "http://localhost:8085/get?key=replication-test"
```

**Note**: With W=5, the write should wait for all nodes to confirm. All reads should return the same value.

### 3. Test Local Reads (Inconsistency Window)

Test the inconsistency window by reading locally from each node:

```bash
# Local read from Node 1
curl "http://localhost:8080/local_read?key=replication-test"

# Local read from Node 2
curl "http://localhost:8082/local_read?key=replication-test"

# Local read from Node 3
curl "http://localhost:8083/local_read?key=replication-test"
```

### 4. Test Write Coordinator

Write to different nodes and verify they all become write coordinators:

```bash
# Write via Node 2
curl -X POST http://localhost:8082/set \
  -H "Content-Type: application/json" \
  -d '{"key":"coordinator-test","value":"from-node2"}'

# Write via Node 3
curl -X POST http://localhost:8083/set \
  -H "Content-Type: application/json" \
  -d '{"key":"coordinator-test","value":"from-node3"}'

# Read from any node
curl "http://localhost:8080/get?key=coordinator-test"
```

## Test Concurrent Writes

Test what happens with concurrent writes to the same key:

```bash
# Write to multiple nodes simultaneously
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"concurrent","value":"write1"}' &
  
curl -X POST http://localhost:8082/set \
  -H "Content-Type: application/json" \
  -d '{"key":"concurrent","value":"write2"}' &
  
curl -X POST http://localhost:8084/set \
  -H "Content-Type: application/json" \
  -d '{"key":"concurrent","value":"write3"}' &

wait

# Check final value
curl "http://localhost:8080/get?key=concurrent"
```

## Monitor Logs

Watch replication in real-time:

```bash
# In a separate terminal
docker compose -f docker-compose-leaderless.yml logs -f
```

## Node Endpoints

- **Node 1**: http://localhost:8080
- **Node 2**: http://localhost:8082
- **Node 3**: http://localhost:8083
- **Node 4**: http://localhost:8084
- **Node 5**: http://localhost:8085

## Expected Behavior

### W=5, R=1 Configuration:
- **Writes**: Must wait for all 5 nodes to confirm (slower but consistent)
- **Reads**: Return local value immediately (fast but may be stale)
- **Inconsistency Window**: Possible during replication delays
- **Write Coordinator**: The node receiving the write coordinates replication

## Stop the Cluster

```bash
docker compose -f docker-compose-leaderless.yml down
```

## Next Steps

1. ✅ Basic read/write working
2. Test replication across all nodes
3. Test inconsistency window with local reads
4. Test concurrent writes
5. Run consistency tests from `tests/` directory
6. Run load tests from `loadtester/` directory



