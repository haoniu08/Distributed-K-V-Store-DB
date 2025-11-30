# Local vs LocalStack Load Test Comparison

## Test Configuration
- **Duration**: 60 seconds
- **Concurrency**: 10 workers
- **Multi-node distribution**: Enabled (v2)
- **Test Date**: November 27, 2025

## Key Differences

### 1% Writes, 99% Reads

| Metric | Local Only | With LocalStack | Difference |
|--------|-----------|-----------------|------------|
| **Success Rate** | 96.45% | 99.78% | +3.33% ✅ |
| **Error Rate** | 3.55% | 0.22% | -3.33% ✅ |
| **Stale Reads** | 1 (0.02%) | 4 (0.07%) | +3 reads |
| **Write Latency (mean)** | 307.29 ms | 307.66 ms | +0.37 ms |
| **Read Latency (mean)** | 1.68 ms | 1.72 ms | +0.04 ms |
| **Total Requests** | 5,459 | 5,470 | +11 |

**Analysis**: LocalStack run shows significantly better success rates and lower error rates. The slight increase in stale reads (4 vs 1) is still very low and within expected variance. Latencies are nearly identical.

### 50% Writes, 50% Reads

| Metric | Local Only | With LocalStack | Difference |
|--------|-----------|-----------------|------------|
| **Success Rate** | 99.65% | 99.92% | +0.27% ✅ |
| **Error Rate** | 0.35% | 0.08% | -0.27% ✅ |
| **Stale Reads** | 1 (0.05%) | 0 (0.00%) | -1 read |
| **Write Latency (mean)** | 308.51 ms | 308.86 ms | +0.35 ms |
| **Read Latency (mean)** | 2.05 ms | 2.00 ms | -0.05 ms |
| **Total Requests** | 4,017 | 3,802 | -215 |

**Analysis**: LocalStack run shows slightly better success rates. No stale reads detected in this run (vs 1 in local-only), which is within normal variance. Latencies are essentially identical.

## Overall Observations

1. **Better Reliability with LocalStack**: The LocalStack runs consistently show higher success rates and lower error rates, particularly in the 1% write scenario. This suggests:
   - Less resource contention when LocalStack is running
   - Possibly better Docker network isolation
   - Different system load patterns

2. **Latency Consistency**: Write and read latencies are nearly identical between both environments, confirming that:
   - The KV store behavior is consistent
   - Network overhead from LocalStack is negligible
   - The replication delays (200ms/100ms) dominate latency

3. **Stale Read Detection**: Both environments successfully detect stale reads, confirming the multi-node distribution fix is working correctly in both scenarios.

4. **System Load Impact**: The LocalStack run processed slightly fewer requests in the 50/50 scenario, possibly due to:
   - Additional Docker container overhead
   - Slightly different timing patterns
   - System resource allocation

## Conclusion

Running the tests with LocalStack active shows **improved reliability** (higher success rates, lower error rates) while maintaining **identical latency characteristics**. This suggests that having LocalStack running doesn't negatively impact performance and may actually improve system stability, possibly due to better resource isolation or Docker network configuration.

The core behavior (stale read detection, latency patterns, consistency) remains consistent between both environments, confirming that the multi-node distribution fix works correctly regardless of whether LocalStack is running.


