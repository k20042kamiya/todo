<script setup lang="ts">
/**
 * TodoItem.vue - 個別のTODOアイテム
 *
 * スクリーンショットでの各TODOの構成:
 *   [●] [✓] タイトル
 *        📅 2026年2月22日  [個人]
 *
 *   ● = 期限超過インジケーター（ピンクの丸）
 *   ✓ = 完了チェックボックス
 *
 * ===== 日付のフォーマット =====
 *
 * バックエンドから返される日付は ISO 8601 形式の文字列:
 *   "2026-02-22T00:00:00Z"
 *
 * これを "2026年2月22日" のように表示するには:
 *   new Date(dateString).toLocaleDateString('ja-JP', {
 *     year: 'numeric',
 *     month: 'long',
 *     day: 'numeric',
 *   })
 */
import type { Todo } from '@/types/todo'

// TODO: defineProps で個別の todo を受け取る
// const props = defineProps<{
//   todo: Todo
// }>()

// TODO: defineEmits でイベントを定義する
// const emit = defineEmits<{
//   toggle: []    // 完了状態の切り替え
//   edit: []      // 編集
//   delete: []    // 削除
// }>()

// ===== defineProps と defineEmits =====
// props で親（TodoList）から個別のTODOデータを受け取り、
// emits でユーザー操作（完了切替、編集、削除）を親に通知します。
// emits の [] は「引数なし」を意味します。
// TodoItem は「どのTODOか」を知る必要がなく、親が v-for のスコープで
// どのTODOに対する操作かを判断します。
const props = defineProps<{
  todo: Todo
}>()

const emit = defineEmits<{
  toggle: []
  edit: []
  delete: []
}>()

/**
 * 期限超過かどうかを判定する
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   function isOverdue(): boolean {
 *     if (!props.todo.due_date || props.todo.is_completed) return false
 *     return new Date(props.todo.due_date) < new Date()
 *   }
 */
// ===== 期限超過の判定ロジック =====
// 以下の3つの条件を全て満たす場合に「期限超過」と判定します:
// 1. due_date が設定されている（null でない）
// 2. まだ完了していない（is_completed が false）
// 3. 期日が現在日時より前（過去）
// !props.todo.due_date は「due_date が null または undefined の場合 true」を意味し、
// その場合は早期リターンで false を返します（期日がなければ超過しようがない）。
function isOverdue(): boolean {
  if (!props.todo.due_date || props.todo.is_completed) return false
  return new Date(props.todo.due_date) < new Date()
}

/**
 * 日付を「YYYY年M月D日」形式にフォーマットする
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   function formatDate(dateString: string): string {
 *     return new Date(dateString).toLocaleDateString('ja-JP', {
 *       year: 'numeric',
 *       month: 'long',
 *       day: 'numeric',
 *     })
 *   }
 */
// ===== toLocaleDateString で日付をフォーマット =====
// new Date(dateString) で ISO 8601 文字列を Date オブジェクトに変換し、
// toLocaleDateString('ja-JP', ...) で日本語形式の文字列に変換します。
// 'ja-JP' は日本語ロケールを指定し、year/month/day のオプションで表示形式を制御します。
// 例: "2026-02-22T00:00:00Z" → "2026年2月22日"
function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}
</script>

