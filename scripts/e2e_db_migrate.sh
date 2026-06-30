#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="vts-e2e-postgres"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

echo "Applying migrations to E2E Postgres container (${CONTAINER_NAME})..."
for f in supabase/migrations/*.sql; do
  echo "  -> ${f}"
  docker exec -i "${CONTAINER_NAME}" psql -U postgres -v ON_ERROR_STOP=1 < "${f}"
done

echo "Verifying demo attempt..."
docker exec "${CONTAINER_NAME}" psql -U postgres -v ON_ERROR_STOP=1 -c "
  SELECT id, status
  FROM attempts
  WHERE id = '00000000-0000-4000-8000-000000000001';
"
