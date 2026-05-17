# LEARN-031: 読み取りモデル再構築コマンド（Projectorリプレイ）

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | データ不整合が発生したとき、イベントストアから読み取りモデルを再構築してシステムを自己修復できるようにする |
| **難易度** | 上級 |
| **ブランチ** | `enhance/031-projector-replay` |
| **依存タスク** | LEARN-019 |

## この課題で学ぶこと

- **プロジェクションの再構築**: イベントソーシングの真価 — イベントが全て残っているので「巻き戻し→再生」でいつでも状態を再構築できる
- **大量データのバッチ処理**: ページネーションを使って全イベントを安全に処理する
- **Projectorの状態管理**: どこまで処理したかを `projector_state` テーブルで管理するパターン

## 事前準備

**読んでおくファイル:**
- `internal/infrastructure/projector/runner.go` — CatchUp の仕組みを理解する（**必須**）
- `cmd/projector/main.go` — 現在のProjector起動方法
- `internal/infrastructure/postgres/event_store.go` — イベントストアのクエリ方法

**理解しておく概念:**
- **イベントソーシング**: 状態ではなくイベントの履歴を保存するアーキテクチャ。「今の状態」はイベントを最初から再生することで復元できる
- **プロジェクション**: イベントを元に読み取りモデル（テーブル）を構築するプロセス

## 実装手順

### ステップ1: 現在のCatchUp処理を理解する

`internal/infrastructure/projector/runner.go` の `CatchUp` 関数（または類似の関数）を読み、どのようにイベントを処理しているかを把握してください。

特に確認すべき点:
- `projector_state` テーブルはすでにあるか？（あれば何を記録しているか）
- どのSQLで「未処理のイベント」を取得しているか

### ステップ2: リプレイコマンドの設計

リプレイの処理フロー:
1. 読み取りモデルのデータを削除（TRUNCATE または DELETE）
2. `projector_state` をリセット（処理済みの位置を0に戻す）
3. 通常のCatchUpを実行（全イベントを最初から再処理）

### ステップ3: `cmd/replay/main.go` を作成

新しいコマンドとして `cmd/replay/main.go` を作成します:

```go
package main

import (
    "context"
    "log/slog"
    "os"

    // 必要なパッケージをimport
)

func main() {
    ctx := context.Background()

    // 1. DB接続を確立
    // （cmd/projector/main.go を参考にする）

    // 2. 対象のProjectorを選択
    // コマンドライン引数でProjector名を指定できると良い
    // 例: go run ./cmd/replay requests

    // 3. 読み取りモデルをリセット
    slog.Info("読み取りモデルをリセットします...")
    // テーブルのTRUNCATEなど

    // 4. projector_stateをリセット
    slog.Info("Projectorの状態をリセットします...")

    // 5. 通常のCatchUpを実行
    slog.Info("イベントの再処理を開始します...")
    // runner.CatchUp(ctx, ...) を呼ぶ

    slog.Info("リプレイ完了")
    os.Exit(0)
}
```

**ポイント:** `cmd/projector/main.go` をコピーして修正するのが最も早いです。

### ステップ4: 読み取りモデルのリセット処理

どのテーブルをリセットするかは再構築対象のProjectorによります。

sqlcクエリ `db/query/projector_management.sql` に以下を追加:
```sql
-- name: ResetProjectorState :exec
DELETE FROM projector_state WHERE projector_name = $1;

-- name: TruncateRequests :exec
TRUNCATE TABLE requests RESTART IDENTITY CASCADE;
-- （Projectrorに応じて適切なテーブルを選択）
```

**注意:** TRUNCATE は全データを消します。本番環境では使用前に確認が必要です。

### ステップ5: 動作確認（実際にデータを壊して復旧する）

これが一番の学びです:

```bash
# 1. 通常の状態でデータを確認
go run ./cmd/ponsu &
# GraphQLでデータを作成・確認

# 2. 意図的にrequestsテーブルのデータを破壊
# (直接SQL DELETEで一部削除など)

# 3. リプレイコマンドで復旧
go run ./cmd/replay requests

# 4. データが復元されたことを確認
```

### 最終ステップ: 動作確認

```bash
go build ./cmd/replay/...
go vet ./...
```

## 受け入れ条件

- [ ] `go run ./cmd/replay [projector名]` でリプレイが実行できる
- [ ] 読み取りモデルが削除され、イベントストアから再構築される
- [ ] 大量データでも動作する（ページング処理がある）
- [ ] 進捗がログに出力される

## 詰まったときのヒント

**Q: DBへの接続方法が分からない**
→ `cmd/projector/main.go` を開いてください。まったく同じ方法でDB接続できます。

**Q: `projector_state` テーブルがない**
→ runner.go を読んで、どのようにProjectorの進捗を管理しているかを確認してください。もしなければ、シンプルにテーブルをDELETEするだけでも動作します。

**Q: どのテーブルをリセットすればいいか分からない**
→ `requests_projector.go` が何のテーブルを更新しているかを確認してください。そのテーブルをリセットします。

## 参考ファイル

- `cmd/projector/main.go` — 起動パターンの参考（ほぼコピーでOK）
- `internal/infrastructure/projector/runner.go` — CatchUpの実装
- `internal/infrastructure/projector/requests_projector.go` — どのテーブルを管理しているかの確認

## 設計ポイント（実装後に記録）

> 実装後、「なぜイベントソーシングではリプレイが簡単なのか」「プロダクション環境でリプレイするときの注意点」を `docs/design/設計ポイント集.md` に追記してください。
