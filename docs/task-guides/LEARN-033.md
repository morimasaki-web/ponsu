# LEARN-033: デモ用モックデータ生成とローカルストレージ永続化

## 概要

| 項目 | 内容 |
|------|------|
| **目的** | デモモード用のモックデータを生成し、ローカルストレージに永続化して再現性のあるデモ環境を構築する |
| **難易度** | 中級 |
| **ブランチ** | `enhance/033-demo-mock-data` |
| **依存タスク** | なし（フロントエンドのみ） |

## この課題で学ぶこと

- **localStorage/IndexedDB**: ブラウザのクライアントサイドストレージ
- **モックデータの設計**: リアルな見た目のテストデータをどう作るか
- **デモモードとプロダクションモードの切り替え**: 環境変数や状態で分岐する設計

## 事前準備

**読んでおくファイル:**
- `frontend/src/pages/RequestDetail.tsx` — 既存のデモモード実装パターン（`appMode === 'demo'` の分岐）を確認
- `frontend/src/` 内の他のページファイル — デモデータがどこで定義されているかを確認

## 実装手順

### ステップ1: 現在のデモモード実装を理解する

`frontend/src/pages/RequestDetail.tsx` を開いて `appMode === 'demo'` の分岐を探してください。

既存のデモデータ（`demoRequestsSeed` 等）がどこで定義されているかも確認します。

### ステップ2: `frontend/src/demo/` ディレクトリ構造を設計する

```
frontend/src/demo/
├── data.ts       # モックデータの定義
├── storage.ts    # localStorageの読み書き
└── index.ts      # エクスポート
```

### ステップ3: モックデータを定義する（`data.ts`）

```typescript
// frontend/src/demo/data.ts

export interface DemoRequest {
  id: string
  title: string
  description: string
  status: 'pending' | 'approved' | 'rejected' | 'returned'
  submitterName: string
  createdAt: string
  updatedAt: string
}

export const initialDemoRequests: DemoRequest[] = [
  {
    id: 'demo-001',
    title: '備品購入申請（ノートPC）',
    description: '開発用ノートPC MacBook Pro 14インチを購入したい。',
    status: 'pending',
    submitterName: '山田 太郎',
    createdAt: '2026-03-01T09:00:00Z',
    updatedAt: '2026-03-01T09:00:00Z',
  },
  {
    id: 'demo-002',
    title: '出張申請（東京出張）',
    description: '取引先訪問のため東京出張を申請します。',
    status: 'approved',
    submitterName: '鈴木 花子',
    createdAt: '2026-02-28T10:00:00Z',
    updatedAt: '2026-02-28T15:30:00Z',
  },
  {
    id: 'demo-003',
    title: '契約書承認（ソフトウェアライセンス）',
    description: 'Adobe Creative Cloudのチームプランを契約したい。',
    status: 'rejected',
    submitterName: '田中 次郎',
    createdAt: '2026-02-25T14:00:00Z',
    updatedAt: '2026-02-26T10:00:00Z',
  },
]
```

### ステップ4: localStorageの読み書き（`storage.ts`）

```typescript
// frontend/src/demo/storage.ts

const STORAGE_KEY = 'ponsu_demo_requests'

export function loadDemoRequests(): DemoRequest[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return initialDemoRequests
    return JSON.parse(raw) as DemoRequest[]
  } catch {
    return initialDemoRequests
  }
}

export function saveDemoRequests(requests: DemoRequest[]): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(requests))
}

export function resetDemoRequests(): void {
  localStorage.removeItem(STORAGE_KEY)
}
```

**ポイント:** `try/catch` が重要です。localStorageが無効な環境（プライベートモード等）でも動作するようにします。

### ステップ5: デモモード判定ロジックを確認・整理する

`frontend/src/` でどこでデモモードを判定しているかを確認して、`demo/index.ts` からエクスポートを整理します。

### ステップ6: リセット機能をUIに追加する

デモ用のリセットボタンを追加します（例: ヘッダーやフッターに配置）:

```tsx
import { resetDemoRequests } from '../demo'

function DemoResetButton() {
  const handleReset = () => {
    resetDemoRequests()
    window.location.reload()
  }
  return (
    <button onClick={handleReset} style={{ /* スタイル */ }}>
      デモをリセット
    </button>
  )
}
```

### 最終ステップ: 動作確認

1. デモモードで起動する
2. 申請の状態を変更する
3. ページをリロードしても状態が維持されることを確認
4. リセットボタンで初期状態に戻ることを確認

## 受け入れ条件

- [ ] デモモードで申請一覧が表示される
- [ ] 状態変更がlocalStorageに保存される（リロードしても維持される）
- [ ] リセットボタンで初期状態に戻せる

## 詰まったときのヒント

**Q: `localStorage.getItem` が null を返す**
→ 初回アクセス時は `null` が返ります。`initialDemoRequests` をデフォルト値として返すのが正しいです。

**Q: JSON.parse でエラーが出る**
→ `try/catch` でラップして、エラー時は初期データを返すようにしてください。

**Q: デモモードかどうかの判定がどこにあるか分からない**
→ `grep -r "appMode\|demo" frontend/src/` で探してください。

## 参考ファイル

- `frontend/src/pages/RequestDetail.tsx` — 既存のデモモード実装
- MDN Web Docs: [Window.localStorage](https://developer.mozilla.org/ja/docs/Web/API/Window/localStorage)

## 設計ポイント（実装後に記録）

> デモモードの設計（localStorageの限界・IndexedDBとの違いなど）について `docs/design/設計ポイント集.md` に追記してください。
