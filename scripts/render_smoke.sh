#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

if [[ -z "${API_BASE:-}" ]]; then
  echo "ERROR: API_BASE is required (e.g. https://my-api.onrender.com)"
  exit 1
fi

echo "Running Render post-deploy smoke against ${API_BASE}..."
API_BASE="${API_BASE}" node scripts/e2e_smoke_api.mjs
