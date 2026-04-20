<script setup lang="ts">
/**
 * TodoList.vue - TODOリスト（TodoItem の一覧を描画する）
 *
 * このコンポーネントは「中間コンポーネント」の役割を持つ。
 * 親（TodoView）から受け取った todos 配列を v-for で繰り返し描画し、
 * 各 TodoItem からのイベントを親に中継する。
 *
 * データフロー:
 *   TodoView → (props) → TodoList → (props) → TodoItem
 *   TodoItem → (emit) → TodoList → (emit) → TodoView
 */
import type { Todo } from '@/types/todo'
// TODO: TodoItem コンポーネントをインポートする
// import TodoItem from '@/components/TodoItem.vue'

// ===== 子コンポーネントのインポート =====
// .vue ファイルをインポートすると、そのコンポーネントをテンプレートで使えるようになります。
// Vue3 の <script setup> では、インポートしたコンポーネントは自動的に登録されるため、
// 別途 components: { TodoItem } のような登録処理は不要です。
import TodoItem from '@/components/TodoItem.vue'

// TODO: defineProps で todos 配列を受け取る
// const props = defineProps<{
//   todos: Todo[]
// }>()

// TODO: defineEmits でイベントを定義する
// const emit = defineEmits<{
//   toggle: [todo: Todo]
//   edit: [todo: Todo]
//   delete: [id: number]
// }>()

// ===== props と emits でイベントの中継を行う =====
// TodoList は「中間コンポーネント」として、親（TodoView）と子（TodoItem）の間で
// データとイベントを受け渡します。
// props で親からデータを受け取り、emits で子からのイベントを親に中継します。
// このパターンは「イベントのバブリング（伝搬）」と呼ばれることもあります。
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
    <!--
      TODO: v-for で TodoItem を繰り返し描画する

      ヒント:
        <TodoItem
          v-for="todo in props.todos"
          :key="todo.id"
          :todo="todo"
          @toggle="emit('toggle', todo)"
          @edit="emit('edit', todo)"
          @delete="emit('delete', todo.id)"
        />

      v-for="todo in props.todos"
        → props.todos 配列の各要素を todo 変数に入れて繰り返す

      :key="todo.id"
        → 各要素を一意に識別する値。Vueの効率的な更新に必要。

      :todo="todo"
        → 各 TodoItem に個別の todo データを props として渡す

      @toggle="emit('toggle', todo)"
        → TodoItem が toggle イベントを発火したら、
          このコンポーネントも toggle イベントを親に発火する（イベントの中継）
    -->
    <!-- ===== v-for と イベント中継の組み合わせ =====
         v-for で todos 配列をループし、各要素に対して TodoItem コンポーネントを描画します。
         :todo="todo" で各 TodoItem に個別のTODOデータを渡し、
         @toggle, @edit, @delete で子からのイベントを受け取って親に中継しています。
         これにより TodoView → TodoList → TodoItem のデータフローが完成します。 -->
    <TodoItem
      v-for="todo in props.todos"
      :key="todo.id"
      :todo="todo"
      @toggle="emit('toggle', todo)"
      @edit="emit('edit', todo)"
      @delete="emit('delete', todo.id)"
    />

    <!-- TODOがない場合のメッセージ -->
    <!-- TODO: v-if="props.todos.length === 0" で条件分岐する -->
    <!-- ===== v-if で空リスト時のメッセージを表示 =====
         props.todos.length === 0 は「配列の要素数が0（空）」かどうかの判定です。
         TODOが1件もない場合だけこのメッセージが表示されます。 -->
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
