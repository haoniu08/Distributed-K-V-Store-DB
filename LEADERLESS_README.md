# Leaderless Database Implementation

This document describes how to use the Leaderless distributed key-value store.

## Architecture

- **N = 5**: Five equal nodes (no leader)
- **W = N (5)**: All nodes must be updated before write completes
- **R = 1**: Reads return local value immediately (no coordination)
- **Any node** can receive read/write requests
- When a node receives a write, it becomes the **Write Coordinator** for that request
- The Write Coordinator replicates to all other nodes before responding

## Key Characteristics

1. **No Single Point of Failure**: Any node can handle requests
2. **Inconsistency Window**: Reads return local values immediately, so stale reads are possible
3. **Write Coordination**: The node receiving a write coordinates replication to all others
4. **Eventual Consistency**: All nodes will eventually have the same data

## Building

```bash
go build -o leaderless ./cmd/leaderless
```

## Running Nodes

All nodes are equal. Each node needs to know about all other nodes.

**Node 1:**
```bash
./leaderless \
  --node-id=node1 \
  --all-node-addrs=localhost:8080,localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8080
```

**Node 2:**
```bash
./leaderless \
  --node-id=node2 \
  --all-node-addrs=localhost:8080,localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8081
```

**Node 3:**
```bash
./leaderless \
  --node-id=node3 \
  --all-node-addrs=localhost:8080,localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8082
```

**Node 4:**
```bash
./leaderless \
  --node-id=node4 \
  --all-node-addrs=localhost:8080,localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8083
```

**Node 5:**
```bash
./leaderless \
  --node-id=node5 \
  --all-node-addrs=localhost:8080,localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8084
```

## API Endpoints

### Write (POST /set)
**Works on any node - that node becomes the Write Coordinator**

```bash
# Write to Node 1
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test","value":"hello"}'

# Write to Node 3 (Node 3 becomes coordinator)
curl -X POST http://localhost:8082/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test2","value":"world"}'
```

**Behavior:**
- The node receiving the write becomes the Write Coordinator
- Coordinator writes locally first
- Coordinator replicates to all other 4 nodes
- Coordinator waits for all nodes to confirm (W=N)
- Only then returns 201-Created to client
- Takes time due to replication delays (~1-2 seconds)

### Read (GET /get)
**Works on any node - returns local value immediately (R=1)**

```bash
# Read from Node 1
curl "http://localhost:8080/get?key=test"

# Read from Node 5
curl "http://localhost:8084/get?key=test"
```

**Behavior:**
- Returns local value immediately
- No coordination with other nodes
- May return stale value if replication hasn't completed
- This is the expected inconsistency window

### Local Read (GET /local_read)
**For testing inconsistency windows**

```bash
curl "http://localhost:8081/local_read?key=test"
```

### Health Check (GET /health)
```bash
curl http://localhost:8080/health
```

Returns: `{"status":"healthy","mode":"leaderless","node_id":"node1","time":"..."}`

## Replication Delays

The implementation includes delays to simulate real-world conditions:

- **Write Coordinator → Other Node**: Coordinator sleeps 200ms after each message
- **Node Receiving Write**: Node sleeps 100ms when receiving update before responding
- **Total time for W=5**: ~(200ms * 4) + (100ms * 4) = ~1.2 seconds minimum

## Testing Inconsistency Window

The inconsistency window can be demonstrated:

```bash
# Write to Node 1
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"inconsistent","value":"new_value"}'

# Immediately read from Node 5 (might return old value or "key not found")
curl "http://localhost:8084/get?key=inconsistent"

# Wait a few seconds, then read again (should have new value)
sleep 3
curl "http://localhost:8084/get?key=inconsistent"
```

## Example Workflow

1. **Client writes to Node 2:**
   - Node 2 becomes Write Coordinator
   - Node 2 writes locally
   - Node 2 replicates to Nodes 1, 3, 4, 5
   - All nodes confirm
   - Node 2 returns 201-Created to client

2. **Client reads from Node 4:**
   - Node 4 returns its local value immediately
   - If replication completed, returns new value
   - If replication still in progress, may return old value or "key not found"

3. **Client writes to Node 5:**
   - Node 5 becomes Write Coordinator
   - Same process as above

## Docker Deployment

You can use Docker Compose to deploy all 5 nodes:

```bash
docker-compose -f docker-compose-leaderless.yml up -d
```

## Differences from Leader-Follower

| Feature | Leader-Follower | Leaderless |
|---------|----------------|------------|
| Write Target | Only Leader | Any Node |
| Read Coordination | Depends on R | None (R=1) |
| Consistency | Strong (with W=5) | Eventual |
| Single Point of Failure | Yes (Leader) | No |
| Inconsistency Window | Small | Larger |

## Notes

- All writes must replicate to all nodes (W=N)
- Reads return local values immediately (R=1)
- Inconsistency windows are expected and acceptable
- Any node can become Write Coordinator
- Version numbers ensure eventual consistency

