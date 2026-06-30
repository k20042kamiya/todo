<script setup lang="ts">
const props = defineProps<{
  remaining: number
  completed: number
  percentage: number
}>()
</script>

<template>
  <div class="stats-container">
    <div class="stats-badges">
      <div class="stat-badge">
        <span class="badge-count badge-remaining">{{ props.remaining }}</span>
        <span class="badge-label">残り</span>
      </div>
      <div class="stat-badge">
        <span class="badge-count badge-completed">{{ props.completed }}</span>
        <span class="badge-label">完了</span>
      </div>
    </div>

    <div class="progress-container">
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: props.percentage + '%' }"></div>
      </div>
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
