#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

API_BIN=""
API_PID=""

cleanup() {
  echo "Cleaning up..."
  if [[ -n "${API_PID:-}" ]]; then
    kill "${API_PID}" 2>/dev/null || true
    wait "${API_PID}" 2>/dev/null || true
  fi
  if [[ -n "${API_BIN:-}" && -f "${API_BIN}" ]]; then
    rm -f "${API_BIN}"
  fi
  rm -rf apps/web/test-results apps/web/playwright-report 2>/dev/null || true
  pnpm e2e:db:stop >/dev/null 2>&1 || true
}
trap cleanup EXIT

pnpm e2e:db:stop >/dev/null 2>&1 || true

# Wait until the previous container is fully removed and the port is free.
for i in $(seq 1 30); do
  if [[ -z "$(docker ps -aq --filter name=vts-e2e-postgres 2>/dev/null)" ]]; then
    if ! ss -tln 2>/dev/null | grep -q ':5434 '; then
      break
    fi
  fi
  sleep 1
done

pnpm e2e:db:start
pnpm e2e:db:migrate

echo "Building API server..."
export DATABASE_URL="postgres://postgres:postgres@localhost:5434/postgres"
export JWT_SIGNING_KEY="demo-signing-key-32-bytes-long!!"
export REFRESH_TOKEN_KEY="demo-refresh-key-32-bytes-long!!"
export FRONTEND_ORIGINS="http://127.0.0.1:5173"
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

echo "Waiting for API /readyz..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:8080/readyz >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

if ! curl -sf http://localhost:8080/readyz >/dev/null 2>&1; then
  echo "ERROR: API /readyz did not become ready" >&2
  exit 1
fi

echo "Running Playwright browser E2E..."
pnpm web:e2e
