import { firebaseAuthClient } from '@/lib/firebaseAuthClient'
import { mockAuthClient } from '@/lib/mockAuthClient'

// アプリが必要とする認証ユーザーの最小限の形（firebase/auth の User はこれを満たす）
export interface AuthUser {
  uid: string
  email: string | null
  getIdToken(): Promise<string>
}

export interface AuthClient {
  onAuthChanged(callback: (user: AuthUser | null) => void): void
  getCurrentUser(): AuthUser | null
  login(email: string, password: string): Promise<void>
  register(email: string, password: string): Promise<void>
  logout(): Promise<void>
}

// VITE_USE_AUTH_MOCK=true のときはFirebaseに接続せずモック認証を使う（開発ビルド限定）
// import.meta.env.DEV の判定により、本番ビルドでは環境変数を誤設定してもモックは有効にならず、
// モック実装のコードもバンドルから除外される
export const authClient: AuthClient =
  import.meta.env.DEV && import.meta.env.VITE_USE_AUTH_MOCK === 'true'
    ? mockAuthClient
    : firebaseAuthClient
