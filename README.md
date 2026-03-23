# Go Load Balancer

## Overview
A simple reverse proxy load balancer built in Go to explore request routing, concurrency, and backend health management.

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