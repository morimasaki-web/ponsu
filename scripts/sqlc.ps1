Param(
  [ValidateSet('generate','version')]
  [string]$Command = 'generate'
)

$ErrorActionPreference = 'Stop'

function Invoke-Step([string]$Name, [scriptblock]$Body) {
  Write-Host "==> $Name" -ForegroundColor Cyan
  & $Body
  if ($LASTEXITCODE -ne 0) {
    throw "Step failed ($Name) with exit code $LASTEXITCODE"
  }
}

# Keep toolchain stable (same philosophy as verify.ps1)
$env:GOTOOLCHAIN = "go1.24.11"

function Get-Env([string]$Key, [string]$Default) {
  $item = Get-Item "Env:$Key" -ErrorAction SilentlyContinue
  if ($null -eq $item) { return $Default }
  if ([string]::IsNullOrWhiteSpace($item.Value)) { return $Default }
  return $item.Value
}

$mode = Get-Env 'PONSU_SQLC_MODE' 'docker' # docker | host

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$localSqlcPath = (Join-Path $repoRoot 'tools/sqlc/sqlc.exe')

$sqlc = Get-Command sqlc -ErrorAction SilentlyContinue

function Invoke-Sqlc([string[]]$SqlcArgs) {
  if ($mode -eq 'docker') {
    $dockerArgs = @(
      'run', '--rm',
      '-v', "${repoRoot}:/src",
      '-w', '/src',
      'sqlc/sqlc:1.27.0'
    ) + $SqlcArgs

    & docker @dockerArgs
    return
  }

  if (Test-Path $localSqlcPath) {
    & $localSqlcPath @SqlcArgs
    return
  }

  if ($null -ne $sqlc) {
    & sqlc @SqlcArgs
    return
  }

  # Fallback (may be unstable on some Windows setups due to WASM parser)
  & go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 @SqlcArgs
}

switch ($Command) {
  'generate' {
    Invoke-Step "sqlc generate ($mode)" { Invoke-Sqlc @('generate', '-f', './sqlc.yaml') }
  }
  'version' {
    Invoke-Step "sqlc version ($mode)" { Invoke-Sqlc @('version') }
  }
}
