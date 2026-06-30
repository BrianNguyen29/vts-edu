#!/usr/bin/env bash
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

API_BIN="$(mktemp)"
cd apps/api
go build -o "${API_BIN}" ./cmd/server
cd "${ROOT_DIR}"

echo "Starting API server..."
"${API_BIN}" &
API_PID=$!

echo "Running API smoke tests..."
node scripts/e2e_smoke_api.mjs
