# Go Load Balancer

## Overview

Reverse proxy load balancer in Go with active health checks, health-aware routing, and passive failure detection.

The system routes requests across multiple backends, avoids unhealthy servers, and updates backend health dynamically based on both periodic checks and real request failures.

## Why I built this

I wanted to build a small but realistic systems component that models how production load balancers handle routing, health checks, and failure scenarios.

This project focuses on correctness, failure handling, and state management rather than just request forwarding.

## Features

- Reverse proxy forwarding to upstream backends
- Round-robin load balancing
- Active health checks (`/health` endpoint)
- Health-aware backend selection
- Passive failure detection (mark backend unhealthy on proxy errors)
- Fail-fast behavior using upstream timeouts
- Returns `503 Service Unavailable` when no backends are healthy
- Backend abstraction via interface

## Architecture

```text
Client → Load Balancer → Backend Servers
```

## Project Structure

```text
cmd/loadbalancer/   # load balancer entrypoint
cmd/backend/        # simple backend server for testing
internal/balancer/  # load balancing logic
internal/server/    # backend abstraction and proxying
scripts/            # load and failure test helpers
```

## How it works

1. The load balancer maintains a set of backend servers and their health state.
2. An initial synchronous health check ensures correct routing at startup.
3. A background goroutine continuously updates backend health.
4. Each request selects the next healthy backend using round-robin.
5. Requests are proxied using `httputil.ReverseProxy`.
6. If a backend fails during request forwarding, it is marked unhealthy (passive check).
7. If no healthy backends are available, a `503` response is returned.

## Running the project

### 1. Start backend servers

Run multiple backend instances:

```bash
PORT=8081 go run ./cmd/backend
PORT=8082 go run ./cmd/backend
PORT=8083 go run ./cmd/backend
```

Each backend exposes:

- / → sample response
- /health → health check endpoint

### 2. Start load balancer

```bash
go run ./cmd/loadbalancer
```

### 3. Test

```bash
curl http://localhost:8080
```

Requests should be distributed across backends

## Running Tests

This project includes both unit tests and integration tests to validate correctness and failover behavior.

### Run all tests

From the project root directory:

```bash
go test ./... -v
```

### What is covered

- Round-robin load balancing correctness
- Skipping unhealthy backends
- Behavior when no backends are available
- Proxy failure handling and failover across requests

## Load Testing

Basic load testing can be performed using `hey`:

```bash
hey -n 1000 -c 10 http://localhost:8080
```

This simulates concurrent traffic and helps evaluate:

- requests per second (RPS)
- latency distribution (p50, p95, p99)
- error rates under load

> _This project focuses on correctness and failure handling rather than raw throughput. Load testing is used to validate behavior under concurrent traffic._

### Scenarios to try

- All backends healthy → baseline throughput and latency
- One backend down → verify failover behavior
- All backends down → expect `503 Service Unavailable`

### Installation

Mac:

```bash
brew install hey
```

Linux:

```bash
go install github.com/rakyll/hey@latest
```

Windows (PowerShell):

```powershell
go install github.com/rakyll/hey@latest
```

> Make sure your Go bin directory is in your **PATH** after using go install.

## Failure Testing

This project includes scripts to simulate backend failure during active load.

### Scripts

- `scripts/run_failover_test.sh`  
  Starts local backends and the load balancer, schedules backend termination, and runs a timed load test.

- `scripts/kill_backend_after.sh`  
  Terminates a backend listener on a given port after a configured delay.

### Example: fail backend `8083` during load

```bash
./scripts/run_failover_test.sh 8083 2 5s 100
```

### Example: kill a backend manually

You can also terminate a backend process directly after a delay:

```bash
./scripts/kill_backend_after.sh 8083 2
```

This will:

- wait for 2 seconds
- find any process listening on port `8083`
- terminate it gracefully

Useful for testing failure handling without running a full load test.

## Observability

This project exposes Prometheus metrics to help analyze system behavior under load and failure.

### Metrics Endpoint

Metrics are available at:

```bash
http://localhost:8080/metrics
```

### Key Metrics

- `lb_requests_total`  
  Total requests handled by the load balancer, labeled by method, path, status, and backend.

- `lb_request_duration_seconds`  
  Request latency histogram for measuring p50, p95, and p99.

- `lb_proxy_errors_total`  
  Total proxy errors per backend.

- `lb_backend_health`  
  Backend health status (1 = healthy, 0 = unhealthy).

### Running Prometheus Locally

Start Prometheus using Docker:

```bash
docker compose up
```

Then open:

```text
http://localhost:9090
```

### Example Queries

Requests per second (RPS):

```promql
rate(lb_requests_total[10s])
```

Per-backend traffic distribution:

```promql
sum by (backend) (rate(lb_requests_total[10s]))
```

Backend health:

```promql
lb_backend_health
```

### What to Observe

- Traffic distribution across backends
- Failover behavior when a backend goes down
- Error spikes during failure windows
- Latency changes under load

> Observability is used here to validate correctness, understand system behavior, and reason about performance under failure scenarios.

## Current Limitations

- No retry logic for failed requests (fail-fast on upstream failure)
- No safe retry for idempotent requests
- No rate limiting or backpressure control
- No observability (metrics, tracing, structured logs)
- Static backend configuration (hardcoded)

## Future Improvements

- Safe retry for idempotent requests (e.g., GET)
- Config-driven backend management (JSON/env)
- Per-backend concurrency limits (bulkheading)
- Metrics (request count, latency, error rates)
- Additional load balancing strategies (least connections, weighted)
