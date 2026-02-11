# GraphQL機能追加ガイド - PonSuアーキテクチャ解説

## はじめに

PonSuでは **Event Sourcing（イベントソーシング）** と **GraphQL** を組み合わせたアーキテクチャを採用しています。このドキュメントでは、LEARN-022/022F（コメント機能追加）を例に、新しいクエリ機能を追加する全体的な流れと、各コンポーネントの役割を解説します。

## なぜこのアーキテクチャなのか？

### RESTとの比較

| 観点 | REST | GraphQL + Event Sourcing |
|------|------|--------------------------|
| データ取得 | 複数エンドポイント、Over-fetching/Under-fetching | 1エンドポイント、必要なデータのみ |
| 監査証跡 | 別途実装が必要 | イベントストアに全履歴が残る |
| データ整合性 | アプリケーションロジックで管理 | イベント駆動で自動的に保証 |
| 学習コスト | 低い | **高い**（トレードオフ） |
| 初期実装コスト | 低い | **高い**（今回実感した点） |
| 長期保守性 | 中 | 高（履歴追跡、データ再構築が可能） |

### PonSuがこのアーキテクチャを選んだ理由

1. **監査要件**: 申請・承認フローの全履歴を記録する必要がある
2. **データ整合性**: 複数の読み取りモデルを同じイベントストアから自動生成できる
3. **柔軟性**: フロントエンドが必要なデータを自由に取得できる
4. **学習機会**: モダンなアーキテクチャパターンを実践できる

**確かに初期コストは高いですが**、監査証跡や履歴管理が必要なドメインでは、長期的にメリットが大きくなります。

---

## アーキテクチャ全体像

```
【書き込みフロー】
Frontend (GraphQL Mutation)
    ↓
GraphQL Resolver (リゾルバ)
    ↓
UseCase Layer (Service)
    ↓
Domain Layer (Aggregate)
    ↓
Event Store (events テーブル) ← すべての変更をイベントとして記録
    ↓
Projector (プロジェクター) ← イベントを購読
    ↓
Read Model (request_comments テーブル) ← クエリ用の最適化されたビュー

【読み取りフロー】
Frontend (GraphQL Query)
    ↓
GraphQL Resolver
    ↓
Read Model から直接取得 ← 高速
```

### 重要な概念

#### 1. イベントソーシング

**通常のCRUD**: データベースの状態を直接更新する
```sql
-- 従来の方法
UPDATE comments SET content = '更新後' WHERE id = 123;
-- 履歴は失われる（更新前の内容は消える）
```

**イベントソーシング**: 変更を「イベント」として記録する
```sql
-- イベントソーシング
INSERT INTO event_store (event_type, payload) VALUES 
('RequestCommented', '{"comment_id": "...", "content": "新しいコメント"}');
-- 全履歴が残る
```

**メリット**:
- 完全な監査証跡（いつ、誰が、何を変更したか）
- データの再構築が可能（イベントを再生すれば過去の状態を復元できる）
- タイムトラベルデバッグ（特定時点の状態を再現できる）

**デメリット**:
- 実装コストが高い ← **今回感じた点**
- クエリが複雑になる（イベントから状態を復元する必要がある）

#### 2. Projector（プロジェクター）とは？

**役割**: イベントストアから読み取り用のビュー（Read Model）を自動生成する

```go
// internal/infrastructure/projector/request_comments_projector.go
func (r *RequestCommentsProjector) HandleEvent(ctx context.Context, event Event) error {
    if event.EventType == "RequestCommented" {
        // イベントペイロードを解析
        var payload request.RequestCommentedPayload
        json.Unmarshal(event.Payload, &payload)
        
        // Read Model (request_comments テーブル) に挿入
        // ← これがプロジェクションの本質！
        queries.InsertComment(ctx, InsertCommentParams{
            ID:        payload.CommentID,
            Content:   payload.Content,
            // ...
        })
    }
    return nil
}
```

**プロジェクターの特徴**:
- **イベント駆動**: 新しいイベントが発生すると自動的に実行される
- **冪等性**: 同じイベントを複数回処理しても結果が同じ
- **非同期**: イベント発行とプロジェクション更新は分離できる（PonSuでは同期実装）

