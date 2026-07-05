import { createRouter, createWebHistory } from 'vue-router'
import { watch } from 'vue'
import TodoView from '@/views/TodoView.vue'
import LoginView from '@/views/LoginView.vue'
import { useAuth } from '@/composables/useAuth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'todo',
      component: TodoView,
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
    },
  ],
})

const { user, loading } = useAuth()

// ログアウトや認証切れでユーザーがいなくなったら、どの画面にいてもログイン画面へ遷移する
watch(user, (currentUser) => {
  if (!currentUser && !loading.value && router.currentRoute.value.name !== 'login') {
    router.push({ name: 'login' })
  }
})

router.beforeEach(async (to) => {
  if (to.name === 'login') {
    return
  }

  if (loading.value) {
    await new Promise<void>((resolve) => {
      const unwatch = watch(loading, (val) => {
        if (!val) {
          unwatch()  // メモリリーク防止
          resolve()
        }
      })
    })
  }

  if (!user.value) {
    return { name: 'login' }
  }
})

export default router
