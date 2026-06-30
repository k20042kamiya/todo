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

router.beforeEach(async (to, from, next) => {
  const { user, loading } = useAuth()

  if (to.name === 'login') {
    next()
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
    next({ name: 'login' })
  } else {
    next()
  }
})

export default router
