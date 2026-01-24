import { useEffect, useMemo, useState } from 'react'
import { HashRouter, Link, Route, Routes } from 'react-router-dom'
import { appMode, fetchMe, type Me } from './graphql'
import { createGraphqlClient } from './gql/client'
import { applyTheme, getStoredTheme, type ThemeMode } from './theme'
import About from './pages/About'
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
  const [demoResetNonce, setDemoResetNonce] = useState(0)

  const resetDemo = () => {
    if (appMode !== 'demo') return
    const ok = window.confirm('デモの状態を初期化しますか？')
    if (!ok) return
    setDemoResetNonce((v) => v + 1)
    window.location.hash = '#/'
  }

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
        <h2>デモガイド</h2>
        <p className="note">
          デモでは Requests の一覧→詳細→操作→監査ログ追記の流れを、モックデータで体験できます。
        </p>
        <ul>
          <li>
            <Link to="/requests">Requests</Link> を開く
          </li>
          <li>任意の申請の詳細へ</li>
          <li>提出→承認/却下/差し戻し→再提出（状態遷移を確認）</li>
        </ul>
        <p style={{ marginBottom: 0 }}>
          <Link to="/about">ガイド（詳しい説明）を見る</Link>
        </p>
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
                <Link to="/about">Guide</Link>
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
                <div className="actions">
                  <button className="btn btn--ghost" onClick={resetDemo}>
                    デモを初期化
                  </button>
                  <p className="note">デモは認証/DBなし（静的）</p>
                </div>
              )}
            </div>
          </header>

          <Routes key={demoResetNonce}>
            <Route path="/" element={home} />
            <Route path="/requests" element={<RequestsList />} />
            <Route path="/requests/new" element={<RequestNew />} />
            <Route path="/requests/:id" element={<RequestDetail />} />
            <Route path="/about" element={<About onResetDemo={resetDemo} />} />
          </Routes>
        </div>
      </HashRouter>
    </Provider>
  )
}
