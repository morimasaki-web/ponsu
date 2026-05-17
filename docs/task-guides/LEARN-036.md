# LEARN-036: ローディング状態の統一とスケルトンUI

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | データ読み込み中のUXを改善し、スケルトンスクリーンで待ち時間を自然に見せる |
| **難易度** | 初級〜中級 |
| **ブランチ** | `enhance/036-loading-skeleton` |
| **依存タスク** | なし |

## この課題で学ぶこと

- **スケルトンスクリーン**: コンテンツの形を模したローディングプレースホルダー
- **CSSアニメーション**: `@keyframes` でシマーアニメーションを実装する
- **コンポーネントの再利用**: 汎用スケルトンをページごとに組み合わせる

## 事前準備

**読んでおくファイル:**
- `frontend/src/pages/` の各ページ — `fetching` 状態でどんな表示をしているか確認
- `frontend/src/` の CSS/スタイル設定 — CSSの書き方（モジュール/Tailwind等）を確認

## 実装手順

### ステップ1: 汎用スケルトンコンポーネントを作成する

```tsx
// frontend/src/components/Skeleton.tsx

interface SkeletonProps {
  width?: string
  height?: string
  borderRadius?: string
  className?: string
}

export function Skeleton({
  width = '100%',
  height = '1rem',
  borderRadius = '4px',
  className,
}: SkeletonProps) {
  return (
    <div
      className={className}
      style={{
        width,
        height,
        borderRadius,
        background: 'linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%)',
        backgroundSize: '200% 100%',
        animation: 'shimmer 1.5s infinite',
      }}
    />
  )
}
```

CSSアニメーションを追加（グローバルCSSかCSSモジュール）:
```css
@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}
```

**ポイント:** `background-position` をアニメーションさせることで光が流れるような効果を実現します。

### ステップ2: 申請一覧用スケルトンを作成する

```tsx
// frontend/src/components/RequestListSkeleton.tsx
import { Skeleton } from './Skeleton'

export function RequestListSkeleton() {
  return (
    <div>
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} style={{ padding: '16px', borderBottom: '1px solid #eee' }}>
          <Skeleton height="1.2rem" width="60%" />
          <div style={{ marginTop: '8px' }}>
            <Skeleton height="0.9rem" width="30%" />
          </div>
        </div>
      ))}
    </div>
  )
}
```

### ステップ3: 各ページにスケルトンを組み込む

申請一覧ページで `fetching` 中にスケルトンを表示:

```tsx
// 変更前
if (fetching) return <p>読み込み中...</p>

// 変更後
if (fetching) return <RequestListSkeleton />
```

申請詳細ページ用のスケルトンも同様に作成します:

```tsx
// frontend/src/components/RequestDetailSkeleton.tsx
export function RequestDetailSkeleton() {
  return (
    <div style={{ padding: '24px' }}>
      <Skeleton height="2rem" width="50%" />  {/* タイトル */}
      <div style={{ marginTop: '16px' }}>
        <Skeleton height="1rem" width="100%" />
        <Skeleton height="1rem" width="90%" style={{ marginTop: '8px' }} />
        <Skeleton height="1rem" width="80%" style={{ marginTop: '8px' }} />
      </div>
    </div>
  )
}
```

### ステップ4: デモモードでもスケルトンを表示する（任意）

デモモードでは実際の遅延がないため、スケルトンが表示されません。学習のためにモックの遅延を追加してもよいです:

```typescript
// デモモードで人工的な遅延を追加（開発・確認用）
await new Promise(resolve => setTimeout(resolve, 800))
```

### 最終ステップ: 動作確認

1. ブラウザのDevToolsで「Slow 3G」にネットワーク速度を変更
2. ページをリロードしてスケルトンが表示されることを確認
3. アニメーションが滑らかに動くことを確認

## 受け入れ条件

- [ ] ローディング中にスケルトンが表示される（「読み込み中...」テキストの代わり）
- [ ] シマーアニメーションが動作する
- [ ] 申請一覧・申請詳細の2画面以上で統一されている

## 詰まったときのヒント

**Q: アニメーションが動かない**
→ `@keyframes shimmer` がグローバルCSSに定義されているか確認してください。CSSモジュールを使っている場合は `:global(@keyframes shimmer)` とする必要があります。

**Q: スケルトンのサイズが合わない**
→ 実際のコンテンツのサイズを DevTools で確認して、`width`/`height` を合わせてください。完全に一致しなくてもOKです。

**Q: `Array.from({ length: 5 })` の意味が分からない**
→ 長さ5の配列を作る方法です。`[...Array(5)]` と同じです。`.map((_, i) => ...)` でインデックスをkeyに使います。

## 参考ファイル

- `frontend/src/pages/` — 現在の `fetching` 状態の処理（置き換え対象）
- [スケルトンスクリーンのUIパターン解説](https://uxdesign.cc/what-you-should-know-about-skeleton-screens-a820c45a571a)

## 設計ポイント（実装後に記録）

> スケルトンUIがユーザー体験にどう影響するか、ローディングスピナーとの使い分けについて `docs/design/設計ポイント集.md` に追記してください。
