# LEARN-029: Projectorのエラーハンドリングとリトライ機能

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | Projectorでエラーが発生したとき自動でリトライし、一時的な障害から自動復旧できるようにする |
| **難易度** | 中級〜上級 |
| **ブランチ** | `enhance/029-projector-resilience` |
| **依存タスク** | LEARN-019 |

## この課題で学ぶこと

- **指数バックオフ**: リトライ間隔を徐々に伸ばす戦略（1s → 2s → 4s）
- **デッドレターキューの概念**: リトライし続けても失敗したイベントをどう扱うか
- **構造化ログ**: `slog` でエラー情報を機械可読な形で記録するGoの作法

## 事前準備

**読んでおくファイル:**
- `internal/infrastructure/projector/runner.go` — Projectorの実行ループを理解する（**必須**）
- `internal/infrastructure/projector/requests_projector.go` — 具体的なProjector実装の参考
- `cmd/projector/main.go` — Projectorの起動方法を確認

**理解しておく概念:**
- **Projector**: イベントストアのイベントを読み取り、読み取りモデル（テーブル）を構築するプロセス
- **指数バックオフ**: 1回目の失敗後1秒待ち、2回目は2秒、3回目は4秒待つリトライ戦略

## 実装手順

### ステップ1: 現在のrunner.goを完全に理解する

`internal/infrastructure/projector/runner.go` を開いて、以下を把握してください:
- `Runner` 構造体のフィールド
- イベントをどのように処理しているか
- 現在のエラーハンドリング（おそらくエラーをログして次に進む or パニックする）

このファイルを修正するため、**現在の挙動を完全に把握する**ことが重要です。

### ステップ2: リトライロジックの設計

リトライが必要なエラーとそうでないエラーを区別します:

```go
// リトライすべきエラー: 一時的な障害（DBの一時的な切断など）
// リトライしないエラー: バグ由来の恒久的な失敗（ペイロードのデコード失敗など）
```

今回は「すべてのエラーを最大3回リトライする」シンプルな実装を目指します。

### ステップ3: リトライ関数を実装する

`internal/infrastructure/projector/runner.go` に以下のような関数を追加します（参考構造）:

```go
// withRetry は最大maxRetry回、指数バックオフでfnを実行する
func withRetry(ctx context.Context, maxRetry int, fn func() error) error {
    var err error
    for i := 0; i < maxRetry; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        if i < maxRetry-1 {
            // 指数バックオフ: 1s, 2s, 4s
            wait := time.Duration(1<<uint(i)) * time.Second
            slog.Warn("projector: リトライします",
                "attempt", i+1,
                "wait", wait,
                "error", err,
            )
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(wait):
            }
        }
    }
    return fmt.Errorf("最大リトライ回数(%d)を超えました: %w", maxRetry, err)
}
```

**ポイント:** `1<<uint(i)` は2の累乗（i=0で1, i=1で2, i=2で4）を計算するビットシフトです。

### ステップ4: `failed_events` テーブルを追加（リトライ失敗の記録）

新しいマイグレーションファイルを作成します:

```sql
-- up.sql
CREATE TABLE IF NOT EXISTS failed_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    event_id        UUID NOT NULL,
    projector_name  TEXT NOT NULL,
    error_message   TEXT NOT NULL,
    retry_count     INT NOT NULL DEFAULT 0,
    failed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

```sql
-- down.sql
DROP TABLE IF EXISTS failed_events;
```

### ステップ5: リトライ失敗時のエラー記録

3回リトライしても失敗した場合、`failed_events` テーブルに記録するロジックを追加します。

sqlcクエリ `db/query/failed_events.sql` を作成:
```sql
-- name: InsertFailedEvent :exec
INSERT INTO failed_events (org_id, event_id, projector_name, error_message, retry_count)
VALUES ($1, $2, $3, $4, $5);
```

`sqlc generate` を実行してコードを生成してください。

### ステップ6: テストを書く

`internal/infrastructure/projector/runner_test.go` にテストを追加します。

最低限テストすべきこと:
- `withRetry` は成功した場合にリトライしないか
- `withRetry` は一時的なエラーをリトライするか
- 3回失敗後にエラーを返すか

モックを使ってエラーを意図的に発生させます:
```go
func TestWithRetry(t *testing.T) {
    callCount := 0
    err := withRetry(context.Background(), 3, func() error {
        callCount++
        if callCount < 3 {
            return errors.New("一時的なエラー")
        }
        return nil // 3回目で成功
    })
    // callCount が 3 になっていることを確認
}
```

### 最終ステップ: 動作確認

```bash
go test ./internal/infrastructure/projector/...
go vet ./...
```

## 受け入れ条件

- [ ] 一時的なエラーはリトライされる（最大3回）
- [ ] リトライ間隔が指数バックオフになっている（1s, 2s, 4s）
- [ ] 3回失敗したイベントが `failed_events` テーブルに記録される
- [ ] `go test ./internal/infrastructure/projector/...` が通る

## 詰まったときのヒント

**Q: `1<<uint(i)` の意味が分からない**
→ ビットシフト演算です。`1<<0=1`, `1<<1=2`, `1<<2=4` となります。`math.Pow(2, float64(i))` でも同じ結果ですが、整数の場合はビットシフトが慣用的です。

**Q: `slog` の使い方が分からない**
→ `slog.Warn("メッセージ", "key1", value1, "key2", value2)` の形式でキーバリューペアを渡します。Go 1.21から標準ライブラリに入りました。

**Q: `select` の `ctx.Done()` は何をしているか**
→ コンテキストがキャンセルされたとき（サーバー停止など）にリトライを中断するためです。`time.After` だけだとサーバー停止時も待ち続けてしまいます。

## 参考ファイル

- `internal/infrastructure/projector/runner.go` — 修正対象（現在の実装）
- `internal/infrastructure/projector/requests_projector.go` — Projectorの実装例
- `migrations/` — マイグレーションの書き方の参考

## 設計ポイント（実装後に記録）

> 実装完了後、「なぜ指数バックオフなのか」「デッドレターキューをどう使うか」などを `docs/design/設計ポイント集.md` に追記してください。
