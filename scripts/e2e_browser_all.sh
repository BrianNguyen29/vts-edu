#!/usr/bin/env bash
# Cross-browser E2E runner. Spins up the same DB + API as
# scripts/e2e_browser.sh, then runs Playwright with the full
# Chromium + Firefox + WebKit matrix.
#
# WebKit requires system libraries (libgtk-4, libgraphene-1.0,
# libxslt, libevent-2.1, libopus, libgstallocators, …) that are not
# always present on developer machines. The script probes each
# browser before invoking Playwright and prints the missing libs for
# WebKit instead of failing silently. Chromium and Firefox typically
# work without extra system packages.
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

# Make sure the requested browsers are present. Chromium is downloaded
# by `pnpm e2e:install`; Firefox + WebKit require an explicit install.
echo "Ensuring Playwright browsers are installed..."
pnpm --filter @vts-edu/web exec playwright install chromium firefox webkit || {
  echo "WARN: playwright install failed; falling back to chromium-only" >&2
}

# Probe WebKit's host deps. The browser binary is downloaded but
# libgtk-4 and friends are required to actually launch it.
webkit_ok=1
if ! pnpm --filter @vts-edu/web exec node -e "
import('@playwright/test').then(async ({ webkit }) => {
  try { const b = await webkit.launch({ headless: true }); await b.close(); }
  catch (e) { process.stderr.write(e.message); process.exit(1); }
}).catch((e) => { process.stderr.write(e.message); process.exit(1); });
" 2>/dev/null; then
  webkit_ok=0
fi

pnpm e2e:db:stop >/dev/null 2>&1 || true
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

export PLAYWRIGHT_BROWSERS=1

if [[ "${webkit_ok}" -eq 0 ]]; then
  echo "WARN: WebKit host dependencies missing; running chromium + firefox only" >&2
  echo "      Re-run after: pnpm --filter @vts-edu/web exec playwright install --with-deps webkit" >&2
  pnpm web:e2e --project=chromium --project=firefox
else
  echo "Running Playwright cross-browser matrix (chromium + firefox + webkit)..."
  pnpm web:e2e
fi
