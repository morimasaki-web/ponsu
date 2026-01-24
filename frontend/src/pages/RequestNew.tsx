    import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useMutation } from 'urql'
import { appMode } from '../graphql'
import { CreateRequestDocument } from '../gql/graphql'

export default function RequestNew() {
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const [{ fetching }, createRequest] = useMutation(CreateRequestDocument)

  const canSubmit = appMode === 'prod' && title.trim().length > 0 && !fetching

  return (
    <section className="card">
      <h2>新規申請（GraphQL）</h2>

      {appMode !== 'prod' && (
        <p className="note">デモモードでは新規作成は無効です。</p>
      )}

      {errorMsg && <p className="error">{errorMsg}</p>}

      <div className="kv">
        <span className="kvKey">タイトル</span>
        <input
          className="input"
          value={title}
          placeholder="例: 経費精算（1月）"
          onChange={(e) => setTitle(e.target.value)}
          disabled={appMode !== 'prod' || fetching}
        />
      </div>

      <div className="actions">
        <Link className="btn btn--ghost" to="/requests">
          一覧へ戻る
        </Link>
        <button
          className="btn"
          disabled={!canSubmit}
          onClick={async () => {
            setErrorMsg(null)
              const res = await createRequest({
                title: title.trim(),
                workflowTemplateID: null,
              })
            if (res.error) {
              setErrorMsg(res.error.message)
              return
            }
            const id = res.data?.createRequest?.id
            if (!id) {
              setErrorMsg('作成に失敗しました（idが取得できません）')
              return
            }
            navigate(`/requests/${encodeURIComponent(id)}`)
          }}
        >
          作成
        </button>
      </div>
    </section>
  )
}
