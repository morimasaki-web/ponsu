import { Link } from 'react-router-dom'
import { demoRequestsSeed, formatStatus } from '../demoData'
import { appMode } from '../graphql'
import { useQuery } from 'urql'
import { RequestsDocument, SearchRequestsQuery } from '../gql/graphql'
import { ErrorBanner } from '../components/ErrorBanner'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { RequestSearchBar } from '../components/RequestSearchBar'
import { useState } from 'react'

export default function RequestsList() {
  useDocumentTitle('Requests List | Ponsu')

  // プロダクションモード: GraphQL検索結果を管理
  const [searchResults, setSearchResults] = useState<SearchRequestsQuery['searchRequests'] | null>(null)

  // デモモード: フィルタリングされた申請一覧を管理
  const [filteredDemoRequests, setFilteredDemoRequests] = useState<typeof demoRequestsSeed>(demoRequestsSeed)

  // デフォルトの全件取得（検索していない時用）
  const [{ data, fetching, error }] = useQuery({
    query: RequestsDocument,
    variables: { limit: 50, offset: 0 },
    pause: appMode === 'demo',
  })

  // 表示するデータ：検索結果があればそれを、なければ全件を表示
  const displayData = searchResults !== null ? searchResults : data?.requests || []

  if (appMode === 'prod') {
    return (
      <section className="card">
        <h2>申請一覧（GraphQL）</h2>
        <div className="actions" style={{ marginBottom: 10 }}>
          <Link className="btn" to="/requests/new">
            新規作成
          </Link>
        </div>

        <RequestSearchBar onSearchResultsChange={setSearchResults} />

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
                  <th>タイトル</th>
                  <th>ステータス</th>
                  <th>更新日時</th>
                </tr>
              </thead>
              <tbody>
                {displayData.map((r) => (
                  <tr key={r.id}>
                    <td className="mono">{r.id}</td>
                    <td>
                      <Link to={`/requests/${encodeURIComponent(r.id)}`}>
                        {r.title}
                      </Link>
                    </td>
                    <td>
                      <span className={`badge badge--${r.status}`}>
                        {r.status}
                      </span>
                    </td>
                    <td className="mono">
                      {new Date(r.updatedAt).toLocaleString()}
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
      <h2>申請一覧（デモ）</h2>
      <p className="note">
        これは静的デモです。データはブラウザ内の一時状態で、サーバやDBには保存されません。
      </p>

      <RequestSearchBar 
        demoData={demoRequestsSeed}
        onDemoFilteredChange={setFilteredDemoRequests}
      />

      <div className="tableWrap">
        <table className="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>タイトル</th>
              <th>ステータス</th>
              <th>更新日時</th>
            </tr>
          </thead>
          <tbody>
            {filteredDemoRequests.map((r) => (
              <tr key={r.id}>)
                <td className="mono">{r.id}</td>
                <td>
                  <Link to={`/requests/${encodeURIComponent(r.id)}`}>
                    {r.title}
                  </Link>
                </td>
                <td>
                  <span className={`badge badge--${r.status}`}>
                    {formatStatus(r.status)}
                  </span>
                </td>
                <td className="mono">{new Date(r.updatedAt).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
