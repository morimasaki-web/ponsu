Param(
  [ValidateSet('up','down','version','force','goto','drop')]
  [string]$Command = 'up',

  # Number of steps for up/down
  [int]$Steps = 0
)

$ErrorActionPreference = 'Stop'

function Invoke-Step([string]$Name, [scriptblock]$Body, [int[]]$OkExitCodes = @(0)) {
  Write-Host "==> $Name" -ForegroundColor Cyan
  & $Body
  if ($OkExitCodes -notcontains $LASTEXITCODE) {
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


$pgHost = Get-Env 'PONSU_PG_HOST' '127.0.0.1'
$pgPort = Get-Env 'PONSU_PG_PORT' '5432'
$pgUser = Get-Env 'PONSU_PG_USER' 'ponsu'
$pgPassword = Get-Env 'PONSU_PG_PASSWORD' 'ponsu'
$pgDb = Get-Env 'PONSU_PG_DB' 'ponsu'
$pgSSLMode = Get-Env 'PONSU_PG_SSLMODE' 'disable'

$mode = Get-Env 'PONSU_MIGRATE_MODE' 'docker' # docker | host

$databaseUrl = "postgres://${pgUser}:${pgPassword}@${pgHost}:${pgPort}/${pgDb}?sslmode=${pgSSLMode}"
$migrationsAbsPath = (Resolve-Path (Join-Path $PSScriptRoot '..\migrations')).Path
$migrationsAbsPathFwd = ($migrationsAbsPath -replace '\\', '/')
$sourceUrl = "file://$migrationsAbsPathFwd"

Write-Host "Mode: $mode" -ForegroundColor DarkGray
Write-Host "DB(host): $databaseUrl" -ForegroundColor DarkGray
Write-Host "Migrations: $migrationsAbsPath" -ForegroundColor DarkGray
Write-Host "Source: $sourceUrl" -ForegroundColor DarkGray

$migrate = Get-Command migrate -ErrorAction SilentlyContinue

# Build common args (host mode)
$common = @(
  '-source', $sourceUrl,
  '-database', $databaseUrl
)

function Invoke-Migrate([string[]]$MigrateArgs) {
  if ($mode -eq 'docker') {
    $dockerDbUrl = "postgres://${pgUser}:${pgPassword}@postgres:5432/${pgDb}?sslmode=${pgSSLMode}"
    $dockerArgs = @(
      'run', '--rm',
      '--network', 'ponsu_default',
      '-v', "${migrationsAbsPath}:/migrations",
      'migrate/migrate:v4.17.1',
      '-path', '/migrations',
      '-database', $dockerDbUrl
    ) + $MigrateArgs

    & docker @dockerArgs
    return
  }

  $finalArgs = $common + $MigrateArgs

  if ($null -ne $migrate) {
    & migrate @finalArgs
    return
  }

  # Fallback: no global install required (slower)
  & go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1 @finalArgs
}

switch ($Command) {
  'up' {
    if ($Steps -gt 0) {
      Invoke-Step "migrate up $Steps" { Invoke-Migrate (@('up', "$Steps")) }
    } else {
      Invoke-Step 'migrate up' { Invoke-Migrate (@('up')) }
    }
  }
  'down' {
    if ($Steps -gt 0) {
      Invoke-Step "migrate down $Steps" { Invoke-Migrate (@('down', "$Steps")) }
    } else {
      Invoke-Step 'migrate down (1)' { Invoke-Migrate (@('down', '1')) }
    }
  }
  'version' {
    # `migrate version` returns exit code 1 when no migration has been applied yet.
    Invoke-Step 'migrate version' { Invoke-Migrate (@('version')) } @(0, 1)
  }
  'force' {
    if ($Steps -le 0) { throw "Steps must be a positive version for 'force'" }
    Invoke-Step "migrate force $Steps" { Invoke-Migrate (@('force', "$Steps")) }
  }
  'goto' {
    if ($Steps -le 0) { throw "Steps must be a positive version for 'goto'" }
    Invoke-Step "migrate goto $Steps" { Invoke-Migrate (@('goto', "$Steps")) }
  }
  'drop' {
    Invoke-Step 'migrate drop (force)' { Invoke-Migrate (@('drop', '-f')) }
  }
}
