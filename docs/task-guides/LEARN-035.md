# LEARN-035: フロントエンドのエラー境界とトースト通知の統一

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | GraphQLエラーやネットワークエラーを統一的に処理し、ユーザーに分かりやすく表示する |
| **難易度** | 中級 |
| **ブランチ** | `enhance/035-error-toast` |
| **依存タスク** | なし |

## この課題で学ぶこと

- **React Error Boundary**: 予期しないエラーをキャッチしてクラッシュを防ぐReactの仕組み
- **トースト通知**: 画面を遷移させずに通知を表示するUI パターン
- **グローバルエラーハンドリング**: アプリ全体で一貫したエラー処理を実現する

## 事前準備

**読んでおくファイル:**
- `frontend/src/App.tsx` — アプリのルートコンポーネント（ここに追加する）
- `frontend/src/components/ErrorBanner.tsx`（あれば） — 既存のエラー表示を確認
- `frontend/src/pages/` 内のページ — 現在のエラーハンドリングの状況

## 実装手順

### ステップ1: `react-hot-toast` をインストール

```bash
cd frontend
npm install react-hot-toast
```

### ステップ2: `ErrorBoundary.tsx` を実装する

Error BoundaryはクラスコンポーネントでないといけないReactの制約があります（Hooks不可）:

```tsx
// frontend/src/components/ErrorBoundary.tsx
import React from 'react'

interface Props {
  children: React.ReactNode
  fallback?: React.ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info)
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? (
        <div>
          <h2>予期しないエラーが発生しました</h2>
          <p>{this.state.error?.message}</p>
          <button onClick={() => this.setState({ hasError: false })}>
            再試行
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
```

**ポイント:** Error Boundaryはクラスコンポーネントである必要があります（React 18時点）。`getDerivedStateFromError` でエラーを検知し、`render()` でフォールバックUIを表示します。

### ステップ3: GraphQLエラーをトーストに変換するヘルパーを作成

```typescript
// frontend/src/lib/error.ts
import toast from 'react-hot-toast'

export function handleGraphQLError(error: Error | undefined) {
  if (!error) return
  
  // GraphQLエラーから適切なメッセージを生成
  const message = error.message || '操作に失敗しました'
  toast.error(message, {
    duration: 4000,
    position: 'top-right',
  })
}
```

### ステップ4: `App.tsx` に `Toaster` と `ErrorBoundary` を追加

```tsx
// frontend/src/App.tsx
import { Toaster } from 'react-hot-toast'
import { ErrorBoundary } from './components/ErrorBoundary'

function App() {
  return (
    <ErrorBoundary>
      {/* 既存のルーターやプロバイダー */}
      <Router>
        {/* ... */}
      </Router>
      <Toaster /> {/* ← トースト表示用コンポーネント */}
    </ErrorBoundary>
  )
}
```

### ステップ5: 各画面でエラーハンドリングを統一する

既存のGraphQLエラー処理を `handleGraphQLError` に置き換えます:

```tsx
// 変更前
const [{ data, error }] = useQuery({ query: RequestsDocument })
if (error) {
  return <div>エラーが発生しました: {error.message}</div>
}

// 変更後
const [{ data, error }] = useQuery({ query: RequestsDocument })
useEffect(() => {
  handleGraphQLError(error)
}, [error])
```

### 最終ステップ: 動作確認

1. ネットワークを切断した状態でGraphQLクエリを実行
2. トースト通知が表示されることを確認
3. JavaScriptエラーを意図的に発生させてError Boundaryがキャッチするか確認

## 受け入れ条件

- [ ] GraphQLエラー時にトースト通知が表示される
- [ ] 致命的なエラーはError Boundaryがキャッチしてフォールバックを表示する
- [ ] エラーメッセージが日本語で分かりやすい

## 詰まったときのヒント

**Q: Error BoundaryがHooksで書けないと言われた**
→ Error Boundaryは現時点（React 18）でクラスコンポーネントである必要があります。React 19からHooksで書けるようになる予定です。

**Q: `toast.error` がどのモジュールから来るか分からない**
→ `import toast from 'react-hot-toast'` です。`react-hot-toast` はdefault exportがトースト関数です。

**Q: `<Toaster>` をどこに置けばいいか**
→ アプリのルートに1つだけ置きます（`App.tsx` の中が一般的）。

## 参考ファイル

- [react-hot-toast公式ドキュメント](https://react-hot-toast.com/)
- `frontend/src/App.tsx` — ルートコンポーネント（追加先）
- [React Error Boundary公式ドキュメント](https://react.dev/reference/react/Component#catching-rendering-errors-with-an-error-boundary)

## 設計ポイント（実装後に記録）

> エラーの種類（ネットワークエラー・GraphQLエラー・JSエラー）をどう区別して処理したかを `docs/design/設計ポイント集.md` に追記してください。
