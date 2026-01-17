# Benchmark Comparison Report

**Generated:** 2026-01-08 19:58:10 EST

**Comparing 3 benchmark runs**

## Run Overview

These tests were run from a local workstation to the specified target URLs. The benchmarks include frontend asset measurements and load testing results. The target system was a Nanode shared VM at Linode with 1 CPU Core, 25 GB Storage, and 1 GB RAM. It was running Ubuntu 24.04.3 with Caddy as the proxy. Two versions of the application were tested: 0.17.0-beta and 0.20.0-beta, both deployed using Docker containers, and accessing a MariaDB database hosted on the same Linode instance outside of Docker, with default configurations.

The real purpose of this test is to demonstrate that a small self-hosted instance can handle moderate load while delivering a responsive frontend experience.

This table summarizes each benchmark run included in this comparison. The **Overall** status indicates whether all tests passed (✅), some tests showed degraded performance (⚠️), or critical tests failed (❌).

| # | Timestamp | Target | Version | Overall |
|---|-----------|--------|---------|--------|
| 1 | 2026-01-08 21:58 | URL1 | 0.20.0-beta | ✅ pass |
| 2 | 2026-01-08 22:48 | URL2 | 0.17.0-beta | ✅ pass |
| 3 | 2026-01-09 00:51 | URL1 | 0.20.0-beta | ✅ pass |

## Frontend Assets Comparison

Frontend metrics measure the size and load time of the web application's static assets (HTML, JavaScript, CSS). These directly impact user experience, especially on slower connections or mobile devices.

- **Total Size (KB)**: Combined size of all frontend assets in kilobytes. Smaller bundles load faster and use less bandwidth. Size increases may indicate new features or inefficient bundling.
- **Total Time (ms)**: Time to download all frontend assets in milliseconds. Affected by both bundle size and server response time.

| Metric | Run 1 | Run 2 | Run 3 | Δ (Last vs First) |
|--------|-------:|-------:|-------:|---------------:|
| Total Size (KB) | 920.09 | 878.94 | 920.09 |⚪ ~0 |
| Total Time (ms) | 182.75 | 214.56 | 181.00 |🟢 -1.75 (-1.0%) |

### Individual Asset Performance

This table shows the size and load time for each individual frontend asset. Large or slow-loading assets are good candidates for optimization.

| Asset | Run 1 Size (KB) | Run 1 Time (ms) | Run 2 Size (KB) | Run 2 Time (ms) | Run 3 Size (KB) | Run 3 Time (ms) |
|-------|--------------:|---------------:|--------------:|---------------:|--------------:|---------------:|
| `index.html` | 1.84 | 58.75 | 1.45 | 79.12 | 1.84 | 57.62 |
| `/assets/index-Cv1xBUjg.js` | 251.23 | 61.98 | - | - | 251.23 | 63.51 |
| `/assets/index-ncZg0FPA.css` | 667.02 | 62.02 | - | - | 667.02 | 59.87 |
| `/assets/index-Bh11ylPu.js` | - | - | 215.46 | 74.08 | - | - |
| `/assets/index-3xSvSqMX.css` | - | - | 662.04 | 61.36 | - | - |

## Load Test Comparison

Load testing simulates multiple concurrent users making requests to measure how the server performs under stress. These metrics are critical for understanding capacity and identifying performance bottlenecks.

### Configuration & Throughput

- **Concurrent**: Number of simultaneous users (goroutines) making requests during the test.
- **Duration (sec)**: How long the load test ran in seconds.
- **Total Requests**: Total number of HTTP requests made during the test.
- **Successful**: Number of requests that returned HTTP 2xx responses.
- **Failed**: Number of requests that failed or returned error responses.
- **RPS (Requests Per Second)**: Throughput measure showing how many requests the server can handle per second. Higher values indicate better capacity.
- **Success Rate**: Percentage of requests that completed successfully. Values below 99% may indicate server overload.

### Latency Distribution

- **Min Latency**: Fastest response time observed during the test.
- **p50 Latency (50th Percentile)**: The median response time—50% of requests completed faster than this value. Represents typical user experience.
- **p95 Latency (95th Percentile)**: 95% of requests completed faster than this value. Helps identify slower outliers that affect some users.
- **p99 Latency (99th Percentile)**: 99% of requests completed faster than this value. Reveals worst-case scenarios and tail latency issues.
- **Max Latency**: Slowest response time observed during the test.
- **Avg Latency**: Arithmetic mean of all response times. Can be skewed by outliers, so percentiles are often more meaningful.

| Metric | Run 1 | Run 2 | Run 3 | Δ (Last vs First) |
|--------|-------:|-------:|-------:|---------------:|
| Concurrent | 50 | 50 | 50 | - |
| Duration (sec) | 300 | 300 | 300 | - |
| Total Requests | 194317 | 185300 | 223885 | - |
| Successful | 194267 | 185250 | 223835 | - |
| Failed | 50 | 50 | 50 | - |
| RPS | **647.72** | **617.66** | **746.28** |🟢 +98.56 (+15.2%) |
| Success Rate | 99.97% | 99.97% | 99.98% | - |
| Min Latency (ms) | 0.03 | 0.43 | 8.64 |🔴 +8.61 (+26078.8%) |
| p50 Latency (ms) | 73.25 | 76.14 | 64.67 |🟢 -8.58 (-11.7%) |
| p95 Latency (ms) | 108.49 | 117.10 | 82.75 |🟢 -25.73 (-23.7%) |
| p99 Latency (ms) | 145.74 | 158.68 | 103.16 |🟢 -42.58 (-29.2%) |
| Max Latency (ms) | 960.84 | 767.41 | 652.38 |🟢 -308.45 (-32.1%) |
| Avg Latency (ms) | **77.18** | **80.94** | **66.99** |🟢 -10.20 (-13.2%) |


### Legend

- **Δ (Delta)**: Change from first run to last run
- 🟢 Improvement (faster/smaller)
- 🔴 Regression (slower/larger)
- ⚪ No significant change

### Threshold Configuration

- p95 Latency Max: 200 ms
- p99 Latency Max: 500 ms
- Error Rate Max: 0.5%
- RPS Minimum: 50
- Health Response Max: 100 ms

---

*Comparison report generated by actalog-bench at 2026-01-08 19:58:10 EST*
