import { useEffect, useMemo, useState } from 'react'
import { HashRouter, Link, Route, Routes } from 'react-router-dom'
import { appMode, fetchMe, type Me } from './graphql'
import { createGraphqlClient } from './gql/client'
import { applyTheme, getStoredTheme, type ThemeMode } from './theme'
import RequestDetail from './pages/RequestDetail'
import RequestNew from './pages/RequestNew'
import RequestsList from './pages/RequestsList'
import { Provider } from 'urql'

export default function App() {
  const graphqlClient = useMemo(() => createGraphqlClient(), [])
  const [me, setMe] = useState<Me | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [theme, setTheme] = useState<ThemeMode>(() => getStoredTheme())

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    fetchMe()
      .then((v) => {
        if (cancelled) return
        setMe(v)
        setError(null)
      })
      .catch((e) => {
        if (cancelled) return
        setError(String(e?.message ?? e))
        setMe(null)
      })
      .finally(() => {
        if (cancelled) return
        setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  const home = (
    <>
      <section className="card">
        <h2>ログイン情報</h2>
        {loading && <p>読み込み中...</p>}
        {error && (
          <div>
            <p className="error">{error}</p>
            {appMode === 'prod' && (
              <p>
                未ログインの場合は <a href="/auth/login">ログイン</a> してください。
              </p>
            )}
          </div>
        )}
        {me && (
          <dl className="dl">
            <dt>userID</dt>
            <dd>{me.userID}</dd>
            <dt>orgID</dt>
            <dd>{me.orgID}</dd>
            <dt>role</dt>
            <dd>{me.role}</dd>
            <dt>name</dt>
            <dd>{me.name ?? '-'}</dd>
            <dt>email</dt>
            <dd>{me.email ?? '-'}</dd>
          </dl>
        )}
      </section>

      <section className="card">
        <h2>次のステップ</h2>
        <ul>
          <li>デモ: 画面遷移と操作フローをモックデータで再現</li>
          <li>正規版: GraphQLクライアント導入＋型付け</li>
          <li>正規版: Requests一覧/詳細/監査のSPA画面と操作</li>
        </ul>
      </section>
    </>
  )

  return (
    <Provider value={graphqlClient}>
      <HashRouter>
        <div className="container">
          <header className="header">
            <div>
              <h1 className="title">PonSu Demo (SPA)</h1>
              <p className="subtitle">mode: {appMode}</p>
            </div>

            <div className="headerRight">
              <nav className="nav">
                <Link to="/">Home</Link>
                <Link to="/requests">Requests</Link>
              </nav>

              <label className="themeLabel">
                Theme
                <select
                  className="select"
                  value={theme}
                  onChange={(e) => {
                    const v = e.target.value as ThemeMode
                    setTheme(v)
                    applyTheme(v)
                  }}
                >
                  <option value="system">system</option>
                  <option value="light">light</option>
                  <option value="dark">dark</option>
                </select>
              </label>

              {appMode === 'prod' ? (
                <nav className="nav">
                  <a href="/auth/login">ログイン</a>
                  <a href="/auth/logout">ログアウト</a>
                  <a href="/playground">GraphQL Playground</a>
                </nav>
              ) : (
                <p className="note">デモは認証/DBなし（静的）</p>
              )}
            </div>
          </header>

          <Routes>
            <Route path="/" element={home} />
            <Route path="/requests" element={<RequestsList />} />
            <Route path="/requests/new" element={<RequestNew />} />
            <Route path="/requests/:id" element={<RequestDetail />} />
          </Routes>
        </div>
      </HashRouter>
    </Provider>
  )
}
