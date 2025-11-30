# Lock Options for Leaderless Mode

## Current State (No Lock)

### How It Works
- Any node can receive a write request and become the Write Coordinator
- Multiple nodes can simultaneously coordinate writes to the same key
- Each coordinator increments version independently
- Last write to finish "wins" (overwrites previous writes)

### Problems
- **Write Conflicts**: Two simultaneous writes to same key can conflict
- **Lost Updates**: One write may overwrite another without detection
- **False Writes**: Older version can overwrite newer version if timing is wrong
- **No Conflict Detection**: System doesn't know when conflicts occur

### Expected Results
- **Stale Reads**: Low rate (0.01-0.1%) - reads happen during replication
- **Write Conflicts**: Possible but hard to detect without tracking
- **Success Rate**: High (96-100%) - writes usually succeed
- **Latency**: ~300ms for writes (W=5, must wait for all nodes)

---

## Option 1: Distributed Lock (External Service)

### How It Works
- Use external service (DynamoDB, Redis, etcd, Consul) for distributed locks
- Before writing, coordinator acquires lock for the key
- Only one coordinator can hold lock at a time
- Release lock after replication completes

### Implementation Approach
```
1. Coordinator receives write request
2. Acquire distributed lock for key (with timeout)
3. If lock acquired:
   - Write locally
   - Replicate to all nodes
   - Release lock
4. If lock fails:
   - Return error or retry
```

### Expected Results
- **Write Conflicts**: Eliminated (only one write at a time per key)
- **Stale Reads**: Same as current (reads still happen during replication)
- **Success Rate**: Slightly lower (lock acquisition can fail)
- **Latency**: Higher (~350-400ms) - adds lock acquisition overhead
- **False Writes**: Eliminated (no concurrent writes possible)

### Pros
- ✅ Prevents all write conflicts
- ✅ Simple to understand
- ✅ Works across all nodes

### Cons
- ❌ Additional service dependency
- ❌ Higher latency (lock acquisition)
- ❌ Potential deadlocks if locks not released
- ❌ Single point of failure (lock service)

---

## Option 2: In-Memory Per-Key Lock (Single Node)

### How It Works
- Each node maintains a map of mutexes (one per key)
- Before writing locally, acquire lock for that key
- Only prevents concurrent writes on the same node
- Does NOT prevent conflicts across different nodes

### Implementation Approach
```
1. Coordinator receives write request
2. Acquire in-memory lock for key
3. Write locally
4. Replicate to all nodes
5. Release lock
```

### Expected Results
- **Write Conflicts**: Still possible (different nodes can still conflict)
- **Stale Reads**: Same as current
- **Success Rate**: Same as current
- **Latency**: Same as current (minimal overhead)
- **False Writes**: Still possible across nodes

### Pros
- ✅ Prevents concurrent writes on same node
- ✅ Very low overhead
- ✅ No external dependencies

### Cons
- ❌ Doesn't prevent cross-node conflicts
- ❌ Limited usefulness for leaderless mode

---

## Option 3: Optimistic Locking (Version-Based)

### How It Works
- Before writing, coordinator reads current version from all nodes
- Only write if version hasn't changed
- If version changed, retry or return conflict error
- Uses version numbers to detect conflicts

### Implementation Approach
```
1. Coordinator receives write request
2. Read current version from all nodes (or quorum)
3. If versions match:
   - Increment version
   - Write locally
   - Replicate to all nodes
4. If versions don't match:
   - Return conflict error or retry
```