**なぜ必要？**
- イベントストアは履歴記録に最適化されているが、クエリには不向き
- プロジェクターで「クエリに最適化されたテーブル」を自動生成する
- 同じイベントストアから複数のビューを作れる（例: 統計用、検索用、表示用）

#### 3. GraphQL Resolver（リゾルバ）とは？

**役割**: GraphQLクエリ/ミューテーションと実際のデータ取得/更新ロジックを繋ぐ

```go
// internal/interface/graphql/graph/schema.resolvers.go

// Query リゾルバ - データ取得
func (r *queryResolver) Comments(ctx context.Context, requestID string) ([]*model.Comment, error) {
    // 1. Read Model から取得
    rows, err := queries.ListCommentsByRequest(ctx, requestID)
    
    // 2. GraphQL型に変換
    var comments []*model.Comment
    for _, row := range rows {
        comments = append(comments, &model.Comment{
            ID:      row.ID,
            Content: row.Content,
            // ...
        })
    }
    return comments, nil
}

// Mutation リゾルバ - データ更新
func (r *mutationResolver) AddComment(ctx context.Context, requestID string, content string) (*model.Comment, error) {
    // 1. イベントを発行（Service層を呼ぶ）
    commentID, err := service.AddComment(ctx, requestID, content)
    // ← この中でイベントストアに追加 + プロジェクター実行
    
    // 2. 新規作成されたコメントを取得
    comment, err := queries.GetComment(ctx, commentID)
    
    // 3. GraphQL型に変換して返す
    return &model.Comment{...}, nil
}
```

**リゾルバの責務**:
- GraphQLスキーマとバックエンドロジックの橋渡し
- 認証・認可のチェック（viewer から user_id 取得など）
- データ型の変換（DB型 → GraphQL型）
- エラーハンドリング（GraphQLエラー形式に変換）

**なぜ分離されている？**
- GraphQLスキーマ（API仕様）とビジネスロジックを分離
- 同じビジネスロジックを複数のAPIから再利用できる
- テストしやすい（リゾルバとロジックを別々にテストできる）

---

## 機能追加の全体フロー

LEARN-022/022F（コメント機能）を例に、実際の手順を詳しく説明します。

### Phase 1: バックエンド - イベント定義（LEARN-022）

#### ステップ1: イベント型を定義

**ファイル**: `internal/domain/request/events.go`

```go
const EventTypeRequestCommented = "RequestCommented"

type RequestCommentedPayload struct {
    CommentID   string    `json:"comment_id"`
    UserID      string    `json:"user_id"`
    Content     string    `json:"content"`
    CommentedAt time.Time `json:"commented_at"`
}

func (p RequestCommentedPayload) Validate() error {
    if p.CommentID == "" {
        return errors.New("comment_id is required")
    }
    // ... バリデーション
    return nil
}
```

**ポイント**:
- イベント名は**過去形**（何が起きたか）
- ペイロードは**不変**（後から変更できない）
- バリデーションを実装（不正なイベントを防ぐ）

#### ステップ2: マイグレーション作成（Read Model用テーブル）

**ファイル**: `migrations/0011_request_comments.up.sql`

```sql
CREATE TABLE IF NOT EXISTS request_comments (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    request_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (org_id, request_id) REFERENCES requests(org_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_request_comments_request 
ON request_comments(org_id, request_id, created_at);
```

**ポイント**:
- イベントストアとは**別のテーブル**（Read Model）
- クエリ性能のためのインデックス追加
- 外部キー制約でデータ整合性を保証

#### ステップ3: SQLクエリ定義（sqlc）

**ファイル**: `db/query/comments.sql`

```sql
-- name: ListCommentsByRequest :many
SELECT id, org_id, request_id, user_id, content, created_at
FROM request_comments
WHERE org_id = $1 AND request_id = $2
ORDER BY created_at ASC;

-- name: InsertComment :one
INSERT INTO request_comments (id, org_id, request_id, user_id, content, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, org_id, request_id, user_id, content, created_at;
```

**実行**: `sqlc generate` で型安全なGoコードが自動生成される

#### ステップ4: Projector実装

**ファイル**: `internal/infrastructure/projector/request_comments_projector.go`

