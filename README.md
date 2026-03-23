# Go Load Balancer

## Overview

A simple reverse proxy load balancer built in Go to explore request routing, concurrency, and backend health management.

## Why I built this

I wanted a hands-on Go project around networking and systems concepts rather than a CRUD app. This project helped me practice:

- Go project structure with `cmd/` and `internal/`
- reverse proxying with `net/http/httputil`
- round-robin load balancing
- health-aware routing
- background health checks and shared state coordination

## Features

- Reverse proxy forwarding to upstream backends
- Round-robin load balancing
- Active health checks (`/health` endpoint)
- Health-aware backend selection
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

1. The load balancer maintains a list of backend servers.
2. A background goroutine periodically checks backend health.
3. Each request selects the next healthy backend using round-robin.
4. Requests are proxied to the selected backend.
5. If no backend is healthy, a `503` response is returned.

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

- No retry or failover logic on request failure
- No failover for in-flight requests
- No rate limiting or backpressure control
- No observability (metrics/log aggregation)
- Static backend configuration (hardcoded)

## Future Improvements

- Config-driven backend management (JSON/env)
- Retry with exponential backoff and jitter
- Passive health checks (marking failures on proxy errors)
- Metrics (request count, latency, error rates)
- Additional load balancing strategies (least connections, weighted)
