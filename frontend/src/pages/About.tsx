import { Link } from 'react-router-dom'
import { appMode } from '../graphql'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

type Props = {
  onResetDemo?: () => void
}

export default function About({ onResetDemo }: Props) {
  useDocumentTitle('About | Ponsu')

  return (
    <section className="card">
      <h2>デモの見方（ガイド）</h2>

      <p className="note">
        このSPAは「承認フロー（Requests）」の画面遷移と操作を、デモ用のモックデータで体験できるようにしたものです。
      </p>

      <div className="cardSub">
        <h3>操作フロー（おすすめ）</h3>
        <ol>
          <li>
            <Link to="/requests">Requests</Link> で一覧を開く
          </li>
          <li>任意の申請をクリックして詳細へ</li>
          <li>
            <span className="mono">提出 → 承認/却下/差し戻し → 再提出</span> の順で操作
          </li>
          <li>監査ログ（audit）が追記されることを確認</li>
        </ol>
      </div>

      {appMode === 'demo' && (
        <div className="cardSub">
          <h3>デモの注意点</h3>
          <ul>
            <li>認証/サーバ/DBへはアクセスしません（静的ホスティング前提）</li>
            <li>操作はブラウザ内のローカル状態を更新するだけです</li>
            <li>
              状態が分からなくなったら「デモを初期化」で初期状態に戻せます
            </li>
          </ul>

          <div className="actions" style={{ marginTop: 10 }}>
            <Link className="btn btn--ghost" to="/">
              Homeへ戻る
            </Link>
            <button
              className="btn btn--warn"
              onClick={() => onResetDemo?.()}
              disabled={!onResetDemo}
              title={onResetDemo ? '' : 'デモモードでのみ利用できます'}
            >
              デモを初期化
            </button>
          </div>
        </div>
      )}

      {appMode === 'prod' && (
        <div className="cardSub">
          <h3>prodモードについて</h3>
          <p className="note">
            prodモードでは GraphQL + OIDC（Cookie）でバックエンドと連携します。
          </p>
        </div>
      )}
    </section>
  )
}
