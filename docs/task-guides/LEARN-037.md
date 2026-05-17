# LEARN-037: フロントエンドテスト環境セットアップ（Vitest + Testing Library）

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | フロントエンドのテスト環境を構築し、コンポーネントテストを実行できるようにする |
| **難易度** | 中級 |
| **ブランチ** | `enhance/037-frontend-testing` |
| **依存タスク** | なし |

## この課題で学ぶこと

- **Vitest**: Viteベースの高速テストランナー（JestのAPIと互換）
- **React Testing Library**: コンポーネントをユーザー視点でテストするライブラリ
- **urqlのモッキング**: GraphQLクライアントをテストでモックする方法

## 事前準備

**読んでおくファイル:**
- `frontend/vite.config.ts` — Viteの設定を確認する
- `frontend/package.json` — 現在の依存関係を確認する
- `frontend/src/pages/RequestDetail.tsx` — テスト対象コンポーネントの例

## 実装手順

### ステップ1: 必要なパッケージをインストール

```bash
cd frontend
npm install -D vitest @vitest/ui jsdom
npm install -D @testing-library/react @testing-library/jest-dom
npm install -D @testing-library/user-event
```

### ステップ2: `vitest.config.ts` を作成

`frontend/vitest.config.ts` を新規作成:

```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,           // describe, it, expect をimportなしで使える
    environment: 'jsdom',    // ブラウザ環境をシミュレート
    setupFiles: './src/test/setup.ts',
  },
})
```

### ステップ3: `src/test/setup.ts` を作成

```typescript
import '@testing-library/jest-dom'
// jest-domのmatcher（toBeInTheDocument等）を追加
```

### ステップ4: `package.json` にテストスクリプトを追加

```json
{
  "scripts": {
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest --coverage"
  }
}
```

### ステップ5: シンプルなサンプルテストを作成

まず動作確認として `frontend/src/test/sample.test.tsx` を作成:

```typescript
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

// シンプルなコンポーネントのテスト
function Greeting({ name }: { name: string }) {
  return <p>こんにちは、{name}！</p>
}

describe('Greeting', () => {
  it('名前を表示する', () => {
    render(<Greeting name="山田" />)
    expect(screen.getByText('こんにちは、山田！')).toBeInTheDocument()
  })
})
```

```bash
cd frontend && npm test
```

テストが通ることを確認してください。

### ステップ6: urqlのモッキング

GraphQLクライアント（urql）を使うコンポーネントのテストには、urqlをモックする必要があります。

`frontend/src/test/urql-mock.ts` を作成:

```typescript
import { vi } from 'vitest'

// useQuery をモックする
export function mockUseQuery(result: object) {
  const useQuery = vi.fn(() => [{ data: result, fetching: false, error: undefined }, vi.fn()])
  vi.mock('urql', () => ({ useQuery, useMutation: vi.fn() }))
  return useQuery
}
```

**ポイント:** urqlのモックは複雑になりがちです。まずシンプルなコンポーネントのテストから始めて、徐々にGraphQLコンポーネントのテストに進んでください。

### ステップ7: 既存コンポーネントのテストを1つ追加

`frontend/src/components/` 内のシンプルなコンポーネントのテストを書いてみてください。

例として `ErrorBanner.tsx` があれば:

```typescript
import { render, screen } from '@testing-library/react'
import { ErrorBanner } from '../components/ErrorBanner'

describe('ErrorBanner', () => {
  it('エラーメッセージを表示する', () => {
    render(<ErrorBanner message="何か問題が発生しました" />)
    expect(screen.getByText('何か問題が発生しました')).toBeInTheDocument()
  })
})
```

### 最終ステップ: 動作確認

```bash
cd frontend
npm test          # テストを実行
npm run test:ui   # ブラウザUIでテスト結果を確認（任意）
```

## 受け入れ条件

- [ ] `npm test` でテストが実行できる
- [ ] `@testing-library/react` でコンポーネントをテストできる
- [ ] サンプルテストが通る
- [ ] urqlのクエリ/ミューテーションをモックできる（1つ以上の実例）

## 詰まったときのヒント

**Q: `globals: true` にしているのに `describe is not defined` エラーが出る**
→ `vitest.config.ts` と `vite.config.ts` が別々にあることを確認してください。Vitestは `vitest.config.ts` を優先します。

**Q: jsdomエラーが出る**
→ `npm install -D jsdom` を実行してください。

**Q: `@testing-library/jest-dom` の matcher が使えない**
→ `setupFiles` に `setup.ts` が正しく指定されているか確認してください。

**Q: urqlのモックが難しい**
→ まずurqlを使わないシンプルなコンポーネントのテストから始めてください。環境構築の確認が目的なので、複雑なテストは後回しでOKです。

## 参考ファイル

- `frontend/vite.config.ts` — 既存のVite設定（vitest.config.tsの参考）
- `frontend/src/pages/RequestDetail.tsx` — テスト対象候補のコンポーネント
- [Vitest公式ドキュメント](https://vitest.dev/)
- [Testing Library公式ドキュメント](https://testing-library.com/docs/react-testing-library/intro/)

## 設計ポイント（実装後に記録）

> テスト環境を構築して気づいたこと（どのコンポーネントがテストしやすいか、urqlのモックのコツなど）を `docs/design/設計ポイント集.md` に追記してください。
