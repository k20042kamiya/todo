<script setup lang="ts">
/**
 * TodoView.vue - メインのTODO管理ページ
 *
 * このコンポーネントがページ全体を構成する「親コンポーネント」。
 * 子コンポーネント（TodoStats, TodoFilter 等）を組み合わせてUIを構築する。
 *
 * ===== コンポーネントの親子関係 =====
 *
 * TodoView（親）
 *   ├── TodoStats     → 統計情報（残り/完了/進捗）
 *   ├── TodoFilter    → フィルタータブ
 *   ├── TodoList      → TODOリスト
 *   │   └── TodoItem  → 各TODO（TodoListが繰り返し描画）
 *   └── TodoFormModal → 作成/編集フォーム
 *
 * ===== props と emits =====
 *
 * props（プロップス）: 親 → 子 にデータを渡す仕組み
 *   親: <TodoStats :remaining="3" :completed="5" />
 *   子: const props = defineProps<{ remaining: number; completed: number }>()
 *
 * emits（エミッツ）: 子 → 親 にイベントを通知する仕組み
 *   子: emit('changeFilter', 'completed')
 *   親: <TodoFilter @change-filter="setFilter" />
 *
 * ===== onMounted() =====
 *
 * コンポーネントがDOMに追加された直後に実行されるライフサイクルフック。
 * 初期データの読み込み（API呼び出し）に使う。
 *
 *   onMounted(() => {
 *     fetchTodos()  // ページ表示時にTODO一覧を取得
 *   })
 */
import { ref, onMounted } from 'vue'
import type { Todo, CreateTodoRequest, UpdateTodoRequest } from '@/types/todo'

// TODO: 子コンポーネントをインポートする
// import TodoStats from '@/components/TodoStats.vue'
// import TodoFilter from '@/components/TodoFilter.vue'
// import TodoList from '@/components/TodoList.vue'
// import TodoFormModal from '@/components/TodoFormModal.vue'

// ===== 子コンポーネントのインポート =====
// 各コンポーネントは一つの「UIパーツ」として独立しています。
// これらを組み合わせてページ全体を構築するのがVueのコンポーネント設計の基本です。
// <script setup> ではインポートするだけで自動的にテンプレートで使えるようになります。
import TodoStats from '@/components/TodoStats.vue'
import TodoFilter from '@/components/TodoFilter.vue'
import TodoList from '@/components/TodoList.vue'
import TodoFormModal from '@/components/TodoFormModal.vue'

// TODO: Composable をインポートして使う
// import { useTodos } from '@/composables/useTodos'
// import { useTodoFilter } from '@/composables/useTodoFilter'
// import { useAuth } from '@/composables/useAuth'
// import { useRouter } from 'vue-router'
//
// const {
//   fetchTodos, addTodo, editTodo, removeTodo, toggleComplete, removeCompleted
// } = useTodos()
// const {
//   currentFilter, filteredTodos, remainingCount, completedCount, progressPercentage, setFilter
// } = useTodoFilter()
// const { logout } = useAuth()
// const router = useRouter()

// ===== Composable の初期化 =====
// 各 Composable から必要な関数・状態を分割代入で取り出します。
// useTodos() → CRUD操作の関数群
// useTodoFilter() → フィルタリングと統計情報
// useAuth() → ログアウト機能
// useRouter() → プログラムからのページ遷移
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

// ---- ローカル状態（このコンポーネントだけで使う状態） ----
const showModal = ref(false)
const editingTodo = ref<Todo | null>(null)

// TODO: ページ表示時にTODO一覧を取得する
// onMounted(() => {
//   fetchTodos()
// })

// ===== onMounted ライフサイクルフック =====
// onMounted はコンポーネントがブラウザのDOMに追加された直後に1回だけ実行されます。
// ここで fetchTodos() を呼ぶことで、ページが表示された瞬間にバックエンドから
// TODO一覧を取得し、画面に表示します。
// これはWebアプリの一般的なパターンで、「初期データの読み込み」と呼ばれます。
onMounted(() => {
  fetchTodos()
})

// ---- イベントハンドラー ----

/** 新規作成モーダルを開く */
function openCreateForm() {
  editingTodo.value = null
  showModal.value = true
}

/** 編集モーダルを開く */
function openEditForm(todo: Todo) {
  editingTodo.value = todo
  showModal.value = true
}

/** モーダルを閉じる */
function closeModal() {
  showModal.value = false
  editingTodo.value = null
  error.value = null
}

/**
 * TODO保存処理（新規作成 or 更新）
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   async function handleSave(data: CreateTodoRequest | UpdateTodoRequest) {
 *     if (editingTodo.value) {
 *       // 編集モード: 既存のTODOを更新
 *       await editTodo(editingTodo.value.id, data as UpdateTodoRequest)
 *     } else {
 *       // 新規作成モード
 *       await addTodo(data as CreateTodoRequest)
 *     }
 *     closeModal()
 *   }
 *
 *   ※ CreateTodoRequest と UpdateTodoRequest は @/types/todo.ts で定義済み
 */
// ===== 型アサーション（as）の使い方 =====
// data は CreateTodoRequest | UpdateTodoRequest のユニオン型ですが、
// editingTodo.value の有無で実際の型が確定します。
// 「as UpdateTodoRequest」は TypeScript に「この値はUpdateTodoRequest型だよ」と伝える
// 型アサーション（型の断定）です。実行時には何も起きず、コンパイル時の型チェック用です。
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

/**
 * ログアウト処理
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   async function handleLogout() {
 *     await logout()
 *     router.push('/login')
 *   }
 */
