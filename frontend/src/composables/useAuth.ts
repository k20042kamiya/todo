import { ref } from 'vue'
import type { User } from 'firebase/auth'
import { auth } from '@/lib/firebase'
import {
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signOut,
  onAuthStateChanged,
} from 'firebase/auth'

const user = ref<User | null>(null)
const loading = ref(true)

onAuthStateChanged(auth, (firebaseUser) => {
  user.value = firebaseUser
  loading.value = false
})

export function useAuth() {
  async function login(email: string, password: string): Promise<void> {
    await signInWithEmailAndPassword(auth, email, password)
  }

  async function register(email: string, password: string): Promise<void> {
    await createUserWithEmailAndPassword(auth, email, password)
  }

  async function logout(): Promise<void> {
    await signOut(auth)
  }

  return {
    user,
    loading,
    login,
    register,
    logout,
  }
}
