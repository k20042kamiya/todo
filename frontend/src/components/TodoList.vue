<script setup lang="ts">
import type { Todo } from '@/types/todo'
import TodoItem from '@/components/TodoItem.vue'

const props = defineProps<{
  todos: Todo[]
}>()

const emit = defineEmits<{
  toggle: [todo: Todo]
  edit: [todo: Todo]
  delete: [id: number]
}>()
</script>

<template>
  <div class="todo-list">
    <TodoItem
      v-for="todo in props.todos"
      :key="todo.id"
      :todo="todo"
      @toggle="emit('toggle', todo)"
      @edit="emit('edit', todo)"
      @delete="emit('delete', todo.id)"
    />

    <div v-if="props.todos.length === 0" class="todo-list-empty">
      タスクがありません
    </div>
  </div>
</template>

<style scoped>
.todo-list {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.todo-list-empty {
  padding: 40px;
  text-align: center;
  color: #999;
  font-size: 14px;
}
</style>
