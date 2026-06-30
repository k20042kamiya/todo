<script setup lang="ts">
/**
 * TodoFormModal.vue - TODO作成/編集用のモーダルフォーム
 *
 * ===== モーダルとは？ =====
 *
 * 画面全体を半透明の背景（オーバーレイ）で覆い、
 * その上にフォームを表示するUIパターン。
 * ユーザーはモーダルを閉じるまで背面のUIを操作できない。
 *
 * ===== このコンポーネントの動作 =====
 *
 * 1. 「+ 新しいタスク」ボタン → todo=null で表示（新規作成モード）
 * 2. TODOをクリック → todo=既存データ で表示（編集モード）
 * 3. フォームに入力して保存 → save イベントを親に通知
 * 4. ×ボタンやオーバーレイクリック → close イベントを親に通知
 *
 * ===== watch() とは？ =====
 *
 * リアクティブな値の変化を監視する関数。
 * 値が変わった時に処理を実行できる。
 *
 *   watch(ソース, (新しい値, 古い値) => {
 *     // 値が変わった時の処理
 *   })
 *
 * ここでは props.todo が変わった時にフォームの初期値を設定するために使う。
 */
import { ref, watch } from 'vue'
import type { Todo } from '@/types/todo'

// TODO: defineProps で props を定義する
// const props = defineProps<{
//   todo: Todo | null   // null なら新規作成、Todo なら編集
// }>()

// TODO: defineEmits でイベントを定義する
// const emit = defineEmits<{
//   save: [data: { title: string; content: string; due_date?: string }]
//   close: []
// }>()

// ===== props で新規/編集モードを切り替え =====
// todo が null なら「新規作成モード」、Todo オブジェクトなら「編集モード」です。
// TypeScript のユニオン型 Todo | null で、どちらの値も受け取れるようにしています。
// emits の save イベントには、フォームの入力値をオブジェクトとして渡します。
const props = defineProps<{
  todo: Todo | null
}>()

const emit = defineEmits<{
  save: [data: { title: string; content: string; due_date?: string }]
  close: []
}>()

// ---- フォームの入力値 ----
const title = ref('')
const content = ref('')
const dueDate = ref('')

// TODO: props.todo の値に応じてフォームの初期値を設定する
//
// ヒント（即時実行パターン）:
//   if (props.todo) {
//     title.value = props.todo.title
//     content.value = props.todo.content ?? ''
//     // 日付は <input type="date"> 用に "YYYY-MM-DD" 形式に変換する
//     dueDate.value = props.todo.due_date
//       ? props.todo.due_date.split('T')[0]   // "2026-02-22T00:00:00Z" → "2026-02-22"
//       : ''
//   }
//
// または watch を使う方法:
//   watch(() => props.todo, (newTodo) => {
//     if (newTodo) {
//       title.value = newTodo.title
//       content.value = newTodo.content ?? ''
//       dueDate.value = newTodo.due_date ? newTodo.due_date.split('T')[0] : ''
//     } else {
//       title.value = ''
//       content.value = ''
//       dueDate.value = ''
//     }
//   }, { immediate: true })
//
//   immediate: true → watch登録直後にも1回実行される

// ===== watch で props の変化を監視してフォーム初期値を設定 =====
// watch(() => props.todo, ...) は props.todo の値が変わるたびに実行されます。
// { immediate: true } オプションを付けると、watch 登録直後にも1回実行されます。
// これにより、コンポーネントが表示された瞬間にフォームに初期値が設定されます。
//
// 編集モード（newTodo が Todo オブジェクト）の場合:
//   - 既存のタイトル、内容、期日をフォームに設定
//   - due_date は "2026-02-22T00:00:00Z" → "2026-02-22" に変換（split('T')[0]）
//     split('T') は文字列を 'T' で分割して配列にし、[0] で最初の要素を取得
//
// 新規作成モード（newTodo が null）の場合:
//   - フォームを空にリセット
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

/**
 * フォーム送信処理
 *
 * TODO: この関数を実装してください
 *
 * ヒント:
 *   function handleSubmit() {
 *     if (!title.value.trim()) return  // タイトルが空なら何もしない
 *     // trim() は文字列の前後の空白を除去する
 *
 *     emit('save', {
 *       title: title.value,
 *       content: content.value,
 *       // dueDate が空でなければ ISO 8601 形式に変換して送信
 *       due_date: dueDate.value ? new Date(dueDate.value + 'T23:59:59').toISOString() : undefined,
 *     })
 *   }
 *
 *   toISOString() は日付を "2026-02-22T00:00:00.000Z" 形式に変換する
 */
