<script setup lang="ts">
import type { Todo } from '@/types/todo'
import { isPastDue } from '@/lib/date'

const props = defineProps<{
  todo: Todo
}>()

const emit = defineEmits<{
  toggle: []
  edit: []
  delete: []
}>()

function isOverdue(): boolean {
  if (!props.todo.due_date || props.todo.is_completed) return false
  return isPastDue(props.todo.due_date)
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}
</script>

<template>
  <div class="todo-item" :class="{ completed: props.todo.is_completed }">
    <div class="overdue-dot" :class="{ hidden: !isOverdue() }"></div>

    <div
      class="todo-checkbox"
      :class="{ checked: props.todo.is_completed }"
      @click="emit('toggle')"
    >
      <span v-if="props.todo.is_completed" class="check-icon">&#10003;</span>
    </div>

    <div class="todo-info" @click="emit('edit')">
      <span class="todo-title" :class="{ completed: props.todo.is_completed }">{{ props.todo.title }}</span>

      <div class="todo-meta">
        <span v-if="props.todo.due_date" class="todo-due-date">
          &#128197; {{ formatDate(props.todo.due_date) }}
        </span>
      </div>
    </div>

    <div class="todo-actions">
      <button class="btn-action" @click.stop="emit('edit')">&#9998;</button>
      <button class="btn-action btn-delete" @click.stop="emit('delete')">&#128465;</button>
    </div>
  </div>
</template>

<style scoped>
.todo-item {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  gap: 12px;
  border-bottom: 1px solid #f5f5f5;
  transition: background-color 0.2s;
}

.todo-item:last-child {
  border-bottom: none;
}

.todo-item:hover {
  background-color: #fafafa;
}

.overdue-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #f5a0a0;
  flex-shrink: 0;
}

.overdue-dot.hidden {
  visibility: hidden;
}

.todo-checkbox {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  border: 2px solid #ddd;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.2s;
}

.todo-checkbox.checked {
  background-color: #e86c50;
  border-color: #e86c50;
}

.check-icon {
  color: white;
  font-size: 14px;
  font-weight: 700;
}

.todo-info {
  flex: 1;
  min-width: 0; /* テキストが溢れた時に省略できるようにする */
  cursor: pointer;
}

.todo-title {
  font-size: 14px;
  color: #2c2c2c;
  display: block;
}

.todo-title.completed {
  color: #bbb;
  text-decoration: line-through;
}

.todo-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
}

.todo-due-date {
  font-size: 12px;
  color: #999;
  display: flex;
  align-items: center;
  gap: 4px;
}

.todo-actions {
  display: flex;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s;
}

.todo-item:hover .todo-actions {
  opacity: 1;
}

.btn-action {
  background: none;
  border: none;
  color: #999;
  cursor: pointer;
  padding: 4px;
  font-size: 16px;
}

.btn-action:hover {
  color: #e86c50;
}

.btn-delete:hover {
  color: #e53e3e;
}
</style>
