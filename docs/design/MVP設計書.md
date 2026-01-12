# Ponsu MVP System Design Specification

## 1. プロダクトコンセプト
- **名称**: Ponsu (ポンス)
- **ひとことで**: 「ポンッ」と承認、スッと流れる。紙・印鑑・稟議を最小ステップでデジタル化する軽量ワークフロー。
- **差別化**: 最初から全てをデジタル化せず、監査証跡（誰がいつ承認したか）をコアにしたスモールスタート設計。

---

## 2. 技術スタック
- **Language**: Go (1.22+)
- **Database**: PostgreSQL (JSONBを活用したスキーマレス・リレーショナル・ハイブリッド)
- **Auth**: OIDC (Google/Microsoft 365)
- **API**: GraphQL
- **Architecture**: 
    - クリーンアーキテクチャ（Domain, Usecase, Interface, Infrastructure）
    - AIによるコード生成と型安全を重視した `sqlc` によるDB操作
    - SSR（サーバーサイドレンダリング）を先に実装し、後からSPAクライアントを追加できるよう API を分離

### 2.1 採用方針（MVP確定）
- DBアクセス層: `sqlc`
- API方式: GraphQL（MVPから採用）
- 監査証跡: イベントソーシング（イベントストア + プロジェクション）
- UI: SPA化を見据えたSSR（まずはSSRで運用投入を優先）
- 認可: RBAC

---

## 3. MVPの目的と成功指標
### 3.1 MVPの目的
- 「承認が起きた事実（誰が・いつ・何を・どう判断したか）」を、最小の運用負荷で記録できる状態を作る。
- 稟議を“全部デジタル化”する前に、承認ログと承認体験の核を作ってスモールスタートできる状態にする。

### 3.2 成功指標（MVP段階）
- 申請作成〜承認（承認/却下）までが1つのプロダクト内で完結する
- 監査証跡（承認ログ）が後から追える（検索・参照が可能）
- 1社（1組織）での試験運用ができる（ユーザー招待、最低限の権限、運用導線）

---

## 4. スコープ定義
### 4.1 MVPに含める（In Scope）
- 組織（テナント）
- ユーザー（OIDCログイン）
- 申請（タイトル/説明/添付/任意の項目）
- 承認フロー（テンプレ＋実行インスタンス）
- 承認アクション（承認/却下/差し戻しのうち最低「承認/却下」）
- 監査証跡（いつ誰が何をしたかの履歴）
- 通知（Slack通知を最低1つ）
- テンプレ作成UI（管理者向け）

MVP後半（段階的に追加）
- 差し戻し（`returned`）
- 再提出（`resubmitted`）

### 4.2 MVPに含めない（Out of Scope / 後回し）
- 高度な権限（ABAC、複雑な条件分岐、職位/金額/部門連携の自動ルーティング）
- 電子署名/タイムスタンプ/改ざん検知の強化（MVPは“追跡可能”を優先）
- 外部システム連携（会計/ERP/ワークフロー既存製品）
- 多言語対応
- モバイルネイティブアプリ
- 高度な通知運用（複雑な購読ルール、通知テンプレの多言語、ユーザーごとの細かな購読設定）

---

## 5. 想定ユーザーとユースケース
### 5.1 ペルソナ
- 申請者: 日常的に稟議/申請を出す（購買/経費/出張など）
- 承認者: 上長/責任者（承認の意思決定が主業務ではない）
- 管理者: 組織/テンプレ/ユーザー管理を行う（情シス/総務）

### 5.2 代表ユースケース
1. 申請者が申請を作成して提出
2. 承認者が通知から詳細を開き、承認/却下
3. 管理者が承認テンプレを作り、承認者を割り当て
4. 監査対応で「いつ誰が承認したか」を参照・出力

---

## 6. 意思決定ポイント（選択肢・メリデメ・推奨）

