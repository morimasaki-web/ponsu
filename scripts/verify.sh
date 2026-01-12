#!/usr/bin/env bash
set -euo pipefail

skip_vuln="${SKIP_VULN:-0}"
skip_lint="${SKIP_LINT:-0}"
skip_tidy="${SKIP_TIDY:-0}"

# Force a stable toolchain to avoid tool incompatibilities with newer local Go.
export GOTOOLCHAIN="go1.24.11"

step() {
  echo "==> $1"
  shift
  "$@"
}

if [ "$skip_tidy" = "0" ]; then
  step "go mod tidy" go mod tidy -go=1.22.0
fi

step "go test" go test ./...
step "go vet" go vet ./...

if [ "$skip_lint" = "0" ]; then
  if command -v golangci-lint >/dev/null 2>&1; then
    step "golangci-lint" golangci-lint run ./...
  else
    step "golangci-lint (go run fallback)" go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run ./...
  fi
fi

if [ "$skip_vuln" = "0" ]; then
  if command -v govulncheck >/dev/null 2>&1; then
    step "govulncheck" govulncheck ./...
  else
    step "govulncheck (go run fallback)" go run golang.org/x/vuln/cmd/govulncheck@latest ./...
  fi
fi

echo "OK"
