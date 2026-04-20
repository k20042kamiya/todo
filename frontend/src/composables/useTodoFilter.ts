/**
 * useTodoFilter - TODOのフィルタリングと統計情報を管理するComposable
 *
 * ===== computed() とは？ =====
 *
 * 他のリアクティブな値から「自動的に計算される値」を作る関数。
 * 依存する値が変わると、computed の値も自動的に再計算される。
 *
 *   const count = ref(3)
 *   const doubled = computed(() => count.value * 2)  // → 6
 *   count.value = 5
 *   // doubled.value は自動的に 10 になる
 *
 * computed は「読み取り専用」。直接 .value = xxx で書き換えることはできない。
 * 元となるデータ（ref）を変えることで、computed も連動して変わる。
 */
import { ref, computed } from 'vue'
import type { FilterType } from '@/types/todo'
import { useTodos } from './useTodos'

export function useTodoFilter() {
  const { todos } = useTodos()
  const currentFilter = ref<FilterType>('all')

  /**
   * 現在のフィルターに基づいてTODOを絞り込む
   *
   * TODO: computed() を使って実装してください
   *
   * ヒント:
   *   const filteredTodos = computed(() => {
   *     switch (currentFilter.value) {
   *       case 'all':
   *         return todos.value
   *       case 'incomplete':
   *         return todos.value.filter(t => !t.is_completed)
   *       case 'completed':
   *         return todos.value.filter(t => t.is_completed)
   *       case 'overdue':
   *         // 期限超過: 期日が今日より前 かつ 未完了
   *         return todos.value.filter(t => {
   *           if (!t.due_date || t.is_completed) return false
   *           return new Date(t.due_date) < new Date()
   *         })
   *       default:
   *         return todos.value
   *     }
   *   })
   *
   *   new Date() は現在の日時を取得する
   *   new Date(文字列) は文字列から日時オブジェクトを作成する
   *   日時同士は < > で比較できる
   */
  // ===== computed() でフィルタリングを自動化 =====
  // computed() の中で currentFilter.value と todos.value を参照しているため、
  // どちらかの値が変わると filteredTodos も自動的に再計算されます。
  // これが Vue3 の「リアクティビティ」の強力な点です。
  // switch文は if/else if の代わりに使える分岐構文で、
  // 比較する値のパターンが多い場合に読みやすくなります。
  const filteredTodos = computed(() => {
    switch (currentFilter.value) {
      case 'all':
        return todos.value
      case 'incomplete':
        // ===== filter() で条件に合う要素だけを抽出 =====
        // !t.is_completed は「未完了のもの」を意味します。
        // filter() は元の配列を変更せず、新しい配列を返します。
        return todos.value.filter(t => !t.is_completed)
      case 'completed':
        return todos.value.filter(t => t.is_completed)
      case 'overdue':
        // ===== 期限超過の判定ロジック =====
        // 期限超過とは「期日が設定されていて、まだ未完了で、期日が過去」のTODOです。
        // new Date(t.due_date) で期日を Date オブジェクトに変換し、
        // new Date()（現在日時）と比較しています。
        return todos.value.filter(t => {
          if (!t.due_date || t.is_completed) return false
          return new Date(t.due_date) < new Date()
        })
      default:
        return todos.value
    }
  })

  /**
   * 未完了のTODO数
   *
   * TODO: computed() で実装してください
   *
   * ヒント:
   *   const remainingCount = computed(() => {
   *     return todos.value.filter(t => !t.is_completed).length
   *   })
   *
   *   .length は配列の要素数を返すプロパティ
   */
  // ===== computed で統計値を自動計算 =====
  // todos が変更されるたびに自動的に再計算されるので、
  // 手動で「TODOが増えたからカウントを更新する」といった処理は不要です。
  // filter() で条件に合う要素だけの配列を作り、.length でその数を取得しています。
  const remainingCount = computed(() => {
    return todos.value.filter(t => !t.is_completed).length
  })

  /**
   * 完了済みのTODO数
   *
   * TODO: computed() で実装してください
   */
  const completedCount = computed(() => {
    return todos.value.filter(t => t.is_completed).length
  })

  /**
   * 進捗率（パーセンテージ: 0〜100）
   *
   * TODO: computed() で実装してください
   *
   * ヒント:
   *   const progressPercentage = computed(() => {
   *     const total = todos.value.length
   *     if (total === 0) return 0               // 0除算を防ぐ
   *     return Math.round(completedCount.value / total * 100)
   *   })
   *
   *   Math.round() は四捨五入する関数
   *   ※ completedCount.value の .value を忘れないこと（computed も ref と同様に .value が必要）
   */
  // ===== 0除算の防止 =====
  // TODOが0件の場合、0で割り算するとNaN（Not a Number）になってしまいます。
  // そのため total === 0 の場合は早期リターンで 0 を返します。
  // Math.round() は四捨五入する関数で、例えば 66.666... → 67 になります。
  // computed の値にアクセスするときも .value が必要な点に注意してください。
  const progressPercentage = computed(() => {
    const total = todos.value.length
    if (total === 0) return 0
    return Math.round(completedCount.value / total * 100)
  })

  /** フィルターを切り替える */
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
