import type { Todo, TodoListResponse, CreateTodoRequest, UpdateTodoRequest } from '@/types/todo'
import { authClient } from '@/lib/authClient'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

// APIがエラーを返したときにステータスとサーバーからのメッセージを保持する
export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// バックエンドは {"error": "..."} 形式でエラーメッセージを返す。
// JSONでない・形式が違う場合はステータスコードからメッセージを組み立てる
async function readErrorMessage(response: Response): Promise<string> {
  try {
    const body = await response.json()
    if (typeof body?.error === 'string' && body.error !== '') {
      return body.error
    }
  } catch {
    // ボディがJSONでない場合はフォールバックへ
  }
  return `サーバーエラー (${response.status})`
}

async function fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
  const user = authClient.getCurrentUser()
  if (!user) {
    throw new Error('ユーザーがログインしていません')
  }

  const token = await user.getIdToken()

  const response = await fetch(url, {
    ...options,
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      ...options.headers,
    },
  })

  if (!response.ok) {
    throw new ApiError(response.status, await readErrorMessage(response))
  }

  return response
}

export async function fetchTodos(): Promise<Todo[]> {
  const response = await fetchWithAuth(`${BASE_URL}/todos`)
  const data: TodoListResponse = await response.json()
  return data.todos
}

export async function createTodo(data: CreateTodoRequest): Promise<Todo> {
  const response = await fetchWithAuth(`${BASE_URL}/todos`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
  const todo: Todo = await response.json()
  return todo
}

export async function updateTodo(id: number, data: UpdateTodoRequest): Promise<Todo> {
  const response = await fetchWithAuth(`${BASE_URL}/todos/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
  const todo: Todo = await response.json()
  return todo
}

export async function deleteTodo(id: number): Promise<void> {
  await fetchWithAuth(`${BASE_URL}/todos/${id}`, {
    method: 'DELETE',
  })
}
