#!/usr/bin/env bash
set -euo pipefail

BACKEND_PORTS=(8081 8082 8083)
LB_PORT=8080
FAIL_PORT="${1:-8083}"
FAIL_DELAY="${2:-2}"
DURATION="${3:-5s}"
CONCURRENCY="${4:-100}"

PIDS=()

cleanup() {
  echo
  echo "Cleaning up..."
  for pid in "${PIDS[@]:-}"; do
    kill -TERM "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT

for port in "${BACKEND_PORTS[@]}" "$LB_PORT"; do
  if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "Port $port is already in use"
    exit 1
  fi
done

BACKENDS_CSV=$(printf "http://localhost:%s," "${BACKEND_PORTS[@]}")
BACKENDS_CSV="${BACKENDS_CSV%,}"

echo "Starting backends..."
for port in "${BACKEND_PORTS[@]}"; do
  PORT="$port" go run ./cmd/backend >"/tmp/backend_${port}.log" 2>&1 &
  PIDS+=("$!")
done

sleep 1

echo "Starting load balancer..."
BACKENDS="$BACKENDS_CSV" go run ./cmd/loadbalancer >"/tmp/load_balancer.log" 2>&1 &
PIDS+=("$!")

sleep 1

echo "Scheduling backend failure on port $FAIL_PORT after ${FAIL_DELAY}s..."
./scripts/kill_backend_after.sh "$FAIL_PORT" "$FAIL_DELAY" &

echo "Running load test: hey -z $DURATION -c $CONCURRENCY http://localhost:$LB_PORT"
hey -z "$DURATION" -c "$CONCURRENCY" "http://localhost:$LB_PORT"