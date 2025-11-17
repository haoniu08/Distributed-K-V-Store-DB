#!/usr/bin/env python3
"""
Create sample data for testing visualization script.
This generates mock load test results for demonstration.
"""

import csv
import json
import os
import sys
from datetime import datetime, timedelta
import random

def create_sample_data(output_dir):
    """Create sample results.csv and summary.json for testing."""
    os.makedirs(output_dir, exist_ok=True)
    
    # Create sample CSV data
    csv_path = os.path.join(output_dir, "results.csv")
    start_time = datetime.now() - timedelta(seconds=60)
    
    with open(csv_path, 'w', newline='') as f:
        writer = csv.writer(f)
        writer.writerow(['timestamp', 'type', 'key', 'latency_ms', 'success', 'is_stale', 'version'])
        
        # Generate 1000 sample requests
        for i in range(1000):
            timestamp = start_time + timedelta(milliseconds=i * 60)
            req_type = 'read' if random.random() < 0.9 else 'write'
            key = f"key_{random.randint(0, 99)}"
            
            # Simulate latency (reads faster, writes slower)
            if req_type == 'read':
                latency = random.gauss(50, 10)  # Mean 50ms, std 10ms
            else:
                latency = random.gauss(1200, 200)  # Mean 1200ms, std 200ms
            
            latency = max(10, latency)  # Minimum 10ms
            success = random.random() > 0.05  # 95% success rate
            is_stale = req_type == 'read' and random.random() < 0.1  # 10% stale reads
            version = random.randint(1, 10)
            
            writer.writerow([
                timestamp.isoformat() + 'Z',
                req_type,
                key,
                f"{latency:.2f}",
                str(success).lower(),
                str(is_stale).lower(),
                version
            ])
    
    # Create sample summary
    summary_path = os.path.join(output_dir, "summary.json")
    summary = {
        "config": "Leader-Follower W=5 R=1 (10% writes, 90% reads) - SAMPLE DATA",
        "total_requests": 1000,
        "total_writes": 100,
        "total_reads": 900,
        "successful_requests": 950,
        "failed_requests": 50,
        "stale_reads": 90,
        "write_latency": {
            "min_ms": 800.0,
            "max_ms": 1800.0,
            "mean_ms": 1200.0,
            "median_ms": 1180.0,
            "p50_ms": 1180.0,
            "p95_ms": 1600.0,
            "p99_ms": 1750.0,
            "p999_ms": 1790.0
        },
        "read_latency": {
            "min_ms": 20.0,
            "max_ms": 100.0,
            "mean_ms": 50.0,
            "median_ms": 48.0,
            "p50_ms": 48.0,
            "p95_ms": 75.0,
            "p99_ms": 90.0,
            "p999_ms": 98.0
        },
        "start_time": start_time.isoformat() + 'Z',
        "end_time": datetime.now().isoformat() + 'Z',
        "duration": "60s"
    }
    
    with open(summary_path, 'w') as f:
        json.dump(summary, f, indent=2)
    
    print(f"✓ Sample data created in: {output_dir}")
    print(f"  - {csv_path}")
    print(f"  - {summary_path}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        output_dir = "results/sample"
    else:
        output_dir = sys.argv[1]
    
    create_sample_data(output_dir)
    print(f"\nNow you can generate graphs:")
    print(f"  python3 loadtester/visualize.py {output_dir}")





