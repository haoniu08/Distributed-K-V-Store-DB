# Leaderless Lock Mechanism Design

## Problem Statement

In leaderless mode, multiple nodes can receive write requests for the same key simultaneously. Without coordination, this leads to:

1. **Lost Updates**: Last write wins, earlier writes are lost
2. **Race Conditions**: Concurrent writes create inconsistent state
3. **Version Conflicts**: Different nodes may have different versions

## Solution: Distributed Locking

### High-Level Approach

**Before a node can become Write Coordinator for a key, it must:**
1. Acquire a distributed lock for that key
2. Hold the lock during replication
3. Release the lock after replication completes

This ensures only one node can coordinate writes for a given key at a time.

---

## Lock Service Architecture

### Option 1: DynamoDB-Based Locking (Recommended)

**Why DynamoDB:**
- Managed AWS service (no infrastructure to manage)
- Built-in TTL (automatic lock expiration)
- High availability and durability
- Low latency
- Conditional writes for atomic operations

**Lock Table Design:**
```
Table: kv-locks
Partition Key: key (String)

Attributes:
- lock_holder: node_id (String) - Which node holds the lock
- lock_timestamp: timestamp (Number) - When lock was acquired
- ttl: expiration_time (Number) - DynamoDB TTL attribute
- lock_version: version (Number) - For optimistic locking
```

**Lock Operations:**

1. **Acquire Lock:**
   - Conditional PutItem: key doesn't exist OR ttl expired
   - Set lock_holder = node_id
   - Set ttl = current_time + lock_duration
   - Return success/failure

2. **Release Lock:**
   - DeleteItem: Only if lock_holder == node_id
   - Ensures only lock holder can release

3. **Extend Lock (Heartbeat):**
   - UpdateItem: Extend ttl if lock_holder == node_id
   - Used if write takes longer than expected

**Lock Duration:**
- Default: 5 seconds (sufficient for W=N replication)
- Extendable: Up to 30 seconds maximum
- Automatic expiration: DynamoDB TTL feature

### Option 2: ElastiCache (Redis) Based Locking

**Why Redis:**
- Very fast (in-memory)
- Low latency
- Well-established locking patterns (Redlock)

**Lock Operations:**
- SETNX (SET if Not eXists) for acquisition
- EXPIRE for TTL
- Lua scripts for atomic operations

**Trade-off:** Requires Redis cluster for HA, more infrastructure to manage

### Option 3: Consensus-Based (Raft)

**Why Raft:**
- No external dependencies
- Self-contained
- Strong consistency

**Trade-off:** Complex implementation, higher latency

---

## Lock Acquisition Flow

### Write Request with Locking

```
1. Node receives write(key, value) request
   ↓
2. Attempt to acquire lock for key
   ├─ Success → Continue to step 3
   └─ Failure → Wait and retry (exponential backoff)
                 OR return error to client
   ↓
3. Lock acquired - Node becomes Write Coordinator
   ↓
4. Write locally
   ↓
5. Replicate to all other nodes (W=N)
   ↓
6. Wait for all confirmations
   ↓
7. Release lock
   ↓
8. Return success to client
```

### Lock Retry Strategy

**Exponential Backoff:**
- Initial wait: 100ms
- Backoff multiplier: 2x
- Maximum wait: 2 seconds
- Maximum retries: 10
- Total timeout: 5 seconds

**Why:**
- Reduces lock contention
- Allows other writes to complete
- Prevents thundering herd

---

## Lock Contention Scenarios

### Scenario 1: Concurrent Writes to Same Key

**Timeline:**
```
T0: Node 1 receives write(key_A, value_1)
T0: Node 2 receives write(key_A, value_2)  (concurrent)
T0: Node 1 acquires lock (first)
T0: Node 2 attempts lock → fails, waits
T1: Node 1 completes write, releases lock
T1: Node 2 acquires lock (retry)
T2: Node 2 completes write, releases lock
```

**Result:** Writes are serialized, no lost updates

### Scenario 2: Different Keys (No Contention)

**Timeline:**
```
T0: Node 1 receives write(key_A, value_1)
T0: Node 2 receives write(key_B, value_2)  (different key)
T0: Node 1 acquires lock(key_A)
T0: Node 2 acquires lock(key_B)  (different lock, no conflict)
T1: Both complete in parallel
```

**Result:** No contention, parallel processing

### Scenario 3: Node Failure During Write

**Timeline:**
```
T0: Node 1 acquires lock(key_A)
T0: Node 1 starts replication
T1: Node 1 crashes (network partition, failure)
T5: Lock expires (TTL)
T5: Node 2 can acquire lock(key_A)
T6: Node 2 completes write
```

**Result:** Automatic recovery via TTL expiration

---

## Implementation Design

### New Components

#### 1. Lock Service Interface

```go
// internal/leaderless/lock.go

type LockService interface {
    // Acquire lock for a key
    AcquireLock(ctx context.Context, key string, nodeID string, duration time.Duration) (bool, error)
    
    // Release lock (only if held by this node)
    ReleaseLock(ctx context.Context, key string, nodeID string) error
    
    // Extend lock duration (heartbeat)
    ExtendLock(ctx context.Context, key string, nodeID string, duration time.Duration) error
}
```

#### 2. DynamoDB Lock Service Implementation