function handleSubmit() {
  // ===== バリデーションとイベント発火 =====
  // trim() は文字列の前後の空白を除去する関数です。
  // "  " のように空白だけの入力を防ぐために使います。
  // !title.value.trim() が true（空文字列）の場合は早期リターンで何もしません。
  //
  // emit('save', { ... }) で親コンポーネントに保存データを送信します。
  // due_date は <input type="date"> の値が "YYYY-MM-DD" 形式なので、
  // new Date().toISOString() で ISO 8601 形式に変換してバックエンドに送ります。
  // dueDate が空文字列の場合は undefined にして、サーバーに期日なしとして送信します。
  if (!title.value.trim()) return

  emit('save', {
    title: title.value,
    content: content.value,
    due_date: dueDate.value ? new Date(dueDate.value).toISOString() : undefined,
  })
}
</script>

<template>
  <!--
    モーダルの構造:
    1. overlay（半透明の背景）→ クリックでモーダルを閉じる
    2. content（白いカード）→ フォーム本体

    @click.self は「その要素自体がクリックされた時だけ」発火する。
    子要素（カード）のクリックでは発火しない。
    これにより、背景クリックでモーダルを閉じ、カード内のクリックでは閉じない。
  -->
  <!-- TODO: @click.self="emit('close')" を追加する -->
  <!-- ===== @click.self 修飾子 =====
       .self 修飾子は「この要素自体がクリックされた時だけ」イベントを発火します。
       子要素（modal-content）をクリックしても発火しません。
       これにより「背景（オーバーレイ）クリックでモーダルを閉じる」が実現できます。 -->
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content">
      <!-- TODO: props.todo の有無でタイトルを切り替える -->
      <!-- ヒント: {{ props.todo ? 'タスクを編集' : '新しいタスク' }} -->
      <h2 class="modal-title">{{ props.todo ? 'タスクを編集' : '新しいタスク' }}</h2>

      <!-- TODO: @submit.prevent="handleSubmit" を追加する -->
      <form @submit.prevent="handleSubmit">
        <!-- タイトル入力 -->
        <div class="form-group">
          <label class="form-label">タイトル</label>
          <!-- TODO: v-model="title" を追加してフォーム入力とバインドする -->
          <input
            v-model="title"
            type="text"
            class="form-input"
            placeholder="タスクのタイトル"
            maxlength="100"
          />
        </div>

        <!-- 内容入力 -->
        <div class="form-group">
          <label class="form-label">内容</label>
          <!--
            <textarea> は複数行のテキスト入力欄。
            TODO: v-model="content" を追加する
          -->
          <textarea
            v-model="content"
            class="form-textarea"
            placeholder="タスクの詳細（任意）"
          ></textarea>
        </div>

        <!-- 期日入力 -->
        <div class="form-group">
          <label class="form-label">期日</label>
          <!--
            type="date" はブラウザ標準の日付選択UIを表示する。
            値は "YYYY-MM-DD" 形式の文字列。
            TODO: v-model="dueDate" を追加する
          -->
          <input
            v-model="dueDate"
            type="date"
            class="form-input"
          />
        </div>

        <!-- ボタンエリア -->
        <div class="modal-actions">
          <!-- TODO: @click="emit('close')" を追加する -->
          <!-- type="button" を指定しないと、form内のbuttonはデフォルトで submit になる -->
          <button type="button" class="btn-cancel" @click="emit('close')">キャンセル</button>
          <button type="submit" class="btn-save">
            <!-- TODO: props.todo の有無でテキストを切り替える -->
            {{ props.todo ? '更新' : '保存' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;    /* 画面に固定（スクロールしても動かない） */
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);  /* 半透明の黒 */
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;      /* 他の要素より前面に表示 */
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
  display: block;     /* block要素にして改行させる */
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
  font-family: inherit; /* 親要素のフォントを継承 */
  outline: none;        /* フォーカス時のデフォルト枠線を消す */
  transition: border-color 0.2s;
}

.form-input:focus,
.form-textarea:focus {
  border-color: #e86c50;  /* フォーカス時にオレンジの枠線 */
}

.form-textarea {
  resize: vertical;    /* 縦方向のみリサイズ可能 */
  min-height: 80px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;  /* 右寄せ */
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
