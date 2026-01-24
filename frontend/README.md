# PonSu Frontend (SPA)

ポートフォリオ向けに「静的デモ（A）」を前提としたSPAです。Vite + React + TypeScript。

## モード

- `VITE_APP_MODE=demo`（デフォルト推奨）
  - バックエンドへ**一切アクセスしない**
  - 認証/DB/課金要素を排除し、モックデータで画面の動きだけ確認
- `VITE_APP_MODE=prod`
  - ローカル開発や将来の正規版向け
  - `/graphql` と `/auth` をバックエンドへプロキシして動作

## 開発手順（静的デモ）

1) フロントエンド（このディレクトリ）

- `cd frontend`
- `npm install`
- `.env` を作成
  - PowerShell: `Copy-Item .env.example .env`
  - cmd: `copy .env.example .env`
- `npm run dev`

Vite dev server は `http://localhost:5173` で起動します。
（設定により `http://127.0.0.1:5173` でもアクセスできるはずです。開けない場合は `vite.config.ts` の `server.host` を確認してください）

## 正規版（prod）: バックエンドへのプロキシ

- `VITE_APP_MODE=prod` にして起動します（`.env` を編集）
- `vite.config.ts` で `/graphql` と `/auth` をバックエンドへプロキシします
- バックエンドのオリジンを変える場合は環境変数 `PONSU_BACKEND_ORIGIN` を使ってください
  - 例: `PONSU_BACKEND_ORIGIN=http://localhost:8080`

### OIDCログインで `invalid state` が出る場合

OIDCは `/auth/login` 開始時に `ponsu_oidc_state` cookie を保存し、`/auth/callback` で同じcookieを読み取って検証します。
そのため **ブラウザ上でアクセスしているホスト名が混在すると失敗**します（例: `http://localhost:5173` でログイン開始 → `http://127.0.0.1:8080/auth/callback` に戻る）。

- 対策: 使うホスト名を統一する（`localhost` か `127.0.0.1` のどちらかに揃える）
- さらに確実: `PONSU_OIDC_REDIRECT_URL` を「実際にブラウザが受ける callback URL」に固定し、OIDC側の許可リダイレクトURLも同じ値にする
  - 例（Vite経由で完結させる）: `PONSU_OIDC_REDIRECT_URL=http://localhost:5173/auth/callback`
  - 例（backend直）: `PONSU_OIDC_REDIRECT_URL=http://127.0.0.1:8080/auth/callback`

## 動作確認

- demo: 画面にダミーの `me` 情報が表示される
- prod: 画面の「ログイン」から OIDC ログイン → `me` 情報が表示される

