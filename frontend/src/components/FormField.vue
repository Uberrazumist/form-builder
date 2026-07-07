<template>
  <div class="form-group">
    <label :for="id">
      <Icon :name="icon" />
      {{ label }}
    </label>
    <input
      :id="id"
      :type="type"
      :value="modelValue"
      @input="$emit('update:modelValue', $event.target.value)"
      :required="required"
      :minlength="minlength"
      :placeholder="placeholder"
    />
    <span v-if="hint" class="hint">{{ hint }}</span>
  </div>
</template>

<script setup>
import Icon from './Icon.vue'

defineProps({
  id: String,
  type: String,
  icon: String,
  label: String,
  placeholder: String,
  hint: String,
  required: Boolean,
  minlength: [String, Number],
  modelValue: String
})

defineEmits(['update:modelValue'])
</script>

<style scoped>
.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
label {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text);
}
input {
  padding: 0.75rem 0.95rem;
  font-size: 0.95rem;
  background: var(--bg);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  transition: all 0.2s;
  color: var(--text);
}
input:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--surface);
  box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1);
}
.hint {
  font-size: 0.78rem;
  color: var(--text-muted);
}
</style>