```go
type DynamoDBLockService struct {
    client    *dynamodb.Client
    tableName string
    nodeID    string
}

func (s *DynamoDBLockService) AcquireLock(ctx context.Context, key string, nodeID string, duration time.Duration) (bool, error) {
    // Conditional PutItem:
    // - Key doesn't exist, OR
    // - TTL has expired
    // Set lock_holder = nodeID
    // Set TTL = now + duration
}
```

#### 3. Integration with Replication Manager

```go
// internal/leaderless/replication.go

func (rm *ReplicationManager) WriteWithCoordination(key, value string) (*WriteResult, error) {
    // 1. Acquire lock
    acquired, err := rm.lockService.AcquireLock(ctx, key, rm.config.NodeID, 5*time.Second)
    if !acquired {
        return nil, fmt.Errorf("failed to acquire lock")
    }
    defer rm.lockService.ReleaseLock(ctx, key, rm.config.NodeID)
    
    // 2. Write locally
    version, err := rm.store.Set(key, value)
    
    // 3. Replicate to all nodes
    // ... (existing replication logic)
    
    // 4. Lock released automatically via defer
    return result, nil
}
```

---

## AWS Integration Points

### 1. DynamoDB Setup

**Table Creation:**
- Table name: `kv-locks`
- Partition key: `key` (String)
- TTL attribute: `ttl` (Number)
- Billing: On-demand or provisioned capacity

**Access Pattern:**
- Write: PutItem (lock acquisition)
- Read: GetItem (lock check)
- Delete: DeleteItem (lock release)
- Update: UpdateItem (lock extension)

### 2. IAM Permissions

**Required Permissions:**
```json
{
  "Effect": "Allow",
  "Action": [
    "dynamodb:PutItem",
    "dynamodb:GetItem",
    "dynamodb:DeleteItem",
    "dynamodb:UpdateItem",
    "dynamodb:Query"
  ],
  "Resource": "arn:aws:dynamodb:region:account:table/kv-locks"
}
```

### 3. Service Discovery

**AWS Cloud Map:**
- Register each node as a service
- Automatic health checks
- DNS-based discovery
- Multi-AZ support

**Node Registration:**
- Each node registers itself on startup
- Updates health status
- Deregisters on shutdown

---

## Performance Considerations

### Lock Acquisition Overhead

**Expected Latency:**
- DynamoDB PutItem: ~5-10ms (same region)
- Total overhead: ~10-20ms per write
- Acceptable for most use cases

### Lock Contention

**Low Contention Scenario:**
- Different keys → No contention
- Parallel processing
- Minimal overhead

**High Contention Scenario:**
- Same key, multiple writers
- Sequential processing (by design)
- Retry overhead
- May need to optimize (e.g., request queuing)

### Optimization Strategies

1. **Lock-Free Reads:**
   - Reads don't need locks (R=1, local read)
   - Only writes require locks

2. **Key Partitioning:**
   - Partition locks by key prefix
   - Reduce contention on lock table

3. **Lock Batching:**
   - Batch multiple key locks
   - Reduce DynamoDB calls

---

## Failure Handling

### Node Failure During Lock Hold

**Scenario:** Node crashes while holding lock

**Solution:** TTL expiration
- Lock expires automatically (5 seconds)
- Other nodes can acquire lock
- No manual cleanup needed

### Network Partition

**Scenario:** Node partitioned from DynamoDB

**Solution:** Lock expiration + retry
- Cannot acquire/extend lock
- Lock expires
- Other nodes can proceed
- Partitioned node retries when network recovers

### DynamoDB Unavailability

**Scenario:** DynamoDB service down

**Solution:** Fallback mechanism
- Retry with exponential backoff
- Return error to client
- Log for monitoring
- Consider fallback to local locking (degraded mode)

---

## Monitoring and Observability

### Key Metrics

1. **Lock Acquisition:**
   - Success rate
   - Average latency
   - Contention rate (retries)

2. **Lock Hold Duration:**
   - Average hold time
   - Maximum hold time
   - TTL expiration rate

3. **Lock Contention:**
   - Number of retries
   - Wait time for lock
   - Contention by key

### CloudWatch Metrics

- `LockAcquisitionSuccess`
- `LockAcquisitionFailure`
- `LockAcquisitionLatency`
- `LockHoldDuration`
- `LockContentionRetries`

---

## Alternative: Optimistic Locking

### Approach

Instead of explicit locks, use version numbers:

1. **Read current version** before write
2. **Write with version check**: Only if version matches
3. **Retry on conflict**: If version changed, retry

**Pros:**
- No lock service needed
- Better for low contention
- Simpler implementation

**Cons:**
- Retries on conflict
- May waste work
- Less predictable

**Use Case:** When write conflicts are rare

---

## Recommendation

**Use DynamoDB-based locking** because:

1. ✅ Simple to implement
2. ✅ AWS managed service
3. ✅ Automatic TTL expiration
4. ✅ High availability
5. ✅ Low latency
6. ✅ No additional infrastructure

**Implementation Priority:**
1. Add lock service interface
2. Implement DynamoDB lock service
3. Integrate with write handler
4. Add retry logic
5. Add monitoring

---

*This design provides a solid foundation for implementing distributed locking in leaderless mode on AWS.*





