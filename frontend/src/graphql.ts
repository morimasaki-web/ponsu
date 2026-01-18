export type Me = {
  userID: string
  orgID: string
  role: string
  name?: string | null
  email?: string | null
}

export type AppMode = 'demo' | 'prod'

export const appMode: AppMode =
  (import.meta.env.VITE_APP_MODE as AppMode | undefined) ?? 'prod'

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export async function fetchMe(): Promise<Me> {
  if (appMode === 'demo') {
    // ポートフォリオ用デモ: 認証/DB/課金要素を一切呼ばず、画面の動きだけ確認する。
    await sleep(250)
    return {
      userID: 'demo-user-0001',
      orgID: 'demo-org-0001',
      role: 'admin',
      name: 'Demo User',
      email: 'demo@example.invalid',
    }
  }

  const res = await fetch('/graphql', {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    // OIDCログインはCookieベース想定のため、同一オリジン/proxy経由でcookieを送る
    credentials: 'include',
    body: JSON.stringify({
      query: `query Me { me { userID orgID role name email } }`,
    }),
  })

  if (!res.ok) {
    // 401のときは /auth/login への導線を出す想定
    throw new Error(`GraphQL HTTP error: ${res.status}`)
  }

  const json = await res.json()
  if (json.errors?.length) {
    throw new Error(json.errors[0]?.message ?? 'GraphQL error')
  }
  return json.data.me as Me
}
