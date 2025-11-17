# Distributed Key-Value Store Database

A distributed in-memory key-value store implementing Leader-Follower and Leaderless replication strategies.

## Project Structure

- `cmd/kv-service/` - Basic KV service implementation
- `cmd/leader-follower/` - Leader-Follower database implementation
- `cmd/leaderless/` - Leaderless database implementation
- `internal/kvstore/` - Core KV store logic
- `internal/api/` - HTTP handlers
- `internal/models/` - Data models
- `loadtester/` - Load testing client
- `tests/` - Unit and integration tests

## Phase 1: Basic KV Service

### Prerequisites

- Go 1.21 or later
- Docker (optional, for containerized deployment)

### Setup

1. Initialize Go module (if not already done):
```bash
go mod init github.com/yourusername/distributed-kv-store
```

2. Download dependencies:
```bash
go mod download
go mod tidy
```

### Building

```bash
go build -o kv-service ./cmd/kv-service
```

### Running

```bash
./kv-service
# Or with custom port:
PORT=8080 ./kv-service
```

### API Endpoints

- `POST /set` - Set a key-value pair
  ```bash
  curl -X POST http://localhost:8080/set \
    -H "Content-Type: application/json" \
    -d '{"key":"mykey","value":"myvalue"}'
  ```
  Returns: 201 Created
  ```json
  {
    "key": "mykey",
    "value": "myvalue",
    "version": 1,
    "status": "created"
  }
  ```

- `GET /get?key=mykey` - Get a value by key
  ```bash
  curl "http://localhost:8080/get?key=mykey"
  ```
  Returns: 200 OK with value, or 404 Not Found
  ```json
  {
    "key": "mykey",
    "value": "myvalue",
    "version": 1
  }
  ```

- `GET /local_read?key=mykey` - Local read (for testing inconsistency windows)
  ```bash
  curl "http://localhost:8080/local_read?key=mykey"
  ```
  Returns: 200 OK with local value, or 404 Not Found

- `GET /health` - Health check
  ```bash
  curl http://localhost:8080/health
  ```
  Returns: 200 OK
  ```json
  {
    "status": "healthy",
    "time": "2024-01-01T00:00:00Z"
  }
  ```

### Docker

```bash
# Build
docker build -t kv-service .

# Run
docker run -p 8080:8080 kv-service

# Run with custom port
docker run -p 9000:8080 -e PORT=8080 kv-service
```

### Testing

```bash
# Test set
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test","value":"hello"}'

# Test get
curl "http://localhost:8080/get?key=test"

# Test local_read
curl "http://localhost:8080/local_read?key=test"

# Test health
curl http://localhost:8080/health
```

## Next Steps

- Phase 2: Implement Leader-Follower database with replication strategies
- Phase 3: Implement Leaderless database
- Phase 4: Write consistency tests
- Phase 5: Create load testing client
