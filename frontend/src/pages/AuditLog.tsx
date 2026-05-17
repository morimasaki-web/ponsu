import { AuditSearchBar } from '../components/AuditSearchBar'
import { AuditLogsDocument, AuditLogsQuery } from '../gql/graphql'
import { Link } from 'react-router-dom'
import { ErrorBanner } from '../components/ErrorBanner'
import { appMode } from '../graphql'
import { useQuery } from 'urql'
import { useState } from 'react'
import { demoAuditLogsSeed, DemoAuditLog } from '../demoData'

/**
 * アクション名を日本語に変換
 */
function getActionLabel(action: string): string {
  const actionMap: Record<string, string> = {
    'create': '作成',
    'submit': '提出',
    'approve': '承認',
    'reject': '却下',
    'return': '差し戻し',
    'resubmit': '再提出',
    'request.created': '申請作成',
    'request.submitted': '申請提出',
    'request.approved': '申請承認',
    'request.rejected': '申請却下',
    'request.returned': '申請差し戻し',
    'request.resubmitted': '申請再提出',
  }
  return actionMap[action] || action
}

/**
 * アクション名をCSS安全なクラス名に変換
 */
function getActionClass(action: string): string {
  // "request.submitted" → "request-submitted"
  return action.replace(/\./g, '-')
}

export default function AuditLog() {
  const [searchResults, setSearchResults] = useState<AuditLogsQuery['auditLogs'] | null>(null)

  // デフォルトの全件取得（検索していない時用）
  const [{ data, fetching, error }] = useQuery({
    query: AuditLogsDocument,
    variables: { 
      requestID: null, 
      actorUserID: null, 
      action: null,
      occurredAtStart: null,
      occurredAtEnd: null,
      limit: 50, 
      offset: 0 
    },
    pause: appMode === 'demo',
  })

  // 表示するデータ：検索結果があればそれを、なければ全件を表示
  const displayData = searchResults !== null ? searchResults : data?.auditLogs || []

  if (appMode === 'prod') {
      return (
        <section className="card">
          <h2>監査ログ</h2>

          <AuditSearchBar onSearchResultsChange={setSearchResults} />

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
          {displayData.length > 0 && (
            <div className="tableWrap">
              <table className="table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>ユーザー</th>
                    <th>アクション</th>
                    <th>日時</th>
                  </tr>
                </thead>
                <tbody>
                  {displayData.map((log) => (
                    <tr key={log.id}>
                      <td className="mono" style={{ fontSize: '0.85em' }}>{log.id.slice(0, 8)}...</td>
                      <td className="mono" style={{ fontSize: '0.85em' }}>
                        {log.actorUserID ? `${log.actorUserID.slice(0, 8)}...` : '(system)'}
                      </td>
                      <td>
                        <span className={`badge badge--${getActionClass(log.action)}`}>
                          {getActionLabel(log.action)}
                        </span>
                      </td>
                      <td className="mono">
                        {new Date(log.occurredAt).toLocaleString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )
    }

    return (
      <section className="card">
        <h2>監査ログ（デモ）</h2>

        <AuditSearchBar 
          onSearchResultsChange={setSearchResults} 
          demoData={demoAuditLogsSeed}
        />

        {displayData.length > 0 && (
          <div className="tableWrap">
            <table className="table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>ユーザー</th>
                  <th>アクション</th>
                  <th>日時</th>
                </tr>
              </thead>
              <tbody>
                {displayData.map((log) => (
                  <tr key={log.id}>
                    <td className="mono" style={{ fontSize: '0.85em' }}>{log.id}</td>
                    <td className="mono" style={{ fontSize: '0.85em' }}>
                      {log.actorUserID ? log.actorUserID : '(system)'}
                    </td>
                    <td>
                      <span className={`badge badge--${getActionClass(log.action)}`}>
                        {getActionLabel(log.action)}
                      </span>
                    </td>
                    <td className="mono">
                      {new Date(log.occurredAt).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        
        {displayData.length === 0 && (
          <p className="note">監査ログが見つかりませんでした。</p>
        )}
      </section>
    )
  }