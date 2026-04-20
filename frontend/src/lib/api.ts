/**
 * APIクライアント - バックエンドとの通信を行うモジュール
 *
 * ===== fetch API の基礎 =====
 *
 * fetch() はブラウザ標準のHTTP通信関数。
 * サーバーにリクエストを送り、レスポンスを受け取る。
 *
 * 基本の使い方:
 *   const response = await fetch(url, {
 *     method: 'GET',                      // HTTPメソッド（GET/POST/PUT/DELETE）
 *     headers: {                           // リクエストヘッダー
 *       'Content-Type': 'application/json',  // 送信データの形式
 *       'Authorization': 'Bearer トークン',   // 認証情報
 *     },
 *     body: JSON.stringify(data),          // リクエストボディ（POST/PUTの場合）
 *   })
 *
 *   const data = await response.json()    // レスポンスをJSONとしてパース
 *
 * ===== async/await の基礎 =====
 *
 * async 関数は非同期処理（時間のかかる処理）を扱う関数。
 * await をつけると、その処理が完了するまで待ってくれる。
 *
 *   async function fetchData() {
 *     const response = await fetch(url)  // ← ここで通信完了を待つ
 *     const data = await response.json() // ← ここでパース完了を待つ
 *     return data
 *   }
 */
import type { Todo, TodoListResponse, CreateTodoRequest, UpdateTodoRequest } from '@/types/todo'
import { auth } from '@/lib/firebase'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

/**
 * 認証トークン付きのfetchラッパー関数
 *
 * 全てのAPIリクエストに Firebase の認証トークンを付与する。
 * バックエンドはこのトークンを検証してユーザーを識別する。
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   1. auth.currentUser でログイン中のユーザーを取得する（null ならエラー）
 *   2. await auth.currentUser.getIdToken() で認証トークン（JWT文字列）を取得する
 *   3. fetch() を呼び出す。headersに以下を設定:
 *      - 'Authorization': `Bearer ${token}`  ← テンプレートリテラル（バッククォート）で文字列に変数を埋め込む
 *      - 'Content-Type': 'application/json'
 *      - ...options.headers  ← 呼び出し元から追加ヘッダーがあればマージ
 *   4. response.ok が false なら throw new Error(`API error: ${response.status}`)
 *   5. response を返す
 */
async function fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
  // ===== ステップ1: ログイン中のユーザーを確認 =====
  // auth.currentUser は現在ログインしているユーザーのオブジェクトです。
  // ログインしていない場合は null になるので、その場合はエラーにします。
  // API通信には必ず認証が必要なので、未ログインでの通信は許可しません。
  if (!auth.currentUser) {
    throw new Error('ユーザーがログインしていません')
  }

  // ===== ステップ2: 認証トークン（JWT）を取得 =====
  // getIdToken() は Firebase が発行する JWT（JSON Web Token）を取得します。
  // このトークンは「このユーザーは確かにログインしている」という証明書のようなものです。
  // バックエンドはこのトークンを検証して、リクエストが正当なものか確認します。
  const token = await auth.currentUser.getIdToken()

  // ===== ステップ3: fetch でAPIリクエストを送信 =====
  // Authorization ヘッダーに「Bearer トークン」形式でトークンを付与します。
  // 「Bearer」は「持参人」という意味で、「このトークンを持っている人を認証してください」
  // という意味のHTTP認証スキームです。
  // ...options.headers はスプレッド構文で、呼び出し元から追加のヘッダーがあればマージします。
  const response = await fetch(url, {
    ...options,
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      ...options.headers,
    },
  })

  // ===== ステップ4: レスポンスのエラーチェック =====
  // response.ok は HTTPステータスコードが 200〜299 の場合に true になります。
  // 400番台（クライアントエラー）や500番台（サーバーエラー）の場合は false になるので、
  // その場合はエラーをスローして呼び出し側の catch で処理させます。
  if (!response.ok) {
    throw new Error(`API error: ${response.status}`)
  }

  return response
}

/**
 * TODO一覧を取得する
 * GET /api/v1/todos
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   const response = await fetchWithAuth(`${BASE_URL}/todos`)
 *   const data: TodoListResponse = await response.json()
 *   return data.todos
 */
export async function fetchTodos(): Promise<Todo[]> {
  // ===== GETリクエストでTODO一覧を取得 =====
  // fetchWithAuth を使うことで、自動的に認証トークンが付与されます。
  // GETリクエストはデフォルトなので method の指定は不要です。
  // response.json() でレスポンスボディをJSONオブジェクトに変換します。
  // バックエンドは { todos: [...] } という形式で返すので、data.todos で配列を取り出します。
  const response = await fetchWithAuth(`${BASE_URL}/todos`)
  const data: TodoListResponse = await response.json()
  return data.todos
}

/**
 * TODOを新規作成する
 * POST /api/v1/todos
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   const response = await fetchWithAuth(`${BASE_URL}/todos`, {
 *     method: 'POST',
 *     body: JSON.stringify(data),  // オブジェクトをJSON文字列に変換
 *   })
 *   const todo: Todo = await response.json()
 *   return todo
 */
export async function createTodo(data: CreateTodoRequest): Promise<Todo> {
  // ===== POSTリクエストで新しいTODOを作成 =====
  // method: 'POST' で「新しいデータを作成する」ことをサーバーに伝えます。
  // JSON.stringify(data) はJavaScriptのオブジェクトをJSON文字列に変換する関数です。
  // HTTPリクエストのボディには文字列しか送れないため、この変換が必要です。
  // サーバーは作成されたTODOオブジェクトをレスポンスとして返します。
  const response = await fetchWithAuth(`${BASE_URL}/todos`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
  const todo: Todo = await response.json()
  return todo
}

/**
 * TODOを更新する
 * PUT /api/v1/todos/:id
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   fetchWithAuth(`${BASE_URL}/todos/${id}`, {
 *     method: 'PUT',
 *     body: JSON.stringify(data),
 *   })
 */
export async function updateTodo(id: number, data: UpdateTodoRequest): Promise<Todo> {
  // ===== PUTリクエストで既存のTODOを更新 =====
  // URLに /${id} を含めることで、「どのTODOを更新するか」をサーバーに伝えます。
  // method: 'PUT' は「既存のデータを置き換える」ことを意味します。
  // テンプレートリテラル（バッククォート `）の中で ${変数名} を使うと、
  // 変数の値が文字列に埋め込まれます。例: id=1 なら "/api/v1/todos/1" になります。
  const response = await fetchWithAuth(`${BASE_URL}/todos/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
  const todo: Todo = await response.json()
  return todo
}

/**
 * TODOを削除する
 * DELETE /api/v1/todos/:id
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   await fetchWithAuth(`${BASE_URL}/todos/${id}`, {
 *     method: 'DELETE',
 *   })
 *   // 削除APIは204 No Contentを返すので、response.json() は不要
 */
export async function deleteTodo(id: number): Promise<void> {
  // ===== DELETEリクエストでTODOを削除 =====
  // method: 'DELETE' は「データを削除する」ことを意味します。
  // 削除APIは成功時に 204 No Content を返すのが一般的で、
  // レスポンスボディが空なので response.json() を呼ぶ必要はありません。
  // 戻り値の型が Promise<void>（何も返さない）なのもそのためです。
  await fetchWithAuth(`${BASE_URL}/todos/${id}`, {
    method: 'DELETE',
  })
}
