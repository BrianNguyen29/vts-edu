#!/usr/bin/env bash
# Bounded load / concurrency harness for the attempt runtime.
#
# This script is OPTIONAL and MANUAL. It is intentionally not part of
# `pnpm check`, not in CI, and not run automatically. The smoke suite
# (`pnpm e2e:smoke`) and the browser E2E suite (`pnpm e2e:browser`)
# already cover correctness; the load harness exists to catch race
# conditions and backend misbehaviour under small bursts.
#
# Run with:
#   pnpm e2e:load           # start DB + apply migrations + start API + run load
#   pnpm e2e:load:direct    # run load against an already-running API on $API_BASE
#
# The harness exercises only the seeded demo attempt + a freshly-created
# draft per scenario. It does not modify production data.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

API_BIN=""

cleanup() {
  echo "Cleaning up..."
  if [[ -n "${API_PID:-}" ]]; then
    kill "${API_PID}" 2>/dev/null || true
    wait "${API_PID}" 2>/dev/null || true
  fi
  if [[ -n "${API_BIN:-}" && -f "${API_BIN}" ]]; then
    rm -f "${API_BIN}"
  fi
  if [[ -n "${RESOURCE_LOCAL_DIR:-}" && -d "${RESOURCE_LOCAL_DIR}" ]]; then
    rm -rf "${RESOURCE_LOCAL_DIR}"
  fi
  ./scripts/e2e_db_stop.sh || true
}
trap cleanup EXIT

./scripts/e2e_db_start.sh
./scripts/e2e_db_migrate.sh

echo "Building API server..."
export DATABASE_URL="postgres://postgres:postgres@localhost:5434/postgres"
export JWT_SIGNING_KEY="demo-signing-key-32-bytes-long!!"
export REFRESH_TOKEN_KEY="demo-refresh-key-32-bytes-long!!"
export FRONTEND_ORIGINS="http://localhost:5173"
export ACCESS_TOKEN_TTL="15m"
export REFRESH_TOKEN_TTL="7d"
export PORT="8080"
export DB_SKIP="false"
# Disable the global rate limiter so concurrency scenarios can fire bursts
# without being throttled. The default is already "false" in the API, but
# setting it explicitly here documents the load test's intent and protects
# the harness from a future default flip.
export RATE_LIMIT_ENABLED="false"

API_BIN="$(mktemp)"
RESOURCE_LOCAL_DIR="$(mktemp -d)"
export RESOURCE_STORAGE_TYPE="local"
export RESOURCE_LOCAL_PATH="${RESOURCE_LOCAL_DIR}"
cd apps/api
go build -o "${API_BIN}" ./cmd/server
cd "${ROOT_DIR}"

echo "Starting API server..."
"${API_BIN}" &
API_PID=$!

echo "Running load / concurrency harness..."
node scripts/e2e_load_api.mjs
