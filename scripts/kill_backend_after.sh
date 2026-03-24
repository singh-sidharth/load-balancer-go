#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <port> <delay_seconds>"
  exit 1
fi

PORT="$1"
DELAY="$2"

echo "Will kill backend on port $PORT after ${DELAY}s..."

sleep "$DELAY"

PIDS=$(lsof -nP -iTCP:"$PORT" -sTCP:LISTEN -t || true)

if [ -z "$PIDS" ]; then
  echo "No listening process found on port $PORT"
  exit 0
fi

echo "Killing backend listener(s) on port $PORT: $PIDS"

for pid in $PIDS; do
  if kill -0 "$pid" 2>/dev/null; then
    kill -TERM "$pid"
    echo "Killed PID $pid"
  else
    echo "PID $pid already exited"
  fi
done