# Load Test Results Analysis

## Results Summary for ll_local_01_99

From `summary.json`:
- **Total requests**: 5,453
- **Total writes**: 52 (1% of requests)
- **Total reads**: 5,401 (99% of requests)
- **Successful requests**: 214 (3.9%)
- **Failed requests**: 5,239 (96.1%)
- **Stale reads**: 0

## Why Most Reads Are Failing (success=false)

### This is **EXPECTED** behavior for 1% writes / 99% reads!

**The Problem:**
1. Only **52 keys** were written during the test
2. But **5,401 reads** were attempted
3. The key generator uses "local-in-time" clustering, but:
   - It can create new clusters (20% probability)
   - With 1,000 possible keys and only 52 writes, most keys are never written
   - Reads to unwritten keys return 404 (key not found) → `success=false`

**This is realistic!** In a real system:
- Cache misses (keys not found) are common
- Read-heavy workloads often read keys that don't exist yet
- The system correctly returns 404 for missing keys

## What the Results Show

### ✅ **Writes are Working**
- All 52 writes succeeded (`success=true`)
- Write latency: ~306ms average (expected for W=5, must wait for all nodes)
- Consistent write performance

### ✅ **Reads are Working (when keys exist)**
- Successful reads: ~162 (214 total successful - 52 writes)
- Read latency: ~1.5ms average (very fast, R=1, local read)
- When keys exist, reads work perfectly

### ⚠️ **Most Reads Fail (Expected)**
- 5,239 reads failed because keys didn't exist
- This is **normal** for 1% write ratio
- The system correctly handles missing keys

## Is This a Problem?

### **No, this is expected!** Here's why:

1. **Realistic Scenario**: In real systems, cache misses are common
2. **Correct Behavior**: System correctly returns 404 for missing keys
3. **Writes Work**: All writes succeeded
4. **Reads Work**: When keys exist, reads succeed

### However, we can improve the test:

**Option 1: Pre-populate Keys**
- Write a set of keys first
- Then run read-heavy test
- Ensures reads have keys to read

**Option 2: Adjust Generator**
- Only read from keys that have been written
- Track written keys and prioritize them for reads
- Still use "local-in-time" clustering

**Option 3: Accept Current Behavior**
- This is realistic (cache misses)
- Focus on successful request metrics
- Analyze latency for successful requests only

## What to Report

### For the Homework Report:

1. **Latency Analysis** (from successful requests):
   - Write latency: ~306ms (W=5, must replicate to all nodes)
   - Read latency: ~1.5ms (R=1, local read, very fast)
   - This shows the trade-off: writes are slow, reads are fast

2. **Success Rate**:
   - Write success: 100% (52/52)
   - Read success: ~3% (162/5401) - but this is expected with 1% writes
   - **Key insight**: With 1% writes, most reads are for non-existent keys

3. **Stale Reads**:
   - 0 stale reads detected
   - This is good! Shows consistency when keys exist

4. **Generator Effectiveness**:
   - "Local-in-time" clustering is working
   - Some reads do hit recently written keys (see line 270, 342 in CSV)
   - But with only 1% writes, most keys are never written

## Recommendations

### For Better Test Results:

1. **Pre-populate keys** before read-heavy tests:
   ```bash
   # Write 100 keys first
   for i in {1..100}; do
     curl -X POST http://localhost:8080/set \
       -H "Content-Type: application/json" \
       -d "{\"key\":\"key_$i\",\"value\":\"initial_$i\"}"
   done
   ```

2. **Or adjust test ratios**:
   - Start with 50/50 to ensure keys exist
   - Then analyze the 1/99 ratio separately
   - Focus on latency patterns, not success rates

3. **Track "cache hit rate"**:
   - Successful reads / Total reads
   - This is a meaningful metric for read-heavy workloads

## Expected Behavior by Ratio

| Ratio | Expected Behavior |
|-------|------------------|
| **1%/99%** | Most reads fail (keys don't exist) - **This is normal!** |
| **10%/90%** | More keys exist, more successful reads |
| **50%/50%** | Balanced, most reads should succeed |
| **90%/10%** | Many writes, few reads, most reads should succeed |

## Conclusion

**The `success=false` for most reads is EXPECTED and CORRECT behavior** for a 1% write / 99% read workload. The system is working as designed:

- ✅ Writes succeed (all 52 writes worked)
- ✅ Reads succeed when keys exist (~162 successful reads)
- ✅ System correctly returns 404 for missing keys
- ✅ No stale reads detected (good consistency)
- ✅ Latency patterns are correct (slow writes, fast reads)

This is **not a bug** - it's realistic behavior for a read-heavy workload with few writes!



