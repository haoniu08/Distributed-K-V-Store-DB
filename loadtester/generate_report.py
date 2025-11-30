#!/usr/bin/env python3
"""
Generate comprehensive report for load test results.
Shows latency, stale reads, error rates, and time intervals.
"""

import json
import sys
from pathlib import Path

def load_summary(summary_path):
    """Load summary JSON file."""
    with open(summary_path, 'r') as f:
        return json.load(f)

def format_latency_stats(stats):
    """Format latency statistics for display."""
    if not stats or stats.get('mean_ms') == 0:
        return "N/A"
    return f"Mean: {stats['mean_ms']:.2f}ms, P95: {stats['p95_ms']:.2f}ms, P99: {stats['p99_ms']:.2f}ms"

def generate_report(results_dir):
    """Generate comprehensive report."""
    results_dir = Path(results_dir)
    summary_path = results_dir / "summary.json"
    
    if not summary_path.exists():
        print(f"Error: {summary_path} not found")
        sys.exit(1)
    
    summary = load_summary(summary_path)
    
    print("=" * 80)
    print(f"LOAD TEST REPORT: {summary.get('config', 'Unknown')}")
    print("=" * 80)
    print()
    
    # Overall Statistics
    print("📊 OVERALL STATISTICS")
    print("-" * 80)
    print(f"Total Requests:        {summary.get('total_requests', 0):,}")
    print(f"Total Writes:          {summary.get('total_writes', 0):,}")
    print(f"Total Reads:           {summary.get('total_reads', 0):,}")
    print(f"Duration:              {summary.get('duration', 'N/A')}")
    print()
    
    # Success/Error Rates
    print("✅ SUCCESS & ERROR RATES")
    print("-" * 80)
    success_rate = summary.get('success_rate_percent', 0)
    error_rate = summary.get('error_rate_percent', 0)
    successful_requests = summary.get('successful_requests', 0)
    failed_requests = summary.get('failed_requests', 0)
    
    print(f"Successful Requests:   {successful_requests:,} ({success_rate:.2f}%)")
    print(f"Failed Requests:       {failed_requests:,} ({error_rate:.2f}%)")
    print()
    
    # Stale Reads
    print("🔄 STALE READS")
    print("-" * 80)
    stale_reads = summary.get('stale_reads', 0)
    stale_read_rate = summary.get('stale_read_rate_percent', 0)
    stale_read_rate_of_successful = summary.get('stale_read_rate_of_successful_percent', 0)
    successful_reads = summary.get('successful_reads', 0)
    
    print(f"Total Stale Reads:     {stale_reads:,}")
    print(f"Stale Read Rate:       {stale_read_rate:.2f}% (of all reads)")
    print(f"Stale Read Rate:       {stale_read_rate_of_successful:.2f}% (of successful reads)")
    print(f"Successful Reads:      {successful_reads:,}")
    print()
    
    # Latency Statistics
    print("⏱️  LATENCY STATISTICS")
    print("-" * 80)
    
    write_latency = summary.get('write_latency', {})
    read_latency = summary.get('read_latency', {})
    
    print("Write Latency:")
    if write_latency and write_latency.get('mean_ms', 0) > 0:
        print(f"  Mean:    {write_latency.get('mean_ms', 0):.2f} ms")
        print(f"  Median:  {write_latency.get('median_ms', 0):.2f} ms")
        print(f"  P95:     {write_latency.get('p95_ms', 0):.2f} ms")
        print(f"  P99:     {write_latency.get('p99_ms', 0):.2f} ms")
        print(f"  Min:     {write_latency.get('min_ms', 0):.2f} ms")
        print(f"  Max:     {write_latency.get('max_ms', 0):.2f} ms")
    else:
        print("  No write data")
    print()
    
    print("Read Latency:")
    if read_latency and read_latency.get('mean_ms', 0) > 0:
        print(f"  Mean:    {read_latency.get('mean_ms', 0):.2f} ms")
        print(f"  Median:  {read_latency.get('median_ms', 0):.2f} ms")
        print(f"  P95:     {read_latency.get('p95_ms', 0):.2f} ms")
        print(f"  P99:     {read_latency.get('p99_ms', 0):.2f} ms")
        print(f"  Min:     {read_latency.get('min_ms', 0):.2f} ms")
        print(f"  Max:     {read_latency.get('max_ms', 0):.2f} ms")
    else:
        print("  No read data")
    print()
    
    # Key Insights
    print("💡 KEY INSIGHTS")
    print("-" * 80)
    
    if error_rate > 50:
        print(f"⚠️  High error rate ({error_rate:.2f}%) - many requests failed")
        print("   This is expected for read-heavy workloads with few writes")
    elif error_rate > 10:
        print(f"⚠️  Moderate error rate ({error_rate:.2f}%)")
    else:
        print(f"✅ Low error rate ({error_rate:.2f}%)")
    
    if stale_read_rate_of_successful > 5:
        print(f"⚠️  High stale read rate ({stale_read_rate_of_successful:.2f}%)")
        print("   This indicates inconsistency windows in leaderless mode")
    elif stale_read_rate_of_successful > 0:
        print(f"ℹ️  Some stale reads detected ({stale_read_rate_of_successful:.2f}%)")
    else:
        print("✅ No stale reads detected")
    
    if write_latency and write_latency.get('mean_ms', 0) > 200:
        print(f"⏱️  High write latency ({write_latency.get('mean_ms', 0):.2f}ms)")
        print("   Expected for W=5 (must wait for all nodes)")
    elif write_latency and write_latency.get('mean_ms', 0) > 0:
        print(f"⏱️  Write latency: {write_latency.get('mean_ms', 0):.2f}ms")
    
    if read_latency and read_latency.get('mean_ms', 0) < 5:
        print(f"⚡ Fast read latency ({read_latency.get('mean_ms', 0):.2f}ms)")
        print("   Expected for R=1 (local read)")
    elif read_latency and read_latency.get('mean_ms', 0) > 0:
        print(f"⏱️  Read latency: {read_latency.get('mean_ms', 0):.2f}ms")
    
    print()
    print("=" * 80)
    print(f"Report generated for: {results_dir}")
    print("=" * 80)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 generate_report.py <results_directory>")
        sys.exit(1)
    
    generate_report(sys.argv[1])



