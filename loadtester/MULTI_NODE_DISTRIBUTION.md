# Multi-Node Request Distribution

## Changes Made

### 1. Updated Config Structure
- Added `target_addrs` field (array of node addresses)
- Kept `target_addr` for backward compatibility
- Updated all leaderless config files to use `target_addrs` with all 5 nodes

### 2. Updated Load Tester Logic

#### Write Requests:
- **Random node selection**: Writes are distributed randomly across all 5 nodes
- **Coordinator tracking**: The node that handles a write is tracked as the "coordinator" for that key
- **Purpose**: Ensures writes go to different nodes, making them coordinators

#### Read Requests:
- **Smart node selection**: 
  - 80% chance: Read from a different node than the coordinator
  - 20% chance: Read from the coordinator (for comparison)
- **Stale read detection**: When reading from a non-coordinator node, we can detect if it has stale data
- **Purpose**: Maximizes chance of detecting stale reads during replication window

### 3. Node Addresses

All 5 nodes are now accessible:
- `localhost:8080` - Node 1
- `localhost:8081` - Node 2  
- `localhost:8082` - Node 3
- `localhost:8083` - Node 4
- `localhost:8084` - Node 5

## How It Works

### Write Flow:
1. Load tester randomly selects a node (e.g., Node 3)
2. Sends write request to Node 3
3. Node 3 becomes Write Coordinator
4. Node 3 replicates to all other nodes (takes ~800ms with delays)
5. Load tester tracks: `coordinatorTracker[key] = "localhost:8082"`

### Read Flow:
1. Load tester checks if key has a coordinator
2. If yes, 80% chance to read from a different node (e.g., Node 1)
3. Node 1 might not have received replication yet → **stale read detected!**
4. If no coordinator yet, read from any random node

## Expected Results

### Before (Single Node):
- All requests → Node 1
- Writes → Node 1 (coordinator)
- Reads → Node 1 (same node)
- **Result**: 0 stale reads (Node 1 always has latest)

### After (Multi-Node):
- Writes → Random nodes (distributed)
- Reads → Different nodes (80% of time)
- **Result**: Should see stale reads, especially in 50%/50% ratio

## Testing

Run the tests again:

```bash
cd loadtester
./run_leaderless_tests.sh 60s 10 ../results/ll_local_v2
```

**Expected improvements:**
- ✅ Stale reads detected (especially in 50%/50% ratio)
- ✅ Better distribution of load across nodes
- ✅ More realistic leaderless behavior

## Configuration Files Updated

All leaderless configs now use `target_addrs`:
- `ll_01_99.json`
- `ll_10_90.json`
- `ll_50_50.json`
- `ll_90_10.json`

## Backward Compatibility

The load tester still supports `target_addr` (single address) for:
- Leader-Follower tests (which use single leader)
- Backward compatibility

If both `target_addr` and `target_addrs` are provided, `target_addrs` takes precedence.



