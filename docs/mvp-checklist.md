# MVP受け入れチェックリスト（手動/E2E）

この文書は、MVPの主要フローが「手順通りに再現できる」ことを確認するためのチェックリストです。

## 前提

- PostgreSQL が用意され、マイグレーションが適用済みである
- OIDC が設定済みで、少なくとも2ユーザーでログインできる
  - ユーザーA: admin（承認/却下/差し戻しができる）
  - ユーザーB: member（作成/提出/再提出ができる）
- アプリが起動できる（SSR画面にアクセスできる）

## 共通：テストデータ

- 組織（org）を1つ用意する
- ユーザーA（admin）とユーザーB（member）を同一orgに所属させる

## シナリオ1：作成 → 提出 → 承認 → 監査参照

1. ユーザーBでログイン
2. リクエストを新規作成する（タイトルを入力）
3. 作成したリクエスト詳細を開く
4. 提出（Submit）する
5. ユーザーAでログイン
6. 対象リクエスト詳細を開く
7. 承認（Approve）する
8. 監査（audit trail）が表示でき、以下が時系列で残っていること
   - `request.created`
   - `request.submitted`
   - `request.approved`

期待結果:
- ステータスが `approved` になっている
- 監査が欠けずに表示される

## シナリオ2：作成 → 提出 → 却下 → 監査参照

1. ユーザーBでログイン
2. リクエストを新規作成する（タイトルを入力）
3. 提出（Submit）する
4. ユーザーAでログイン
5. 対象リクエスト詳細を開く
6. 却下（Reject）する（理由を入力）
7. 監査（audit trail）に以下が残っていること
   - `request.created`
   - `request.submitted`
   - `request.rejected`（理由が保存されている）

期待結果:
- ステータスが `rejected` になっている
- 却下理由が監査に残る

## シナリオ3：差し戻し → 再提出

1. ユーザーBでログイン
2. リクエストを新規作成する（タイトルを入力）
3. 提出（Submit）する
4. ユーザーAでログイン
5. 対象リクエスト詳細を開く
6. 差し戻し（Return）する（理由を入力）
7. ユーザーBでログイン
8. 対象リクエスト詳細を開く
9. 再提出（Resubmit）する
10. 監査（audit trail）に以下が残っていること
    - `request.created`
    - `request.submitted`
    - `request.returned`（理由が保存されている）
    - `request.resubmitted`

期待結果:
- 差し戻し後、ステータスが `returned` になっている
- 再提出後、ステータスが `submitted` に戻っている
- RBACが効いている
  - admin以外が差し戻しできない
  - 作成者以外が再提出できない

## シナリオ4：権限（RBAC）チェック

- ユーザーB（member）でログイン中
  - 承認/却下/差し戻しの操作ができない（ボタンが出ない、または403）
- ユーザーA（admin）でログイン中
  - 提出/再提出（作成者制約がある操作）を実行できない（ボタンが出ない、または403）

## 参考（API）

- GraphQL Mutation
  - `returnRequest(id: ID!, reason: String!): Request!`
  - `resubmitRequest(id: ID!): Request!`