```go
type RequestCommentsProjector struct {
    queries *dbgen.Queries
}

func (p *RequestCommentsProjector) HandleEvent(ctx context.Context, evt Event) error {
    // RequestCommented イベントのみ処理
    if evt.EventType != request.EventTypeRequestCommented {
        return nil
    }
    
    // ペイロード解析
    var payload request.RequestCommentedPayload
    if err := json.Unmarshal(evt.Payload, &payload); err != nil {
        return err
    }
    
    // Read Model に挿入（プロジェクション）
    _, err := p.queries.InsertComment(ctx, dbgen.InsertCommentParams{
        ID:        uuid.MustParse(payload.CommentID),
        OrgID:     evt.OrgID,
        RequestID: evt.AggregateID,
        UserID:    uuid.MustParse(payload.UserID),
        Content:   payload.Content,
        CreatedAt: payload.CommentedAt,
    })
    
    return err
}
```

**プロジェクターの登録**:
```go
// cmd/projector/main.go
runners := []projector.Runner{
    projector.NewRequestsProjector(queries),
    projector.NewRequestCommentsProjector(queries), // ← 追加
}
```

**ポイント**:
- イベントを受け取って Read Model を更新
- トランザクション内で実行される（イベント発行と同時）
- エラー時はロールバックされる（データ整合性保証）

#### ステップ5: Service層にビジネスロジック追加

**ファイル**: `internal/usecase/requests/service.go`

```go
func (s *Service) AddComment(ctx context.Context, orgID, actorUserID, requestID uuid.UUID, content string) (uuid.UUID, error) {
    // 1. トランザクション開始
    tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
    if err != nil {
        return uuid.Nil, err
    }
    defer tx.Rollback()
    
    // 2. 権限チェック（組織メンバーのみ）
    _, err = queries.GetMembershipByOrgAndUserID(ctx, ...)
    if err != nil {
        return uuid.Nil, ErrUnauthorized
    }
    
    // 3. イベント発行
    commentID := uuid.New()
    payload := request.RequestCommentedPayload{
        CommentID:   commentID.String(),
        UserID:      actorUserID.String(),
        Content:     content,
        CommentedAt: time.Now(),
    }
    payloadBytes, _ := json.Marshal(payload)
    
    // イベントストアに追加
    eventstore.Append(ctx, tx, orgID, "request", requestID, 
        request.EventTypeRequestCommented, payloadBytes, metadata)
    
    // 4. プロジェクター実行（同期）
    // ← Read Model が自動的に更新される
    projector.CatchUpTx(ctx, tx, orgID)
    
    // 5. コミット
    tx.Commit()
    
    return commentID, nil
}
```

**ポイント**:
- イベント発行 → プロジェクター → Read Model更新が1トランザクション
- エラー時は全てロールバック（原子性保証）

#### ステップ6: GraphQLスキーマ定義

**ファイル**: `internal/interface/graphql/schema.graphqls`

```graphql
type Comment {
  id: ID!
  requestID: ID!
  userID: ID!
  content: String!
  createdAt: Time!
}

extend type Query {
  comments(requestID: ID!): [Comment!]!
}

extend type Mutation {
  addComment(requestID: ID!, content: String!): Comment!
}
```

**実行**: `gqlgen generate` で resolver スタブが生成される

#### ステップ7: Resolver実装

**ファイル**: `internal/interface/graphql/graph/schema.resolvers.go`

```go
// Query
func (r *queryResolver) Comments(ctx context.Context, requestID string) ([]*model.Comment, error) {
    viewer, err := viewerFromContext(ctx)
    if err != nil {
        return nil, err
    }
    
    reqID := uuid.MustParse(requestID)
    
    // Read Model から取得（高速）
    rows, err := r.queries.ListCommentsByRequest(ctx, dbgen.ListCommentsByRequestParams{
        OrgID:     viewer.OrgID,
        RequestID: reqID,
    })
    if err != nil {
        return nil, err
    }
    
    // GraphQL型に変換
    var comments []*model.Comment
    for _, row := range rows {
        comments = append(comments, &model.Comment{
            ID:        row.ID.String(),
            RequestID: row.RequestID.String(),
            UserID:    row.UserID.String(),
            Content:   row.Content,
            CreatedAt: row.CreatedAt,
        })
    }
    
    return comments, nil
}

// Mutation
func (r *mutationResolver) AddComment(ctx context.Context, requestID string, content string) (*model.Comment, error) {
    viewer, err := viewerFromContext(ctx)
    if err != nil {
        return nil, err
    }
    
    reqID := uuid.MustParse(requestID)
    
    // Service層を呼ぶ（イベント発行 + プロジェクション実行）
    commentID, err := r.service.AddComment(ctx, viewer.OrgID, viewer.UserID, reqID, content)
    if err != nil {
        return nil, err
    }
    
    // 作成されたコメントを取得
    row, err := r.queries.GetComment(ctx, dbgen.GetCommentParams{
        OrgID: viewer.OrgID,
        ID:    commentID,
    })
    if err != nil {
        return nil, err
    }
    
    return &model.Comment{
        ID:        row.ID.String(),
        RequestID: row.RequestID.String(),
        UserID:    row.UserID.String(),
        Content:   row.Content,
        CreatedAt: row.CreatedAt,
    }, nil
}
```

