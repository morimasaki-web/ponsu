import { useQuery } from 'urql'
import {
  DashboardSummaryDocument,
  CountRequestsByStatusDocument,
  CountRequestsByMonthDocument,
  AvgTimeToApprovalDocument,
} from '../gql/graphql'
import { appMode } from '../graphql'
import { ErrorBanner } from '../components/ErrorBanner'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import {
  PieChart,
  Pie,
  Cell,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'

// ステータスごとの色を定義
const STATUS_COLORS: Record<string, string> = {
  draft: '#808080',
  submitted: '#3b82f6',
  approved: '#10b981',
  rejected: '#ef4444',
  returned: '#f59e0b',
}

// デモモード用のモックデータ
const demoSummary = {
  draftCount: 3,
  submittedCount: 5,
  approvedCount: 12,
  rejectedCount: 2,
  totalCount: 22,
  avgApprovalSeconds: 86400, // 1日
}

const demoStatusData = [
  { status: 'draft', count: 3 },
  { status: 'submitted', count: 5 },
  { status: 'approved', count: 12 },
  { status: 'rejected', count: 2 },
]

const demoMonthData = [
  { month: '2026-01-01T00:00:00Z', count: 8 },
  { month: '2026-02-01T00:00:00Z', count: 14 },
]

export default function Dashboard() {
  useDocumentTitle('Dashboard | Ponsu')

  // GraphQLクエリ（デモモードでは実行しない）
  const [summaryResult] = useQuery({
    query: DashboardSummaryDocument,
    pause: appMode === 'demo',
  })

  const [statusResult] = useQuery({
    query: CountRequestsByStatusDocument,
    pause: appMode === 'demo',
  })

  const [monthResult] = useQuery({
    query: CountRequestsByMonthDocument,
    pause: appMode === 'demo',
  })

  const [avgResult] = useQuery({
    query: AvgTimeToApprovalDocument,
    pause: appMode === 'demo',
  })

  // データの取得（デモモードとプロダクションモードで切り替え）
  const summary = appMode === 'demo' ? demoSummary : summaryResult.data?.dashboardSummary
  const statusData = appMode === 'demo' ? demoStatusData : statusResult.data?.countRequestsByStatus || []
  const monthData = appMode === 'demo' ? demoMonthData : monthResult.data?.countRequestsByMonth || []
  const avgData = appMode === 'demo' 
    ? { avgSeconds: 86400, sampleCount: 12 } 
    : avgResult.data?.avgTimeToApproval

  // ローディング状態
  const isLoading = appMode === 'prod' && (
    summaryResult.fetching || statusResult.fetching || monthResult.fetching || avgResult.fetching
  )

  // エラー状態
  const error = summaryResult.error || statusResult.error || monthResult.error || avgResult.error

  // 平均承認時間を人間が読める形式に変換
  const formatApprovalTime = (seconds: number): string => {
    if (seconds === 0) return '0秒'
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    
    const parts: string[] = []
    if (days > 0) parts.push(`${days}日`)
    if (hours > 0) parts.push(`${hours}時間`)
    if (minutes > 0) parts.push(`${minutes}分`)
    
    return parts.join(' ') || '1分未満'
  }

  // ステータス名を日本語に変換
  const formatStatusLabel = (status: string): string => {
    const labels: Record<string, string> = {
      draft: '下書き',
      submitted: '提出済み',
      approved: '承認済み',
      rejected: '却下',
      returned: '差し戻し',
    }
    return labels[status] || status
  }

  // 月のフォーマット（2026-01 形式に）
  const formatMonth = (dateString: string): string => {
    const date = new Date(dateString)
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
  }

  return (
    <section className="card">
      <h2>ダッシュボード</h2>

      {appMode === 'demo' && (
        <div className="info-banner" style={{ marginBottom: 20 }}>
          📊 デモモード: サンプルデータを表示しています
        </div>
      )}

      {isLoading && <p>読み込み中...</p>}

      {error && (
        <ErrorBanner
          message={error.message}
          action={
            <>
              未ログインの場合は <a href="/auth/login">ログイン</a> してください。
            </>
          }
        />
      )}

      {summary && (
        <>
          {/* サマリー数値 */}
          <div style={{ 
            display: 'grid', 
            gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', 
            gap: 20, 
            marginBottom: 40 
          }}>
            <div className="stat-card">
              <div className="stat-label">総申請数</div>
              <div className="stat-value">{summary.totalCount}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">下書き</div>
              <div className="stat-value" style={{ color: STATUS_COLORS.draft }}>
                {summary.draftCount}
              </div>
            </div>
            <div className="stat-card">
              <div className="stat-label">提出済み</div>
              <div className="stat-value" style={{ color: STATUS_COLORS.submitted }}>
                {summary.submittedCount}
              </div>
            </div>
            <div className="stat-card">
              <div className="stat-label">承認済み</div>
              <div className="stat-value" style={{ color: STATUS_COLORS.approved }}>
                {summary.approvedCount}
              </div>
            </div>
            <div className="stat-card">
              <div className="stat-label">却下</div>
              <div className="stat-value" style={{ color: STATUS_COLORS.rejected }}>
                {summary.rejectedCount}
              </div>
            </div>
            <div className="stat-card">
              <div className="stat-label">平均承認時間</div>
              <div className="stat-value" style={{ fontSize: '1.2rem' }}>
                {formatApprovalTime(summary.avgApprovalSeconds)}
              </div>
              {avgData && avgData.sampleCount > 0 && (
                <div style={{ fontSize: '0.8rem', color: '#666', marginTop: 4 }}>
                  （{avgData.sampleCount}件のサンプル）
                </div>
              )}
            </div>
          </div>

          {/* ステータス別の円グラフ */}
          <div style={{ marginBottom: 40 }}>
            <h3>ステータス別申請数</h3>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={statusData}
                  dataKey="count"
                  nameKey="status"
                  cx="50%"
                  cy="50%"
                  outerRadius={80}
                  label={(entry: any) => `${formatStatusLabel(entry.status)}: ${entry.count}`}
                >
                  {statusData.map((entry) => (
                    <Cell 
                      key={`cell-${entry.status}`} 
                      fill={STATUS_COLORS[entry.status] || '#999'} 
                    />
                  ))}
                </Pie>
                <Tooltip 
                  formatter={(value: number | undefined, name: string | undefined) => {
                    if (value === undefined || name === undefined) return ['', '']
                    return [value, formatStatusLabel(name)]
                  }}
                />
                <Legend 
                  formatter={(value: string) => formatStatusLabel(value)}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>

          {/* 月別申請数の折れ線グラフ */}
          <div>
            <h3>月別申請数推移</h3>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={monthData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="month" 
                  tickFormatter={formatMonth}
                />
                <YAxis />
                <Tooltip 
                  labelFormatter={(label: any) => {
                    if (typeof label === 'string') return formatMonth(label)
                    return String(label)
                  }}
                  formatter={(value: number | undefined) => {
                    if (value === undefined) return ['', '']
                    return [value, '申請数']
                  }}
                />
                <Legend formatter={() => '申請数'} />
                <Line 
                  type="monotone" 
                  dataKey="count" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  dot={{ r: 4 }}
                  activeDot={{ r: 6 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </>
      )}
    </section>
  )
}
