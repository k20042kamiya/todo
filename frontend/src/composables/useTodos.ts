/**
 * useTodos - TODO一覧のCRUD操作を管理するComposable
 *
 * ===== CRUD とは？ =====
 * データの基本操作4つの頭文字:
 *   C = Create（作成）  → POST /todos
 *   R = Read（読み取り） → GET /todos
 *   U = Update（更新）  → PUT /todos/:id
 *   D = Delete（削除）  → DELETE /todos/:id
 *
 * ===== try/catch/finally の基礎 =====
 *
 * エラーが発生する可能性のあるコードを安全に実行する仕組み:
 *
 *   try {
 *     // エラーが起きるかもしれない処理
 *     const data = await api.fetchTodos()
 *   } catch (error) {
 *     // エラーが起きた時の処理
 *     console.error('エラー:', error)
 *   } finally {
 *     // 成功でも失敗でも必ず実行される処理
 *     loading.value = false
 *   }
 */
import { ref } from 'vue'
import type { Todo, CreateTodoRequest, UpdateTodoRequest } from '@/types/todo'
import * as api from '@/lib/api'

// ---- グローバル状態（全コンポーネントで共有） ----
const todos = ref<Todo[]>([])
const loading = ref(false)

export function useTodos() {
  /**
   * TODO一覧をバックエンドから取得する
   *
   * TODO: この関数を実装してください
   *
   * ヒント:
   *   async function fetchTodos(): Promise<void> {
   *     loading.value = true
   *     try {
   *       todos.value = await api.fetchTodos()
   *     } catch (error) {
   *       console.error('TODO一覧の取得に失敗:', error)
   *     } finally {
   *       loading.value = false
   *     }
   *   }
   */
  async function fetchTodos(): Promise<void> {
    // ===== try/catch/finally パターン =====
    // API通信はネットワークエラーやサーバーエラーが起きる可能性があるため、
    // 必ず try/catch で囲みます。
    // loading.value を true にすることで、画面に「読み込み中」の表示ができます。
    // finally ブロックは成功しても失敗しても必ず実行されるので、
    // loading を false に戻す処理を書くのに最適です。
    loading.value = true
    try {
      todos.value = await api.fetchTodos()
    } catch (error) {
      console.error('TODO一覧の取得に失敗:', error)
    } finally {
      loading.value = false
    }
  }

  /**
   * 新しいTODOを作成する
   *
   * TODO: この関数を実装してください
   *
   * ヒント:
   *   1. api.createTodo(data) を await で呼ぶ → 作成された Todo が返る
   *   2. 返ってきた Todo を todos.value の配列に追加する
   *
   *   配列に要素を追加する方法:
   *     todos.value.push(newTodo)         // 末尾に追加
   *     todos.value = [...todos.value, newTodo]  // スプレッド構文で新しい配列を作成
   */
  async function addTodo(data: CreateTodoRequest): Promise<void> {
    try {
      // ===== APIを呼んでサーバー側でTODOを作成 =====
      // api.createTodo はサーバーに新しいTODOを作成するリクエストを送ります。
      // サーバーが作成したTODO（IDやcreated_at等が付与された完全なデータ）を返すので、
      // それを todos 配列に追加します。
      // push() は配列の末尾に要素を追加するメソッドです。
      const newTodo = await api.createTodo(data)
      todos.value.push(newTodo)
    } catch (error) {
      console.error('TODOの作成に失敗:', error)
    }
  }

  /**
   * TODOを更新する
   *
   * TODO: この関数を実装してください
   *
   * ヒント:
   *   1. api.updateTodo(id, data) を呼ぶ → 更新された Todo が返る
   *   2. todos.value 内の該当 Todo を新しいデータで置き換える
   *
   *   配列の要素を置き換える方法（map を使う）:
   *     todos.value = todos.value.map(t => t.id === id ? updatedTodo : t)
   *
   *   map() は配列の各要素に関数を適用して新しい配列を作る:
   *     [1, 2, 3].map(n => n * 2)  →  [2, 4, 6]
   */
  async function editTodo(id: number, data: UpdateTodoRequest): Promise<void> {
    try {
      // ===== 配列内の要素を置き換える（mapパターン） =====
      // map() は配列の各要素に関数を適用して「新しい配列」を作ります。
      // 三項演算子（条件 ? A : B）で、IDが一致する要素だけを更新後のデータに置き換え、
      // それ以外はそのまま残します。
      // 例: todos = [{id:1,...}, {id:2,...}] で id=2 を更新すると
      //     → [{id:1,...}, {更新されたid:2のデータ}]
      const updatedTodo = await api.updateTodo(id, data)
      todos.value = todos.value.map(t => t.id === id ? updatedTodo : t)
    } catch (error) {
      console.error('TODOの更新に失敗:', error)
    }
  }

  /**
   * TODOを削除する
   *
   * TODO: この関数を実装してください
   *
   * ヒント:
   *   1. api.deleteTodo(id) を呼ぶ
   *   2. todos.value から該当 Todo を除外する
   *
   *   配列から要素を除外する方法（filter を使う）:
   *     todos.value = todos.value.filter(t => t.id !== id)
   *
   *   filter() は条件に合う要素だけの新しい配列を作る:
   *     [1, 2, 3].filter(n => n !== 2)  →  [1, 3]
   */
  async function removeTodo(id: number): Promise<void> {
    try {
      // ===== 配列から要素を除外する（filterパターン） =====
      // まずサーバー側で削除を行い、成功したらローカルの配列からも除外します。
      // filter() は条件が true の要素だけを残した新しい配列を作ります。
      // t.id !== id は「削除対象のID以外」という条件なので、
      // 削除対象だけが除外された新しい配列になります。
      await api.deleteTodo(id)
      todos.value = todos.value.filter(t => t.id !== id)
    } catch (error) {
      console.error('TODOの削除に失敗:', error)
    }
  }

  /**
   * TODOの完了/未完了を切り替える
   *
   * TODO: この関数を実装してください
   *
   * ヒント:
   *   1. 現在の todo の値から UpdateTodoRequest を作る
   *   2. is_completed を反転させる（true → false、false → true）
   *
   *   const data: UpdateTodoRequest = {
   *     title: todo.title,
   *     content: todo.content ?? '',   // ?? はnull合体演算子（nullなら右辺の値を使う）
   *     due_date: todo.due_date ?? undefined,
   *     is_completed: !todo.is_completed,  // ! で真偽値を反転
   *   }
   *   await editTodo(todo.id, data)
   */
  async function toggleComplete(todo: Todo): Promise<void> {
    // ===== null合体演算子 ?? と論理否定演算子 ! =====
    // ?? （null合体演算子）: 左辺が null または undefined の場合に右辺の値を使います。
    //   例: null ?? '' → ''、'hello' ?? '' → 'hello'
    // ! （論理否定演算子）: true を false に、false を true に反転します。
    //   例: !true → false、!false → true
    // UpdateTodoRequest の型に合わせて、null になりうるフィールドを適切に変換しています。
    const data: UpdateTodoRequest = {
      title: todo.title,
      content: todo.content ?? '',
      due_date: todo.due_date ?? undefined,
      is_completed: !todo.is_completed,
    }
    await editTodo(todo.id, data)
  }

  /**
   * 完了済みTODOを一括削除する
   *
   * TODO: この関数を実装してください
   *
   * ヒント:
   *   1. todos.value から is_completed === true のものを抽出する
   *   2. 各完了TODOに対して api.deleteTodo() を呼ぶ
   *   3. Promise.all() で並行実行する（全ての削除を同時に実行）
   *
   *   const completedTodos = todos.value.filter(t => t.is_completed)
   *   await Promise.all(completedTodos.map(t => api.deleteTodo(t.id)))
   *   todos.value = todos.value.filter(t => !t.is_completed)
   *
   *   ※ バックエンドに一括削除APIがないため、個別に削除する
   */
  async function removeCompleted(): Promise<void> {
    try {
      // ===== Promise.all で複数の非同期処理を並行実行 =====
      // Promise.all() は複数の Promise を同時に実行し、全てが完了するのを待ちます。
      // 例えば完了済みTODOが3件あれば、3つの削除リクエストが同時に送信されます。
      // 1件ずつ順番に削除するよりも高速です。
      // map() で各TODOを api.deleteTodo(t.id) という Promise に変換し、
      // それらを Promise.all() に渡しています。
      const completedTodos = todos.value.filter(t => t.is_completed)
      await Promise.all(completedTodos.map(t => api.deleteTodo(t.id)))
      // サーバー側の削除が成功したら、ローカルの配列からも完了済みを除外
      todos.value = todos.value.filter(t => !t.is_completed)
    } catch (error) {
      console.error('完了済みTODOの削除に失敗:', error)
    }
  }

  return {
    todos,
    loading,
    fetchTodos,
    addTodo,
    editTodo,
    removeTodo,
    toggleComplete,
    removeCompleted,
  }
}