---

### Phase 2: フロントエンド - UI実装（LEARN-022F）

#### ステップ1: GraphQLクエリ定義

**ファイル**: `frontend/src/gql/comments.graphql`

```graphql
query RequestComments($requestID: ID!) {
  comments(requestID: $requestID) {
    id
    requestID
    userID
    content
    createdAt
  }
}

mutation AddComment($requestID: ID!, $content: String!) {
  addComment(requestID: $requestID, content: $content) {
    id
    requestID
    userID
    content
    createdAt
  }
}
```

**実行**: `npm run codegen` で型と Document が自動生成される

#### ステップ2: コンポーネント実装

**ファイル**: `frontend/src/components/CommentList.tsx`

```tsx
import { useQuery } from 'urql';
import { RequestCommentsDocument } from '../gql/graphql';

export function CommentList({ requestID }: { requestID: string }) {
  // GraphQL Query実行
  const [{ data, fetching, error }] = useQuery({
    query: RequestCommentsDocument,
    variables: { requestID },
    requestPolicy: 'cache-and-network', // リアルタイム更新
  });

  if (fetching) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div className="commentList">
      {data.comments.map(comment => (
        <div key={comment.id} className="comment">
          <div className="commentAuthor">{comment.userID}</div>
          <div className="commentContent">{comment.content}</div>
          <div className="commentDate">
            {new Date(comment.createdAt).toLocaleString()}
          </div>
        </div>
      ))}
    </div>
  );
}
```

**ファイル**: `frontend/src/components/CommentForm.tsx`

```tsx
import { useMutation } from 'urql';
import { AddCommentDocument } from '../gql/graphql';

export function CommentForm({ requestID }: { requestID: string }) {
  const [content, setContent] = useState('');
  const [{ fetching }, addComment] = useMutation(AddCommentDocument);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const result = await addComment({ requestID, content });
    
    if (result.error) {
      alert('Failed to add comment');
      return;
    }
    
    setContent(''); // フォームクリア
  };

  return (
    <form onSubmit={handleSubmit} className="commentForm">
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="コメントを入力..."
        disabled={fetching}
      />
      <button type="submit" disabled={fetching || !content.trim()}>
        送信
      </button>
    </form>
  );
}
```

#### ステップ3: ページに統合

**ファイル**: `frontend/src/pages/RequestDetail.tsx`

```tsx
import { CommentList } from '../components/CommentList';
import { CommentForm } from '../components/CommentForm';

export function RequestDetail() {
  const { id } = useParams();
  
  return (
    <div>
      {/* 申請詳細 */}
      <RequestInfo requestID={id} />
      
      {/* コメントセクション */}
      <section className="commentsSection">
        <h2>コメント</h2>
        <CommentList requestID={id} />
        <CommentForm requestID={id} />
      </section>
    </div>
  );
}
```

---

## まとめ: なぜこんなに多くのコードが必要なのか？

### レイヤー別の役割整理

| レイヤー | ファイル例 | 責務 |
|----------|-----------|------|
| **Domain** | `events.go`, `aggregate.go` | ビジネスルール、イベント定義 |
| **Infrastructure - Event Store** | `event_store.go` | イベント永続化 |
| **Infrastructure - Projector** | `request_comments_projector.go` | イベント → Read Model 変換 |
| **Infrastructure - Database** | `comments.sql`, sqlc生成コード | Read Model CRUD |
| **UseCase** | `service.go` | トランザクション管理、権限チェック |
| **Interface - GraphQL** | `schema.graphqls`, `resolvers.go` | APIエンドポイント |
| **Frontend** | `CommentList.tsx`, `CommentForm.tsx` | UI表示、ユーザー操作 |

