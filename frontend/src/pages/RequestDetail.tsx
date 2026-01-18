import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  demoRequestsSeed,
  formatStatus,
  nowIso,
  type DemoRequest,
  type RequestAudit,
  type RequestStatus,
} from '../demoData'

function canSubmit(status: RequestStatus) {
  return status === 'draft'
}
function canApprove(status: RequestStatus) {
  return status === 'submitted' || status === 'resubmitted'
}
function canReject(status: RequestStatus) {
  return status === 'submitted' || status === 'resubmitted'
}
function canReturn(status: RequestStatus) {
  return status === 'submitted' || status === 'resubmitted'
}
function canResubmit(status: RequestStatus) {
  return status === 'returned'
}

export default function RequestDetail() {
  const params = useParams()
  const id = decodeURIComponent(params.id ?? '')

  const seed = useMemo(
    () => demoRequestsSeed.find((r) => r.id === id) ?? null,
    [id],
  )

  const [req, setReq] = useState<DemoRequest | null>(() =>
    seed ? structuredClone(seed) : null,
  )

  const pushAudit = (a: Omit<RequestAudit, 'at'>) => {
    setReq((prev) => {
      if (!prev) return prev
      const at = nowIso()
      return {
        ...prev,
        updatedAt: at,
        audit: [...prev.audit, { at, ...a }],
      }
    })
  }

  const setStatus = (status: RequestStatus, action: string, note?: string) => {
    setReq((prev) => {
      if (!prev) return prev
      return { ...prev, status }
    })
    pushAudit({ actor: 'Demo User', action, note })
  }

  if (!req) {
    return (
      <section className="card">
        <h2>申請詳細</h2>
        <p className="error">申請が見つかりませんでした。</p>
        <p>
          <Link to="/requests">一覧へ戻る</Link>
        </p>
      </section>
    )
  }

  return (
    <div className="stack">
      <section className="card">
        <h2>申請詳細（デモ）</h2>
        <p className="note">
          操作は疑似的にローカル状態を更新します（DB保存なし）。
        </p>

        <div className="row">
          <div>
            <div className="kv">
              <span className="kvKey">ID</span>
              <span className="mono">{req.id}</span>
            </div>
            <div className="kv">
              <span className="kvKey">タイトル</span>
              <span>{req.title}</span>
            </div>
            <div className="kv">
              <span className="kvKey">作成者</span>
              <span>{req.createdBy}</span>
            </div>
            <div className="kv">
              <span className="kvKey">ステータス</span>
              <span className={`badge badge--${req.status}`}>
                {formatStatus(req.status)}
              </span>
            </div>
          </div>

          <div className="actions">
            <Link className="btn btn--ghost" to="/requests">
              一覧へ戻る
            </Link>
            <button
              className="btn"
              disabled={!canSubmit(req.status)}
              onClick={() => setStatus('submitted', '提出')}
            >
              提出
            </button>
            <button
              className="btn btn--ok"
              disabled={!canApprove(req.status)}
              onClick={() => setStatus('approved', '承認')}
            >
              承認
            </button>
            <button
              className="btn btn--danger"
              disabled={!canReject(req.status)}
              onClick={() => setStatus('rejected', '却下')}
            >
              却下
            </button>
            <button
              className="btn btn--warn"
              disabled={!canReturn(req.status)}
              onClick={() =>
                setStatus('returned', '差し戻し', '添付が不足しています（デモ）')
              }
            >
              差し戻し
            </button>
            <button
              className="btn"
              disabled={!canResubmit(req.status)}
              onClick={() => setStatus('resubmitted', '再提出')}
            >
              再提出
            </button>
          </div>
        </div>

        <div className="cardSub">
          <h3>概要</h3>
          <p>{req.summary}</p>
        </div>
      </section>

      <section className="card">
        <h2>監査ログ（デモ）</h2>
        <ol className="audit">
          {req.audit.map((a, i) => (
            <li key={`${a.at}-${i}`} className="auditItem">
              <div className="auditTop">
                <span className="mono">{new Date(a.at).toLocaleString()}</span>
                <span className="auditAction">{a.action}</span>
                <span className="auditActor">{a.actor}</span>
              </div>
              {a.note && <div className="auditNote">{a.note}</div>}
            </li>
          ))}
        </ol>
      </section>
    </div>
  )
}
