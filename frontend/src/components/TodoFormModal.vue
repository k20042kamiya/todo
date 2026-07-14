<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Todo } from '@/types/todo'

const props = defineProps<{
  todo: Todo | null
}>()

const emit = defineEmits<{
  save: [data: { title: string; content: string | null; due_date?: string }]
  close: []
}>()

const title = ref('')
const content = ref('')
const dueDate = ref('')

watch(() => props.todo, (newTodo) => {
  if (newTodo) {
    title.value = newTodo.title
    content.value = newTodo.content ?? ''
    dueDate.value = newTodo.due_date ? newTodo.due_date.split('T')[0] : ''
  } else {
    title.value = ''
    content.value = ''
    dueDate.value = ''
  }
}, { immediate: true })

function handleSubmit() {
  if (!title.value.trim()) return

  emit('save', {
    title: title.value,
    // 未入力は「値なし」として null で送る（DB上の NULL と '' を区別する方針）
    content: content.value || null,
    due_date: dueDate.value || undefined,
  })
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content">
      <h2 class="modal-title">{{ props.todo ? 'タスクを編集' : '新しいタスク' }}</h2>

      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label class="form-label">タイトル</label>
          <input
            v-model="title"
            type="text"
            class="form-input"
            placeholder="タスクのタイトル"
            maxlength="100"
          />
        </div>

        <div class="form-group">
          <label class="form-label">内容</label>
          <textarea
            v-model="content"
            class="form-textarea"
            placeholder="タスクの詳細（任意）"
          ></textarea>
        </div>

        <div class="form-group">
          <label class="form-label">期日</label>
          <input
            v-model="dueDate"
            type="date"
            class="form-input"
          />
        </div>

        <div class="modal-actions">
          <button type="button" class="btn-cancel" @click="emit('close')">キャンセル</button>
          <button type="submit" class="btn-save">
            {{ props.todo ? '更新' : '保存' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 16px;
  padding: 32px;
  width: 90%;
  max-width: 480px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.modal-title {
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 24px;
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

.form-input,
.form-textarea {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus,
.form-textarea:focus {
  border-color: #e86c50;
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.btn-cancel {
  padding: 10px 24px;
  border: 1px solid #e0e0e0;
  background: white;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  color: #666;
}

.btn-cancel:hover {
  background-color: #f9f9f9;
}

.btn-save {
  padding: 10px 24px;
  border: none;
  background-color: #e86c50;
  color: white;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.btn-save:hover {
  background-color: #d55a40;
}
</style>
