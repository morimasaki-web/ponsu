$ErrorActionPreference = 'Stop'

function Invoke-Step([string]$Name, [scriptblock]$Body) {
  Write-Host "==> $Name" -ForegroundColor Cyan
  & $Body
  if ($LASTEXITCODE -ne 0) {
    throw "Step failed ($Name) with exit code $LASTEXITCODE"
  }
}

# Run dbcheck inside Docker network to avoid Windows host auth/network quirks.
Invoke-Step 'dbcheck (docker)' {
  docker run --rm --network ponsu_default `
    -e PONSU_PG_HOST=postgres `
    -e PONSU_PG_PORT=5432 `
    -e PONSU_PG_USER=ponsu `
    -e PONSU_PG_PASSWORD=ponsu `
    -e PONSU_PG_DB=ponsu `
    -e PONSU_PG_SSLMODE=disable `
    -v "${PSScriptRoot}\..:/src" `
    -w /src `
    golang:1.24 `
    go run ./cmd/dbcheck
}
