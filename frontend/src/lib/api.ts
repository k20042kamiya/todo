import type { Todo, TodoListResponse, CreateTodoRequest, UpdateTodoRequest } from '@/types/todo'
import { auth } from '@/lib/firebase'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

async function fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
  if (!auth.currentUser) {
    throw new Error('ユーザーがログインしていません')
  }

  const token = await auth.currentUser.getIdToken()

  const response = await fetch(url, {
    ...options,
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      ...options.headers,
    },
  })

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`)
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