// ===== ログアウト後にログインページへ遷移 =====
// logout() でFirebaseの認証状態をクリアした後、
// router.push('/login') でログインページに遷移します。
// await を使うことで、ログアウト処理が完了してから遷移を行います。
async function handleLogout() {
  await logout()
  router.push('/login')
}

/**
 * 今日の日付を「YYYY年M月D日曜日」形式で取得する
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   JavaScript の Date オブジェクトを使う:
 *
 *   function getFormattedDate(): string {
 *     const today = new Date()
 *     return today.toLocaleDateString('ja-JP', {
 *       year: 'numeric',    // 年を数字で表示
 *       month: 'long',      // 月を「2月」のように表示
 *       day: 'numeric',     // 日を数字で表示
 *       weekday: 'long',    // 曜日を「日曜日」のように表示
 *     })
 *   }
 *   // → "2026年2月15日日曜日"
 */
// ===== toLocaleDateString でロケールに応じた日付表示 =====
// toLocaleDateString は Date オブジェクトをロケール（地域）に応じた文字列に変換します。
// 'ja-JP' を指定すると日本語形式になり、各オプションで表示する情報を制御できます。
// この関数はテンプレートから {{ getFormattedDate() }} で呼び出されます。
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
    <!-- ===== ヘッダーエリア ===== -->
    <header class="page-header">
      <!-- TODO: 今日の日付を表示する -->
      <!-- ヒント: {{ getFormattedDate() }} で関数の戻り値を表示できる -->
      <p class="current-date">{{ getFormattedDate() }}</p>

      <div class="title-row">
        <h1 class="page-title">タスク管理</h1>
        <!-- TODO: ログアウトボタンを追加する -->
        <!-- ヒント: <button class="btn-logout" @click="handleLogout">ログアウト</button> -->
        <button class="btn-logout" @click="handleLogout">ログアウト</button>
      </div>
    </header>

    <!-- ===== 統計情報エリア ===== -->
    <!-- TODO: TodoStats コンポーネントを配置する -->
    <!--
      コンポーネントの使い方:
      <TodoStats
        :remaining="remainingCount"
        :completed="completedCount"
        :percentage="progressPercentage"
      />
      ※ :属性名 は v-bind の省略記法。変数の値を props として渡す。
    -->
    <!-- ===== 子コンポーネントに props を渡す =====
         :remaining="remainingCount" は v-bind:remaining="remainingCount" の省略記法です。
         useTodoFilter の computed 値を子コンポーネントに渡しています。
         computed 値が変わると、自動的に子コンポーネントも再描画されます。 -->
    <TodoStats
      :remaining="remainingCount"
      :completed="completedCount"
      :percentage="progressPercentage"
    />

    <!-- ===== アクションボタンエリア ===== -->
    <div class="action-buttons">
      <!-- TODO: @click="openCreateForm" を追加する -->
      <button class="btn-new-task" @click="openCreateForm">+ 新しいタスク</button>
    </div>

    <!-- ===== フィルターエリア ===== -->
    <!-- TODO: TodoFilter コンポーネントを配置する -->
    <!--
      <TodoFilter
        :current-filter="currentFilter"
        @change-filter="setFilter"
        @delete-completed="removeCompleted"
      />
      ※ @イベント名 は v-on の省略記法。子コンポーネントからのイベントを受け取る。
      ※ props名はキャメルケース(currentFilter)でもケバブケース(current-filter)でもOK
    -->
    <!-- ===== props のケバブケース表記 =====
         HTMLテンプレートでは :current-filter のようにケバブケース（ハイフン区切り）で書きます。
         Vue3 が自動的にキャメルケース（currentFilter）に変換してくれます。
         @change-filter は子コンポーネントの emit('changeFilter', ...) に対応します。 -->
    <TodoFilter
      :current-filter="currentFilter"
      @change-filter="setFilter"
      @delete-completed="removeCompleted"
    />

    <!-- ===== TODOリストエリア ===== -->
    <!-- TODO: TodoList コンポーネントを配置する -->
    <!--
      <TodoList
        :todos="filteredTodos"
        @toggle="toggleComplete"
        @edit="openEditForm"
        @delete="removeTodo"
      />
    -->
    <!-- ===== filteredTodos で表示データを制御 =====
         filteredTodos は useTodoFilter の computed で、
         現在のフィルター設定に応じて絞り込まれたTODO配列です。
         フィルターを切り替えるだけで、表示されるTODOリストが自動的に変わります。 -->
    <TodoList
      :todos="filteredTodos"
      @toggle="toggleComplete"
      @edit="openEditForm"
      @delete="removeTodo"
    />

    <!-- ===== 作成/編集モーダル ===== -->
    <!-- TODO: TodoFormModal を条件付きで表示する -->
    <!--
      <TodoFormModal
        v-if="showModal"
        :todo="editingTodo"
        @save="handleSave"
        @close="closeModal"
      />
      ※ v-if でモーダルの表示/非表示を切り替える
    -->
    <!-- ===== v-if でモーダルの表示/非表示を制御 =====
         v-if="showModal" は showModal が true の場合のみコンポーネントを描画します。
         false の場合はDOMから完全に削除されます（v-show とは異なり、非表示ではなく削除）。
         :todo="editingTodo" で編集対象のTODOを渡し、null なら新規作成モードになります。 -->
    <div v-if="error" class="error-banner">{{ error }}</div>

    <TodoFormModal
      v-if="showModal"
      :todo="editingTodo"
      @save="handleSave"
      @close="closeModal"
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

.error-banner {
  background-color: #fff0ed;
  border: 1px solid #e86c50;
  color: #c0392b;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  margin-bottom: 12px;
}
</style>
