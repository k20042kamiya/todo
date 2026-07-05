<script setup lang="ts">
const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <div class="dialog-overlay">
    <div class="dialog" role="alertdialog" aria-modal="true" aria-labelledby="session-expired-title">
      <h2 id="session-expired-title" class="dialog-title">ログインの有効期限切れ</h2>
      <p class="dialog-message">
        ログインの有効期限が切れました。<br />
        お手数ですが、再度ログインしてください。
      </p>
      <button class="dialog-button" @click="emit('close')">OK</button>
    </div>
  </div>
</template>

<style scoped>
.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: white;
  border-radius: 16px;
  padding: 32px;
  width: 100%;
  max-width: 360px;
  text-align: center;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.dialog-title {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 12px;
}

.dialog-message {
  font-size: 14px;
  color: #666;
  line-height: 1.7;
  margin-bottom: 24px;
}

.dialog-button {
  background-color: #e86c50;
  color: white;
  border: none;
  padding: 10px 40px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.dialog-button:hover {
  background-color: #d55a40;
}
</style>
