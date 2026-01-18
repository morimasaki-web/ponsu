import { Link } from 'react-router-dom'
import { demoRequestsSeed, formatStatus } from '../demoData'

export default function RequestsList() {
  return (
    <section className="card">
      <h2>申請一覧（デモ）</h2>
      <p className="note">
        これは静的デモです。データはブラウザ内の一時状態で、サーバやDBには保存されません。
      </p>

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
            {demoRequestsSeed.map((r) => (
              <tr key={r.id}>
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
