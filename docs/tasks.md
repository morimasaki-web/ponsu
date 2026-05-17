# Ponsu タスク管理

> このファイルは `docs/tasks.yaml` をMarkdown形式にリニューアルしたものです。
> 過去のアーカイブは `docs/tasks_mvp_archive.yaml` / `docs/tasks_learn_archive.yaml` を参照してください。

## 方針（MVP確定）

- DB: PostgreSQL
- DBアクセス: sqlc
- API: GraphQL (gqlgen)
- 監査/状態管理: Event Sourcing（event_store + 同期プロジェクション）
- UI: SPA化を見据えたSSR
- 認可: RBAC
- 通知: Slack（Incoming Webhook）
- 添付: MinIO（S3互換）
- マイグレーション: golang-migrate/migrate

---

## MVPタスク

### 進行中

- [ ] **MVP-120** `feature/mvp-120-portfolio-demo` [high]  
  ポートフォリオ: デモ掲載（URL/画像/説明/技術スタック）  
  _ポートフォリオにデモ版を掲載し、機能・設計・使い方が伝わる形に整える。_
  - 受け入れ条件:
    - ポートフォリオからデモへ到達できる
    - 主要機能（承認/却下/差し戻し/再提出/監査）が説明されている

- [ ] **MVP-121** `fix/mvp-121-go-upgrade` [high]  
  セキュリティ: Go標準ライブラリ脆弱性対応（Go 1.24.13+ または 1.25.7+）  
  _govulncheckで検出された脆弱性（crypto/tls, net/url）に対応する。_
  - 段階的対応:
    - [x] go.mod toolchain を go1.24.13 に更新（GO-2026-4337対応）
    - [ ] Go 1.25.7以上へアップグレード（全脆弱性対応）
  - 受け入れ条件:
    - GitHub Actions の govulncheck が通る
    - govulncheck ./... で脆弱性が0件
    - すべてのテストが通る

### 未着手

- [ ] **MVP-122** `feature/mvp-122-demo-guide-reset` [medium] depends: MVP-121, MVP-119  
  デモ: 操作フローのガイド/リセット導線（モック状態の再現性）  
  _デモ閲覧者が迷わず試せるように、サンプルデータ/操作ガイドと「状態リセット」導線を用意する。_
  - 受け入れ条件:
    - 初見でも「どこを触れば良いか」が分かる
    - 状態が壊れてもワンクリックで初期状態に戻せる

- [ ] **MVP-123** `feature/mvp-123-spa-auth-flow` [medium] depends: MVP-112, MVP-010  
  正規版: SPAのログイン/ログアウト導線（OIDCと整合）  
  _prodモードで、SPAからOIDCログイン/ログアウトを自然に扱えるようにする。_
  - 受け入れ条件:
    - 未ログイン時にログインへ誘導できる
    - ログアウト後に適切な画面へ遷移し、キャッシュされたデータが残らない

- [ ] **MVP-124** `feature/mvp-124-spa-error-ux` [medium] depends: MVP-112, MVP-118  
  正規版: GraphQLエラー/ローディング表示の統一（UX最小改善）  
  _GraphQL/ネットワークエラーやローディングを画面全体で統一して扱う。_
  - 受け入れ条件:
    - エラー時に「何が起きたか/次にどうするか」が表示される
    - 主要画面でローディング状態が統一され、操作不能が分かる

---

## 学習用エンハンスタスク

> ブランチ命名規則: `enhance/NNN-short-slug`  
> 目的: PonSuの設計を深く理解し、自分で機能追加できる状態になる

### 完了済み

- [x] **LEARN-018** `enhance/018-event-store-query`  
  監査: 集約単位のイベント履歴取得API実装

- [x] **LEARN-019** `enhance/019-user-stats-projector`  
  分析: ユーザー別申請統計の自動集計機能追加

- [x] **LEARN-020** `enhance/020-graphql-field-resolver`  
  API: 申請者情報の入れ子取得対応

- [x] **LEARN-021** `enhance/021-rbac-permission`  
  権限: 申請データエクスポート権限の追加

- [x] **LEARN-022** `enhance/022-new-event-type`  
  機能: 申請コメント機能の追加（イベント定義）

- [x] **LEARN-022F** `enhance/022f-comment-ui`  
  Frontend: 申請コメント機能のUI実装

- [x] **LEARN-023** `enhance/023-request-search`  
  検索: 申請の検索・フィルタリング機能実装

- [x] **LEARN-023F** `enhance/023f-search-ui`  
  Frontend: 申請検索UIとフィルター機能

- [x] **LEARN-023F-DEMO** `enhance/023f-demo-search`  
  Frontend: デモモード用の検索機能実装

