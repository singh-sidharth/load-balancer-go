#!/usr/bin/env bash
set -euo pipefail

PIDS=()

cleanup() {
  echo
  echo "Stopping local services..."
  for pid in "${PIDS[@]:-}"; do
    kill -TERM "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT

for port in 8081 8082 8083 8080; do
  if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "Port $port is already in use"
    exit 1
  fi
done

echo "Starting backends..."
PORT=8081 go run ./cmd/backend > /tmp/backend_8081.log 2>&1 &
PIDS+=("$!")

PORT=8082 go run ./cmd/backend > /tmp/backend_8082.log 2>&1 &
PIDS+=("$!")

PORT=8083 go run ./cmd/backend > /tmp/backend_8083.log 2>&1 &
PIDS+=("$!")

sleep 1

echo "Starting load balancer..."
BACKENDS="http://localhost:8081,http://localhost:8082,http://localhost:8083" \
go run ./cmd/loadbalancer > /tmp/load_balancer.log 2>&1 &
PIDS+=("$!")

sleep 1

echo
echo "Services are up:"
echo "  backends: 8081, 8082, 8083"
echo "  load balancer: 8080"
echo "  metrics: http://localhost:8080/metrics"
echo
echo "Press Ctrl+C to stop everything."

wait