### Expected Results
- **Write Conflicts**: Detected and handled (retry or error)
- **Stale Reads**: Same as current
- **Success Rate**: Lower (conflicts cause retries/failures)
- **Latency**: Higher (~350-400ms) - adds version check overhead
- **False Writes**: Prevented (won't overwrite newer version)

### Pros
- ✅ Detects conflicts before writing
- ✅ Prevents false writes
- ✅ No external dependencies

### Cons
- ❌ Higher latency (version check)
- ❌ Lower success rate (conflicts cause failures)
- ❌ Requires retry logic
- ❌ Race conditions still possible

---

## Option 4: Pessimistic Locking (Quorum-Based)

### How It Works
- Coordinator requests locks from a quorum of nodes (e.g., 3 out of 5)
- Only proceed if quorum grants lock
- Write and replicate, then release locks
- If quorum unavailable, fail the write

### Implementation Approach
```
1. Coordinator receives write request
2. Request locks from quorum of nodes (e.g., 3/5)
3. If quorum grants locks:
   - Write locally
   - Replicate to all nodes
   - Release locks
4. If quorum unavailable:
   - Return error
```

### Expected Results
- **Write Conflicts**: Prevented (quorum ensures exclusivity)
- **Stale Reads**: Same as current
- **Success Rate**: Lower (quorum must be available)
- **Latency**: Higher (~350-400ms) - adds lock acquisition
- **False Writes**: Prevented

### Pros
- ✅ Prevents write conflicts
- ✅ Works with quorum (more resilient than full lock)
- ✅ No external dependencies

### Cons
- ❌ Higher latency
- ❌ Lower success rate (quorum requirement)
- ❌ More complex implementation
- ❌ Deadlock risk if locks not released

---

## Option 5: Timestamp-Based Last-Write-Wins (LWW)

### How It Works
- Each write includes a timestamp (from coordinator)
- When replicating, compare timestamps
- Keep the write with the latest timestamp
- Accept that some writes may be lost

### Implementation Approach
```
1. Coordinator receives write request
2. Generate timestamp (coordinator's clock)
3. Write locally with timestamp
4. Replicate to all nodes with timestamp
5. Each node compares timestamps and keeps latest
```

### Expected Results
- **Write Conflicts**: Handled automatically (latest wins)
- **Stale Reads**: Same as current
- **Success Rate**: High (writes always succeed)
- **Latency**: Same as current (~300ms)
- **False Writes**: Prevented (timestamp comparison)

### Pros
- ✅ Simple conflict resolution
- ✅ No lock overhead
- ✅ Writes always succeed

### Cons
- ❌ May lose updates (older writes discarded)
- ❌ Requires clock synchronization
- ❌ Clock skew can cause issues

---

## Option 6: Hybrid: Version + Timestamp

### How It Works
- Combine version numbers with timestamps
- Version for ordering, timestamp for conflict resolution
- When versions conflict, use timestamp as tiebreaker

### Implementation Approach
```
1. Coordinator receives write request
2. Read current version from nodes
3. Generate timestamp
4. Write with (version, timestamp) tuple
5. Replicate to all nodes
6. Nodes compare both version and timestamp
```

### Expected Results
- **Write Conflicts**: Better detection and resolution
- **Stale Reads**: Same as current
- **Success Rate**: High
- **Latency**: Slightly higher (~320ms)
- **False Writes**: Prevented

### Pros
- ✅ Better conflict resolution
- ✅ Prevents false writes
- ✅ Reasonable latency

### Cons
- ❌ More complex
- ❌ Requires clock synchronization

---

## Recommended Testing Order

1. **Current (No Lock)** - Baseline
2. **Timestamp-Based LWW** - Simple improvement, no locks
3. **Optimistic Locking** - Good balance of safety and performance
4. **Distributed Lock** - Maximum safety, higher latency
5. **Pessimistic Quorum Lock** - Balanced approach

---

## Expected Impact Summary

| Option | Write Conflicts | False Writes | Stale Reads | Latency | Success Rate | Complexity |
|--------|----------------|--------------|-------------|---------|--------------|------------|
| **No Lock (Current)** | Possible | Possible | Low (0.01-0.1%) | ~300ms | High (96-100%) | Low |
| **Distributed Lock** | Eliminated | Eliminated | Same | ~350-400ms | Medium (90-95%) | Medium |
| **In-Memory Lock** | Still possible | Still possible | Same | ~300ms | Same | Low |
| **Optimistic Lock** | Detected | Prevented | Same | ~350-400ms | Medium (90-95%) | Medium |
| **Pessimistic Quorum** | Prevented | Prevented | Same | ~350-400ms | Medium (85-95%) | High |
| **Timestamp LWW** | Auto-resolved | Prevented | Same | ~300ms | High (96-100%) | Low |
| **Version + Timestamp** | Better resolved | Prevented | Same | ~320ms | High (95-99%) | Medium |

---

## Key Metrics to Track

For each lock option, measure:
1. **Write Conflict Rate**: % of writes that conflict
2. **False Write Rate**: % of writes that overwrite newer data
3. **Stale Read Rate**: % of reads that return old data
4. **Success Rate**: % of writes that succeed
5. **Write Latency**: P50, P95, P99 latencies
6. **Read Latency**: P50, P95, P99 latencies
7. **Error Rate**: % of requests that fail


