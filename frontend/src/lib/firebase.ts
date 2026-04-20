/**
 * Firebase の初期化設定
 *
 * Firebase はGoogleが提供するバックエンドサービス。
 * ここでは Authentication（ユーザー認証）機能を使う。
 *
 * 公式ドキュメント: https://firebase.google.com/docs/web/setup
 *
 * 設定手順:
 *   1. Firebase コンソール (https://console.firebase.google.com/) にアクセス
 *   2. プロジェクト設定 > ウェブアプリ を選択
 *   3. 表示される設定値を .env ファイルに記入する
 */
import { initializeApp } from 'firebase/app'
import { getAuth } from 'firebase/auth'

// import.meta.env.VITE_xxx で .env ファイルの値を取得できる
// VITE_ プレフィックスが必要（Viteのセキュリティ仕様）
const firebaseConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
}

const app = initializeApp(firebaseConfig)
export const auth = getAuth(app)
