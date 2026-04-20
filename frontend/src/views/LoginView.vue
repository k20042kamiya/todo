<script setup lang="ts">
/**
 * LoginView.vue - ログイン/ユーザー登録ページ
 *
 * ===== ref() でフォームの状態を管理する =====
 *
 * フォームの入力値やUI状態を ref() で管理する。
 * テンプレートで v-model="変数" を使うと、input要素と変数が双方向バインドされる。
 *
 * ===== v-model とは？ =====
 *
 * <input v-model="email" />
 * ↓ これは以下の省略記法:
 * <input :value="email" @input="email = $event.target.value" />
 *
 * つまり:
 *   - input の値として email の値を表示する
 *   - ユーザーが入力すると email の値が自動的に更新される
 *   - email の値が変わると input の表示も更新される
 *   → 「双方向バインディング」
 */
import { ref } from 'vue'
// TODO: useAuth と useRouter をインポートする
// import { useAuth } from '@/composables/useAuth'
// import { useRouter } from 'vue-router'

// ===== Composable と Router のインポート・初期化 =====
// useAuth() は先ほど実装した認証Composableで、login/register関数を提供します。
// useRouter() は Vue Router が提供するComposableで、プログラムからページ遷移するための
// router オブジェクトを返します。router.push('/') でTODOページに遷移できます。
import { useAuth } from '@/composables/useAuth'
import { useRouter } from 'vue-router'

// TODO: composable と router を初期化する
// const { login, register } = useAuth()
// const router = useRouter()

// ===== 分割代入（Destructuring）で必要な関数だけを取り出す =====
// useAuth() が返すオブジェクト { user, loading, login, register, logout } の中から、
// このコンポーネントで使う login と register だけを取り出しています。
// これを「分割代入」と呼びます。
const { login, register } = useAuth()
const router = useRouter()

// ---- フォームの状態 ----
// TODO: 以下の ref を使ってフォームの入力値を管理する
const email = ref('')
const password = ref('')
const errorMessage = ref('')
const isLoginMode = ref(true) // true: ログインモード、false: 登録モード

/**
 * フォーム送信処理
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   async function handleSubmit() {
 *     errorMessage.value = ''   // エラーメッセージをクリア
 *     try {
 *       if (isLoginMode.value) {
 *         await login(email.value, password.value)
 *       } else {
 *         await register(email.value, password.value)
 *       }
 *       // 成功したらTODOページに遷移
 *       router.push('/')
 *     } catch (error) {
 *       // エラーメッセージを表示
 *       errorMessage.value = 'ログインに失敗しました'
 *     }
 *   }
 */
async function handleSubmit() {
  // ===== フォーム送信の基本パターン =====
  // 1. まずエラーメッセージをクリアして前回のエラー表示を消す
  // 2. try ブロックで認証処理を実行
  // 3. isLoginMode の値に応じてログインか登録かを分岐
  // 4. 成功したら router.push('/') でTODOページに遷移
  // 5. 失敗したら catch でエラーメッセージを表示
  errorMessage.value = ''
  try {
    if (isLoginMode.value) {
      await login(email.value, password.value)
    } else {
      await register(email.value, password.value)
    }
    // ===== プログラムによるページ遷移 =====
    // router.push(パス) でプログラムからページを遷移させることができます。
    // ブラウザの戻るボタンでこの遷移は「戻る」ことができます。
    router.push('/')
  } catch (error) {
    errorMessage.value = 'ログインに失敗しました'
  }
}

/** ログイン/登録モードを切り替える */
function toggleMode() {
  isLoginMode.value = !isLoginMode.value
  errorMessage.value = ''
}
</script>

