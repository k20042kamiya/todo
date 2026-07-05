<script setup lang="ts">
defineProps<{
  message: string
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <Teleport to="body">
    <div class="error-alert-overlay">
      <div class="error-alert-dialog" role="alertdialog" aria-modal="true" aria-labelledby="error-alert-title">
        <h2 id="error-alert-title" class="error-alert-title">エラー</h2>
        <p class="error-alert-message">{{ message }}</p>
        <button class="error-alert-button" @click="emit('close')">OK</button>
      </div>
    </div>
  </Teleport>
</template>

<!-- Teleport 先 (body) では scoped styles が適用されない場合があるため、
     固有のクラス名プレフィックス (error-alert-) を付けた非 scoped スタイルを使う -->
<style>
.error-alert-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  /* モーダル・ダイアログ (z-index: 1000) より常に前面に表示する */
  z-index: 2000;
}

.error-alert-dialog {
  background: white;
  border-radius: 16px;
  padding: 32px;
  width: 100%;
  max-width: 360px;
  text-align: center;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.error-alert-title {
  font-size: 18px;
  font-weight: 700;
  color: #c0392b;
  margin-bottom: 12px;
}

.error-alert-message {
  font-size: 14px;
  color: #666;
  line-height: 1.7;
  margin-bottom: 24px;
}

.error-alert-button {
  background-color: #e86c50;
  color: white;
  border: none;
  padding: 10px 40px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.error-alert-button:hover {
  background-color: #d55a40;
}
</style>
