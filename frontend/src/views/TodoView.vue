<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Todo, CreateTodoRequest, UpdateTodoRequest } from '@/types/todo'
import TodoStats from '@/components/TodoStats.vue'
import TodoFilter from '@/components/TodoFilter.vue'
import TodoList from '@/components/TodoList.vue'
import TodoFormModal from '@/components/TodoFormModal.vue'
import ErrorAlertDialog from '@/components/ErrorAlertDialog.vue'
import { useTodos } from '@/composables/useTodos'
import { useTodoFilter } from '@/composables/useTodoFilter'
import { useAuth } from '@/composables/useAuth'
import { useRouter } from 'vue-router'

const {
  fetchTodos, addTodo, editTodo, removeTodo, toggleComplete, removeCompleted, error
} = useTodos()
const {
  currentFilter, filteredTodos, remainingCount, completedCount, progressPercentage, setFilter
} = useTodoFilter()
const { logout } = useAuth()
const router = useRouter()

const showModal = ref(false)
const editingTodo = ref<Todo | null>(null)

onMounted(() => {
  fetchTodos()
})

function openCreateForm() {
  editingTodo.value = null
  showModal.value = true
}

function openEditForm(todo: Todo) {
  editingTodo.value = todo
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingTodo.value = null
  error.value = null
}

async function handleSave(data: CreateTodoRequest | UpdateTodoRequest) {
  try {
    if (editingTodo.value) {
      await editTodo(editingTodo.value.id, data as UpdateTodoRequest)
    } else {
      await addTodo(data as CreateTodoRequest)
    }
    closeModal()
  } catch {
    // error.value は useTodos 側でセット済み、モーダルはそのまま表示
  }
}

async function handleLogout() {
  await logout()
  router.push('/login')
}

function getFormattedDate(): string {
  const today = new Date()
  return today.toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  })
}
</script>

<template>
  <div class="todo-page">
    <!-- ヘッダー -->
    <header class="page-header">
      <p class="current-date">{{ getFormattedDate() }}</p>
      <div class="title-row">
        <h1 class="page-title">タスク管理</h1>
        <button class="btn-logout" @click="handleLogout">ログアウト</button>
      </div>
    </header>

    <TodoStats
      :remaining="remainingCount"
      :completed="completedCount"
      :percentage="progressPercentage"
    />

    <div class="action-buttons">
      <button class="btn-new-task" @click="openCreateForm">+ 新しいタスク</button>
    </div>

    <TodoFilter
      :current-filter="currentFilter"
      @change-filter="setFilter"
      @delete-completed="removeCompleted"
    />

    <TodoList
      :todos="filteredTodos"
      @toggle="toggleComplete"
      @edit="openEditForm"
      @delete="removeTodo"
    />

    <TodoFormModal
      v-if="showModal"
      :todo="editingTodo"
      @save="handleSave"
      @close="closeModal"
    />

    <ErrorAlertDialog
      v-if="error"
      :message="error"
      @close="error = null"
    />
  </div>
</template>

<style scoped>
.page-header {
  margin-bottom: 24px;
}

.current-date {
  color: #e86c50;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 4px;
}

.title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-size: 32px;
  font-weight: 700;
  letter-spacing: 4px;
}

.btn-logout {
  background: none;
  border: 1px solid #e0e0e0;
  color: #999;
  padding: 6px 16px;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
}

.btn-logout:hover {
  color: #e86c50;
  border-color: #e86c50;
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.btn-new-task {
  background-color: #e86c50;
  color: white;
  border: none;
  padding: 10px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-new-task:hover {
  background-color: #d55a40;
}
</style>
