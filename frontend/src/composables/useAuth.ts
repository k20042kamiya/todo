/**
 * useAuth - Firebase認証を管理するComposable
 *
 * ===== Composable（コンポーザブル）とは？ =====
 *
 * Vue 3 の Composition API を使った「ロジックの再利用パターン」。
 * "use" で始まる関数で、リアクティブな状態とそれを操作する関数をまとめたもの。
 *
 * 例えば、ログインユーザーの情報はヘッダー、TODOページ、設定ページなど
 * 複数の場所で必要になる。Composableにまとめることで、
 * どのコンポーネントからでも同じ状態にアクセスできる。
 *
 * ===== ref() とは？ =====
 *
 * ref() はリアクティブな変数を作る関数。
 * 「リアクティブ」= 値が変わると、それを使っているUIが自動的に再描画される。
 *
 *   const count = ref(0)      // リアクティブな変数を作成
 *   count.value = 1           // .value でアクセス・更新する
 *   // ※ テンプレート（HTML）内では .value は不要。自動で展開される。
 *
 * ===== このComposableの役割 =====
 *
 * - user: 現在のログインユーザー情報（未ログインならnull）
 * - loading: Firebase の認証状態を復元中かどうか
 * - login(): メール/パスワードでログイン
 * - register(): 新規ユーザー登録
 * - logout(): ログアウト
 */
import { ref } from 'vue'
import type { User } from 'firebase/auth'
import { auth } from '@/lib/firebase'

// TODO: firebase/auth から認証関連の関数をインポートする
// ヒント:
// import {
//   signInWithEmailAndPassword,      // メール/パスワードでログイン
//   createUserWithEmailAndPassword,  // メール/パスワードで新規登録
//   signOut,                         // ログアウト
//   onAuthStateChanged,              // 認証状態の変化を監視するリスナー
// } from 'firebase/auth'

// ===== Firebase認証関数のインポート =====
// firebase/auth モジュールから、認証に必要な4つの関数をインポートしています。
// これらはFirebaseが提供する「すぐ使える認証機能」で、
// 自分でパスワードのハッシュ化やセッション管理を実装する必要がありません。
import {
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signOut,
  onAuthStateChanged,
} from 'firebase/auth'

// ---- グローバル状態 ----
// Composable関数の「外」に定義すると、全コンポーネントで同じインスタンスを共有する（シングルトン）
// Composable関数の「中」に定義すると、呼び出すたびに新しいインスタンスが作られる
const user = ref<User | null>(null)
const loading = ref(true)

// TODO: onAuthStateChanged で Firebase の認証状態を監視する
//
// onAuthStateChanged は「コールバック関数」を登録する仕組み。
// Firebaseが認証状態の変化を検知すると、登録した関数が自動的に呼ばれる。
//
// ヒント:
//   onAuthStateChanged(auth, (firebaseUser) => {
//     // firebaseUser: ログイン中なら User オブジェクト、未ログインなら null
//     user.value = firebaseUser
//     loading.value = false
//   })
//
// これを Composable関数の外（モジュールのトップレベル）に書くことで、
// アプリ起動時に1回だけ登録される。

// ===== 認証状態の監視を登録 =====
// onAuthStateChanged はFirebaseの認証状態が変化するたびに呼ばれるリスナーです。
// - ページを開いた時（Firebaseが保存済みのログイン情報を復元した時）
// - ログインした時
// - ログアウトした時
// これをモジュールのトップレベル（関数の外）に書くことで、アプリ全体で1回だけ登録されます。
// もし useAuth() 関数の中に書くと、useAuth() が呼ばれるたびに重複登録されてしまいます。
onAuthStateChanged(auth, (firebaseUser) => {
  // firebaseUser: ログイン中なら User オブジェクト、未ログインなら null
  user.value = firebaseUser
  loading.value = false
})

export function useAuth() {
  /**
   * ログイン処理
   *
   * TODO: この関数を実装してください
   *
   * ヒント:
   *   await signInWithEmailAndPassword(auth, email, password)
   *
   *   ・成功すると onAuthStateChanged のコールバックが発火し、user.value が更新される
   *   ・失敗すると例外がスローされる（呼び出し側で try/catch する）
   */
  async function login(email: string, password: string): Promise<void> {
    // ===== なぜ await だけで良いのか =====
    // signInWithEmailAndPassword は Promise を返す非同期関数です。
    // await で完了を待ち、成功すれば onAuthStateChanged が自動的に発火して
    // user.value が更新されます。戻り値を変数に入れる必要はありません。
    // エラーが起きた場合は例外がスローされ、呼び出し側の try/catch でキャッチします。
    await signInWithEmailAndPassword(auth, email, password)
  }

  /**
   * ユーザー登録処理
   *
   * TODO: この関数を実装してください
   *
   * ヒント:
   *   await createUserWithEmailAndPassword(auth, email, password)
   *
   *   ・登録成功後、自動的にログイン状態になる
   *   ・onAuthStateChanged のコールバックが発火し、user.value が更新される
   */
  async function register(email: string, password: string): Promise<void> {
    // ===== 登録と同時にログインされる =====
    // createUserWithEmailAndPassword は新しいユーザーを作成し、
    // 同時にそのユーザーでログインした状態にしてくれます。
    // そのため、登録後に別途 login() を呼ぶ必要はありません。
    await createUserWithEmailAndPassword(auth, email, password)
  }

  /**
   * ログアウト処理
   *
   * TODO: この関数を実装してください
   *
   * ヒント:
   *   await signOut(auth)
   *
   *   ・ログアウト後、onAuthStateChanged のコールバックが発火し、user.value が null になる
   *   ・ログアウト後にログインページへ遷移させたい場合は、
   *     呼び出し側で router.push('/login') する
   */
  async function logout(): Promise<void> {
    // ===== signOut で認証状態をクリア =====
    // signOut を呼ぶと、Firebaseがブラウザに保存していた認証情報が削除され、
    // onAuthStateChanged のコールバックが発火して user.value が null になります。
    // ページ遷移（router.push('/login')）はこの関数の呼び出し側で行います。
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
