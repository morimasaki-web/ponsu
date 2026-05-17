# LEARN-026: 通知テンプレートのカスタマイズ機能

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | 組織ごとに通知メッセージをカスタマイズ可能にし、導入時の柔軟性を向上させる |
| **難易度** | 中級 |
| **ブランチ** | `enhance/026-notification-template` |
| **依存タスク** | なし |

## この課題で学ぶこと

- **`text/template`**: Go標準ライブラリのテンプレートエンジン（変数埋め込み・条件分岐）
- **DBへのテンプレート保存**: 設定をDBで管理するパターン
- **デフォルト値とカスタム値の優先順位**: DBに値がなければデフォルトを使うパターン

## 事前準備

**読んでおくファイル:**
- `internal/infrastructure/notify/slack.go` — 現在の通知実装を理解する
- `migrations/` 内の最新マイグレーションファイル — up/downの書き方を確認
- `db/query/` 内の任意のSQLファイル — sqlcクエリの書き方を確認
- `internal/infrastructure/postgres/` 内の任意のファイル — DBアクセス層の実装パターン

**理解しておく概念:**
- `text/template`: `{{.FieldName}}` で変数を埋め込むGoのテンプレートエンジン
- sqlc: SQLを書くとGoのコードを自動生成するツール（`go run github.com/sqlc-dev/sqlc/cmd/sqlc generate`）

## 実装手順

### ステップ1: 既存の通知実装を読む

まず `internal/infrastructure/notify/slack.go` を開いて、どのようにSlack通知を送っているかを理解してください。

確認すべき点:
- どんな情報（イベント種別・申請タイトル等）を使ってメッセージを組み立てているか
- メッセージの文字列はハードコードされているか

### ステップ2: マイグレーションで `notification_templates` テーブルを追加

`migrations/` 内の最新番号を確認して、次の番号でファイルを作成します。

例: 最新が `0010_xxx.up.sql` なら `0011_notification_templates.up.sql` を作成。

**up.sql の内容:**
```sql
CREATE TABLE IF NOT EXISTS notification_templates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL,
    event_type TEXT NOT NULL,
    template_body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, event_type)
);
```

**down.sql の内容:**
```sql
DROP TABLE IF EXISTS notification_templates;
```

マイグレーション実行:
```bash
bash scripts/migrate.sh up
# または Windows
scripts\migrate.cmd up
```

### ステップ3: sqlcクエリを追加

`db/query/notification_templates.sql` を新規作成して、以下のクエリを追加してください:

```sql
-- name: GetNotificationTemplate :one
SELECT id, org_id, event_type, template_body, created_at, updated_at
FROM notification_templates
WHERE org_id = $1 AND event_type = $2
LIMIT 1;

-- name: UpsertNotificationTemplate :one
INSERT INTO notification_templates (org_id, event_type, template_body)
VALUES ($1, $2, $3)
ON CONFLICT (org_id, event_type)
DO UPDATE SET template_body = EXCLUDED.template_body, updated_at = NOW()
RETURNING id, org_id, event_type, template_body, created_at, updated_at;
```

**sqlcでコード生成:**
```bash
go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
```

`internal/infrastructure/dbgen/` に新しいファイルが生成されることを確認してください。

### ステップ4: テンプレートレンダリング関数を実装

`internal/infrastructure/notify/template.go` を新規作成します。

```go
package notify

import (
    "bytes"
    "text/template"
)

// デフォルトテンプレート（組織がカスタマイズしていない場合）
var defaultTemplates = map[string]string{
    "RequestSubmitted": "新しい申請が届きました: {{.Title}}（申請者: {{.SubmitterName}}）",
    "RequestApproved":  "申請が承認されました: {{.Title}}",
    "RequestRejected":  "申請が却下されました: {{.Title}}",
}

// TemplateData はテンプレートに渡すデータ
type TemplateData struct {
    Title         string
    SubmitterName string
    // 必要に応じてフィールドを追加
}

// RenderTemplate はテンプレート文字列にデータを埋め込む
func RenderTemplate(templateBody string, data TemplateData) (string, error) {
    tmpl, err := template.New("notification").Parse(templateBody)
    if err != nil {
        return "", fmt.Errorf("テンプレートのパースに失敗: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("テンプレートの実行に失敗: %w", err)
    }

    return buf.String(), nil
}

// GetTemplate はカスタムテンプレートがなければデフォルトを返す
func GetTemplate(customTemplate *string, eventType string) string {
    if customTemplate != nil && *customTemplate != "" {
        return *customTemplate
    }
    if def, ok := defaultTemplates[eventType]; ok {
        return def
    }
    return "通知: {{.Title}}"
}
```

**ポイント:** `text/template` は `{{.フィールド名}}` で構造体のフィールドを埋め込めます。

### ステップ5: slack.go を修正してテンプレートを使うように変更

`internal/infrastructure/notify/slack.go` を修正して、DBからテンプレートを取得し、なければデフォルトを使うように変更します。

```go
// メッセージ組み立て部分を変更
// 変更前: ハードコードされた文字列
// 変更後: GetTemplate でテンプレートを取得 → RenderTemplate でデータを埋め込む
```

**ポイント:** DBアクセスには `dbgen.Queries` を使います。既存のコードを参考にしてください。

### ステップ6: テストを書く

`internal/infrastructure/notify/template_test.go` を作成します。

最低限テストすべきこと:
- `RenderTemplate` がテンプレートに正しくデータを埋め込めるか
- カスタムテンプレートがある場合はそちらが使われるか
- カスタムテンプレートがない場合はデフォルトが使われるか

### 最終ステップ: 動作確認

```bash
go test ./internal/infrastructure/notify/...
go vet ./...
```

## 受け入れ条件

- [ ] `notification_templates` テーブルが作成される（マイグレーション up/down 動作）
- [ ] sqlc generate が成功する
- [ ] テンプレートからメッセージを生成できる
- [ ] カスタムテンプレートが優先される（DBに値があればそちらを使う）
- [ ] `go test ./internal/infrastructure/notify/...` が通る

## 詰まったときのヒント

**Q: `text/template` の書き方が分からない**
→ `template.New("name").Parse("Hello, {{.Name}}!")` でパース、`.Execute(w, data)` で実行します。`data` は構造体でOK。

**Q: sqlcのSQLが構文エラーになる**
→ `db/query/requests.sql` の書き方をそのまま真似てみてください。コメント `-- name: XXX :one` が重要です。

**Q: DBからテンプレートを取得する方法が分からない**
→ `internal/infrastructure/postgres/` の既存ファイルで `q.GetXxx(ctx, ...)` のパターンを探してください。

**Q: マイグレーション番号の確認方法**
→ `ls migrations/` で最新番号を確認してください。

## 参考ファイル

- `internal/infrastructure/notify/slack.go` — 現在の通知実装（ここを修正する）
- `db/query/requests.sql` — sqlcクエリの書き方の参考
- `migrations/` — マイグレーションファイルの書き方の参考
- `internal/infrastructure/postgres/event_store.go` — DBアクセスパターンの参考

## 設計ポイント（実装後に記録）

> 実装完了後、気づいた設計判断・トレードオフ・学びを `docs/design/設計ポイント集.md` に追記してください。
> 記録する内容（最小）: 背景 / 決定 / 理由 / トレードオフ / 学び
