# Poster Generator Lambda - Load Test Results

**Date:** January 3, 2026
**Region:** ap-south-1
**Lambda Memory:** 512MB
**Architecture:** ARM64 (Graviton2)

---

## Test 1: Warm Lambda (100 Requests)

**Configuration:**
- Requests: 100
- Concurrency: 10
- Lambda State: Warm

**Results:**

| Metric | Value |
|--------|-------|
| Total Time | 10.33 secs |
| Requests/sec | 9.68 |
| Avg Response | 980ms |
| Fastest | 399ms |
| Slowest | 3.16s |
| Success Rate | 100% |

**Latency Distribution:**
| Percentile | Time |
|------------|------|
| 50% | 776ms |
| 75% | 1.32s |
| 90% | 1.56s |
| 95% | 2.10s |
| 99% | 3.16s |

**CloudWatch Metrics:**
- Invocations: 102
- Duration: Min 829ms, Avg 954ms, Max 1,078ms
- Concurrent Executions: 1-2
- Throttles: 0
- Errors: 0

---

## Test 2: Cold Start Lambda (500 Requests)

**Configuration:**
- Requests: 500
- Concurrency: 50
- Lambda State: Cold (forced via env variable update)

**Results:**

| Metric | Value |
|--------|-------|
| Total Time | 42.71 secs |
| Requests/sec | 11.71 |
| Avg Response | 3.86s |
| Fastest | 473ms |
| Slowest | 17.22s |
| Success Rate | 100% |

**Latency Distribution:**
| Percentile | Time |
|------------|------|
| 50% | 3.32s |
| 75% | 4.25s |
| 90% | 6.77s |
| 95% | 9.77s |
| 99% | 14.49s |

**Response Time Histogram:**
```
0.47s - 2.15s  [122] ████████████████████████
2.15s - 3.82s  [202] ████████████████████████████████████████
3.82s - 5.50s  [101] ████████████████████
5.50s - 7.17s  [28]  ██████
7.17s - 8.85s  [15]  ███
8.85s - 10.5s  [12]  ██
10.5s - 12.2s  [8]   ██
12.2s - 13.9s  [6]   █
13.9s - 15.5s  [3]   █
15.5s - 17.2s  [2]   █
```

**CloudWatch Metrics:**
- Invocations: 500+
- Duration: Min 355ms, Avg 898ms, Max 1,725ms
- Concurrent Executions: **10** (auto-scaled)
- Throttles: 0
- Errors: 0

---

## Comparison Summary

| Metric | Warm (100 req) | Cold (500 req) | Delta |
|--------|----------------|----------------|-------|
| Avg Response | 980ms | 3.86s | +294% |
| Fastest | 399ms | 473ms | +19% |
| Slowest | 3.16s | 17.22s | +445% |
| Throughput | 9.7 req/s | 11.7 req/s | +21% |
| Concurrency | 1-2 | 10 | +400% |
| Success Rate | 100% | 100% | - |
| Errors | 0 | 0 | - |
| Throttles | 0 | 0 | - |

---

## Key Observations

1. **Cold Start Impact:** Initial cold starts added ~3-17s latency to requests
2. **Auto-Scaling:** Lambda successfully scaled to 10 concurrent instances under load
3. **Stability:** Zero errors and zero throttling across all tests
4. **Warm Performance:** Once warmed, Lambda responded in ~400-900ms consistently
5. **Throughput:** Higher concurrency actually improved overall throughput (9.7 → 11.7 req/s)

---

## API Endpoint

```
POST https://w1qbo9xwwf.execute-api.ap-south-1.amazonaws.com/generate
Content-Type: application/json

{
  "name": "Test User",
  "avatarUrl": "https://i.pravatar.cc/300"
}
```

---

## Commands Used

```bash
# Warm test
hey -n 100 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"name": "Test User", "avatarUrl": "https://i.pravatar.cc/300"}' \
  https://w1qbo9xwwf.execute-api.ap-south-1.amazonaws.com/generate

# Force cold start
aws lambda update-function-configuration \
  --function-name PosterStack-PosterGeneratorFF0D0D0A-x35C8d7zXo0f \
  --environment "Variables={GO_ENV=production,FORCE_COLD=$(date +%s)}" \
  --profile techpix

# Cold start test
hey -n 500 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -d '{"name": "Test User", "avatarUrl": "https://i.pravatar.cc/300"}' \
  https://w1qbo9xwwf.execute-api.ap-south-1.amazonaws.com/generate
```
