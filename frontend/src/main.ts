/**
 * main.ts - アプリケーションのエントリーポイント
 *
 * Vueアプリの初期化を行う:
 * 1. createApp()  - Vueアプリケーションインスタンスを作成
 * 2. app.use()    - プラグイン（Router等）を登録
 * 3. app.mount()  - 指定のDOM要素にアプリをマウント（描画開始）
 */
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './assets/main.css'

const app = createApp(App)
app.use(router)
app.mount('#app')
