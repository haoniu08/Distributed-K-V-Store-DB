# Load Testing Plan for Leaderless Mode

## Overview

This plan covers load testing the leaderless KV store in two environments:
1. **Pure Local (Docker Compose)** - Recommended, easier to set up
2. **LocalStack** - Optional, if ECS is available

## Requirements from Homework10.md

- Test 4 read-write ratios: 1%/99%, 10%/90%, 50%/50%, 90%/10%
- Record latency for each request
- Record number of stale reads
- Create graphs:
  1. Distribution of latency for reads and writes (show long tail)
  2. Distribution of time intervals between reading/writing same key
- Ensure "local-in-time" key generation

## Current Status

✅ **Already Have:**
- Load tester code (`loadtester/`)
- Config files for leaderless (`ll_01_99.json`, `ll_10_90.json`, `ll_50_50.json`, `ll_90_10.json`)
- Visualization script (`visualize.py`)
- Docker Compose setup (`docker-compose-leaderless.yml`)
- Generator with "local-in-time" key clustering

## Plan: Option 1 - Pure Local (Docker Compose) ⭐ Recommended

### Why This Approach?
- ✅ Simplest to set up
- ✅ No AWS/LocalStack dependencies
- ✅ Faster iteration
- ✅ Matches your current working setup
- ✅ All 5 nodes accessible on localhost

### Steps

#### 1. Start Leaderless Cluster
```bash
# Start all 5 nodes
docker compose -f docker-compose-leaderless.yml up -d

# Verify all nodes are running
docker compose -f docker-compose-leaderless.yml ps

# Test health
curl http://localhost:8080/health
curl http://localhost:8082/health
curl http://localhost:8084/health
```

#### 2. Update Load Tester Configs

The load tester needs to know it can send requests to any node. For leaderless:
- **Writes**: Can go to any node (that node becomes Write Coordinator)
- **Reads**: Can go to any node (returns local value, R=1)

Update configs to use round-robin or random node selection.

#### 3. Run Load Tests

For each of the 4 ratios:

```bash
# 1% writes, 99% reads
go run loadtester/main.go \
  -config loadtester/configs/ll_01_99.json \
  -duration 60s \
  -concurrency 10 \
  -output results/ll_local_01_99

# 10% writes, 90% reads
go run loadtester/main.go \
  -config loadtester/configs/ll_10_90.json \
  -duration 60s \
  -concurrency 10 \
  -output results/ll_local_10_90

# 50% writes, 50% reads
go run loadtester/main.go \
  -config loadtester/configs/ll_50_50.json \
  -duration 60s \
  -concurrency 10 \
  -output results/ll_local_50_50

# 90% writes, 10% reads
go run loadtester/main.go \
  -config loadtester/configs/ll_90_10.json \
  -duration 60s \
  -concurrency 10 \
  -output results/ll_local_90_10
```

#### 4. Generate Graphs

```bash
# Generate all graphs
cd loadtester
python3 visualize.py ../results/ll_local_01_99
python3 visualize.py ../results/ll_local_10_90
python3 visualize.py ../results/ll_local_50_50
python3 visualize.py ../results/ll_local_90_10
```

#### 5. Analyze Results

- Check `summary.json` for:
  - Total requests
  - Average latency (reads/writes)
  - P95, P99 latencies
  - Number of stale reads
  - Stale read percentage

## Plan: Option 2 - LocalStack (Optional)

### Why This Approach?
- Tests in a more AWS-like environment
- Good for learning AWS ECS
- **But**: ECS/ECR may not be fully available in Community Edition

### Prerequisites

1. **LocalStack Pro/Student Edition** (for full ECS support)
   - Sign up: https://www.localstack.cloud/localstack-for-students
   - Or use Community Edition (limited ECS support)

2. **Terraform Configuration for LocalStack**
   - Adapt existing Terraform to use LocalStack endpoints
   - Or use simpler approach: deploy containers directly

### Steps

#### Option 2A: LocalStack with ECS (If Available)

1. **Configure Terraform for LocalStack**:
   ```hcl
   provider "aws" {
     endpoints {
       ecs = "http://localhost:4566"
       ecr = "http://localhost:4566"
       # ... other endpoints
     }
     skip_credentials_validation = true
     skip_metadata_api_check     = true
   }
   ```

2. **Deploy to LocalStack**:
   ```bash
   export AWS_ENDPOINT_URL=http://localhost:4566
   terraform apply
   ```

3. **Run Load Tests** against LocalStack ALB endpoint

#### Option 2B: LocalStack with Direct Containers (Simpler)

1. **Start LocalStack** (for other services if needed)
2. **Use Docker Compose** for the actual nodes (same as Option 1)
3. **Run Load Tests** against Docker Compose endpoints

**Note**: This is essentially the same as Option 1, just with LocalStack running in the background.

## Recommended Approach

**Use Option 1 (Pure Local/Docker Compose)** because:
1. ✅ Already working
2. ✅ Simpler setup
3. ✅ Faster testing
4. ✅ No AWS dependencies
5. ✅ Meets all requirements

**Use Option 2 (LocalStack)** only if:
- You want to test AWS-specific features
- You need to demonstrate AWS deployment
- You have LocalStack Pro/Student with full ECS support

## Implementation Checklist

### Phase 1: Update Load Tester for Leaderless

- [ ] Update load tester to support multiple target nodes (round-robin)
- [ ] Ensure writes can go to any node
- [ ] Ensure reads can go to any node
- [ ] Track which node handled each request (for stale read detection)

### Phase 2: Run Tests

- [ ] Test 1% writes / 99% reads
- [ ] Test 10% writes / 90% reads
- [ ] Test 50% writes / 50% reads
- [ ] Test 90% writes / 10% reads

### Phase 3: Generate Results

- [ ] Generate latency distribution graphs (reads)
- [ ] Generate latency distribution graphs (writes)
- [ ] Generate time interval graphs
- [ ] Calculate stale read statistics

### Phase 4: Analysis

- [ ] Document how generator ensures "local-in-time" keys
- [ ] Analyze latency patterns
- [ ] Analyze stale read patterns
- [ ] Compare results across different ratios

## Key Metrics to Track

1. **Latency**:
   - Average read latency
   - Average write latency
   - P95, P99 percentiles
   - Long tail analysis

2. **Stale Reads**:
   - Total stale reads
   - Stale read percentage
   - Which nodes returned stale data

3. **Time Intervals**:
   - Distribution of intervals between same-key operations
   - Clustering effectiveness

4. **Throughput**:
   - Requests per second
   - Successful vs failed requests

## Expected Behavior

### Leaderless (W=5, R=1):
- **Writes**: High latency (must wait for all 5 nodes)
- **Reads**: Low latency (returns local value immediately)
- **Stale Reads**: Possible during replication window
- **Write-Heavy (90/10)**: High write latency, few stale reads
- **Read-Heavy (1/99)**: Low read latency, more stale reads possible

## Next Steps

1. **Update load tester** to support leaderless mode properly
2. **Run all 4 test configurations**
3. **Generate graphs**
4. **Write analysis**

Would you like me to:
1. Update the load tester code for leaderless mode?
2. Create a script to run all tests?
3. Set up LocalStack testing (if you want)?



