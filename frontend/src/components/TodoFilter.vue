<script setup lang="ts">
/**
 * TodoFilter.vue - フィルタータブ + 完了削除ボタン
 *
 * ===== defineEmits の使い方 =====
 *
 * defineEmits は子 → 親にイベントを通知する仕組みを定義する。
 *
 *   const emit = defineEmits<{
 *     changeFilter: [filter: FilterType]   // イベント名: [引数の型]
 *     deleteCompleted: []                   // 引数なしのイベント
 *   }>()
 *
 *   // イベントを発火する:
 *   emit('changeFilter', 'completed')
 *   emit('deleteCompleted')
 *
 * ===== 動的クラスバインディング =====
 *
 * :class は要素のCSSクラスを動的に切り替えるディレクティブ。
 *
 *   <button :class="{ active: isActive }">
 *   → isActive が true なら class="active" が付く
 *
 *   <button :class="['base', { active: isActive }]">
 *   → 常に class="base"、isActive が true なら class="base active"
 */
import type { FilterType } from '@/types/todo'

// TODO: defineProps で現在のフィルターを受け取る
// const props = defineProps<{
//   currentFilter: FilterType
// }>()

// TODO: defineEmits でイベントを定義する
// const emit = defineEmits<{
//   changeFilter: [filter: FilterType]
//   deleteCompleted: []
// }>()

// ===== defineProps と defineEmits の組み合わせ =====
// このコンポーネントは「データを受け取る（props）」と「イベントを通知する（emits）」の
// 両方を行います。これがVueの「単方向データフロー」の基本パターンです。
// - props: 親 → 子（データが流れる方向）
// - emits: 子 → 親（イベントが通知される方向）
// データは常に親から子へ流れ、子が親のデータを直接変更することはできません。
// 子は emit でイベントを発火し、親がそのイベントを受けてデータを変更します。
const props = defineProps<{
  currentFilter: FilterType
}>()

const emit = defineEmits<{
  changeFilter: [filter: FilterType]
  deleteCompleted: []
}>()

// フィルターの選択肢（テンプレートの v-for で使う）
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
      <!--
        TODO: v-for でフィルターボタンを繰り返し描画する

        v-for は配列の各要素に対して要素を繰り返し描画するディレクティブ:
          <button v-for="filter in filters" :key="filter.key">
            {{ filter.label }}
          </button>

        :key は各要素を一意に識別するための属性（必須）。
        Vueが効率的にDOMを更新するために使う。

        追加で以下も実装する:
        - @click="emit('changeFilter', filter.key)"  → クリックでフィルター切替
        - :class="{ active: props.currentFilter === filter.key }"
          → 選択中のフィルターに active クラスを付ける
      -->
      <!-- ===== v-for による繰り返し描画 =====
           v-for="filter in filters" で、filters 配列の各要素を filter 変数に入れて繰り返します。
           :key は Vueが各要素を追跡するための一意な識別子で、必ず指定する必要があります。
           @click で emit を呼び、親コンポーネントにフィルター変更を通知します。
           :class で現在選択中のフィルターに active クラスを付け、スタイルを変更します。 -->
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

    <!-- TODO: @click="emit('deleteCompleted')" を追加する -->
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