<template>
  <!--
    HTMLの基礎:
    <div>  → 汎用的なブロック要素（箱）。レイアウトのグループ分けに使う
    <h1>   → 見出し（h1が最大、h6が最小）
    <form> → フォーム。@submit.prevent でページリロードなしに送信処理できる
    <label> → 入力欄のラベル（何を入力するかの説明）
    <input> → テキスト入力欄
    <button> → ボタン
    <p>     → 段落（テキストのブロック）
    <span>  → インライン要素（テキストの一部をグループ化）

    class="xxx" → CSSのスタイルを適用するための名前
  -->
  <div class="login-page">
    <div class="login-card">
      <!-- TODO: isLoginMode に応じてタイトルを切り替える -->
      <!-- ヒント: {{ isLoginMode ? 'ログイン' : 'ユーザー登録' }} -->
      <!-- 三項演算子: 条件 ? trueの値 : falseの値 -->
      <!-- ===== テンプレート内での三項演算子 =====
           {{ }} の中にJavaScript式を書くと、その結果がHTMLに表示されます。
           三項演算子で isLoginMode が true なら 'ログイン'、false なら 'ユーザー登録' を表示します。 -->
      <h1 class="login-title">{{ isLoginMode ? 'ログイン' : 'ユーザー登録' }}</h1>

      <!-- TODO: エラーメッセージを条件付きで表示する -->
      <!-- ヒント: <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p> -->
      <!-- v-if="条件" → 条件が truthy の時だけ要素を描画する -->
      <!-- 空文字列 '' は falsy なので、エラーがない時は非表示になる -->
      <!-- ===== v-if による条件付きレンダリング =====
           v-if はVueのディレクティブで、条件が truthy の時だけDOM要素を描画します。
           errorMessage が空文字列 '' の場合は falsy なので、この要素は描画されません。
           エラーが発生して errorMessage に値が入ると、自動的に表示されます。 -->
      <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>

      <!-- TODO: @submit.prevent="handleSubmit" を <form> に追加する -->
      <!-- .prevent は event.preventDefault() の省略記法。 -->
      <!-- フォーム送信時のページリロードを防ぐ。 -->
      <!-- ===== @submit.prevent の仕組み =====
           @submit はフォーム送信イベントをリッスンするディレクティブです。
           .prevent 修飾子を付けると、ブラウザのデフォルト動作（ページリロード）を防止します。
           SPAではページを丸ごとリロードせず、JavaScriptで画面更新するため .prevent が必須です。 -->
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label class="form-label">メールアドレス</label>
          <!-- TODO: v-model="email" を追加して入力値とバインドする -->
          <!-- ===== v-model で双方向バインディング =====
               v-model="email" を追加することで、input の値と email 変数が同期されます。
               ユーザーがキーボードで入力すると email.value が更新され、
               email.value をプログラムで変更すると input の表示も更新されます。 -->
          <input
            v-model="email"
            type="email"
            class="form-input"
            placeholder="example@mail.com"
          />
        </div>

        <div class="form-group">
          <label class="form-label">パスワード</label>
          <!-- TODO: v-model="password" を追加 -->
          <input
            v-model="password"
            type="password"
            class="form-input"
            placeholder="パスワードを入力"
          />
        </div>

        <button type="submit" class="btn-login">
          <!-- TODO: isLoginMode に応じてボタンテキストを切り替える -->
          {{ isLoginMode ? 'ログイン' : '登録' }}
        </button>
      </form>

      <p class="toggle-text">
        <!-- TODO: isLoginMode に応じてテキストを切り替える -->
        <!-- ヒント: {{ isLoginMode ? 'アカウントをお持ちでない方' : 'すでにアカウントをお持ちの方' }} -->
        {{ isLoginMode ? 'アカウントをお持ちでない方' : 'すでにアカウントをお持ちの方' }}
        <button class="toggle-link" @click="toggleMode">
          <!-- TODO: isLoginMode に応じてリンクテキストを切り替える -->
          {{ isLoginMode ? '新規登録はこちら' : 'ログインはこちら' }}
        </button>
      </p>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;     /* 縦方向の中央寄せ */
  justify-content: center; /* 横方向の中央寄せ */
  min-height: 80vh;        /* ビューポート高さの80% */
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
