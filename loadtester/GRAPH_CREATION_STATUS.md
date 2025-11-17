# Graph Creation Status

## Current Status

✅ **Sample data created**: `results/sample/results.csv` and `results/sample/summary.json`

⚠️ **Python dependencies needed**: `matplotlib` and `numpy` are not installed

## To Create Graphs

### Option 1: Install Python Dependencies

**For macOS (system Python):**
```bash
pip3 install --user matplotlib numpy
```

**Or use a virtual environment (recommended):**
```bash
python3 -m venv venv
source venv/bin/activate
pip install matplotlib numpy
```

**Or with system packages flag (if needed):**
```bash
pip3 install --break-system-packages matplotlib numpy
```

### Option 2: Generate Graphs After Installing

Once matplotlib is installed, run:

```bash
# Generate graphs from sample data
python3 loadtester/visualize.py results/sample

# Or generate graphs from real load test data
python3 loadtester/visualize.py results/<your-test-directory>
```

## What's Ready

1. ✅ **Sample data created** - `results/sample/` contains mock load test data
2. ✅ **Visualization script ready** - `loadtester/visualize.py` is ready to use
3. ✅ **Helper scripts created**:
   - `create_sample_data.py` - Creates sample data for testing
   - `create_all_graphs.sh` - Creates graphs for all existing results
   - `run_test_and_visualize.sh` - Runs test and creates graphs automatically

## Graph Locations

Once graphs are created, they will be saved as:
- `results/sample/latency_distribution.png`
- `results/sample/time_intervals.png`

## Next Steps

1. **Install matplotlib** (see options above)
2. **Run visualization**:
   ```bash
   python3 loadtester/visualize.py results/sample
   ```
3. **View graphs**:
   ```bash
   open results/sample/latency_distribution.png
   open results/sample/time_intervals.png
   ```

## For Real Load Test Data

To create graphs from actual load tests:

1. **Start database nodes**
2. **Run load test**:
   ```bash
   ./loadtester-bin \
     --config=loadtester/configs/lf_w5_r1_10_90.json \
     --duration=60s \
     --output=results/my_test
   ```
3. **Generate graphs**:
   ```bash
   python3 loadtester/visualize.py results/my_test
   ```

## Summary

- ✅ Sample data: Ready
- ✅ Scripts: Ready  
- ⚠️ Dependencies: Need to install matplotlib/numpy
- 📊 Graphs: Will be created after installing dependencies





