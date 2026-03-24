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
