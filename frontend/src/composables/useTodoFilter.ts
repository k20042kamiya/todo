import { ref, computed } from 'vue'
import type { FilterType } from '@/types/todo'
import { useTodos } from './useTodos'

export function useTodoFilter() {
  const { todos } = useTodos()
  const currentFilter = ref<FilterType>('all')

  const filteredTodos = computed(() => {
    switch (currentFilter.value) {
      case 'all':
        return todos.value
      case 'incomplete':
        return todos.value.filter(t => !t.is_completed)
      case 'completed':
        return todos.value.filter(t => t.is_completed)
      case 'overdue':
        return todos.value.filter(t => {
          if (!t.due_date || t.is_completed) return false
          return new Date(t.due_date) < new Date()
        })
      default:
        return todos.value
    }
  })

  const remainingCount = computed(() => {
    return todos.value.filter(t => !t.is_completed).length
  })

  const completedCount = computed(() => {
    return todos.value.filter(t => t.is_completed).length
  })

  const progressPercentage = computed(() => {
    const total = todos.value.length
    if (total === 0) return 0
    return Math.round(completedCount.value / total * 100)
  })

  function setFilter(filter: FilterType): void {
    currentFilter.value = filter
  }

  return {
    currentFilter,
    filteredTodos,
    remainingCount,
    completedCount,
    progressPercentage,
    setFilter,
  }
}
