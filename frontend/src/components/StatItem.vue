<template>
  <div class="stat">
    <span class="stat-num">{{ num }}</span>
    <span class="stat-label">{{ label }}</span>
  </div>
  <div v-if="!isLast" class="stat-divider"></div>
</template>

<script setup>
defineProps({
  num: String,
  label: String,
  isLast: Boolean
})
</script>

<style scoped>
.stat {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}
.stat-num {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--primary);
}
.stat-label {
  font-size: 0.82rem;
  color: var(--text-muted);
}
.stat-divider {
  width: 1px;
  height: 36px;
  background: var(--border);
}
</style>
