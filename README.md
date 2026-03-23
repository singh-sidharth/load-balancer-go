# Go Load Balancer

## Overview
A simple reverse proxy load balancer built in Go to explore request routing, concurrency, and backend health management.

## Why I built this

I wanted a hands-on Go project around networking and systems concepts rather than a CRUD app. This project helped me practice:

- Go project structure with `cmd/` and `internal/`
- reverse proxying with `net/http/httputil`
- round-robin backend selection
- shared state coordination for request routing

## Features
- Round-robin load balancing
- Reverse proxy using net/http
- Backend abstraction via interface

## Architecture
Client → Load Balancer → Backend Servers

## Current Limitations
- No health checks
- No retries or timeouts
- Static backend configuration

## Future Improvements
- Active + passive health checks
- Config-driven backends
- Retry with backoff
- Metrics and observability