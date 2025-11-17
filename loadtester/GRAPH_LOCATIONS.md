# Graph/Chart Locations

## Where Graphs Are Saved

The `visualize.py` script creates graphs in the **results directory** that you specify when running the script.

### Graph Files Created

When you run:
```bash
python3 loadtester/visualize.py <results_directory>
```

The script creates **two PNG files** in that results directory:

1. **`latency_distribution.png`**
   - Shows latency distribution for reads (left) and writes (right)
   - Includes P95 and P99 percentile markers
   - Histogram format

2. **`time_intervals.png`**
   - Shows distribution of time intervals between writes and reads of the same key
   - Log scale for better visibility
   - Histogram format

### Example

If you run:
```bash
python3 loadtester/visualize.py results/lf_w5_r1_01_99
```

The graphs will be saved as:
- `results/lf_w5_r1_01_99/latency_distribution.png`
- `results/lf_w5_r1_01_99/time_intervals.png`

### Full Path Structure

```
results/
├── lf_w5_r1_01_99/
│   ├── results.csv
│   ├── summary.json
│   ├── latency_distribution.png    ← Graph 1
│   └── time_intervals.png          ← Graph 2
├── lf_w5_r1_10_90/
│   ├── results.csv
│   ├── summary.json
│   ├── latency_distribution.png    ← Graph 1
│   └── time_intervals.png          ← Graph 2
└── ...
```

### Finding Your Graphs

After running load tests and visualization:

1. **Check the output directory** you specified:
   ```bash
   ls -la results/*/
   ```

2. **Or search for PNG files**:
   ```bash
   find results -name "*.png"
   ```

3. **View a specific graph**:
   ```bash
   open results/lf_w5_r1_01_99/latency_distribution.png  # macOS
   # or
   xdg-open results/lf_w5_r1_01_99/latency_distribution.png  # Linux
   ```

### Note

- Graphs are only created **after** you run the visualization script
- You need to run `visualize.py` for each results directory
- The script requires the `results.csv` file to exist in the directory





