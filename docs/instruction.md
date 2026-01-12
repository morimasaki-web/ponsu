
# PonSu 開発ガイド（開発環境・実行・検証）

このドキュメントは、PonSu を開発・検証するための「最低限の前提」と「迷わない手順」をまとめたものです。

## 1. 前提（Windows）

### 1.1 必須ツール
- Go（推奨: 1.22+）
- Git
- Docker Desktop（PostgreSQL / MinIO をローカルで立てるため）

確認コマンド:

```powershell
go version
git --version
docker version
docker compose version
```

#### Goが見つからない場合
PowerShellで `go` が見つからない場合は、Goのインストールと PATH 設定が未完了です。
- いったん新しいターミナルを開き直す
- `where go` で解決先を確認する

```powershell
where go
```

### 1.2 推奨ツール（任意）
- VS Code
- Bash 環境（Git Bash など）

## 2. ローカル依存サービス（Postgres / MinIO）

PonSu はMVPで PostgreSQL と MinIO を使う想定です。

事前準備:

```powershell
copy .env.example .env
```

起動:

```powershell
docker compose up -d
```

停止:

```powershell
docker compose down
```

疎通チェック例:

```powershell
docker compose ps
docker compose exec postgres pg_isready -U ponsu -d ponsu
```

MinIO コンソール:
- http://127.0.0.1:9001

## 2.1 DBマイグレーション（golang-migrate/migrate）

PonSu は SQL マイグレーションに `golang-migrate/migrate` を使います。

ポイント:
 - Windows は環境差でホストからDBへ接続しづらいことがあるため、デフォルトは Docker 上の `migrate/migrate` を使って実行します（`docker compose up -d` 済みが前提）。
- `migrate` コマンドをグローバルに入れていなくても、スクリプトがフォールバック実行します。
- Windows の PowerShell 実行ポリシー回避のため、`.cmd` ラッパーを推奨します。

事前条件（Postgres起動）:

```powershell
docker compose up -d
```

実行例:

```powershell
# 最新へ適用
ponsu\scripts\migrate.cmd up

# 直近1つ戻す（デフォルト）
ponsu\scripts\migrate.cmd down

# 現在のバージョン確認
ponsu\scripts\migrate.cmd version
```

補足:
- `up` / `down` のステップ数を指定したい場合は第2引数で渡します（例: `ponsu\scripts\migrate.cmd up 1`）。
- Dockerを使わずホストから直接接続したい場合は `PONSU_MIGRATE_MODE=host` を環境変数に設定してください。
- Bash環境では `bash ponsu/scripts/migrate.sh up` のように実行できます。

## 3. 無料検証（必ず回す）

PonSu の作業フローでは、外部モデルレビューの代わりに「無料の静的解析・脆弱性チェック」を必ず実行します。

### 3.1 まとめて実行（Windows推奨）

Windows では PowerShell の実行ポリシーにより `.ps1` の直接実行がブロックされることがあるため、ラッパーの `.cmd` を使います。

```powershell
ponsu\scripts\verify.cmd
```

個別にスキップしたい場合:

```powershell
ponsu\scripts\verify.cmd -SkipLint -SkipVuln
```

### 3.2 まとめて実行（Bash）

```bash
bash ponsu/scripts/verify.sh
```

### 3.3 verify が実行すること
- `go test ./...`
- `go vet ./...`
- `golangci-lint run ./...`
- `govulncheck ./...`

補足:
- `golangci-lint` / `govulncheck` が未インストールでも、`go run <module>@<version>` でフォールバック実行します（無料・ただし遅め）。

## 4. よくあるトラブル

### 4.1 PowerShellの実行ポリシーで弾かれる
`verify.ps1` を直接実行せず、`verify.cmd` を使ってください。

### 4.2 `go` が見つからない
Go のインストールが未完了か、PATH が通っていません。`go version` と `where go` で確認してください。

### 4.3 Docker が使えない
Docker Desktop が起動しているか確認してください。`docker version` が通ればOKです。

