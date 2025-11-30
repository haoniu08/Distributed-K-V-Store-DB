# Leaderless Concurrency and Global Clocks

## The Problem: Concurrent Writes in Leaderless Mode

### Scenario: Two Simultaneous Writes to Same Key

**Without Coordination:**

```
Time    Node 1                    Node 2
----    --------                  --------
T0      Receives write(key_A, v1)  Receives write(key_A, v2)
        Current version: 5        Current version: 5
T1      Becomes Write Coordinator Becomes Write Coordinator
        Sets version = 6           Sets version = 6
T2      Replicates to Node 2      Replicates to Node 1
T3      Node 2 receives v6          Node 1 receives v6
        (overwrites its own v6!)   (overwrites its own v6!)
```

**Result**: Both writes succeed, but one overwrites the other. Last write wins, but we lose one update.

### Why Version Numbers Alone Don't Solve This

**Current Implementation:**
- Each node maintains a global version counter
- When writing, increment version
- Replicate with new version

**Problem:**
- If two nodes start writing simultaneously, they both read the same current version
- Both increment to the same new version
- Both replicate with the same version
- Last one to finish "wins" (but we lost an update)

**Example:**
```
Node 1: Reads version=5, writes version=6, replicates
Node 2: Reads version=5 (before Node 1's write), writes version=6, replicates
Result: Both have version=6, but different values. Last write wins.
```

## Solution: Global Clock / Timestamp

### How a Global Clock Would Help

**With Global Clock (TrueTime, Logical Clock, etc.):**

```
Time    Node 1                    Node 2
----    --------                  --------
T0      Receives write(key_A, v1) Receives write(key_A, v2)
        Timestamp: T=100           Timestamp: T=101
T1      Becomes Write Coordinator  Becomes Write Coordinator
        Sets version = (T=100, ...) Sets version = (T=101, ...)
T2      Replicates to Node 2      Replicates to Node 1
T3      Node 2 receives T=100     Node 1 receives T=101
        Compares: T=101 > T=100    Compares: T=101 > T=100
        Keeps T=101 (newer)        Keeps T=101 (newer)
```

**Result**: Both writes succeed, but the one with later timestamp wins. No lost updates (we can detect conflicts).

### Types of Global Clocks

#### 1. **TrueTime (Google Spanner)**
- Uses GPS + atomic clocks
- Provides bounded uncertainty
- Guarantees: `TT.now().earliest <= real_time <= TT.now().latest`
- **Pros**: Very accurate, globally consistent
- **Cons**: Requires special hardware, complex

#### 2. **Logical Clocks (Lamport Timestamps)**
- Each node maintains a logical clock
- Increment on local events
- Send clock value with messages
- **Pros**: Simple, no hardware needed
- **Cons**: Doesn't reflect real time, only causality

#### 3. **Vector Clocks**
- Track causality between all nodes
- Each node maintains vector of clocks
- **Pros**: Can detect all causal relationships
- **Cons**: More complex, vector size = number of nodes

#### 4. **Hybrid Logical Clocks (HLC)**
- Combines physical time + logical clock
- Better than pure logical clocks
- **Pros**: Good balance of accuracy and simplicity
- **Cons**: Still requires clock synchronization

## Would a Global Clock Fix the Problem?

### Short Answer: **Partially, but with trade-offs**

### What It Would Fix:
1. ✅ **Conflict Resolution**: Can determine which write is "newer"
2. ✅ **No Lost Updates**: Can detect conflicts and handle them
3. ✅ **Causality**: Can order events correctly

### What It Wouldn't Fix:
1. ❌ **Still Need Coordination**: Two simultaneous writes still need resolution
2. ❌ **Network Partitions**: Clocks can drift during partitions
3. ❌ **Clock Skew**: Clocks on different nodes may not be perfectly synchronized

### What You'd Still Need:
1. **Conflict Detection**: Compare timestamps/versions
2. **Conflict Resolution**: Decide which write wins (or merge)
3. **Coordination**: Even with clocks, you need some coordination

## Better Solutions for Leaderless Mode

### Option 1: Distributed Locking (What We Discussed Earlier)
- Use DynamoDB or similar for locks
- Only one node can write to a key at a time
- **Pros**: Simple, prevents conflicts
- **Cons**: Additional service dependency, latency

### Option 2: Last-Write-Wins (LWW) with Timestamps
- Use timestamps to determine "last" write
- Accept that some writes may be lost
- **Pros**: Simple, fast
- **Cons**: Can lose updates

### Option 3: Conflict-Free Replicated Data Types (CRDTs)
- Use data structures that automatically merge
- No conflicts possible
- **Pros**: No coordination needed
- **Cons**: Limited to specific data types

### Option 4: Quorum with Version Vectors
- Use version vectors to track causality
- Require quorum for writes
- **Pros**: Strong consistency guarantees
- **Cons**: More complex, higher latency

## For Your Current Implementation

### Current Behavior (Acceptable for Homework):
- **W=5, R=1**: All nodes must confirm write
- **No explicit locking**: Concurrent writes possible
- **Version numbers**: Help but don't prevent conflicts
- **Last write wins**: Acceptable for this assignment

### What Happens with Concurrent Writes:
1. Two nodes receive writes for same key simultaneously
2. Both become Write Coordinators
3. Both replicate to all nodes
4. Last one to finish "wins"
5. One update is lost (but this is acceptable for the assignment)

### Why This is OK:
- The assignment focuses on **demonstrating inconsistency windows**
- Concurrent writes are a known limitation
- The system still works correctly for sequential writes
- Load testing will show the behavior

## Recommendation

### For Your Homework:
1. **Keep current implementation** (no global clock needed)
2. **Document the limitation**: Explain that concurrent writes to same key can conflict
3. **Show it works for normal case**: Sequential writes work perfectly
4. **Mention in report**: "Concurrent writes to same key may result in last-write-wins behavior"

### If You Want to Improve (Optional):
1. **Add simple locking**: Use in-memory locks per key (only for same-node concurrency)
2. **Use timestamps**: Add timestamp to version for better conflict resolution
3. **Accept limitation**: Document it clearly

## Summary

**Question**: Would a global clock fix concurrent write conflicts?

**Answer**: 
- **Partially**: It would help resolve conflicts, but wouldn't eliminate the need for coordination
- **Trade-offs**: Adds complexity, clock synchronization issues
- **For homework**: Current implementation is fine - document the limitation
- **For production**: Would need proper conflict resolution (locks, CRDTs, or quorum)

The current leaderless design is **correct for the assignment** - it demonstrates the inconsistency window and eventual consistency, which is the goal!



