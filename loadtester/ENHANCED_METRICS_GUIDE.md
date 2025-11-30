# Enhanced Load Testing Metrics Guide

## Overview

The load tester now tracks and reports comprehensive metrics including:
- ✅ Latency (reads and writes)
- ✅ Time intervals between same-key operations
- ✅ **Stale read counts and rates**
- ✅ **Error/failure counts and rates**

## New Metrics Added

### 1. Stale Read Metrics
- **`stale_reads`**: Total number of stale reads detected
- **`stale_read_rate_percent`**: Percentage of all reads that were stale
- **`stale_read_rate_of_successful_percent`**: Percentage of successful reads that were stale (most meaningful)
- **`successful_reads`**: Total number of successful reads

### 2. Error/Failure Metrics
- **`success_rate_percent`**: Percentage of requests that succeeded
- **`error_rate_percent`**: Percentage of requests that failed
- **`successful_requests`**: Total successful requests
- **`failed_requests`**: Total failed requests

## How to Run Tests

### Run All Leaderless Tests

```bash
cd loadtester
./run_leaderless_tests.sh 60s 10 ../results/ll_local
```

This will:
1. Run all 4 test configurations (1%/99%, 10%/90%, 50%/50%, 90%/10%)
2. Generate graphs with stale read and error rate annotations
3. Generate comprehensive reports for each test

### Run Single Test

```bash
# Run test
go run loadtester/main.go \
  -config loadtester/configs/ll_01_99.json \
  -duration 60s \
  -concurrency 10 \
  -output results/ll_local_01_99

# Generate graph
python3 loadtester/visualize.py results/ll_local_01_99

# Generate report
python3 loadtester/generate_report.py results/ll_local_01_99
```

## Understanding the Metrics

### Stale Reads

**What is a stale read?**
- A read that returns a version older than the latest known version
- Indicates inconsistency window in leaderless mode
- Expected behavior: reads from nodes that haven't received the latest write yet

**Example:**
```
1. Write key_A=value1 (version=5) to Node 1
2. Node 1 replicates to other nodes (takes time)
3. Read key_A from Node 2 (before replication completes)
4. Node 2 returns version=4 (stale!)
```

**What the metrics show:**
- `stale_read_rate_of_successful_percent`: Most meaningful - shows % of successful reads that were stale
- Higher rate = more inconsistency windows detected
- Expected for leaderless mode with R=1

### Error Rates

**What are errors?**
- Failed requests (404 for missing keys, network errors, etc.)
- For read-heavy workloads (1%/99%), many reads fail because keys don't exist
- This is **expected** and **normal** behavior

**What the metrics show:**
- `error_rate_percent`: Overall error rate
- For 1%/99% ratio: Expect high error rate (most keys never written)
- For 50%/50% ratio: Expect low error rate (most keys exist)

## Output Files

### 1. `summary.json`
Contains all metrics:
```json
{
  "config": "Leaderless W=5 R=1 (1% writes, 99% reads)",
  "total_requests": 5453,
  "success_rate_percent": 3.92,
  "error_rate_percent": 96.08,
  "stale_reads": 0,
  "stale_read_rate_percent": 0.0,
  "stale_read_rate_of_successful_percent": 0.0,
  "successful_reads": 162,
  ...
}
```

### 2. `results.csv`
Raw data with columns:
- `timestamp`: When request occurred
- `type`: "read" or "write"
- `key`: Key name
- `latency_ms`: Request latency
- `success`: true/false
- `is_stale`: true/false (for reads)
- `version`: Version number returned

### 3. `latency_distribution.png`
Graph showing:
- Read latency distribution (with P95/P99)
- Write latency distribution (with P95/P99)
- **Stale read rate** annotation on read graph
- **Error rate** annotation on write graph

### 4. `time_intervals.png`
Graph showing time intervals between writes and reads of the same key

### 5. Console Report
Comprehensive text report showing:
- Overall statistics
- Success/error rates
- Stale read metrics
- Latency statistics
- Key insights

## Expected Results by Ratio

### 1% Writes / 99% Reads
- **Error Rate**: High (~96%) - most keys never written
- **Stale Reads**: Low (few writes, less chance of inconsistency)
- **Read Latency**: Very low (~1-2ms, R=1)
- **Write Latency**: High (~300ms, W=5)

### 10% Writes / 90% Reads
- **Error Rate**: Moderate (~50-70%)
- **Stale Reads**: Moderate (more writes = more inconsistency windows)
- **Read Latency**: Low (~1-2ms)
- **Write Latency**: High (~300ms)

### 50% Writes / 50% Reads
- **Error Rate**: Low (~5-10%)
- **Stale Reads**: Higher (balanced workload, more inconsistency windows)
- **Read Latency**: Low (~1-2ms)
- **Write Latency**: High (~300ms)

### 90% Writes / 10% Reads
- **Error Rate**: Very low (~1-2%)
- **Stale Reads**: Lower (few reads, less chance to detect stale)
- **Read Latency**: Low (~1-2ms)
- **Write Latency**: High (~300ms)

## Interpreting Results

### Good Results:
- ✅ Stale reads detected (shows inconsistency windows working)
- ✅ Low read latency (R=1 is fast)
- ✅ High write latency (W=5 is slow, expected)
- ✅ Error rates match expectations for ratio

### Concerning Results:
- ⚠️ No stale reads detected (might need more load or different timing)
- ⚠️ Very high error rate for balanced ratios (might indicate issues)
- ⚠️ Unexpected latency patterns

## Next Steps

1. **Run all 4 test configurations**
2. **Review reports** for each configuration
3. **Compare metrics** across different ratios
4. **Analyze patterns**:
   - How stale read rate changes with write ratio
   - How error rate changes with write ratio
   - Latency patterns for reads vs writes

## Example Report Output

```
================================================================================
LOAD TEST REPORT: Leaderless W=5 R=1 (1% writes, 99% reads)
================================================================================

📊 OVERALL STATISTICS
--------------------------------------------------------------------------------
Total Requests:        5,453
Total Writes:          52
Total Reads:           5,401
Duration:              59.996054833s

✅ SUCCESS & ERROR RATES
--------------------------------------------------------------------------------
Successful Requests:   214 (3.92%)
Failed Requests:       5,239 (96.08%)

🔄 STALE READS
--------------------------------------------------------------------------------
Total Stale Reads:     0
Stale Read Rate:       0.00% (of all reads)
Stale Read Rate:       0.00% (of successful reads)
Successful Reads:      162

⏱️  LATENCY STATISTICS
--------------------------------------------------------------------------------
Write Latency:
  Mean:    306.05 ms
  P95:     309.04 ms
  P99:     309.47 ms

Read Latency:
  Mean:    1.52 ms
  P95:     2.16 ms
  P99:     3.39 ms
```



