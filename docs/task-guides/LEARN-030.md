# LEARN-030: GraphQLエラーレスポンスの統一と構造化

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | GraphQLエラーを一貫した形式で返し、フロントエンドで適切なエラーハンドリングを可能にする |
| **難易度** | 中級 |
| **ブランチ** | `enhance/030-graphql-error-handling` |
| **依存タスク** | なし |

## この課題で学ぶこと

- **gqlgenのエラーハンドリング機構**: `graphql.ErrorPresenter` でエラーをカスタマイズする方法
- **エラーコードとメッセージの設計**: クライアントが扱いやすいエラー構造
- **エクステンションフィールドの活用**: GraphQLエラーにコードや詳細を付加する方法

## 事前準備

**読んでおくファイル:**
- `internal/interface/graphql/graph/schema.resolvers.go` — リゾルバでどんなエラーを返しているかを確認
- `internal/usecase/requests/service.go` — ドメインエラー（ErrNotFound等）の定義を確認
- `gqlgen.yml` — gqlgenの設定ファイル

**理解しておく概念:**
- **GraphQLエラー**: GraphQLのレスポンスは `{ "data": ..., "errors": [...] }` の形式。エラーは配列で返る
- **エクステンション**: GraphQLエラーに `"extensions": {"code": "NOT_FOUND"}` のような追加情報を付加できる

## 実装手順

### ステップ1: 現在のエラー返却を把握する

`internal/interface/graphql/graph/schema.resolvers.go` を開き、いくつかのリゾルバで `return nil, err` の部分を探してください。

現在は `err` をそのまま返しているはずです。これを構造化されたGraphQLエラーに変換します。

### ステップ2: エラー種別を定義する

`internal/interface/graphql/errors.go` を新規作成:

```go
package graphql

import (
    "context"
    "errors"

    "github.com/99designs/gqlgen/graphql"
    "github.com/vektah/gqlparser/v2/gqlerror"

    requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
)

// エラーコード定数
const (
    ErrCodeNotFound     = "NOT_FOUND"
    ErrCodeUnauthorized = "UNAUTHORIZED"
    ErrCodeForbidden    = "FORBIDDEN"
    ErrCodeValidation   = "VALIDATION_ERROR"
    ErrCodeInternal     = "INTERNAL_ERROR"
)

// NewErrorPresenter はgqlgenに渡すエラープレゼンター
// ドメインエラー → GraphQLエラー（コード付き）に変換する
func NewErrorPresenter() graphql.ErrorPresenterFunc {
    return func(ctx context.Context, err error) *gqlerror.Error {
        // gqlerrorに既に変換済みならそのまま返す
        var gqlErr *gqlerror.Error
        if errors.As(err, &gqlErr) {
            return gqlErr
        }

        // ドメインエラーをコードに変換
        code := ErrCodeInternal
        message := "内部エラーが発生しました"

        switch {
        case errors.Is(err, requestsuc.ErrNotFound):
            code = ErrCodeNotFound
            message = "リソースが見つかりません"
        case errors.Is(err, requestsuc.ErrForbidden):
            code = ErrCodeForbidden
            message = "この操作を行う権限がありません"
        }

        return &gqlerror.Error{
            Message: message,
            Extensions: map[string]interface{}{
                "code": code,
            },
        }
    }
}
```

**ポイント:** `errors.Is` は Go のエラーラッピングに対応しています。`fmt.Errorf("...: %w", err)` でラップされたエラーも正しく判定できます。

### ステップ3: エラープレゼンターをgqlgenに登録する

`internal/interface/graphql/` の中でgqlgenのサーバーを設定しているファイル（`server.go` や `handler.go` 等）を探して、エラープレゼンターを登録します:

```go
// gqlgenのサーバー設定にエラープレゼンターを追加
srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))
srv.SetErrorPresenter(NewErrorPresenter())
```

設定ファイルの場所が分からない場合は `grep -r "handler.NewDefaultServer" internal/` で探してください。

### ステップ4: テストを書く

`internal/interface/graphql/errors_test.go` を作成:

```go
func TestErrorPresenter(t *testing.T) {
    presenter := NewErrorPresenter()

    t.Run("ErrNotFoundをNOT_FOUNDに変換する", func(t *testing.T) {
        gqlErr := presenter(context.Background(), requestsuc.ErrNotFound)
        
        if gqlErr.Extensions["code"] != ErrCodeNotFound {
            t.Errorf("expected %s, got %v", ErrCodeNotFound, gqlErr.Extensions["code"])
        }
    })
}
```

### 最終ステップ: 動作確認

```bash
go test ./internal/interface/graphql/...
go build ./...
```

## 受け入れ条件

- [ ] ドメインエラーに適切なコードが付与される（NOT_FOUND, FORBIDDEN等）
- [ ] エラーメッセージがユーザー向けの日本語になっている
- [ ] `go test ./internal/interface/graphql/...` が通る

## 詰まったときのヒント

**Q: gqlgenのエラープレゼンターをどこに登録するかが分からない**
→ `grep -r "handler.NewDefaultServer\|handler.New(" internal/` で設定箇所を探してください。

**Q: `ErrNotFound` がどこで定義されているか分からない**
→ `grep -r "ErrNotFound\|ErrForbidden" internal/usecase/` で確認してください。

**Q: `gqlerror.Error` の構造が分からない**
→ `go doc github.com/vektah/gqlparser/v2/gqlerror Error` で確認するか、既存のimportから辿ってください。

## 参考ファイル

- `internal/interface/graphql/graph/schema.resolvers.go` — エラーを返しているリゾルバ
- `internal/usecase/requests/service.go` — ドメインエラーの定義

## 設計ポイント（実装後に記録）

> GraphQLのエラー設計（クライアントが扱いやすいエラー構造とは何か）について `docs/design/設計ポイント集.md` に追記してください。