- [x] **LEARN-024** `enhance/024-dashboard-stats`  
  統計: ダッシュボード集計クエリの実装

- [x] **LEARN-024F** `enhance/024f-dashboard-ui`  
  Frontend: ダッシュボード画面の実装

- [x] **LEARN-025** `enhance/025-attachment-download`  
  添付: ファイルダウンロード署名URL生成機能

- [x] **LEARN-025F** `enhance/025f-download-button`  
  Frontend: 添付ファイルダウンロードボタンの実装

- [x] **LEARN-027** `enhance/027-audit-query-api`  
  監査: 操作ログの検索・ページング機能実装

- [x] **LEARN-027F** `enhance/027f-audit-log-ui`  
  Frontend: 監査ログ閲覧画面の実装

- [x] **LEARN-028** `enhance/028-idempotency-test`  
  信頼性: 重複リクエスト防止機能のテスト追加

### 未着手

- [x] **LEARN-026** `enhance/026-notification-template` [medium]  
  通知: 通知テンプレートのカスタマイズ機能  
  _組織ごとに通知メッセージをカスタマイズ可能にする（text/template活用）。_
  - 受け入れ条件:
    - テンプレートからメッセージを生成できる
    - カスタムテンプレートが優先される
    - go test が通る

- [ ] **LEARN-029** `enhance/029-projector-resilience` [high] depends: LEARN-019  
  Projector: エラーハンドリングとリトライ機能追加  
  _一時的な障害を自動復旧できるように指数バックオフでリトライする。_
  - 受け入れ条件:
    - 一時的なエラーはリトライされる（最大3回、指数バックオフ）
    - リトライ失敗イベントが記録される
    - go test が通る

- [x] **LEARN-030** `enhance/030-graphql-error-handling` [medium]  
  GraphQL: エラーレスポンスの統一と構造化  
  _gqlgenのエラーハンドリング機構を活用してエラーコードを付与する。_
  - 受け入れ条件:
    - エラーに適切なコードが付与される
    - go test が通る

- [ ] **LEARN-031** `enhance/031-projector-replay` [high] depends: LEARN-019  
  復旧: 読み取りモデル再構築コマンド実装  
  _データ不整合発生時にイベントストアから読み取りモデルを再構築する。_
  - 受け入れ条件:
    - リプレイコマンドでデータを再構築できる
    - 進捗がテーブルに記録される

- [ ] **LEARN-032** `enhance/032-transaction-boundary` [high] depends: LEARN-018  
  信頼性: トランザクション境界の明確化と例外時ロールバック  
  _エラー発生時の部分更新を防止し、データ整合性を保証する。_
  - 受け入れ条件:
    - トランザクション境界が明確
    - エラー時にロールバックされる
    - ドキュメントが追加されている

- [ ] **LEARN-033** `enhance/033-demo-mock-data` [high]  
  デモ: モックデータ生成とローカルストレージ永続化  
  _デモモード用のモックデータを生成し、再現性のあるデモ環境を構築する。_
  - 受け入れ条件:
    - デモモードで動作する
    - モックデータが永続化される
    - リセット機能が動作する

- [ ] **LEARN-034** `enhance/034-frontend-pagination` [medium] depends: LEARN-023  
  Frontend: リクエスト一覧のページネーション実装  
  _大量の申請データを効率的に表示する（カーソルベースページネーション）。_
  - 受け入れ条件:
    - ページネーションが動作する
    - ローディング状態が表示される

- [ ] **LEARN-035** `enhance/035-error-toast` [high]  
  Frontend: エラー境界とトースト通知の統一  
  _GraphQLエラーやネットワークエラーを統一的に処理し、ユーザーに分かりやすく表示する。_
  - 受け入れ条件:
    - エラー時にトースト通知が表示される
    - 致命的エラーはError Boundaryがキャッチする

- [ ] **LEARN-036** `enhance/036-loading-skeleton` [medium]  
  Frontend: ローディング状態の統一とスケルトンUI  
  _データ読み込み中のUXを改善し、スケルトンスクリーンを実装する。_
  - 受け入れ条件:
    - ローディング時にスケルトンが表示される
    - 各画面で統一されている

- [ ] **LEARN-037** `enhance/037-frontend-testing` [medium]  
  Frontend: テスト環境セットアップ（Vitest + Testing Library）  
  _フロントエンドのテスト環境を構築し、コンポーネントテストを実行可能にする。_
  - 受け入れ条件:
    - npm test でテストが実行できる
    - @testing-library/react でコンポーネントをテストできる
    - urqlのクエリ/ミューテーションをモックできる
