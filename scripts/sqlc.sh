#!/usr/bin/env bash
set -euo pipefail

COMMAND="${1:-generate}"

PONSU_SQLC_MODE="${PONSU_SQLC_MODE:-docker}" # docker | host

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ "$PONSU_SQLC_MODE" == "docker" ]]; then
  case "$COMMAND" in
    generate)
      docker run --rm -v "$REPO_ROOT:/src" -w /src sqlc/sqlc:1.27.0 generate -f ./sqlc.yaml
      ;;
    version)
      docker run --rm sqlc/sqlc:1.27.0 version
      ;;
    *)
      echo "Unknown command: $COMMAND" >&2
      exit 2
      ;;
  esac
  exit 0
fi

if command -v sqlc >/dev/null 2>&1; then
  SQLC=(sqlc)
else
  SQLC=(go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0)
fi

case "$COMMAND" in
  generate)
    "${SQLC[@]}" generate -f "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../sqlc.yaml"
    ;;
  version)
    "${SQLC[@]}" version
    ;;
  *)
    echo "Unknown command: $COMMAND" >&2
    exit 2
    ;;
esac
