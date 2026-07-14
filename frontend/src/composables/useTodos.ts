import { ref } from 'vue'
import type { Todo, CreateTodoRequest, UpdateTodoRequest } from '@/types/todo'
import * as api from '@/lib/api'

const todos = ref<Todo[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

export function useTodos() {
  async function fetchTodos(): Promise<void> {
    loading.value = true
    try {
      todos.value = await api.fetchTodos()
    } catch (e) {
      error.value = 'TODO一覧の取得に失敗しました'
      console.error('TODO一覧の取得に失敗:', e)
    } finally {
      loading.value = false
    }
  }

  async function addTodo(data: CreateTodoRequest): Promise<void> {
    try {
      const newTodo = await api.createTodo(data)
      todos.value.push(newTodo)
      error.value = null
    } catch (e) {
      error.value = 'TODOの作成に失敗しました'
      console.error('TODOの作成に失敗:', e)
      throw e
    }
  }

  async function editTodo(id: number, data: UpdateTodoRequest): Promise<void> {
    try {
      const updatedTodo = await api.updateTodo(id, data)
      todos.value = todos.value.map(t => t.id === id ? updatedTodo : t)
      error.value = null
    } catch (e) {
      error.value = 'TODOの更新に失敗しました'
      console.error('TODOの更新に失敗:', e)
      throw e
    }
  }

  async function removeTodo(id: number): Promise<void> {
    try {
      await api.deleteTodo(id)
      todos.value = todos.value.filter(t => t.id !== id)
    } catch (e) {
      error.value = 'TODOの削除に失敗しました'
      console.error('TODOの削除に失敗:', e)
    }
  }

  async function toggleComplete(todo: Todo): Promise<void> {
    // 完了状態の切り替えでは他のフィールドを変化させない（content の null はそのまま送る）
    const data: UpdateTodoRequest = {
      title: todo.title,
      content: todo.content,
      due_date: todo.due_date ?? undefined,
      is_completed: !todo.is_completed,
    }
    try {
      await editTodo(todo.id, data)
    } catch {
      // error.value は editTodo 側でセット済み。イベントハンドラから直接呼ばれるため未処理拒否を防ぐ
    }
  }

  async function removeCompleted(): Promise<void> {
    try {
      const completedTodos = todos.value.filter(t => t.is_completed)
      await Promise.all(completedTodos.map(t => api.deleteTodo(t.id)))
      todos.value = todos.value.filter(t => !t.is_completed)
    } catch (e) {
      error.value = '完了済みTODOの削除に失敗しました'
      console.error('完了済みTODOの削除に失敗:', e)
    }
  }

  return {
    todos,
    loading,
    error,
    fetchTodos,
    addTodo,
    editTodo,
    removeTodo,
    toggleComplete,
    removeCompleted,
  }
}
