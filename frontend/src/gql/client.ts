import { cacheExchange, createClient, fetchExchange, type Client } from 'urql'
import { appMode } from '../graphql'

export function createGraphqlClient(): Client {
  // demoモードはネットワークアクセスしない方針だが、Provider自体はあって問題ない。
  // 実際のクエリ発行は各ページ側で appMode を見て制御する。
  const url = '/graphql'

  return createClient({
    url,
    exchanges: [cacheExchange, fetchExchange],
    // OIDCログインはCookieベース想定のため、同一オリジン or dev-proxy 経由で cookie を送る。
    fetchOptions: () => ({
      credentials: 'include',
      headers: {
        'content-type': 'application/json',
        'x-ponsu-app-mode': appMode,
      },
    }),
  })
}
