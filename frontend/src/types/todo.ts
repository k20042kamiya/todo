/**
 * 型定義ファイル
 *
 * TypeScript の interface はデータの「形」を定義する。
 * 型を定義しておくと:
 *   - エディタの補完が効く
 *   - 間違ったプロパティ名を使うとエラーが出る
 *   - コードが読みやすくなる
 */

/** バックエンドから返されるTODOの型 */
export interface Todo {
  id: number
  user_id: number
  title: string
  content: string | null   // null の可能性がある（string | null はユニオン型）
  due_date: string | null   // ISO 8601形式 例: "2026-02-22T00:00:00Z"
  is_completed: boolean
  created_at: string
  updated_at: string
}

/** TODO一覧レスポンスの型（APIが返すJSON構造に対応） */
export interface TodoListResponse {
  todos: Todo[]
}

/** TODO作成リクエストの型 */
export interface CreateTodoRequest {
  title: string
  content?: string      // ? はオプショナル（省略可能）
  due_date?: string
}

/** TODO更新リクエストの型 */
export interface UpdateTodoRequest {
  title: string
  content: string
  due_date?: string
  is_completed: boolean
}

/** フィルターの種類 */
export type FilterType = 'all' | 'incomplete' | 'completed' | 'overdue'
