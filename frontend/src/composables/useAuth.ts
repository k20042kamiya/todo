import { shallowRef } from 'vue'
import { authClient, type AuthUser } from '@/lib/authClient'

// userは外部SDKのオブジェクトを丸ごと差し替えるだけなのでshallowRefで十分
const user = shallowRef<AuthUser | null>(null)
const loading = shallowRef(true)
// 意図しないログアウト(トークン失効など)を検知したときに立てるフラグ。ダイアログ表示に使う
const sessionExpired = shallowRef(false)
let intentionalLogout = false

authClient.onAuthChanged((authUser) => {
  const hadUser = user.value !== null
  user.value = authUser
  loading.value = false

  if (hadUser && !authUser && !intentionalLogout) {
    sessionExpired.value = true
  }
  intentionalLogout = false
})

export function useAuth() {
  async function login(email: string, password: string): Promise<void> {
    await authClient.login(email, password)
  }

  async function register(email: string, password: string): Promise<void> {
    await authClient.register(email, password)
  }

  async function logout(): Promise<void> {
    intentionalLogout = true
    try {
      await authClient.logout()
    } catch (error) {
      intentionalLogout = false
      throw error
    }
  }

  return {
    user,
    loading,
    sessionExpired,
    login,
    register,
    logout,
  }
}
