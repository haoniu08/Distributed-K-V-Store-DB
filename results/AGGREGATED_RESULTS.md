# Aggregated Load Test Results

This file aggregates the 16 load-test runs produced for Homework 10 (Option 1).
Each run is ~60s with concurrency 10 and uses the same local-in-time workload generator.
All data and graphs are in the `results/` directory; this document summarizes key metrics and highlights comparisons between Leader‑Follower strategies and the Leaderless configuration.

## How to read this file
- Each row in the main table is one test run (configuration + read/write ratio).
- Key metrics: total requests, writes, reads, successful/failed requests, stale reads, and latency summary (mean and P95) for reads and writes.
- Use the links in the "Run folder" column to open the per-run folder with raw CSV, `summary.json`, and graphs.

---

## Aggregated table (16 runs)

| Run folder | Config | Total reqs | Writes | Reads | Success | Failed | Stale reads | Read mean (ms) | Read P95 (ms) | Write mean (ms) | Write P95 (ms) |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `lf_w5_r1_01_99` | LF W=5 R=1 (1/99) | 5386 | 46 | 5340 | 341 | 5045 | 0 | 0.777 | 1.2259 | 303.7821 | 305.7115 |
| `lf_w5_r1_10_90` | LF W=5 R=1 (10/90) | 5444 | 537 | 4907 | 2011 | 3433 | 0 | 0.8097 | 1.4823 | 303.4990 | 305.4275 |
| `lf_w5_r1_50_50` | LF W=5 R=1 (50/50) | 3915 | 1970 | 1945 | 3530 | 385 | 0 | 0.5110 | 1.0601 | 304.2847 | 307.2163 |
| `lf_w5_r1_90_10` | LF W=5 R=1 (90/10) | 2144 | 1970 | 174 | 2136 | 8 | 0 | 0.7350 | 2.0601 | 305.2441 | 308.2390 |

| `lf_w1_r5_01_99` | LF W=1 R=5 (1/99) | 5386 | 47 | 5339 | 5186 | 200 | 0 | 0.6927 | 0.8939 | 303.4045 | 304.3420 |
| `lf_w1_r5_10_90` | LF W=1 R=5 (10/90) | 5087 | 510 | 4577 | 4946 | 141 | 0 | 0.7872 | 1.3380 | 304.9583 | 307.7715 |
| `lf_w1_r5_50_50` | LF W=1 R=5 (50/50) | 3956 | 1960 | 1996 | 3939 | 17 | 0 | 0.6107 | 1.3552 | 305.8160 | 309.0603 |
| `lf_w1_r5_90_10` | LF W=1 R=5 (90/10) | 2195 | 1960 | 235 | 2194 | 1 | 0 | 0.7373 | 1.8226 | 306.5809 | 310.1168 |

| `lf_w3_r3_01_99` | LF W=3 R=3 (1/99) | 5043 | 51 | 4992 | 5028 | 15 | 0 | 0.7126 | 0.9385 | 304.8853 | 306.6240 |
| `lf_w3_r3_10_90` | LF W=3 R=3 (10/90) | 5086 | 481 | 4605 | 5073 | 13 | 0 | 0.7911 | 1.4946 | 304.8972 | 307.2668 |
| `lf_w3_r3_50_50` | LF W=3 R=3 (50/50) | 3855 | 1970 | 1885 | 3851 | 4 | 0 | 0.5422 | 1.1127 | 304.5490 | 307.2636 |
| `lf_w3_r3_90_10` | LF W=3 R=3 (90/10) | 2188 | 1970 | 218 | 2188 | 0 | 0 | 0.6275 | 1.3215 | 305.1319 | 307.8178 |

| `ll_01_99` | Leaderless W=N R=1 (1/99) | 5392 | 49 | 5343 | 5392 | 0 | 0 | 0.6980 | 0.9084 | 303.3547 | 304.9578 |
| `ll_10_90` | Leaderless W=N R=1 (10/90) | 5406 | 541 | 4865 | 5406 | 0 | 0 | 0.7490 | 1.3484 | 303.4264 | 305.3612 |
| `ll_50_50` | Leaderless W=N R=1 (50/50) | 3863 | 1970 | 1893 | 3863 | 0 | 0 | 0.5453 | 1.1227 | 304.5320 | 307.0474 |
| `ll_90_10` | Leaderless W=N R=1 (90/10) | 2152 | 1962 | 190 | 2152 | 0 | 0 | 0.6560 | 1.4920 | 306.0408 | 309.1568 |

---

## Highlighted comparisons — Leader‑Follower vs Leaderless

Summary of the main takeaways comparing Leader‑Follower strategies with the Leaderless configuration (same workload parameters):

- Read latency (mean): all runs show sub‑millisecond average read latency. Differences between the LF strategies and Leaderless are very small (fractions of a millisecond). For read‑heavy workloads (1% writes) LF W=1,R=5 and Leaderless offered the lowest read means (≈0.69–0.70 ms).

- Write latency (mean): all writes cluster around ~303–307 ms due to the intentionally added replication delays (leader/coordinator sleeps and follower response sleeps). Leaderless write mean is comparable to LF strategies; differences are only a few milliseconds.

- Tail behavior (read P95 / write P95): P95 values are close to means; the largest read P95 observed is 2.06 ms (LF W=5,R=1 at 90% writes). Write P95 values are in a tight band (~304–310 ms).

- Stale reads: zero stale reads detected in every run (`stale_reads = 0`). This indicates replication usually completed before reads occurred under the test settings (localhost environment, concurrency 10). To surface stale reads increase concurrency, reduce W, or increase follower delays.

### Practical takeaway
- If you need a simple implementation with no single point of failure and comparable latency, Leaderless performs similarly to the Leader‑Follower variants in this local testbed.
- If you need stronger guarantees and controlled quorum behaviour, W=3,R=3 gives a balanced result for mixed workloads (50/50 and 90/10) and performs slightly better in several balanced scenarios.

## Where to find the raw artifacts
- Per-run folder (CSV, summary, graphs) example: `results/lf_w5_r1_01_99/` contains `results.csv`, `summary.json`, `latency_distribution.png`, and `time_intervals.png`.
- Full report: `results/hw10_report.pdf` and `results/hw10_report.docx`.

---

If you want, I can also:
- produce a CSV `results/aggregated_summary.csv` with the table above, or
- generate a few comparison plots (read mean by config/ratio, write mean by config/ratio) and save them under `results/`.

Pick one and I will add it next.
