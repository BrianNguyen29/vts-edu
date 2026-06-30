#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="vts-e2e-postgres"

echo "Stopping E2E Postgres container..."
docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true
echo "Stopped."
