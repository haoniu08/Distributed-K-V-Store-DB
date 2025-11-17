# Consistency Tests

This directory contains consistency tests for both Leader-Follower and Leaderless database implementations.

## Test Files

- `consistency_test.go` - Test client utilities
- `leader_follower_consistency_test.go` - Leader-Follower consistency tests
- `leaderless_consistency_test.go` - Leaderless consistency tests
- `run_consistency_tests.sh` - Automated test runner
- `test_consistency_manual.sh` - Manual test script using curl

## Running Tests

### Automated Tests (Go tests)

**Run all tests:**
```bash
./tests/run_consistency_tests.sh
```

**Run only Leader-Follower tests:**
```bash
./tests/run_consistency_tests.sh leader-follower
```

**Run only Leaderless tests:**
```bash
./tests/run_consistency_tests.sh leaderless
```

The script will:
1. Build the necessary binaries
2. Start the required nodes
3. Run the Go tests
4. Clean up

### Manual Tests (curl)

**Prerequisites:** Nodes must be running

**For Leader-Follower:**
1. Start Leader-Follower cluster (see LEADER_FOLLOWER_README.md)
2. Run: `./tests/test_consistency_manual.sh`

**For Leaderless:**
1. Start Leaderless cluster (see LEADERLESS_README.md)
2. Run: `./tests/test_consistency_manual.sh`

## Test Cases

### Leader-Follower Tests

1. **Write to Leader, Read from Leader**
   - Write a key-value pair to the Leader
   - After Leader acknowledges, read from Leader
   - **Expected**: Consistent data

2. **Write to Leader, Read from Follower**
   - Write a key-value pair to the Leader
   - After Leader acknowledges, read from a Follower
   - **Expected**: Consistent data (after replication)

3. **Write to Leader, Immediate Local Read from Followers**
   - Write a key-value pair to the Leader
   - Immediately use `local_read` on Followers (within update window)
   - **Expected**: May show inconsistency (key not found or old value)
   - After waiting, all should be consistent

4. **High Load Test**
   - Send multiple writes quickly
   - Check for inconsistencies
   - **Expected**: Some inconsistencies detected at high load

### Leaderless Tests

1. **Write to Random Node, Read from Others (Within Window)**
   - Write to a random node (becomes Write Coordinator)
   - Immediately read from other nodes
   - **Expected**: Inconsistency detected (key not found or old value)

2. **After Coordinator Ack, Read from Coordinator**
   - Write to a random node (becomes Write Coordinator)
   - After Coordinator acknowledges, read from Coordinator
   - **Expected**: Consistent data

3. **After Coordinator Ack, Read from Another Node**
   - Write to a random node (becomes Write Coordinator)
   - After Coordinator acknowledges, wait for replication
   - Read from another node
   - **Expected**: Consistent data

4. **High Load Test**
   - Send multiple writes to random nodes
   - Check for inconsistencies
   - **Expected**: Some inconsistencies detected at high load

## Understanding Results

### Consistency
- **Consistent**: All nodes return the same value with the same version
- **Inconsistent**: Nodes return different values, different versions, or "key not found"

### Inconsistency Window
- **Expected**: During replication, reads may return stale values or "key not found"
- **Acceptable**: This is the expected behavior and demonstrates the inconsistency window
- **After Replication**: All nodes should be consistent

### High Load
- At high load, more inconsistencies are expected
- This demonstrates the trade-off between consistency and performance
- The system should eventually become consistent

## Notes

- Tests use unique keys (timestamp-based) to avoid conflicts
- Tests include appropriate wait times for replication
- Inconsistency detection is logged, not treated as failures (it's expected behavior)
- After replication completes, all nodes should be consistent