### 必要な作業まとめ

**バックエンド（7ステップ）**:
1. イベント型定義
2. マイグレーション作成
3. SQLクエリ定義 + sqlc generate
4. Projector実装
5. Service層にビジネスロジック追加
6. GraphQLスキーマ定義 + gqlgen generate
7. Resolver実装

**フロントエンド（3ステップ）**:
1. GraphQLクエリ定義 + codegen
2. コンポーネント実装（List, Form）
3. ページに統合

**合計10ステップ、約500-1000行のコード** ← 確かに多い！

### これだけの作業が必要な理由

1. **レイヤー分離**: 各レイヤーが独立してテスト可能
2. **型安全性**: sqlc, gqlgen, codegen で自動生成
3. **監査証跡**: すべての変更がイベントとして記録される
4. **データ整合性**: トランザクション + プロジェクターで保証
5. **パフォーマンス**: Read Model でクエリ最適化
6. **保守性**: 各コンポーネントの責務が明確

### REST APIだったら？

```javascript
// REST の場合（シンプルだがトレードオフあり）

// バックエンド
app.post('/api/requests/:id/comments', (req, res) => {
  const { content } = req.body;
  db.query('INSERT INTO comments (request_id, content) VALUES (?, ?)', 
    [req.params.id, content]);
  res.json({ success: true });
});

app.get('/api/requests/:id/comments', (req, res) => {
  const comments = db.query('SELECT * FROM comments WHERE request_id = ?', 
    [req.params.id]);
  res.json(comments);
});

// フロントエンド
fetch(`/api/requests/${id}/comments`, {
  method: 'POST',
  body: JSON.stringify({ content }),
});
```

**RESTの方がシンプル**だが：
- 監査証跡が残らない（誰が、いつ、何を変更したか不明）
- データの再構築ができない
- 複数のエンドポイントが必要
- Over-fetching、Under-fetchingの問題

---

## 開発Tips

### デバッグ方法

1. **イベントストアを確認**:
   ```sql
   SELECT * FROM event_store 
   WHERE aggregate_id = 'request-id' 
   ORDER BY sequence_number;
   ```

2. **Projectorの動作確認**:
   ```sql
   SELECT * FROM request_comments WHERE request_id = 'request-id';
   ```

3. **GraphQL Playground**:
   `http://localhost:8080/graphql` で直接クエリを試す

### よくあるエラー

#### "Cannot query field 'comments'"
- GraphQLスキーマに定義されていない
- `gqlgen generate` を実行していない

#### "permission denied"
- Resolver で権限チェックが必要
- `viewerFromContext` でユーザー情報を取得

#### "duplicate key value"
- Projectorの冪等性が保証されていない
- `ON CONFLICT DO NOTHING` または `UPSERT` を使う

---

## 次のステップ

このガイドを理解したら、以下のタスクにチャレンジしてみましょう：

- **LEARN-023**: 検索・フィルタリング機能（WHERE句の動的構築）
- **LEARN-024**: ダッシュボード統計（GROUP BY集計）
- **LEARN-025**: ファイルダウンロード（署名付きURL）

各タスクで同じパターンを繰り返すことで、アーキテクチャへの理解が深まります。

---

## 参考リンク

- [Event Sourcing パターン - Martin Fowler](https://martinfowler.com/eaaDev/EventSourcing.html)
- [CQRS パターン](https://docs.microsoft.com/ja-jp/azure/architecture/patterns/cqrs)
- [GraphQL 公式ドキュメント](https://graphql.org/learn/)
- [gqlgen - Go GraphQL ライブラリ](https://gqlgen.com/)
- [sqlc - SQL → Go 変換ツール](https://sqlc.dev/)

---

**作成日**: 2026-02-11  
**対象タスク**: LEARN-022, LEARN-022F  
**想定読者**: PonSu開発者（初めてEvent Sourcing + GraphQLに触れる人）
