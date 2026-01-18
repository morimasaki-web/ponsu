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
    port: 5173,
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
    },
  },
})
