#!/usr/bin/env bash
set -e

trap 'kill 0' EXIT

(cd go-backend && go run ./cmd/flourish/ serve) &
(cd frontend && npm run dev) &

wait
