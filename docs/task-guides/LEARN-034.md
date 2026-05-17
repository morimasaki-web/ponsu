# LEARN-034: フロントエンド申請一覧のページネーション

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | 大量の申請データを効率的に表示し、ユーザー体験を向上させる |
| **難易度** | 中級 |
| **ブランチ** | `enhance/034-frontend-pagination` |
| **依存タスク** | LEARN-023（検索機能） |

## この課題で学ぶこと

- **GraphQLのページネーション**: `limit`/`offset` またはカーソルベースの2方式
- **Reactでの「もっと読み込む」パターン**: `fetchMore` で追加データを取得する
- **ローディング状態の管理**: データ取得中のUX

## 事前準備

**読んでおくファイル:**
- `internal/interface/graphql/graph/schema.graphqls` — 現在のrequestsクエリの引数を確認
- `frontend/src/pages/` の申請一覧ページ — 現在の実装を確認
- `frontend/src/gql/` — 現在のGraphQLクエリファイル

## 実装手順

### ステップ1: バックエンドのGraphQLスキーマを確認する

`internal/interface/graphql/graph/schema.graphqls` を開いて、`requests` クエリに `limit`/`offset` の引数があるか確認してください。

もしない場合は LEARN-023（検索機能）で追加されているはずです。

```graphql
# このような形になっているか確認
type Query {
  requests(limit: Int, offset: Int, ...): [Request!]!
}
```

### ステップ2: フロントエンドのGraphQLクエリを更新する

`frontend/src/gql/` にある申請一覧クエリファイルを更新します:

```graphql
query Requests($limit: Int, $offset: Int) {
  requests(limit: $limit, offset: $offset) {
    id
    title
    status
    createdAt
  }
}
```

`npm run codegen` を実行して型を再生成してください。

### ステップ3: ページネーション状態を管理する

申請一覧ページに以下の状態を追加します:

```typescript
const PAGE_SIZE = 20

const [offset, setOffset] = useState(0)
const [allRequests, setAllRequests] = useState<Request[]>([])
const [hasMore, setHasMore] = useState(true)

const [{ data, fetching }] = useQuery({
  query: RequestsDocument,
  variables: { limit: PAGE_SIZE, offset },
})
```

### ステップ4: データを蓄積する

```typescript
useEffect(() => {
  if (data?.requests) {
    if (offset === 0) {
      // 初回 or リセット時は全部入れ替え
      setAllRequests(data.requests)
    } else {
      // 追加読み込み時は末尾に追加
      setAllRequests(prev => [...prev, ...data.requests])
    }
    // 取得件数がPAGE_SIZE未満なら残りなし
    setHasMore(data.requests.length >= PAGE_SIZE)
  }
}, [data, offset])
```

### ステップ5: 「もっと読み込む」ボタンを追加する

```tsx
<div>
  {allRequests.map(req => (
    <RequestListItem key={req.id} request={req} />
  ))}

  {hasMore && (
    <button
      onClick={() => setOffset(prev => prev + PAGE_SIZE)}
      disabled={fetching}
    >
      {fetching ? '読み込み中...' : 'もっと読み込む'}
    </button>
  )}

  {!hasMore && allRequests.length > 0 && (
    <p>全 {allRequests.length} 件を表示しています</p>
  )}
</div>
```

### ステップ6: デモモード対応

デモモードではlocalStorageのデータをページングします:

```typescript
if (appMode === 'demo') {
  const demoData = loadDemoRequests()
  const paged = demoData.slice(offset, offset + PAGE_SIZE)
  // ローカルデータでページングシミュレート
}
```

### 最終ステップ: 動作確認

1. `docker compose up -d` でDB起動
2. テストデータを20件以上作成
3. ページネーションが動作することを確認

## 受け入れ条件

- [ ] 「もっと読み込む」ボタンで追加データを取得できる
- [ ] ローディング状態が表示される
- [ ] 全件表示後にボタンが消える

## 詰まったときのヒント

**Q: `useEffect` の依存配列をどうすればいいか分からない**
→ `[data, offset]` にしてください。`data` か `offset` が変わるたびにEffectが実行されます。

**Q: 重複データが表示される**
→ `offset` が変わっても `data` がキャッシュされた古い値のままのことがあります。`requestPolicy: 'network-only'` をクエリに追加してみてください。

**Q: バックエンドにoffsetパラメータがない**
→ LEARN-023が完了しているか確認してください。もし未実装なら先にそちらを完了させてください。

## 参考ファイル

- `frontend/src/gql/` — 既存のGraphQLクエリの書き方
- `frontend/src/pages/RequestDetail.tsx` — useQueryの使い方の参考

## 設計ポイント（実装後に記録）

> offset/limitとカーソルベースページネーションのトレードオフについて `docs/design/設計ポイント集.md` に追記してください。
