# Leader-Follower Database Implementation

This document describes how to use the Leader-Follower distributed key-value store.

## Architecture

- **N = 5**: One Leader and four Followers
- **All writes** go to the Leader, which replicates to Followers
- **Reads** can go to any node (Leader or Follower)
- **Three replication strategies** are supported:
  1. **W=5, R=1**: Write to all nodes, read from Leader only
  2. **W=1, R=5**: Write to Leader only, read from all nodes (return most recent)
  3. **R=3, W=3**: Quorum - write to 3 nodes, read from 3 nodes (return most recent)

## Building

```bash
go build -o leader-follower ./cmd/leader-follower
```

## Running Nodes

### Leader Node

```bash
./leader-follower \
  --node-id=leader1 \
  --role=leader \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8080
```

### Follower Nodes

**Follower 1:**
```bash
./leader-follower \
  --node-id=follower1 \
  --role=follower \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8081
```

**Follower 2:**
```bash
./leader-follower \
  --node-id=follower2 \
  --role=follower \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8082
```

**Follower 3:**
```bash
./leader-follower \
  --node-id=follower3 \
  --role=follower \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8083
```

**Follower 4:**
```bash
./leader-follower \
  --node-id=follower4 \
  --role=follower \
  --leader-addr=localhost:8080 \
  --follower-addrs=localhost:8081,localhost:8082,localhost:8083,localhost:8084 \
  --port=8084
```

## Configuration

### Setting Replication Parameters

You can configure R and W values via the `/config` endpoint:

**Get current configuration:**
```bash
curl http://localhost:8080/config
```

**Set W=5, R=1 (Strategy 1):**
```bash
curl -X POST http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{"r":1,"w":5}'
```

**Set W=1, R=5 (Strategy 2):**
```bash
curl -X POST http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{"r":5,"w":1}'
```

**Set R=3, W=3 (Strategy 3 - Quorum):**
```bash
curl -X POST http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{"r":3,"w":3}'
```

## API Endpoints

### Write (POST /set)
**Only works on Leader node**

```bash
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test","value":"hello"}'
```

### Read (GET /get)
**Works on any node**

```bash
curl "http://localhost:8080/get?key=test"
curl "http://localhost:8081/get?key=test"  # Can read from follower too
```

### Local Read (GET /local_read)
**For testing inconsistency windows**

```bash
curl "http://localhost:8081/local_read?key=test"
```

### Health Check (GET /health)
```bash
curl http://localhost:8080/health
```

## Replication Delays

The implementation includes delays to simulate real-world conditions and make inconsistency windows easier to observe:

- **Leader → Follower (write)**: Leader sleeps 200ms after each message to a Follower
- **Follower (write)**: Follower sleeps 100ms when receiving update before responding
- **Follower (read)**: Follower sleeps 50ms when receiving read request from Leader before responding
- **Leader (read)**: No delay

## Testing

### Test Strategy 1 (W=5, R=1)

1. Set configuration:
```bash
curl -X POST http://localhost:8080/config -H "Content-Type: application/json" -d '{"r":1,"w":5}'
```

2. Write to Leader:
```bash
curl -X POST http://localhost:8080/set -H "Content-Type: application/json" -d '{"key":"test1","value":"value1"}'
```

3. Read from Leader (should be consistent):
```bash
curl "http://localhost:8080/get?key=test1"
```

4. Read from Follower (should be consistent after replication):
```bash
curl "http://localhost:8081/get?key=test1"
```

### Test Strategy 2 (W=1, R=5)

1. Set configuration:
```bash
curl -X POST http://localhost:8080/config -H "Content-Type: application/json" -d '{"r":5,"w":1}'
```

2. Write to Leader (responds immediately):
```bash
curl -X POST http://localhost:8080/set -H "Content-Type: application/json" -d '{"key":"test2","value":"value2"}'
```

3. Read from any node (reads from all 5 nodes, returns most recent):
```bash
curl "http://localhost:8080/get?key=test2"
```

### Test Strategy 3 (R=3, W=3)

1. Set configuration:
```bash
curl -X POST http://localhost:8080/config -H "Content-Type: application/json" -d '{"r":3,"w":3}'
```

2. Write to Leader (waits for 3 nodes total):
```bash
curl -X POST http://localhost:8080/set -H "Content-Type: application/json" -d '{"key":"test3","value":"value3"}'
```

3. Read from any node (reads from 3 nodes, returns most recent):
```bash
curl "http://localhost:8080/get?key=test3"
```

## Docker Deployment

You can use Docker Compose to deploy all 5 nodes. Create a `docker-compose.yml` file (see example below) and run:

```bash
docker-compose up -d
```

## Notes

- All writes must go to the Leader node
- Reads can go to any node (Leader or Follower)
- The load-test client should know the Leader's address and only send writes to it
- Version numbers are used to determine the most recent value when reading from multiple nodes
- The `local_read` endpoint is useful for testing inconsistency windows during replication

