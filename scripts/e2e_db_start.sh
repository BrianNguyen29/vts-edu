#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="vts-e2e-postgres"
PORT="5434"

echo "Starting E2E Postgres container on port ${PORT}..."
docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true
docker run -d \
  --name "${CONTAINER_NAME}" \
  -e POSTGRES_PASSWORD=postgres \
  -p "${PORT}:5432" \
  postgres:15 >/dev/null

for i in $(seq 1 30); do
  if docker exec "${CONTAINER_NAME}" pg_isready -U postgres >/dev/null 2>&1; then
    echo "Postgres ready on localhost:${PORT}"
    exit 0
  fi
  sleep 1
done

echo "Postgres did not become ready in time" >&2
exit 1
