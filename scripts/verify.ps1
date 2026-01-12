Param(
  [switch]$SkipVuln,
  [switch]$SkipLint,
  [switch]$SkipTidy
)

$ErrorActionPreference = 'Stop'

function Invoke-Step([string]$Name, [scriptblock]$Body) {
  Write-Host "==> $Name" -ForegroundColor Cyan
  & $Body

  # NOTE: External commands don't throw even with $ErrorActionPreference='Stop'.
  if ($LASTEXITCODE -ne 0) {
    throw "Step failed ($Name) with exit code $LASTEXITCODE"
  }
}

# Force a stable toolchain to avoid tool incompatibilities with newer local Go.
# Go will auto-download this toolchain if missing.
$env:GOTOOLCHAIN = "go1.24.11"

if (-not $SkipTidy) {
  Invoke-Step "go mod tidy" {
    go --% mod tidy -go=1.22.0
  }
}

Invoke-Step "go test" { go test ./... }
Invoke-Step "go vet" { go vet ./... }

if (-not $SkipLint) {
  $golangci = Get-Command golangci-lint -ErrorAction SilentlyContinue
  if ($null -ne $golangci) {
    Invoke-Step "golangci-lint" { golangci-lint run ./... }
  } else {
    # Fallback: no global install needed (slower)
    Invoke-Step "golangci-lint (go run fallback)" { go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run ./... }
  }
}

if (-not $SkipVuln) {
  $govuln = Get-Command govulncheck -ErrorAction SilentlyContinue
  if ($null -ne $govuln) {
    Invoke-Step "govulncheck" { govulncheck ./... }
  } else {
    # Fallback: no global install needed (slower)
    Invoke-Step "govulncheck (go run fallback)" { go run golang.org/x/vuln/cmd/govulncheck@latest ./... }
  }
}

Write-Host "OK" -ForegroundColor Green
