# RUN BenchmarkEndToEndBooking

Location : `server/benchmark_test.go`

```
BenchmarkEndToEndBooking
BenchmarkEndToEndBooking-8 26076 45647 ns/op 8255 B/op 96 allocs/op
```

Inference:

| Metric       | Value             | Verdict    |
| ------------ | ----------------- | ---------- |
| Latency      | ~45 µs            | Excellent  |
| Throughput   | ~22k ops/sec/core | Strong     |
| Allocations  | 96/op             | Acceptable |
| Memory       | 8 KB/op           | Good       |
| Architecture | End-to-end HTTP   | Realistic  |

# Running BenchmarkP95Latency

```
Running tool: /opt/homebrew/bin/go test -test.fullpath=true -benchmem -run=^$ -bench ^BenchmarkP95Latency$ github.com/ignius299792458/techkraft-ch-svr

goos: darwin
goarch: arm64
pkg: github.com/ignius299792458/techkraft-ch-svr
cpu: Apple M2
=== RUN   BenchmarkP95Latency
BenchmarkP95Latency

=== Performance Summary ===
Total Requests: 1000
P50 Latency: 1ms
P95 Latency: 2ms (target: 500ms)
P99 Latency: 4ms
========================

=== Performance Summary ===
Total Requests: 1000
P50 Latency: 0ms
P95 Latency: 2ms (target: 500ms)
P99 Latency: 3ms
========================

=== Performance Summary ===
Total Requests: 1000
P50 Latency: 0ms
P95 Latency: 2ms (target: 500ms)
P99 Latency: 3ms
========================

=== Performance Summary ===
Total Requests: 1000
P50 Latency: 0ms
P95 Latency: 2ms (target: 500ms)
P99 Latency: 3ms
========================

=== Performance Summary ===
Total Requests: 1000
P50 Latency: 0ms
P95 Latency: 2ms (target: 500ms)
P99 Latency: 3ms
========================

=== Performance Summary ===
Total Requests: 1000
P50 Latency: 0ms
P95 Latency: 2ms (target: 500ms)
P99 Latency: 3ms
========================

=== Performance Summary ===
Total Requests: 1000
P50 Latency: 0ms
P95 Latency: 2ms (target: 500ms)
P99 Latency: 2ms
========================
BenchmarkP95Latency-8           1000000000               0.02024 ns/op           0.6162 p50_ms           2.335 p95_ms            2.956 p99_ms            1000 total_requests              0 B/op          0 allocs/op
PASS
ok      github.com/ignius299792458/techkraft-ch-svr     0.511s

```
