# デモ（静的ホスティング）手順（Vercel 推奨）

対象: PonSu のポートフォリオ向け「フロントのみ静的デモ」

- フロント: `ponsu/frontend`（Vite + React + TypeScript）
- ルーティング: Hash Router（`/#/`）
- デモ動作: `VITE_APP_MODE=demo`（バックエンドに接続しない）

関連: セキュリティ方針は [docs/demo-security.md](demo-security.md) を参照。

## 0. 前提

- Node.js / npm が利用できること
- 公開は **フロントのみ**（DB/認証/GraphQL は公開しない）

ビルド成果物:
- `ponsu/frontend/dist/`

## 1. 共通（ローカルでビルド）

PowerShell 例:

```powershell
cd ponsu\frontend

# demoモードでビルド（ビルド時に埋め込まれるので必ず設定する）
$env:VITE_APP_MODE = "demo"

npm ci
npm run build
```

確認:
- `ponsu/frontend/dist/index.html` が生成される

注意:
- `VITE_*` はビルド成果物へ埋め込まれるため、秘密情報は入れない（方針は [docs/demo-security.md](demo-security.md)）。

## 2. Vercel（推奨）

sheet-hive 側も Vercel を使っている前提で、PonSu も **Vercel を第一候補** とします。

理由:
- 設定がシンプル（`ponsu/frontend` をそのままビルドして `dist` を配信）
- カスタムヘッダ（CSP 等）の適用がしやすい
- project pages のようなサブパス配信問題を避けやすい

### 2.1 Vercel プロジェクト作成

Project 設定例:
- Framework Preset: `Vite`
- Root Directory: `ponsu/frontend`
- Install Command: `npm ci`
- Build Command: `npm run build`
- Output Directory: `dist`

Environment Variables:
- `VITE_APP_MODE=demo`
- `PONSU_FRONTEND_BASE=/`（通常は不要。サブパス運用する場合のみ）

デプロイ後、`/#/requests` などが表示できればOKです。

### 2.2 （任意）セキュリティヘッダ

`ponsu/frontend/vercel.json` を配置している場合、基本ヘッダ（CSP/`nosniff` 等）を同梱できます。
方針は [docs/demo-security.md](demo-security.md) を参照。

## 3. GitHub Pages（project pages）

GitHub Pages は多くの場合 `https://<user>.github.io/<repo>/` の形（サブパス配下）になるため、assets の参照パスに注意します。

### 2.1 推奨: Actions で `dist` をデプロイ

概要:
- `ponsu/frontend` をビルド
- `dist` を Pages へ公開

ポイント:
- サブパス配下で配信する場合、Vite の `base` を `/repo/` に合わせる
- Hash Router のため、SPAのリライト設定は基本不要（URL は `/#/...`）

`/repo/` を合わせるための環境変数（後述の Vite 設定が入っている前提）:

```powershell
$env:PONSU_FRONTEND_BASE = "/ponsu/"  # 例: リポジトリ名が ponsu の場合
$env:VITE_APP_MODE = "demo"
npm run build
```

次のようなワークフローを `.github/workflows/demo-pages.yml` として用意すると運用が楽です（例）:

```yaml
name: Deploy demo to GitHub Pages
on:
  workflow_dispatch:
  push:
    branches: [main]

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: true

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: ponsu/frontend/package-lock.json

      - name: Build
        working-directory: ponsu/frontend
        env:
          VITE_APP_MODE: demo
          PONSU_FRONTEND_BASE: /ponsu/
        run: |
          npm ci
          npm run build

      - uses: actions/upload-pages-artifact@v3
        with:
          path: ponsu/frontend/dist

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

初回だけ GitHub の設定で Pages を Actions に切り替えます:
- Repository Settings → Pages → Build and deployment → Source: GitHub Actions

### 2.2 手動で上げる場合

GitHub Pages の仕様上、手動アップロードよりも Actions を推奨します。
どうしても手動でやる場合は、`dist` を Pages の公開ソース（`gh-pages` ブランチ等）へ配置します。

## 4. Cloudflare Pages

Cloudflare Pages も静的ホスティング向きです。

Project 設定例:
- Root directory: `ponsu/frontend`
- Build command: `npm run build`
- Build output directory: `dist`
- Environment variables:
  - `VITE_APP_MODE=demo`
  - `PONSU_FRONTEND_BASE=/`（必要なら）

## 5. 公開後の動作確認チェック

- トップが表示される
- 一覧が表示される（`/#/requests`）
- 詳細が表示される（`/#/requests/REQ-0001`）
- 操作（承認/却下/差し戻し/再提出）でローカル状態と監査ログが変化する
- ネットワークタブで外部API呼び出しが発生していない（demo前提）

## 6. トラブルシュート

### 6.1 CSS/JS が 404 になる（GitHub Pagesで多い）
原因:
- `/` 前提でビルドしているのに、`/repo/` 配下で配信している

対処:
- `PONSU_FRONTEND_BASE=/repo/` を設定してビルドする（`/` で終わること）

### 6.2 画面が空白で、コンソールにエラーが出る
- `VITE_APP_MODE` が `demo` になっているか確認
- ブラウザのキャッシュをクリア（特に Service Worker を入れた場合）