### 6.1 DBアクセス層: `sqlc` vs `ent`
| 方式 | 概要 | メリット | デメリット | MVP推奨 |
|---|---|---|---|---|
| `sqlc` | SQLを正とし、型付きコード生成 | SQLが明確、性能予測しやすい、移行しやすい | クエリ管理が増える、関係が複雑だとSQLが肥大化しがち | ◎ |
| `ent` | スキーマ定義を正とするORM | CRUDが速い、関連の取り回しが楽、開発体験が良い | ORM都合が出る、複雑SQL/最適化で詰まりやすい | ○ |

決定: `sqlc` を採用。テンプレ/申請の柔軟項目は JSONB を併用し、クエリは最小限から始める。

### 6.2 API方式: REST vs GraphQL
| 方式 | メリット | デメリット | MVP推奨 |
|---|---|---|---|
| REST | 実装/デバッグが単純、キャッシュしやすい、ツールが豊富 | 画面が増えるとエンドポイントが増える | ◎ |
| GraphQL | 画面都合で取得形を最適化しやすい | 設計/運用コスト増、認可/監査が難しくなりやすい | △ |

決定: GraphQL を採用。認可・監査を難しくしないため、スキーマ設計とリゾルバ層にポリシーを明示し、監査ログは必ずイベントとして記録する。

### 6.3 監査証跡モデル: 追記型ログ vs イベントソーシング
| 方式 | 概要 | メリット | デメリット | MVP推奨 |
|---|---|---|---|---|
| 追記型ログ（append-only） | `approval_events` 等に履歴を追加 | 実装が簡単、監査対応しやすい、後で強化しやすい | 厳密な再生（リプレイ）までは保証しない | ◎ |
| イベントソーシング | 状態はイベント列から再構成 | 改ざん耐性を作りやすい、状態遷移が明確 | 学習/実装/運用コスト大、MVPで過剰 | △ |

決定: イベントソーシングにチャレンジする。MVPでは「イベントストア + 同期プロジェクション（Read Model更新）」で運用複雑性を抑える。

### 6.4 UI実装: SSR（Goテンプレ）/SPA/最初はUIなし
| 方式 | メリット | デメリット | MVP推奨 |
|---|---|---|---|
| SSR（Goテンプレ + htmx等） | 速度・単純さ・デプロイが楽、MVPに強い | UIが複雑になると辛い | ◎ |
| SPA（React等） | 体験が良く拡張しやすい | 初期の構成・認証・型共有が重い | ○ |
| UIなし（APIのみ） | 実装量最小 | 実運用/検証が困難、価値検証が遅い | △ |

決定: SPA化を見据えたSSR。
方針:
- まずSSRで運用投入し、ユーザー検証を回す
- 同じバックエンドで GraphQL API を提供し、将来SPAクライアントを追加してもバックエンドのユースケースを流用できる構造にする

### 6.5 通知方式: Slack / メール / アプリ内
| 方式 | メリット | デメリット | MVP推奨 |
|---|---|---|---|
| Slack Incoming Webhook | 実装が最短、運用が軽い、デモで刺さりやすい | 個人DMなど柔軟性が低い、権限設計はWebhook次第 | ◎ |
| Slack App（Bot） | DM/メンション/インタラクションなど拡張性が高い | OAuth/権限/配布が重い、MVPで工数増 | ○ |
| メール | 外部依存が少なく誰でも使える | 到達性・テンプレ管理が意外と重い | ○ |
| アプリ内通知 | 体験が良い、履歴が残る | 画面/UI/既読管理の実装が必要 | ○ |

推奨: MVPは Slack Incoming Webhook から開始。
将来: Bot化（DM通知や承認ボタン等）が必要になったら移行。

### 6.6 添付保存先: ローカル / MinIO / 外部SaaS
前提: 「デモ運用でお金が発生しない」「受けがいい（説明しやすい/それっぽい）」を優先。

