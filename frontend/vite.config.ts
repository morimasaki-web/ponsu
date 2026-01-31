import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Back-end (Go) は /graphql をログイン必須で提供している。
// SPA開発時は Vite の dev server をフロントにして、/graphql をバックへプロキシする。
export default defineConfig({
  // GitHub Pages（project pages）等で /<repo>/ 配下に置く場合に assets が 404 になりやすい。
  // その場合は PONSU_FRONTEND_BASE="/<repo>/" を設定してビルドする。
  base: process.env.PONSU_FRONTEND_BASE ?? '/',
  plugins: [react()],
  server: {
    // `npm run dev` の表示は localhost になりがちだが、OIDCのcookie/stateはホスト名不一致で壊れやすい。
    // 127.0.0.1 でも開けるようにバインド先を広げておく。
    host: true,
    port: 5173,
    strictPort: true,
    proxy: {
      '/graphql': {
        target: process.env.PONSU_BACKEND_ORIGIN ?? 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/auth': {
        target: process.env.PONSU_BACKEND_ORIGIN ?? 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/playground': {
        target: process.env.PONSU_BACKEND_ORIGIN ?? 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
