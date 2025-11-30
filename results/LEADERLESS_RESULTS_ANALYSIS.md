# Leaderless Load Test Results Analysis

## Test Summary

| Ratio | Total Requests | Success Rate | Error Rate | Stale Reads | Write Latency | Read Latency |
|-------|---------------|--------------|------------|-------------|---------------|--------------|
| **1%/99%** | 5,453 | 10.0% | 90.0% | **0** | ~307ms | ~1.6ms |
| **10%/90%** | 5,463 | 41.6% | 58.4% | **0** | ~307ms | ~1.9ms |
| **50%/50%** | 3,816 | 89.1% | 10.9% | **0** | ~310ms | ~1.7ms |
| **90%/10%** | 2,159 | 99.1% | 0.9% | **0** | ~312ms | ~2.6ms |

## ✅ Expected Results

### 1. **Error Rates** - ✅ As Expected
- **1%/99%**: 90% error rate - Expected! Most keys never written
- **10%/90%**: 58.4% error rate - Expected! Many keys don't exist
- **50%/50%**: 10.9% error rate - Good! Most keys exist
- **90%/10%**: 0.9% error rate - Excellent! Almost all keys exist

### 2. **Write Latency** - ✅ As Expected
- All ratios: ~307-312ms
- **Expected**: W=5 means must wait for all 5 nodes
- Consistent across all ratios (good!)

### 3. **Read Latency** - ✅ As Expected
- All ratios: ~1.6-2.6ms
- **Expected**: R=1 means local read, very fast
- Slightly higher in 90%/10% (more writes = more contention)

### 4. **Success Rates** - ✅ As Expected
- Correlates with write ratio (more writes = more keys exist)
- Pattern is correct

## ⚠️ **UNUSUAL FINDING: Zero Stale Reads**

### The Issue
**All tests show 0 stale reads**, which is unexpected for leaderless mode with R=1.

### Why This Might Happen

#### 1. **All Requests Go to Same Node**
- Current setup: Load tester sends all requests to `localhost:8080` (Node 1)
- **Problem**: If all writes and reads go to the same node, no stale reads possible!
- **Solution**: Need to distribute requests across all 5 nodes

#### 2. **Replication Too Fast**
- Replication completes before reads happen
- With 200ms delay per node, replication takes ~800ms (4 nodes × 200ms)
- But reads might happen after replication completes

#### 3. **Stale Read Detection Logic**
- Client tracks latest known version
- If reads always go to coordinator (Node 1), it always has latest version
- Need to read from different nodes to detect staleness

### How to Verify

Check the CSV files to see which nodes are being accessed:

```bash
# Check if all requests go to same address
grep -o "localhost:[0-9]*" results/ll_local_ll_50_50/results.csv | sort | uniq -c
```

If all requests show `localhost:8080`, that's the problem!

## 🔍 Detailed Analysis by Ratio

### 1% Writes / 99% Reads
- **Error Rate**: 90% - ✅ Expected (most keys never written)
- **Stale Reads**: 0 - ⚠️ Unexpected (but might be due to single-node access)
- **Latency**: Fast reads (~1.6ms), slow writes (~307ms) - ✅ Expected

### 10% Writes / 90% Reads
- **Error Rate**: 58.4% - ✅ Expected
- **Stale Reads**: 0 - ⚠️ Unexpected
- **Latency**: Consistent with other ratios - ✅ Expected

### 50% Writes / 50% Reads
- **Error Rate**: 10.9% - ✅ Good (most keys exist)
- **Stale Reads**: 0 - ⚠️ Unexpected (should see some with balanced workload)
- **Latency**: Slightly higher write latency (~310ms) - ✅ Expected (more contention)

### 90% Writes / 10% Reads
- **Error Rate**: 0.9% - ✅ Excellent
- **Stale Reads**: 0 - ⚠️ Unexpected (but fewer reads = less chance)
- **Latency**: Higher read latency (~2.6ms) - ✅ Expected (more writes = more contention)

## 📊 Performance Patterns

### Write Latency
- **Consistent**: ~307-312ms across all ratios
- **Why**: W=5 always requires waiting for all nodes
- **Good**: No degradation under load

### Read Latency
- **Very Fast**: ~1.6-2.6ms
- **Why**: R=1 means local read, no coordination
- **Pattern**: Slightly higher with more writes (contention)

### Error Rates
- **Pattern**: Correlates perfectly with write ratio
- **1%/99%**: 90% errors (expected - most keys don't exist)
- **90%/10%**: 0.9% errors (excellent - almost all keys exist)

## 🎯 Recommendations

### 1. **Fix Stale Read Detection** (Most Important)

**Problem**: All requests likely go to same node (localhost:8080)

**Solution**: Update load tester to distribute requests across all 5 nodes:
- Round-robin or random node selection
- Track which node handled each request
- Read from different nodes than write coordinator

**Expected Result**: Should see stale reads, especially in 50%/50% ratio

### 2. **Verify Node Distribution**

Check if requests are distributed:
```bash
# Count requests per node (if we track it)
# Or check if all go to same address
```

### 3. **Increase Load/Concurrency**

- Current: 10 concurrent workers
- Try: 20-50 concurrent workers
- More concurrent requests = more chance of stale reads

### 4. **Reduce Replication Delay** (for testing)

- Current: 200ms delay per node
- Try: 50ms delay (faster replication = harder to catch stale reads)
- Or: Keep 200ms but increase concurrency

## 🔄 LocalStack vs Docker Compose

### Would LocalStack Produce Different Results?

**Short Answer**: **Probably NOT significantly different** for these metrics.

### Why:

1. **Same Application Code**
   - Same Go code running
   - Same replication logic
   - Same delays (200ms, 100ms, 50ms)

2. **Same Network Characteristics**
   - LocalStack containers still run locally
   - Network latency is minimal (same machine)
   - No real network delays

3. **Same Load Tester**
   - Same client code
   - Same request patterns
   - Same concurrency

### What WOULD Be Different:

1. **Container Overhead**
   - ECS Fargate might have slightly more overhead
   - But minimal impact on latency

2. **Network Latency** (if on real AWS)
   - Real AWS deployment would have network latency
   - But LocalStack runs locally, so no difference

3. **Resource Constraints**
   - ECS might have different CPU/memory limits
   - But LocalStack emulates locally, so similar

### When LocalStack Would Matter:

1. **Testing AWS-Specific Features**
   - Service discovery (Cloud Map)
   - Load balancer behavior
   - IAM roles, etc.

2. **Testing Deployment Process**
   - Terraform deployment
   - Container registry
   - ECS task definitions

3. **Testing at Scale**
   - Real AWS network conditions
   - Real resource constraints
   - Real multi-AZ deployment

### Recommendation:

**For Load Testing**: Docker Compose is sufficient
- Same application behavior
- Same performance characteristics
- Easier to debug and iterate

**For Deployment Testing**: LocalStack is useful
- Test Terraform configurations
- Test AWS service integration
- Validate deployment process

## 📝 Summary

### ✅ What's Working Well:
1. Error rates match expectations
2. Latency patterns are correct
3. Success rates correlate with write ratios
4. System handles load correctly

### ⚠️ What Needs Attention:
1. **Zero stale reads** - Likely because all requests go to same node
2. Need to distribute requests across all 5 nodes
3. Need to verify stale read detection logic

### 🎯 Next Steps:
1. Update load tester to use multiple nodes
2. Re-run tests and verify stale reads appear
3. Document findings in report

### 🔄 LocalStack:
- **Not necessary** for load testing (Docker Compose is sufficient)
- **Useful** for testing AWS deployment process
- **Results would be similar** (same code, same network conditions)



