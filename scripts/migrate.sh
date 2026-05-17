#!/usr/bin/env bash
set -euo pipefail

COMMAND="${1:-up}"
STEPS="${2:-0}"

PONSU_PG_HOST="${PONSU_PG_HOST:-127.0.0.1}"
PONSU_PG_PORT="${PONSU_PG_PORT:-5432}"
PONSU_PG_USER="${PONSU_PG_USER:-ponsu}"
PONSU_PG_PASSWORD="${PONSU_PG_PASSWORD:-ponsu}"
PONSU_PG_DB="${PONSU_PG_DB:-ponsu}"
PONSU_PG_SSLMODE="${PONSU_PG_SSLMODE:-disable}"

DATABASE_URL="postgres://${PONSU_PG_USER}:${PONSU_PG_PASSWORD}@${PONSU_PG_HOST}:${PONSU_PG_PORT}/${PONSU_PG_DB}?sslmode=${PONSU_PG_SSLMODE}"
MIGRATIONS_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../migrations"

if command -v migrate >/dev/null 2>&1; then
  MIGRATE=(migrate)
else
  MIGRATE=(go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1)
fi

COMMON=(-path "$MIGRATIONS_PATH" -database "$DATABASE_URL")

echo "DB: $DATABASE_URL" >&2
echo "Migrations: $MIGRATIONS_PATH" >&2

case "$COMMAND" in
  up)
    if [[ "$STEPS" != "0" ]]; then
      "${MIGRATE[@]}" "${COMMON[@]}" up "$STEPS"
    else
      "${MIGRATE[@]}" "${COMMON[@]}" up
    fi
    ;;
  down)
    if [[ "$STEPS" != "0" ]]; then
      "${MIGRATE[@]}" "${COMMON[@]}" down "$STEPS"
    else
      "${MIGRATE[@]}" "${COMMON[@]}" down 1
    fi
    ;;
  version)
    "${MIGRATE[@]}" "${COMMON[@]}" version
    ;;
  force|goto)
    if [[ "$STEPS" == "0" ]]; then
      echo "STEPS required for $COMMAND" >&2
      exit 2
    fi
    "${MIGRATE[@]}" "${COMMON[@]}" "$COMMAND" "$STEPS"
    ;;
  drop)
    "${MIGRATE[@]}" "${COMMON[@]}" drop -f
    ;;
  *)
    echo "Unknown command: $COMMAND" >&2
    exit 2
    ;;
esac