<template>
  <!--
    TODO: 完了状態に応じてクラスを動的に追加する
    ヒント: <div class="todo-item" :class="{ completed: props.todo.is_completed }">

    :class に渡すオブジェクトの形式:
      { クラス名: 条件 }
      条件が true のとき、そのクラスが追加される
  -->
  <!-- ===== 動的クラスバインディング =====
       :class="{ completed: props.todo.is_completed }" は、
       is_completed が true の場合に class="todo-item completed" になります。
       CSSで .todo-item.completed にスタイルを定義することで、完了状態の見た目を変えられます。 -->
  <div class="todo-item" :class="{ completed: props.todo.is_completed }">
    <!-- 期限超過インジケーター（ピンクの丸） -->
    <!-- TODO: 期限超過でない場合は hidden クラスを追加して非表示にする -->
    <!-- ヒント: :class="{ hidden: !isOverdue() }" -->
    <!-- ===== 関数の戻り値でクラスを切り替え =====
         isOverdue() の戻り値が false の場合、hidden クラスが追加されて
         CSS の visibility: hidden で非表示になります。
         v-if と違い、要素自体はDOMに存在するため、レイアウトのスペースは保持されます。 -->
    <div class="overdue-dot" :class="{ hidden: !isOverdue() }"></div>

    <!-- 完了チェックボックス -->
    <!-- TODO: @click="emit('toggle')" を追加してクリックで完了切替 -->
    <!-- TODO: :class="{ checked: props.todo.is_completed }" でチェック状態のスタイルを切替 -->
    <!-- ===== @click でイベントを親に通知 =====
         @click="emit('toggle')" はチェックボックスがクリックされた時に
         toggle イベントを親コンポーネント（TodoList）に通知します。
         TodoList はさらに親（TodoView）に中継し、最終的に toggleComplete() が呼ばれます。 -->
    <div
      class="todo-checkbox"
      :class="{ checked: props.todo.is_completed }"
      @click="emit('toggle')"
    >
      <!-- TODO: 完了時にチェックマーク(✓)を表示する -->
      <!-- ヒント: <span v-if="props.todo.is_completed" class="check-icon">✓</span> -->
      <span v-if="props.todo.is_completed" class="check-icon">&#10003;</span>
    </div>

    <!-- TODO情報エリア -->
    <!-- TODO: @click="emit('edit')" を追加してクリックで編集モーダルを開く -->
    <div class="todo-info" @click="emit('edit')">
      <!-- タイトル -->
      <!-- TODO: {{ props.todo.title }} でタイトルを表示する -->
      <!-- TODO: :class="{ completed: props.todo.is_completed }" で完了時のスタイルを適用 -->
      <!-- ===== 完了時の打ち消し線スタイル =====
           :class="{ completed: ... }" で、完了したTODOのタイトルに
           text-decoration: line-through（打ち消し線）と薄い色が適用されます。 -->
      <span class="todo-title" :class="{ completed: props.todo.is_completed }">{{ props.todo.title }}</span>

      <div class="todo-meta">
        <!-- 期日（due_date がある場合のみ表示） -->
        <!-- TODO: v-if="props.todo.due_date" で条件付き表示 -->
        <!-- TODO: {{ formatDate(props.todo.due_date) }} で日付をフォーマットして表示 -->
        <!-- ===== v-if で条件付き表示 =====
             due_date が null の場合は期日の表示自体が不要なので、
             v-if で due_date が存在する場合のみ表示します。
             formatDate() で ISO 形式の日付文字列を日本語形式に変換して表示します。 -->
        <span v-if="props.todo.due_date" class="todo-due-date">
          &#128197; {{ formatDate(props.todo.due_date) }}
        </span>
      </div>
    </div>

    <!-- 操作ボタン（ホバー時に表示される） -->
    <div class="todo-actions">
      <!-- TODO: @click.stop="emit('edit')" を追加 -->
      <!-- .stop は event.stopPropagation() の省略記法。 -->
      <!-- 親要素へのイベント伝播を止める（親のclickが発火するのを防ぐ）。 -->
      <!-- ===== .stop 修飾子でイベント伝播を防止 =====
           @click.stop は「このクリックイベントを親要素に伝播させない」という意味です。
           .stop がないと、ボタンをクリックした時に親の todo-info の @click も発火してしまい、
           編集モーダルが二重に開いてしまいます。 -->
      <button class="btn-action" @click.stop="emit('edit')">&#9998;</button> <!-- 編集アイコン -->

      <!-- TODO: @click.stop="emit('delete')" を追加 -->
      <button class="btn-action btn-delete" @click.stop="emit('delete')">&#128465;</button> <!-- 削除アイコン -->
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

/* 期限超過のピンク丸 */
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

/* チェックボックス */
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

/* TODO情報エリア */
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

/* 操作ボタン（ホバー時に表示） */
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
