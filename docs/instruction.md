
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

補足（推奨）:
- アプリの認証情報（OIDC クライアントシークレット等）は `.env.local` に置くことを推奨します（git管理しない）。
- PonSu の起動時に `.env` と `.env.local` を自動で読み込みます（存在しない場合は無視されます）。

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

dotenv 補足:
- `migrate.cmd` / `migrate.ps1` は `.env` / `.env.local` を読み込みます（OS環境変数が優先で上書きしません）。
- ワークスペース外の dotenv を使いたい場合は `PONSU_DOTENV_FILES`（カンマ `,` / セミコロン `;` 区切り）を設定してください。

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

## 4.1 設計ポイントの蓄積（開発フロー）

PonSu では、実装を進めながら「設計判断」を蓄積します。

- 設計ポイント集: [docs/design/設計ポイント集.md](design/設計ポイント集.md)
- タスクが `done` になったら、そのタスクの設計ポイントを **必ず追記**します。
- 追記する内容（最小）: 背景 / 決定 / 理由 / トレードオフ / 学び
- 目的: 後から意思決定の理由を追えるようにし、プロダクト設計の再現性を上げる。

## 5. ローカル起動（.env.local 利用）

### 5.0 推奨: 秘密情報はワークスペース外へ置く

Copilot / VS Code 拡張は、原理的にワークスペース内のファイルへアクセス可能です。
そのため、**秘密情報を Copilot から“確実に読めない”状態にしたい場合は、秘密ファイルをワークスペース外へ置く**運用にしてください。

PonSu は `PONSU_DOTENV_FILES` を設定すると、指定したパスの dotenv ファイルのみを読み込みます（デフォルトの `.env/.env.local` は読み込みません）。

OIDC を「特定メールだけログイン可」にしたい場合は、`PONSU_OIDC_ALLOWED_EMAILS` を設定します（カンマ `,` / セミコロン `;` 区切り）。

例（Windows / PowerShell）:

```powershell
# 例: ホーム配下に秘密ファイルを置く（このパスは例）
notepad $HOME\.ponsu.secrets.env

# 現在のターミナルセッションだけ有効
$env:PONSU_DOTENV_FILES = "$HOME\.ponsu.secrets.env"

# 例: 特定メールだけ許可（複数指定も可）
$env:PONSU_OIDC_ALLOWED_EMAILS = "morimasaki1990@gmail.com"

# 起動
go run ./cmd/ponsu
```

補足:
- `PONSU_DOTENV_FILES` はカンマ `,` またはセミコロン `;` 区切りで複数指定できます。
- 既にOS環境変数として設定済みのキーは、dotenv で上書きしません。

補足（重要）:
- これは **アプリ側の強制** です。Google 側の「テストユーザー」設定だけでは、状況によって他アカウントがログインできる場合があります。

`.env.local` を用意して起動する例:

```powershell
copy .env.example .env.local
notepad .env.local

# ponsu ルートで起動（.env/.env.local を読み込む）
go run ./cmd/ponsu
```

動作確認:
- http://127.0.0.1:8080/healthz
- http://127.0.0.1:8080/

### 4.1 PowerShellの実行ポリシーで弾かれる
`verify.ps1` を直接実行せず、`verify.cmd` を使ってください。

### 4.2 `go` が見つからない
Go のインストールが未完了か、PATH が通っていません。`go version` と `where go` で確認してください。

### 4.3 Docker が使えない
Docker Desktop が起動しているか確認してください。`docker version` が通ればOKです。

### 4.4 `sqlc generate` が失敗する（Dockerなし / WASM panic）

Windows 環境によっては、`sqlc` を `go run` フォールバックで実行した際に WASM パーサ由来の panic が出ることがあります。

回避策（推奨）:
- `sqlc` の Windows バイナリ（`sqlc.exe`）をダウンロードして `tools/sqlc/sqlc.exe` に配置
- `PONSU_SQLC_MODE=host` を設定して生成を実行

例（PowerShell）:

```powershell
cd ponsu
$env:PONSU_SQLC_MODE = "host"
\scripts\sqlc.cmd generate
```

