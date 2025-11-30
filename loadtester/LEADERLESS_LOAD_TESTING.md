# Leaderless Load Testing Guide

## Quick Start

### 1. Start Leaderless Cluster
```bash
docker compose -f docker-compose-leaderless.yml up -d

# Verify all nodes are running
docker compose -f docker-compose-leaderless.yml ps

# Test health
curl http://localhost:8080/health
```

### 2. Run All Load Tests
```bash
cd loadtester
chmod +x run_leaderless_tests.sh
./run_leaderless_tests.sh 60s 10 results/ll_local
```

This will:
- Run all 4 read-write ratios (1%/99%, 10%/90%, 50%/50%, 90%/10%)
- Generate results in `results/ll_local_*` directories
- Create graphs automatically

### 3. View Results
```bash
# Check summary
cat results/ll_local_ll_01_99/summary.json

# View graphs
open results/ll_local_ll_01_99/latency_distribution.png
open results/ll_local_ll_01_99/time_intervals.png
```

## Current Setup

### Load Tester Status
- ✅ Load tester code exists
- ✅ Config files for leaderless exist
- ✅ Visualization script exists
- ⚠️ Currently targets single node (localhost:8080)

### For Leaderless Mode
The load tester should distribute requests across all 5 nodes:
- **Writes**: Can go to any node (round-robin or random)
- **Reads**: Can go to any node (round-robin or random)

## Two Approaches

### Approach 1: Use Single Node (Current)
- **Pros**: Simple, works immediately
- **Cons**: Doesn't test true leaderless behavior (all nodes receiving requests)
- **Use Case**: Quick testing, basic validation

### Approach 2: Distribute Across All Nodes (Recommended)
- **Pros**: Tests true leaderless behavior, more realistic
- **Cons**: Requires updating load tester
- **Use Case**: Full testing, matches homework requirements

## Implementation Options

### Option A: Update Config to Support Multiple Nodes
Add `target_addrs` array to config:
```json
{
  "name": "Leaderless W=5 R=1 (1% writes, 99% reads)",
  "target_addrs": [
    "localhost:8080",
    "localhost:8082",
    "localhost:8083",
    "localhost:8084",
    "localhost:8085"
  ],
  "write_ratio": 0.01,
  "read_ratio": 0.99
}
```

### Option B: Use Load Balancer
- Set up a simple load balancer (nginx, HAProxy)
- Point load tester to load balancer
- Load balancer distributes to all nodes

### Option C: Update Load Tester Code
- Modify `loadtester/main.go` to support multiple target addresses
- Round-robin or random selection
- Track which node handled each request

## Recommended: Quick Test with Current Setup

You can test immediately with the current setup:

```bash
# 1. Start nodes
docker compose -f docker-compose-leaderless.yml up -d

# 2. Run test (targets node1 on port 8080)
go run loadtester/main.go \
  -config loadtester/configs/ll_01_99.json \
  -duration 60s \
  -concurrency 10 \
  -output results/ll_local_01_99

# 3. Generate graph
python3 loadtester/visualize.py results/ll_local_01_99
```

This will work, but all requests go to node1. For true leaderless testing, we should distribute across all nodes.

## Next Steps

1. **Quick Test**: Use current setup to validate everything works
2. **Full Test**: Update load tester to distribute across all 5 nodes
3. **LocalStack**: Optional, if you want AWS-like environment

Would you like me to:
1. Update the load tester to support multiple nodes?
2. Create a simple load balancer setup?
3. Just run tests with current setup first?



