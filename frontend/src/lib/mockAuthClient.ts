import type { AuthClient, AuthUser } from '@/lib/authClient'

// 開発用のモック認証。Firebaseに接続せず、ブラウザ内だけでログイン状態を管理する。
// 登録ユーザーとログイン状態は localStorage に保存されるためリロード後も維持される。
// 初期ユーザー: test@example.com / password123

const USERS_KEY = 'mock-auth-users'
const SESSION_KEY = 'mock-auth-session'

const DEFAULT_USERS: Record<string, string> = {
  'test@example.com': 'password123',
}

function loadUsers(): Record<string, string> {
  const saved = localStorage.getItem(USERS_KEY)
  return saved ? JSON.parse(saved) : { ...DEFAULT_USERS }
}

function saveUsers(users: Record<string, string>): void {
  localStorage.setItem(USERS_KEY, JSON.stringify(users))
}

function toAuthUser(email: string): AuthUser {
  return {
    uid: `mock-${email}`,
    email,
    getIdToken: async () => 'mock-id-token',
  }
}

// 通信の体感に近づけるための擬似遅延
function delay(ms = 400): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

let currentUser: AuthUser | null = null
let sessionRestored = false
const listeners: Array<(user: AuthUser | null) => void> = []

// モジュール読み込み時に副作用を持たせない（本番ビルドでのtree-shakingを妨げないため）よう、
// 保存済みセッションの復元は初回アクセス時に行う
function restoreSession(): void {
  if (sessionRestored) return
  sessionRestored = true
  const savedEmail = localStorage.getItem(SESSION_KEY)
  if (savedEmail) {
    currentUser = toAuthUser(savedEmail)
  }
}

function setCurrentUser(user: AuthUser | null): void {
  currentUser = user
  if (user?.email) {
    localStorage.setItem(SESSION_KEY, user.email)
  } else {
    localStorage.removeItem(SESSION_KEY)
  }
  listeners.forEach((cb) => cb(user))
}

export const mockAuthClient: AuthClient = {
  onAuthChanged(callback) {
    restoreSession()
    listeners.push(callback)
    // Firebase同様、購読開始時に現在の状態を非同期で通知する
    queueMicrotask(() => callback(currentUser))
  },

  getCurrentUser() {
    restoreSession()
    return currentUser
  },

  async login(email, password) {
    await delay()
    const users = loadUsers()
    if (users[email] !== password) {
      throw new Error('auth/invalid-credential')
    }
    setCurrentUser(toAuthUser(email))
  },

  async register(email, password) {
    await delay()
    const users = loadUsers()
    if (email in users) {
      throw new Error('auth/email-already-in-use')
    }
    users[email] = password
    saveUsers(users)
    setCurrentUser(toAuthUser(email))
  },

  async logout() {
    setCurrentUser(null)
  },
}