| 方式 | メリット | デメリット | MVP推奨 |
|---|---|---|---|
| ローカルファイル（サーバーディスク） | 追加コスト0、実装が簡単、最短で動く | スケール/冗長化が弱い、デプロイ形態で壊れやすい | ○ |
| MinIO（S3互換・セルフホスト） | 追加コスト0（自前）、S3互換で“受けがいい”、将来のS3移行が楽 | 運用対象が増える、バックアップ設計が必要 | ◎ |
| 外部SaaS（S3/R2等） | 運用が楽、可用性が高い | 無料枠超過で課金、デモでも説明がやや面倒 | △ |

推奨: MinIO（S3互換）を採用。
実装方針: `Storage` インターフェースを切り、最初はローカル実装で開発しつつ、デモ環境はMinIOに寄せる。

### 6.7 マイグレーションツール: `migrate` / `goose` / `atlas`
| 方式 | メリット | デメリット | MVP推奨 |
|---|---|---|---|
| `golang-migrate/migrate` | 定番、CI/CDに乗せやすい、言語非依存、SQL運用がシンプル | 仕組みは薄味で、運用規約は自分で決める必要 | ◎ |
| `goose` | Goプロジェクトで使いやすい、SQL/Go両対応 | プロジェクトによって運用流儀が分かれる | ○ |
| `atlas` | 宣言的なスキーマ管理が強い、差分生成が便利 | 学習コスト、導入がやや重い（MVPで過剰になりがち） | △ |

推奨: `golang-migrate/migrate`（SQLマイグレーションを正とする）。

### 6.8 GraphQL実装ライブラリ（Go）: `gqlgen` ほか
| 方式 | メリット | デメリット | MVP推奨 |
|---|---|---|---|
| `99designs/gqlgen` | スキーマファースト、型安全、生成コードが実用的、事例が多い | 設計を雑にするとスキーマが肥大化しやすい | ◎ |
| `graph-gophers/graphql-go` | シンプル、軽い | エコシステム/生成周りが弱くなりがち | ○ |
| 自前実装/薄い実装 | 柔軟 | 罠が多く、MVPには不向き | △ |

推奨: `gqlgen`。

---

## 7. ドメインモデル（概念）
### 7.1 コア概念
- Organization（組織/テナント）
- User（ユーザー）
- Membership（ユーザーの組織所属・ロール）
- WorkflowTemplate（承認テンプレ）
- Request（申請/稟議）
- ApprovalStep（承認ステップ）
- ApprovalAction（承認アクション: approve/reject/…）
- AuditEvent（監査イベント: create/submit/approve/reject/comment/…）
- Attachment（添付）

### 7.2 状態遷移（MVP最小）
段階1（MVP前半）
- RequestStatus: `draft` → `submitted` → (`approved` | `rejected`)

段階2（MVP後半）
- RequestStatus: `draft` → `submitted` → `returned` → `resubmitted` → (`approved` | `rejected`)

注: `resubmitted` を独立ステータスにせず `submitted` に戻す設計も可能。監査ログ表示と整合する方をMVP後半で確定する。

---

## 8. データ設計（PostgreSQL / JSONB併用方針）
### 8.1 テーブル案（MVP）
イベントソーシングのため、Write Model（イベント）と Read Model（参照しやすいテーブル）を分ける。

Write Model（イベントストア）
- `event_store`（全イベントの単一テーブル）

Read Model（プロジェクション）
- `organizations`
- `users`
- `memberships`（org_id, user_id, role）
- `workflow_templates`
- `requests`（申請の現在状態・検索用）
- `request_steps`（テンプレ展開したステップの実体）
- `attachments`
- `request_audit_trail`（申請単位の監査表示用の投影。必要なら）

※ `approval_actions` / `audit_events` 相当はイベントから投影できるが、MVPでは画面表示・検索のために投影テーブルを置くのは許容する。

