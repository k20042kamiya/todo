import { initializeApp } from 'firebase/app'
import {
  getAuth,
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signOut,
  onAuthStateChanged,
  type Auth,
} from 'firebase/auth'
import type { AuthClient } from '@/lib/authClient'

let authInstance: Auth | null = null

// モックモード時にFirebaseを初期化しないよう、初回利用時に初期化する
function getAuthInstance(): Auth {
  if (!authInstance) {
    const app = initializeApp({
      apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
      authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
      projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
    })
    authInstance = getAuth(app)
  }
  return authInstance
}

export const firebaseAuthClient: AuthClient = {
  onAuthChanged(callback) {
    onAuthStateChanged(getAuthInstance(), callback)
  },

  getCurrentUser() {
    return getAuthInstance().currentUser
  },

  async login(email, password) {
    await signInWithEmailAndPassword(getAuthInstance(), email, password)
  },

  async register(email, password) {
    await createUserWithEmailAndPassword(getAuthInstance(), email, password)
  },

  async logout() {
    await signOut(getAuthInstance())
  },
}
