#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

docker run --rm --network ponsu_default \
  -e PONSU_PG_HOST=postgres \
  -e PONSU_PG_PORT=5432 \
  -e PONSU_PG_USER=ponsu \
  -e PONSU_PG_PASSWORD=ponsu \
  -e PONSU_PG_DB=ponsu \
  -e PONSU_PG_SSLMODE=disable \
  -v "$REPO_ROOT:/src" \
  -w /src \
  golang:1.24 \
  go run ./cmd/dbcheck