### 8.2 イベントストア設計（MVP案）
`event_store` の主なカラム案:
- `id` (uuid)
- `org_id` (uuid)
- `aggregate_type` (text) 例: `request`
- `aggregate_id` (uuid)
- `version` (int)（同一aggregate内の連番。楽観ロック）
- `event_type` (text) 例: `request.submitted`
- `payload` (jsonb)（イベントの中身）
- `metadata` (jsonb)（actor_user_id, request_id, ip, user_agent など）
- `occurred_at` (timestamptz)

制約:
- `unique(org_id, aggregate_type, aggregate_id, version)`
- インデックス: `(org_id, aggregate_type, aggregate_id)`, `(org_id, occurred_at)`

プロジェクション更新:
- MVPは「イベント追記 + Read Model更新」を同一トランザクションで実行（同期プロジェクション）
- 将来はアウトボックス/非同期ワーカーに移行可能な形で境界を切る

### 8.2 “柔軟な項目”の持ち方: 正規化 vs JSONB
| 方式 | メリット | デメリット | MVP推奨 |
|---|---|---|---|
| 正規化（列/別テーブル） | 検索・制約が強い | 仕様変更に弱い、開発が重い | ○（固定項目のみ） |
| JSONB（任意項目） | 仕様変更に強い、テンプレ自由度が高い | 検索/制約が弱い、クエリが難しい | ◎（任意項目） |

推奨: 「固定で必要な列」は正規化（例: title, status, created_by, created_at）。「テンプレ依存の入力項目」は JSONB で保持（例: form_data）。

---

## 8.3 MVPで扱う主要イベント（案）
Request（申請）を主なAggregateとし、状態はイベントから導出する。

- `request.created`
- `request.updated`（MVPでは省略可。draft編集を入れるなら）
- `request.submitted`
- `request.approved`
- `request.rejected`
- `request.returned`（MVP後半）
- `request.resubmitted`（MVP後半）
- `attachment.added`

メタデータ（共通）:
- `actor_user_id`
- `actor_role`（必要なら）
- `correlation_id`（リクエスト追跡）

---

## 9. 認証・認可
### 9.1 認証（OIDC）
- Google / Microsoft 365 のどちらか一方から開始してもよい（実運用先に合わせる）
- MVPは「ログインできる」「組織に紐づく」までを最優先

### 9.2 認可（MVPの最小設計）
- 組織ロール: `admin`, `member`
- 操作権限（最小）
    - `admin`: ユーザー招待、テンプレ作成、全申請参照
    - `member`: 自分の申請作成、関与する申請の承認

選択肢: RBAC vs ABAC
| 方式 | メリット | デメリット | MVP推奨 |
|---|---|---|---|
| RBAC | 実装が単純、説明しやすい | 例外が増えると辛い | ◎ |
| ABAC | 条件で柔軟に制御可能 | 実装/検証が難しい、監査観点で説明が難しい | △ |

推奨: MVPはRBAC。ABACは要件が出てから追加。

---

## 10. 主要フロー（画面/操作）
### 10.1 申請作成〜提出
1. 申請一覧 → 「新規作成」
2. テンプレ選択（または固定テンプレ1つで開始）
3. 入力（タイトル/説明/任意項目/添付）
4. 下書き保存（任意）
5. 提出

### 10.2 承認
1. 通知（メール or アプリ内）
2. 申請詳細で内容・添付・履歴を確認
3. 承認/却下
4. 履歴に記録（監査証跡）

### 10.4 差し戻し・再提出（MVP後半）
1. 承認者が「差し戻し」理由を添えて実行
2. 申請者に通知
3. 申請者が修正して再提出
4. 履歴（イベント）として残る

### 10.3 監査ログ閲覧
- 申請詳細で時系列表示
- 管理者向けに検索（期間/申請者/ステータス）

---

## 11. API設計（叩き台）
※ 実装前に `docs/tasks.md` 化してから確定する。

### 11.1 GraphQL スキーマ（叩き台）
MVPでは Query/Mutation を最小に絞り、「申請→提出→承認/却下→参照」を通す。

