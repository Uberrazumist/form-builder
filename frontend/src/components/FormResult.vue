<template>
  <div class="result" :class="{ error: result.error, warning: result.warning }">
    <div class="result-head">
      <Icon :name="result.error ? 'error' : (result.warning ? 'error' : 'check')" />
      <span>{{ result.error ? 'Ошибка' : (result.warning ? 'Предупреждение' : 'Успешно') }}</span>
    </div>
    <pre>{{ JSON.stringify(result, null, 2) }}</pre>
  </div>
</template>

<script setup>
import Icon from './Icon.vue'

defineProps({
  result: Object
})
</script>

<style scoped>
.result {
  margin-top: 1.5rem;
  padding: 1rem 1.1rem;
  background: #eef6ee;
  border: 1px solid #cfe6cf;
  border-radius: var(--radius-sm);
  animation: fadeUp 0.3s ease both;
}
.result.error {
  background: #fdecec;
  border-color: #f5c6c6;
}
.result.warning {
  background: #fff8e1;
  border-color: #ffe082;
}
.result-head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}
.result-head svg {
  color: #2f855a;
}
.result.error .result-head svg {
  color: #c53030;
}
.result.warning .result-head svg {
  color: #d69e2e;
}
pre {
  font-family: 'SF Mono', Menlo, monospace;
  font-size: 0.8rem;
  color: var(--text-muted);
  white-space: pre-wrap;
  word-break: break-all;
}
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
