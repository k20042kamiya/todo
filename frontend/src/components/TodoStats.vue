<script setup lang="ts">
/**
 * TodoStats.vue - 統計情報（残り/完了カウント + プログレスバー）
 *
 * ===== defineProps の使い方 =====
 *
 * defineProps は親コンポーネントから渡されたデータ（props）を受け取る関数。
 * TypeScript のジェネリクス <{ ... }> で型を定義する。
 *
 *   const props = defineProps<{
 *     remaining: number   // 未完了数
 *     completed: number   // 完了数
 *     percentage: number  // 進捗率
 *   }>()
 *
 *   // テンプレートでは props.remaining で値にアクセスする
 *   // script 内でも props.remaining で値にアクセスする
 */

// TODO: defineProps で props を定義する
// ヒント:
// const props = defineProps<{
//   remaining: number
//   completed: number
//   percentage: number
// }>()

// ===== defineProps でpropsの型を定義 =====
// defineProps はVue3のマクロ（特殊な関数）で、親コンポーネントから受け取るデータの型を定義します。
// <{ ... }> の中に TypeScript の型を書くことで、
// 親が渡すデータの型が間違っている場合にエディタがエラーを表示してくれます。
// 例: 親が <TodoStats :remaining="'文字列'" /> と書くとエラーになる（number が期待されるため）。
const props = defineProps<{
  remaining: number
  completed: number
  percentage: number
}>()
</script>

<template>
  <div class="stats-container">
    <div class="stats-badges">
      <!-- 残りタスク数バッジ -->
      <div class="stat-badge">
        <!-- TODO: {{ props.remaining }} で残り数を表示する -->
        <!-- ===== props の値をテンプレートで表示 =====
             {{ props.remaining }} で親から渡された値を表示します。
             props の値が変わると、テンプレートも自動的に再描画されます。 -->
        <span class="badge-count badge-remaining">{{ props.remaining }}</span>
        <span class="badge-label">残り</span>
      </div>
      <!-- 完了タスク数バッジ -->
      <div class="stat-badge">
        <!-- TODO: {{ props.completed }} で完了数を表示する -->
        <span class="badge-count badge-completed">{{ props.completed }}</span>
        <span class="badge-label">完了</span>
      </div>
    </div>

    <!-- プログレスバー -->
    <div class="progress-container">
      <!-- TODO: :style でプログレスバーの幅を動的に設定する -->
      <!--
        :style はインラインスタイルを動的に設定するディレクティブ。
        オブジェクト形式で CSS プロパティを指定する:
          :style="{ width: props.percentage + '%' }"

        これにより、percentage が 75 なら width: 75% になる。
      -->
      <div class="progress-bar">
        <!-- ===== :style で動的にCSSを設定 =====
             :style はオブジェクト形式でCSSプロパティを動的に設定します。
             { width: props.percentage + '%' } は、例えば percentage が 60 なら
             style="width: 60%" というインラインスタイルを生成します。
             これにより、プログレスバーの幅が進捗率に応じて変化します。 -->
        <div class="progress-fill" :style="{ width: props.percentage + '%' }"></div>
      </div>
      <!-- TODO: {{ props.percentage }}% で進捗率を表示する -->
      <span class="progress-text">{{ props.percentage }}%</span>
    </div>
  </div>
</template>

<style scoped>
.stats-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.stats-badges {
  display: flex;
  gap: 16px;
  align-items: center;
}

.stat-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.badge-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  font-size: 13px;
  font-weight: 600;
}

.badge-remaining {
  background-color: #2c2c2c;
  color: white;
}

.badge-completed {
  background-color: #e86c50;
  color: white;
}

.badge-label {
  color: #666;
}

.progress-container {
  display: flex;
  align-items: center;
  gap: 12px;
}

.progress-bar {
  width: 120px;
  height: 6px;
  background-color: #e0e0e0;
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: #e86c50;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 14px;
  font-weight: 600;
  color: #666;
}
</style>
