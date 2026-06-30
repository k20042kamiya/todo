<script setup lang="ts">
import { ref } from 'vue'
import { useAuth } from '@/composables/useAuth'
import { useRouter } from 'vue-router'

const { login, register } = useAuth()
const router = useRouter()

const email = ref('')
const password = ref('')
const errorMessage = ref('')
const isLoginMode = ref(true)

async function handleSubmit() {
  errorMessage.value = ''
  try {
    if (isLoginMode.value) {
      await login(email.value, password.value)
    } else {
      await register(email.value, password.value)
    }
    router.push('/')
  } catch (error) {
    errorMessage.value = 'ログインに失敗しました'
  }
}

function toggleMode() {
  isLoginMode.value = !isLoginMode.value
  errorMessage.value = ''
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="login-title">{{ isLoginMode ? 'ログイン' : 'ユーザー登録' }}</h1>

      <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>

      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label class="form-label">メールアドレス</label>
          <input
            v-model="email"
            type="email"
            class="form-input"
            placeholder="example@mail.com"
          />
        </div>

        <div class="form-group">
          <label class="form-label">パスワード</label>
          <input
            v-model="password"
            type="password"
            class="form-input"
            placeholder="パスワードを入力"
          />
        </div>

        <button type="submit" class="btn-login">
          {{ isLoginMode ? 'ログイン' : '登録' }}
        </button>
      </form>

      <p class="toggle-text">
        {{ isLoginMode ? 'アカウントをお持ちでない方' : 'すでにアカウントをお持ちの方' }}
        <button class="toggle-link" @click="toggleMode">
          {{ isLoginMode ? '新規登録はこちら' : 'ログインはこちら' }}
        </button>
      </p>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 80vh;
}

.login-card {
  background: white;
  border-radius: 16px;
  padding: 40px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.login-title {
  font-size: 24px;
  font-weight: 700;
  text-align: center;
  margin-bottom: 32px;
}

.error-message {
  color: #e86c50;
  font-size: 13px;
  margin-bottom: 16px;
  text-align: center;
}

.form-group {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #666;
  margin-bottom: 6px;
}

.form-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus {
  border-color: #e86c50;
}

.btn-login {
  width: 100%;
  padding: 12px;
  border: none;
  background-color: #e86c50;
  color: white;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  margin-bottom: 16px;
}

.btn-login:hover {
  background-color: #d55a40;
}

.toggle-text {
  text-align: center;
  font-size: 13px;
  color: #999;
}

.toggle-link {
  color: #e86c50;
  cursor: pointer;
  background: none;
  border: none;
  font-size: 13px;
  text-decoration: underline;
}
</style>
