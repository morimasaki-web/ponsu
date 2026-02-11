import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useMutation, useQuery } from 'urql'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { ErrorBanner } from '../components/ErrorBanner'
import {
  demoRequestsSeed,
  demoCommentsSeed,
  formatStatus,
  nowIso,
  type DemoRequest,
  type DemoComment,
  type RequestAudit,
  type RequestStatus,
} from '../demoData'
import { appMode } from '../graphql'
import {
  ApproveRequestDocument,
  MeDocument,
  RejectRequestDocument,
  RequestDocument,
  ResubmitRequestDocument,
  ReturnRequestDocument,
  SubmitRequestDocument,
} from '../gql/graphql'
import { CommentList } from '../components/CommentList'
import { CommentForm } from '../components/CommentForm'

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
  useDocumentTitle('Request Detail | Ponsu')

  const params = useParams()
  const id = decodeURIComponent(params.id ?? '')
  const [commentsKey, setCommentsKey] = useState(0)

  // デモコメント用LocalStorageキー
  const DEMO_COMMENTS_KEY = 'ponsu_demo_comments'

  // デモコメントの初期ロード
  const loadDemoComments = (): DemoComment[] => {
    if (appMode !== 'demo') return []
    try {
      const stored = localStorage.getItem(DEMO_COMMENTS_KEY)
      if (stored) {
        return JSON.parse(stored) as DemoComment[]
      }
    } catch (e) {
      console.error('Failed to load demo comments:', e)
    }
    return [...demoCommentsSeed]
  }

  const [demoComments, setDemoComments] = useState<DemoComment[]>(loadDemoComments)

  // デモコメントをLocalStorageに保存
  const saveDemoComments = (comments: DemoComment[]) => {
    if (appMode !== 'demo') return
    try {
      localStorage.setItem(DEMO_COMMENTS_KEY, JSON.stringify(comments))
    } catch (e) {
      console.error('Failed to save demo comments:', e)
    }
  }

  const [{ data, fetching, error }, reexecuteRequest] = useQuery({
    query: RequestDocument,
    variables: { id },
    pause: appMode === 'demo' || !id,
  })

  const [{ data: meData, fetching: meFetching, error: meError }] = useQuery({
    query: MeDocument,
    pause: appMode === 'demo',
  })

  const [{ fetching: submitFetching }, submitRequest] = useMutation(
    SubmitRequestDocument,
  )
  const [{ fetching: approveFetching }, approveRequest] = useMutation(
    ApproveRequestDocument,
  )
  const [{ fetching: rejectFetching }, rejectRequest] = useMutation(
    RejectRequestDocument,
  )
  const [{ fetching: returnFetching }, returnRequest] = useMutation(
    ReturnRequestDocument,
  )
  const [{ fetching: resubmitFetching }, resubmitRequest] = useMutation(
    ResubmitRequestDocument,
  )

  const [actionError, setActionError] = useState<string | null>(null)

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

  // デモコメント追加ハンドラ
  const handleDemoAddComment = (content: string) => {
    const newComment: DemoComment = {
      id: `CMT-${Date.now()}`,
      requestID: id,
      userID: 'Demo User',
      content,
      createdAt: nowIso(),
    }
    const updated = [...demoComments, newComment]
    setDemoComments(updated)
    saveDemoComments(updated)
  }

  const setStatus = (status: RequestStatus, action: string, note?: string) => {
    setReq((prev) => {
      if (!prev) return prev
      return { ...prev, status }
    })
    pushAudit({ actor: 'Demo User', action, note })
  }

  if (appMode === 'prod') {
    const request = data?.request ?? null
    const me = meData?.me ?? null
    const role = me?.role ?? null
    const isAdmin = role === 'admin'
    const isCreator =
      !!me?.userID && !!request?.createdByUserID && request.createdByUserID === me.userID
    const status = request?.status ?? ''

    const canSubmitProd = status === 'draft' && isCreator
    const canApproveProd =
      (status === 'submitted' || status === 'resubmitted') && isAdmin
    const canRejectProd =
      (status === 'submitted' || status === 'resubmitted') && isAdmin
    const canReturnProd =
      (status === 'submitted' || status === 'resubmitted') && isAdmin
    const canResubmitProd = status === 'returned' && isCreator

    const busy =
      fetching ||
      meFetching ||
      submitFetching ||
      approveFetching ||
      rejectFetching ||
      returnFetching ||
      resubmitFetching

    return (
      <div className="stack">
        <section className="card">
          <h2>申請詳細（GraphQL）</h2>

          {fetching && <p>読み込み中...</p>}
          {error && 
            <ErrorBanner 
              message={error.message}
              action={
                <>
                  未ログインの場合は <a href="/auth/login">ログイン</a>{' '}
                  してください。
                </>
              }
            />
          }

          {meError && <p className="error">{meError.message}</p>}
          {actionError && <p className="error">{actionError}</p>}

          {!fetching && !error && !request && (
            <p className="error">申請が見つかりませんでした。</p>
          )}

          {request && (
            <div className="row">
              <div>
                <div className="kv">
                  <span className="kvKey">ID</span>
                  <span className="mono">{request.id}</span>
                </div>
                <div className="kv">
                  <span className="kvKey">タイトル</span>
                  <span>{request.title}</span>
                </div>
                <div className="kv">
                  <span className="kvKey">ステータス</span>
                  <span className={`badge badge--${request.status}`}>
                    {request.status}
                  </span>
                </div>
                <div className="kv">
                  <span className="kvKey">作成者</span>
                  <span className="mono">{request.createdByUserID ?? '-'}</span>
                </div>
                <div className="kv">
                  <span className="kvKey">更新日時</span>
                  <span className="mono">
                    {new Date(request.updatedAt).toLocaleString()}
                  </span>
                </div>
              </div>

              <div className="actions">
                <Link className="btn btn--ghost" to="/requests">
                  一覧へ戻る
                </Link>

                <button
                  className="btn"
                  disabled={!canSubmitProd || busy}
                  onClick={async () => {
                    if (!request) return
                    setActionError(null)
                    const res = await submitRequest({ id: request.id })
                    if (res.error) {
                      setActionError(res.error.message)
                      return
                    }
                    reexecuteRequest({ requestPolicy: 'network-only' })
                  }}
                >
                  提出
                </button>

                <button
                  className="btn btn--ok"
                  disabled={!canApproveProd || busy}
                  onClick={async () => {
                    if (!request) return
                    setActionError(null)
                    const res = await approveRequest({ id: request.id })
                    if (res.error) {
                      setActionError(res.error.message)
                      return
                    }
                    reexecuteRequest({ requestPolicy: 'network-only' })
                  }}
                >
                  承認
                </button>

                <button
                  className="btn btn--danger"
                  disabled={!canRejectProd || busy}
                  onClick={async () => {
                    if (!request) return
                    const reason = window.prompt('却下理由', '')
                    if (reason == null) return
                    if (reason.trim().length === 0) {
                      setActionError('却下理由を入力してください')
                      return
                    }
                    setActionError(null)
                    const res = await rejectRequest({
                      id: request.id,
                      reason: reason.trim(),
                    })
                    if (res.error) {
                      setActionError(res.error.message)
                      return
                    }
                    reexecuteRequest({ requestPolicy: 'network-only' })
                  }}
                >
                  却下
                </button>

                <button
                  className="btn btn--warn"
                  disabled={!canReturnProd || busy}
                  onClick={async () => {
                    if (!request) return
                    const reason = window.prompt('差し戻し理由', '')
                    if (reason == null) return
                    if (reason.trim().length === 0) {
                      setActionError('差し戻し理由を入力してください')
                      return
                    }
                    setActionError(null)
                    const res = await returnRequest({
                      id: request.id,
                      reason: reason.trim(),
                    })
                    if (res.error) {
                      setActionError(res.error.message)
                      return
                    }
                    reexecuteRequest({ requestPolicy: 'network-only' })
                  }}
                >
                  差し戻し
                </button>

                <button
                  className="btn"
                  disabled={!canResubmitProd || busy}
                  onClick={async () => {
                    if (!request) return
                    setActionError(null)
                    const res = await resubmitRequest({ id: request.id })
                    if (res.error) {
                      setActionError(res.error.message)
                      return
                    }
                    reexecuteRequest({ requestPolicy: 'network-only' })
                  }}
                >
                  再提出
                </button>
              </div>
            </div>
          )}

          {request && me && (
            <p className="note" style={{ marginTop: 10 }}>
              role: <span className="mono">{me.role}</span> / user:
              <span className="mono"> {me.userID}</span>
              {isCreator ? '（作成者）' : ''}
            </p>
          )}

          {request && request.steps.length > 0 && (
            <div className="cardSub">
              <h3>ステップ</h3>
              <div className="tableWrap">
                <table className="table">
                  <thead>
                    <tr>
                      <th>#</th>
                      <th>ラベル</th>
                      <th>ステータス</th>
                      <th>担当</th>
                      <th>更新日時</th>
                    </tr>
                  </thead>
                  <tbody>
                    {request.steps
                      .slice()
                      .sort((a, b) => a.stepIndex - b.stepIndex)
                      .map((s) => (
                        <tr key={s.stepIndex}>
                          <td className="mono">{s.stepIndex}</td>
                          <td>{s.label}</td>
                          <td>
                            <span className={`badge badge--${s.status}`}>
                              {s.status}
                            </span>
                          </td>
                          <td className="mono">{s.assignedToUserID ?? '-'}</td>
                          <td className="mono">
                            {new Date(s.updatedAt).toLocaleString()}
                          </td>
                        </tr>
                      ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </section>
      
        <section className="card">
          <h2>コメント</h2>
          <CommentList key={commentsKey} requestID={id} />
          <CommentForm requestID={id} onSuccess={() => {
            reexecuteRequest({ requestPolicy: 'network-only' })
            setCommentsKey(k => k + 1)
          }} />
        </section>

        {request && (
          <section className="card">
            <h2>監査ログ（GraphQL）</h2>
            <ol className="audit">
              {request.auditTrail
                .slice()
                .sort(
                  (a, b) =>
                    new Date(a.occurredAt).getTime() -
                    new Date(b.occurredAt).getTime(),
                )
                .map((a) => (
                  <li key={a.id} className="auditItem">
                    <div className="auditTop">
                      <span className="mono">
                        {new Date(a.occurredAt).toLocaleString()}
                      </span>
                      <span className="auditAction">{a.action}</span>
                      <span className="auditActor mono">
                        {a.actorUserID ?? '-'}
                      </span>
                    </div>
                    <pre className="mono" style={{ margin: '8px 0 0' }}>
                      {JSON.stringify(a.data, null, 2)}
                    </pre>
                  </li>
                ))}
            </ol>
          </section>
        )}
      </div>
    )
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
        <h2>コメント</h2>
        <CommentList 
          requestID={id} 
          demoComments={demoComments.filter(c => c.requestID === id)} 
        />
        <CommentForm 
          requestID={id} 
          onSuccess={() => {}} 
          onDemoSubmit={handleDemoAddComment}
        />
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
