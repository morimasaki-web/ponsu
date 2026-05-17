# LEARN-032: トランザクション境界の明確化と例外時ロールバック

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | エラー発生時の部分更新を防止し、データ整合性を保証する |
| **難易度** | 中級 |
| **ブランチ** | `enhance/032-transaction-boundary` |
| **依存タスク** | LEARN-018 |

## この課題で学ぶこと

- **「1トランザクション = 1集約」の原則**: 複数の集約をまたぐトランザクションは避ける
- **`defer tx.Rollback()`**: Goでのトランザクション管理のイディオム
- **結果整合性とイベント駆動**: トランザクション外の整合性をどう保つか

## 事前準備

**読んでおくファイル:**
- `internal/usecase/requests/service.go` — トランザクション管理の現状を完全に理解する（**必須**）
- `internal/infrastructure/postgres/event_store.go` — イベント追記のトランザクション処理
- Go公式ドキュメント: `database/sql` パッケージの `BeginTx`, `Commit`, `Rollback`

**理解しておく概念:**
- **トランザクション**: 複数のDB操作を「全部成功 or 全部失敗」にまとめる仕組み
- **`defer`**: 関数終了時に必ず実行される。`defer tx.Rollback()` は「コミット前にエラーがあれば必ずロールバック」を保証する

## 実装手順

### ステップ1: 現在のトランザクション管理を読む

`internal/usecase/requests/service.go` の `applyTransition` 関数（または承認処理の関数）を開いて、以下を確認してください:

1. `BeginTx` でトランザクションを開始しているか
2. `defer tx.Rollback()` が書かれているか
3. 最後に `tx.Commit()` が呼ばれているか
4. エラー時に適切にロールバックされるか

```go
// 良いパターン（これになっているか確認する）
tx, err := s.DB.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer func() { _ = tx.Rollback() }() // ← これが重要

// ... 処理 ...

return tx.Commit() // ← ここに到達しなければRollbackが呼ばれる
```

### ステップ2: トランザクション境界の問題を探す

以下のようなアンチパターンがないかチェックします:

```go
// ❌ 問題パターン1: tx.Rollback() を defer していない
tx, _ := s.DB.BeginTx(ctx, nil)
if err := doSomething(tx); err != nil {
    tx.Rollback() // ← 手動で呼ぶ必要があるが漏れやすい
    return err
}
tx.Commit()

// ❌ 問題パターン2: エラーを無視してcommitしている
result, _ := doSomething(tx) // エラー無視！
tx.Commit() // ← 問題があってもcommitされてしまう
```

### ステップ3: 問題があれば修正する

問題パターンが見つかれば修正してください。

**Goのトランザクション慣用句:**
```go
tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
if err != nil {
    return fmt.Errorf("トランザクション開始失敗: %w", err)
}
defer func() { _ = tx.Rollback() }()
// ↑ Commit()が成功した後にRollback()が呼ばれても、
//   既にcommit済みなのでエラーにならない（PostgreSQLの場合）

// ... 処理 ...

if err := tx.Commit(); err != nil {
    return fmt.Errorf("コミット失敗: %w", err)
}
return nil
```

### ステップ4: トランザクション境界のテストを書く

`internal/usecase/requests/service_test.go`（または新規作成）でトランザクション境界を検証します:

```go
func TestApproveRequest_Rollback(t *testing.T) {
    // テスト内容:
    // 1. 申請を作成
    // 2. 承認処理の途中でエラーが発生するシナリオを作る
    // 3. 申請のステータスが変更されていないことを確認（ロールバックされた）
}
```

**ポイント:** 統合テストになるのでDBが必要です。`docker compose up -d` を事前に実行してください。

### ステップ5: ドキュメントに境界を明記する

`docs/design/設計ポイント集.md` にトランザクション境界について追記してください:

- どの関数がトランザクション境界か
- 「1集約 = 1トランザクション」の原則に従っているか
- プロジェクション（読み取りモデル更新）はトランザクション内か外か

### 最終ステップ: 動作確認

```bash
go test ./internal/usecase/requests/...
go vet ./...
```

## 受け入れ条件

- [ ] トランザクション境界が明確になっている
- [ ] `defer tx.Rollback()` パターンが使われている
- [ ] エラー時にロールバックされることがテストで確認できる
- [ ] トランザクション境界のドキュメントが追加されている

## 詰まったときのヒント

**Q: `defer tx.Rollback()` の後にCommitしても問題ないか**
→ はい、問題ありません。`tx.Commit()` が成功した後に `tx.Rollback()` を呼んでも、PostgreSQLは既にコミット済みのトランザクションのロールバックを無視します（エラーを返しますが、`_ =` で無視して問題なし）。

**Q: どこまでが「1つの集約」か**
→ PonSuでは申請（Request）が1集約です。`requests` テーブル + `event_store`（aggregate_type='request'）への操作が1トランザクションに収まるべきです。

**Q: テストでDBが使えるか分からない**
→ `docker compose up -d` でDBを起動して、`go test -tags=integration ./...` のようにタグを使うか、テスト用DBの接続情報を環境変数で渡します。既存テストの書き方を参考にしてください。

## 参考ファイル

- `internal/usecase/requests/service.go` — メインの修正対象
- `internal/infrastructure/postgres/event_store.go` — イベント追記のトランザクション例

## 設計ポイント（実装後に記録）

> 「なぜ1集約=1トランザクションなのか」「プロジェクションを同期・非同期どちらで行うかのトレードオフ」について `docs/design/設計ポイント集.md` に追記してください。
