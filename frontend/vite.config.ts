import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      // '@/〜' で src/ ディレクトリからの絶対パスインポートが使える
      // 例: import Foo from '@/components/Foo.vue'
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    // 開発中のAPIリクエストをバックエンド(localhost:8080)にプロキシする
    // これにより、CORS問題を回避できる
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
