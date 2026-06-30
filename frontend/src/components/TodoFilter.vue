<script setup lang="ts">
import type { FilterType } from '@/types/todo'

const props = defineProps<{
  currentFilter: FilterType
}>()

const emit = defineEmits<{
  changeFilter: [filter: FilterType]
  deleteCompleted: []
}>()

const filters: { key: FilterType; label: string }[] = [
  { key: 'all', label: 'すべて' },
  { key: 'incomplete', label: '未完了' },
  { key: 'completed', label: '完了済み' },
  { key: 'overdue', label: '期限超過' },
]
</script>

<template>
  <div class="filter-container">
    <div class="filter-tabs">
      <button
        v-for="filter in filters"
        :key="filter.key"
        class="filter-tab"
        :class="{ active: props.currentFilter === filter.key }"
        @click="emit('changeFilter', filter.key)"
      >
        {{ filter.label }}
      </button>
    </div>

    <button class="btn-delete-completed" @click="emit('deleteCompleted')">完了を削除</button>
  </div>
</template>

<style scoped>
.filter-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filter-tabs {
  display: flex;
  gap: 4px;
}

.filter-tab {
  padding: 6px 16px;
  border: none;
  background: transparent;
  border-radius: 20px;
  font-size: 13px;
  cursor: pointer;
  color: #666;
  transition: all 0.2s;
}

.filter-tab.active {
  background-color: #2c2c2c;
  color: white;
}

.filter-tab:hover:not(.active) {
  background-color: #e8e4de;
}

.btn-delete-completed {
  background: none;
  border: none;
  color: #999;
  font-size: 13px;
  cursor: pointer;
  text-decoration: underline;
}

.btn-delete-completed:hover {
  color: #e86c50;
}
</style>
