export type RequestStatus =
  | 'draft'
  | 'submitted'
  | 'approved'
  | 'rejected'
  | 'returned'
  | 'resubmitted'

export type RequestAudit = {
  at: string
  actor: string
  action: string
  note?: string
}

export type DemoComment = {
  id: string
  requestID: string
  userID: string
  content: string
  createdAt: string
}

export type DemoRequest = {
  id: string
  title: string
  status: RequestStatus
  createdAt: string
  updatedAt: string
  createdBy: string
  summary: string
  audit: RequestAudit[]
}

export function formatStatus(status: RequestStatus): string {
  switch (status) {
    case 'draft':
      return '下書き'
    case 'submitted':
      return '提出済み'
    case 'approved':
      return '承認'
    case 'rejected':
      return '却下'
    case 'returned':
      return '差し戻し'
    case 'resubmitted':
      return '再提出'
  }
}

export function nowIso(): string {
  return new Date().toISOString()
}

export const demoRequestsSeed: DemoRequest[] = [
  {
    id: 'REQ-0001',
    title: '備品購入申請（モニター）',
    status: 'submitted',
    createdAt: '2026-01-10T09:12:00.000Z',
    updatedAt: '2026-01-11T10:04:00.000Z',
    createdBy: 'Demo User',
    summary: '開発効率向上のため、27インチモニターを購入したい。',
    audit: [
      {
        at: '2026-01-10T09:12:00.000Z',
        actor: 'Demo User',
        action: '作成',
      },
      {
        at: '2026-01-11T10:04:00.000Z',
        actor: 'Demo User',
        action: '提出',
      },
    ],
  },
  {
    id: 'REQ-0002',
    title: '出張申請（大阪）',
    status: 'returned',
    createdAt: '2026-01-08T03:30:00.000Z',
    updatedAt: '2026-01-09T08:20:00.000Z',
    createdBy: 'Demo User',
    summary: '顧客訪問のため大阪へ出張。交通費と宿泊費の承認を得たい。',
    audit: [
      {
        at: '2026-01-08T03:30:00.000Z',
        actor: 'Demo User',
        action: '作成',
      },
      {
        at: '2026-01-08T04:15:00.000Z',
        actor: 'Demo User',
        action: '提出',
      },
      {
        at: '2026-01-09T08:20:00.000Z',
        actor: 'Approver (demo)',
        action: '差し戻し',
        note: '旅程と見積りの添付をお願いします',
      },
    ],
  },
  {
    id: 'REQ-0003',
    title: '稟議申請（SaaS導入）',
    status: 'approved',
    createdAt: '2026-01-05T12:00:00.000Z',
    updatedAt: '2026-01-06T13:00:00.000Z',
    createdBy: 'Demo User',
    summary: 'セキュリティ強化のため、SaaSの導入を稟議したい。',
    audit: [
      {
        at: '2026-01-05T12:00:00.000Z',
        actor: 'Demo User',
        action: '作成',
      },
      {
        at: '2026-01-05T12:40:00.000Z',
        actor: 'Demo User',
        action: '提出',
      },
      {
        at: '2026-01-06T13:00:00.000Z',
        actor: 'Approver (demo)',
        action: '承認',
      },
    ],
  },
]

export const demoCommentsSeed: DemoComment[] = [
  {
    id: 'CMT-0001',
    requestID: 'REQ-0001',
    userID: 'user-demo-001',
    content: '確認いたしました。承認をお願いします。',
    createdAt: '2026-01-11T10:05:00.000Z',
  },
  {
    id: 'CMT-0002',
    requestID: 'REQ-0002',
    userID: 'approver-demo',
    content: '旅程表と見積りの添付をお願いします。',
    createdAt: '2026-01-09T08:21:00.000Z',
  },
  {
    id: 'CMT-0003',
    requestID: 'REQ-0002',
    userID: 'user-demo-001',
    content: '承知しました。追加で添付いたします。',
    createdAt: '2026-01-09T09:30:00.000Z',
  },
  {
    id: 'CMT-0004',
    requestID: 'REQ-0003',
    userID: 'approver-demo',
    content: '承認しました。導入を進めてください。',
    createdAt: '2026-01-06T13:01:00.000Z',
  },
]
