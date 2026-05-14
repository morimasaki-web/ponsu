# デモ（静的ホスティング）用 セキュリティ方針

対象: PonSu のポートフォリオ向け「フロントのみ静的デモ」（Vite + React + TypeScript）。

このデモは **外部認証 / DB / 課金が発生し得る外部API** を利用しないことで、公開時のリスクを物理的に最小化します。

## 1. スコープ

### 1.1 対象
- `ponsu/frontend` の `demo` モード（`VITE_APP_MODE=demo`）
- 静的ホスティング（GitHub Pages / Vercel / Cloudflare Pages 等）

### 1.2 非対象（本番アプリ側の話）
- OIDC（Client Secretを使う）
- サーバー側のセッション/CSRF
- DB（PostgreSQL）
- GraphQL API（gqlgen）

本番相当（`prod` モード）を公開する場合は、別途「サーバーを含む脅威モデル/運用」を定義すること。

## 2. 脅威モデル（簡易）

デモは静的サイトのため、主なリスクは以下に限定されます。

- JSバンドルに **秘密情報が混入** して公開される
- XSS（自前レンダリング/依存ライブラリ）
- 依存パッケージの脆弱性（`npm audit` 指摘）
- ホスティング設定不備（CSP/ヘッダ/キャッシュ）

逆に、以下のクラスのリスクは **デモの設計により発生しない** ことを目標にします。

- 認証情報の不正利用（そもそも認証機能を呼ばない）
- 個人情報の漏えい（そもそも保存/送信しない）
- 課金（外部APIや有料SaaSに接続しない）

## 3. ルール（必須）

### 3.1 秘密情報は「絶対に」フロントへ入れない
Vite の `VITE_*` 環境変数は **ビルド成果物に埋め込まれます**。したがって、以下は **禁止** です。

- `VITE_CLIENT_SECRET` / `VITE_API_KEY` / `VITE_*TOKEN*` 等の投入
- `.env` / `.env.local` に秘密を置いてデモ公開ビルドする
- `src/` へ認証/秘密をハードコードする

許可されるのは、公開しても問題ない「モード切替」や「公開URL」などのみです。

例:
- ✅ `VITE_APP_MODE=demo`
- ✅ `VITE_PUBLIC_BASE=/`（ホスティング都合）

### 3.2 demoモードはバックエンドに接続しない
デモは「フロントのみ」で完結させます。

- `demo` では GraphQL などのネットワークアクセスを行わない
- 画面データはモック（固定データ/ローカル状態）で表現する

## 4. 公開前チェック（必須）

### 4.1 リポジトリに秘密がないこと
- `ponsu/.env` / `ponsu/.env.local` / `ponsu/frontend/.env*` が **コミットされていない**
- `docs/` に secret が書かれていない

推奨チェック（例）:
- `git grep -n "CLIENT_SECRET\|API_KEY\|TOKEN\|SECRET"`

### 4.2 ビルド成果物に秘密がないこと
- `frontend/dist/` をホストに上げる前提で、`dist/assets/*.js` に secret が含まれない

推奨チェック（例）:
- `grep` でキーワード検索（PowerShellなら `Select-String`）

### 4.3 依存脆弱性の扱いを決める
- 原則: `npm audit` の結果を確認し、更新可能なら更新する
- デモの性質上、すぐに直せない場合は「理由」と「影響範囲」を記録して公開可否を判断する

この判断を曖昧にしないため、公開時点の `npm audit` 結果（件数と概要）を `docs/tasks.yaml` やデプロイ手順書にメモしてよい。

## 5. セキュリティヘッダ / CSP 方針（推奨）

静的ホスティングでは「ヘッダはホスティング側設定」になることが多いです。

### 5.1 推奨ヘッダ
- `Content-Security-Policy`（後述）
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: no-referrer`
- `Permissions-Policy: geolocation=(), microphone=(), camera=()`（不要機能を無効化）

補足:
- `X-Frame-Options` は CSP の `frame-ancestors` と併用/代替の関係。ホスティングの制約に合わせる。

### 5.2 CSP（まずは安全寄りの最小）

デモは外部スクリプトを使わない方針のため、基本は以下の思想で組みます。

- `default-src 'self'`
- `script-src 'self'`
- `style-src 'self'`（必要なら `style-src 'self' 'unsafe-inline'` を検討。まずは不要にしない）
- `img-src 'self' data:`（アイコン/埋め込みに必要なら）
- `connect-src 'none'`（demoでネットワークを禁止したい場合）

注意:
- Vite の構成やホスティングにより、`style-src` 等で調整が必要になる可能性があります。
- GitHub Pages はヘッダ制御が難しいため、CSP を厳密にやる場合は Vercel/Cloudflare Pages 等を検討。

## 6. データ取り扱い

- デモはユーザー入力を永続化しない（DB/外部ストレージなし）
- 保存が必要な場合は `localStorage` に限定し、個人情報を入れない
- 解析ツール（GA等）は原則入れない（入れる場合はプライバシー方針を別途用意）

## 7. インシデント対応（最小）

「秘密が混入した」可能性が出た場合:
1. 直ちに公開を停止（ホスティングの無効化/リダイレクト）
2. 秘密の失効（APIキー/トークンのローテーション）
3. コミット履歴からの除去（必要なら履歴書き換え）
4. 再発防止（チェックリスト強化、CIでの検知導入）

## 8. 関連

- タスク定義: `docs/tasks.yaml` の MVP-115
- デモアプリ: `ponsu/frontend`
