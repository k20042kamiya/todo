/**
 * Vue Router の設定
 *
 * Vue Router は URL に応じて表示するコンポーネント（ページ）を切り替える。
 * SPA（シングルページアプリケーション）なので、ページ遷移時にブラウザのリロードは発生しない。
 *
 * createWebHistory() - ブラウザの History API を使う（URLに # が付かない）
 * routes - URLパスとコンポーネントの対応を定義
 */
import { createRouter, createWebHistory } from 'vue-router'
import { watch } from 'vue'
import TodoView from '@/views/TodoView.vue'
import LoginView from '@/views/LoginView.vue'

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

// TODO: ナビゲーションガード（ルート遷移前の認証チェック）を実装する
//
// ナビゲーションガードとは？
//   ページ遷移の前に実行される関数。
//   未認証のユーザーがTODOページにアクセスしようとしたら、
//   ログインページにリダイレクトするために使う。
//
// ヒント:
//   import { useAuth } from '@/composables/useAuth'
//
//   router.beforeEach((to, from, next) => {
//     // to: 遷移先のルート情報
//     // from: 遷移元のルート情報
//     // next: 遷移を続行する関数
//
//     const { user, loading } = useAuth()
//
//     // ログインページへの遷移はそのまま許可
//     if (to.name === 'login') {
//       next()
//       return
//     }
//
//     // 認証状態の読み込み中は待機が必要
//     // ユーザーが未認証ならログインページにリダイレクト
//     // if (!user.value) {
//     //   next({ name: 'login' })
//     // } else {
//     //   next()
//     // }
//   })
//
// 注意: Firebase の認証状態の復元は非同期なので、
//       loading が true の間は判定を待つ仕組みが必要。
//       まずはガードなしで実装を進め、後から追加するのでもOK。

// ===== ナビゲーションガードの実装 =====
// router.beforeEach は「全てのページ遷移の前」に実行される関数です。
// これにより、未認証のユーザーがTODOページにアクセスしようとした場合に、
// 自動的にログインページにリダイレクトすることができます。
//
// 注意: Firebaseの認証状態の復元は非同期で行われます。
// アプリ起動直後は loading が true で、user が null の状態です。
// loading が false になるまで待ってから認証チェックを行わないと、
// ログイン済みのユーザーまでログインページにリダイレクトされてしまいます。
import { useAuth } from '@/composables/useAuth'

router.beforeEach(async (to, from, next) => {
  const { user, loading } = useAuth()

  // ===== ログインページへの遷移は常に許可 =====
  // ログインページにはどんなユーザーでもアクセスできる必要があるため、
  // 認証チェックをスキップして即座に遷移を許可します。
  if (to.name === 'login') {
    next()
    return
  }

  // ===== loading 中は認証状態の確定を待機 =====
  // Firebaseの認証状態の復元が完了するまで待ちます。
  // watch() で loading の変化を監視し、false になったら resolve して待機を解除します。
  // Promise を使って非同期処理を同期的に待つパターンです。
  if (loading.value) {
    await new Promise<void>((resolve) => {
      const unwatch = watch(loading, (val) => {
        if (!val) {
          unwatch()  // 監視を解除（メモリリーク防止）
          resolve()  // Promise を解決して待機終了
        }
      })
    })
  }

  // ===== 認証チェック =====
  // loading が完了した後、user.value を確認します。
  // null（未ログイン）ならログインページにリダイレクト、
  // User オブジェクト（ログイン済み）なら遷移を許可します。
  if (!user.value) {
    next({ name: 'login' })
  } else {
    next()
  }
})

export default router
