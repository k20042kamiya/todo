export interface Todo {
  id: number
  user_id: number
  title: string
  content: string | null
  due_date: string | null
  is_completed: boolean
  created_at: string
  updated_at: string
}

export interface TodoListResponse {
  todos: Todo[]
}

export interface CreateTodoRequest {
  title: string
  content?: string
  due_date?: string
}

export interface UpdateTodoRequest {
  title: string
  content: string
  due_date?: string
  is_completed: boolean
}

export type FilterType = 'all' | 'incomplete' | 'completed' | 'overdue'