#### Query
- `me: User!`
- `requests(orgId: ID!, status: RequestStatus, limit: Int, cursor: String): RequestConnection!`
- `request(orgId: ID!, id: ID!): Request`
- `workflowTemplates(orgId: ID!): [WorkflowTemplate!]!`

#### Mutation
- `createRequest(input: CreateRequestInput!): Request!`
- `submitRequest(input: SubmitRequestInput!): Request!`
- `approveRequest(input: ApproveRequestInput!): Request!`
- `rejectRequest(input: RejectRequestInput!): Request!`
- `createWorkflowTemplate(input: CreateWorkflowTemplateInput!): WorkflowTemplate!`
- `createWorkflowTemplate(input: CreateWorkflowTemplateInput!): WorkflowTemplate!`（テンプレ作成UI向け）
- `returnRequest(input: ReturnRequestInput!): Request!`（MVP後半）
- `resubmitRequest(input: ResubmitRequestInput!): Request!`（MVP後半）

#### 型（例）
- `type Request { id, orgId, title, description, status, formData, createdBy, createdAt, updatedAt, auditTrail }`
- `type AuditEvent { id, type, occurredAt, actor { id, name }, payload }`

### 11.2 認証（HTTP）
OIDCはHTTPのリダイレクトフローが必要なため、ここはGraphQLではなくHTTPルートで持つ。
- `GET /auth/login`
- `GET /auth/callback`
- `POST /auth/logout`

### 11.3 SSR 画面ルート（例）
SSRのHTML配信は通常のHTTPルートで提供する。
- `GET /`（ダッシュボード/申請一覧）
- `GET /requests/new`
- `GET /requests/:id`

SSRの内部実装は「Usecase直呼び」を基本としつつ、将来SPAが GraphQL を直接叩く形に備えて、Usecase/Domainにロジックを集約する。

---

## 12. 非機能要件（MVPの現実ライン）
### 12.1 セキュリティ
- CSRF対策（Cookieセッションの場合）/ SameSite 設定
- イベントストアは削除不可（原則追記のみ。修正は補正イベントで表現）
- 添付はウイルススキャンは将来（MVPはサイズ制限・拡張子制限・保存先分離）

### 12.2 運用
- DBマイグレーション: `golang-migrate/migrate` を推奨
- バックアップ: Postgresの定期バックアップ（手順をドキュメント化）

### 12.3 監視
- 構造化ログ（request_id, user_id, org_id を持つ）
- メトリクスはMVPでは最低限でも可（エラー率/レイテンシ）

---

## 13. テスト戦略
- Domain/Usecase: ユニットテスト中心
- Infrastructure(DB): 重要クエリは結合テスト（ローカルPostgres）
- API: 主要フロー（作成→提出→承認→参照）のE2E相当を最小本数

イベントソーシング特有の重点:
- イベント→状態再構成（Aggregateのreplay）
- 楽観ロック（version）競合時の挙動
- イベント追記とRead Model更新の同一トランザクション性

---

## 14. リスクと対策（MVP時点）
- 認可の漏れ: ルート単位でポリシーを明示し、テストで担保
- テンプレ自由度の暴走: MVPはテンプレ仕様を最小（固定ステップ or 単純直列）
- 監査/状態の不整合: イベント追記とRead Model更新を同一トランザクションで行う
- GraphQLの過取得/過複雑化: Query/MutationをMVP最小に固定し、認可はリゾルバで必須チェック
- イベント設計の迷走: イベント命名規約とpayloadスキーマを先に固定し、破壊変更は避ける

---

## 15. 未決事項（タスク化候補）
意思決定が必要な項目。ここを `docs/tasks.md` に落としてから実装フェーズへ進む。

1. 通知の詳細（Slack: WebhookかBotか、送信先設定のUI/管理方法）
2. 添付保存の実装詳細（ローカル/MinIO の切替方法、署名URLの要否）
3. イベントストアのスキーマ最終化（イベント命名/metadata項目/スナップショット有無）
4. `returned`/`resubmitted` のステータス設計（`resubmitted` を独立させるか）
