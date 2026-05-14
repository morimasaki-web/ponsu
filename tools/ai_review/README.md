# ai_review

別モデル（別チャット/別LLM）にコードレビューさせるための「レビュー・バンドル」を生成するツールです。

## できること
- 現在の git working tree から差分（staged/unstaged）と状態を集める
- 貼り付け用のレビュー依頼プロンプト（prompt.md）を自動生成する

## 使い方
リポジトリのルート（例: `MyApp/`）で実行します。

```bash
python ponsu/tools/ai_review/prepare_review.py
```

成功すると、生成先ディレクトリのパスが標準出力に表示されます。

例: `.ai-review/20260112_123456_main/`

中身:
- `prompt.md`
- `status.txt`
- `diff_unstaged.patch`
- `diff_staged.patch`
- `files_changed.txt`

## 補足
- APIキーは不要です（別モデルに貼り付ける前提）。
- 生成プロンプトには `ponsu/docs/design/MVP設計書.md` と `ponsu/docs/tasks.yaml` の先頭を短く抜粋して含めます（存在する場合）。
