#!/usr/bin/env python3
"""
Visualization script for load test results.
Creates graphs showing latency distributions and time intervals.
"""

import json
import csv
import sys
import os
from pathlib import Path
import matplotlib.pyplot as plt
import numpy as np

def load_summary(summary_path):
    """Load summary JSON file."""
    with open(summary_path, 'r') as f:
        return json.load(f)

def load_csv(csv_path):
    """Load CSV results file."""
    records = []
    with open(csv_path, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            records.append(row)
    return records

def plot_latency_distribution(records, output_path, config_name):
    """Plot latency distribution for reads and writes."""
    read_latencies = []
    write_latencies = []
    
    for record in records:
        if record['success'] == 'true':
            latency_ms = float(record['latency_ms'])
            if record['type'] == 'read':
                read_latencies.append(latency_ms)
            elif record['type'] == 'write':
                write_latencies.append(latency_ms)
    
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 6))
    
    # Read latency distribution
    if read_latencies:
        ax1.hist(read_latencies, bins=50, alpha=0.7, color='blue', edgecolor='black')
        ax1.set_xlabel('Latency (ms)')
        ax1.set_ylabel('Frequency')
        ax1.set_title(f'Read Latency Distribution\n{config_name}')
        ax1.grid(True, alpha=0.3)
        ax1.axvline(np.percentile(read_latencies, 95), color='red', linestyle='--', label='P95')
        ax1.axvline(np.percentile(read_latencies, 99), color='orange', linestyle='--', label='P99')
        ax1.legend()
    else:
        ax1.text(0.5, 0.5, 'No read data', ha='center', va='center')
        ax1.set_title('Read Latency Distribution')
    
    # Write latency distribution
    if write_latencies:
        ax2.hist(write_latencies, bins=50, alpha=0.7, color='green', edgecolor='black')
        ax2.set_xlabel('Latency (ms)')
        ax2.set_ylabel('Frequency')
        ax2.set_title(f'Write Latency Distribution\n{config_name}')
        ax2.grid(True, alpha=0.3)
        ax2.axvline(np.percentile(write_latencies, 95), color='red', linestyle='--', label='P95')
        ax2.axvline(np.percentile(write_latencies, 99), color='orange', linestyle='--', label='P99')
        ax2.legend()
    else:
        ax2.text(0.5, 0.5, 'No write data', ha='center', va='center')
        ax2.set_title('Write Latency Distribution')
    
    plt.tight_layout()
    plt.savefig(output_path, dpi=150, bbox_inches='tight')
    plt.close()
    print(f"Saved latency distribution: {output_path}")

def plot_time_intervals(records, output_path, config_name):
    """Plot time intervals between reads and writes of the same key."""
    from datetime import datetime
    
    # Group records by key
    key_events = {}
    for record in records:
        key = record['key']
        if key not in key_events:
            key_events[key] = []
        
        timestamp = datetime.fromisoformat(record['timestamp'].replace('Z', '+00:00'))
        key_events[key].append({
            'timestamp': timestamp,
            'type': record['type'],
            'latency_ms': float(record['latency_ms'])
        })
    
    # Calculate intervals between reads and writes of same key
    intervals = []
    for key, events in key_events.items():
        # Sort by timestamp
        events.sort(key=lambda x: x['timestamp'])
        
        # Find intervals between writes and subsequent reads
        for i in range(len(events) - 1):
            if events[i]['type'] == 'write' and events[i+1]['type'] == 'read':
                interval = (events[i+1]['timestamp'] - events[i]['timestamp']).total_seconds() * 1000  # ms
                intervals.append(interval)
    
    if not intervals:
        print(f"No read-write intervals found for {config_name}")
        return
    
    plt.figure(figsize=(10, 6))
    plt.hist(intervals, bins=50, alpha=0.7, color='purple', edgecolor='black')
    plt.xlabel('Time Interval (ms)')
    plt.ylabel('Frequency')
    plt.title(f'Time Intervals Between Writes and Reads (Same Key)\n{config_name}')
    plt.grid(True, alpha=0.3)
    plt.yscale('log')
    plt.tight_layout()
    plt.savefig(output_path, dpi=150, bbox_inches='tight')
    plt.close()
    print(f"Saved time intervals: {output_path}")

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 visualize.py <results_directory>")
        sys.exit(1)
    
    results_dir = Path(sys.argv[1])
    
    csv_path = results_dir / "results.csv"
    summary_path = results_dir / "summary.json"
    
    if not csv_path.exists():
        print(f"Error: {csv_path} not found")
        sys.exit(1)
    
    # Load summary for config name
    config_name = "Unknown"
    if summary_path.exists():
        summary = load_summary(summary_path)
        config_name = summary.get('config', 'Unknown')
    
    # Load CSV data
    records = load_csv(csv_path)
    
    # Create visualizations
    latency_plot = results_dir / "latency_distribution.png"
    interval_plot = results_dir / "time_intervals.png"
    
    plot_latency_distribution(records, latency_plot, config_name)
    plot_time_intervals(records, interval_plot, config_name)
    
    print(f"\nVisualizations created in: {results_dir}")

if __name__ == "__main__":
    main()





