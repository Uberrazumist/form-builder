<template>
  <div class="schedule-builder">
    <div class="form-group">
      <label>Название расписания <span class="required">*</span></label>
      <input type="text" v-model="form.name" placeholder="Например: Расписание Иванова" />
    </div>

    <div class="form-group">
      <label>Тип расписания</label>
      <select v-model="form.type">
        <option value="weekly">Еженедельное</option>
        <option value="daily">Ежедневное</option>
      </select>
    </div>

    <div v-if="form.type === 'weekly'" class="form-group">
      <label>Дни недели</label>
      <div class="days-grid">
        <label v-for="(day, idx) in daysOfWeek" :key="idx" class="day-checkbox">
          <input type="checkbox" :value="idx === 0 ? 7 : idx" v-model="form.days" />
          <span>{{ day }}</span>
        </label>
      </div>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>Время начала</label>
        <input type="time" v-model="form.start_time" />
      </div>
      <div class="form-group">
        <label>Время окончания</label>
        <input type="time" v-model="form.end_time" />
      </div>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>Длительность слота (мин)</label>
        <input type="number" v-model.number="form.slot_duration" min="5" max="240" />
      </div>
      <div class="form-group">
        <label>Перерыв между слотами (мин)</label>
        <input type="number" v-model.number="form.break_between" min="0" max="120" />
      </div>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>Начало действия</label>
        <input type="date" v-model="form.start_date" />
      </div>
      <div class="form-group">
        <label>Окончание действия</label>
        <input type="date" v-model="form.end_date" />
      </div>
    </div>

    <div class="form-group">
      <label>Исключения (блокировка дней)</label>
      <textarea
        v-model="exceptionsText"
        placeholder="Каждая дата на новой строке: 2026-12-31"
        rows="3"
      ></textarea>
      <span class="hint">Укажите даты, когда расписание не действует (праздники, отпуска)</span>
    </div>

    <div v-if="previewSlots.length > 0" class="preview-section">
      <h4>Предпросмотр слотов на {{ previewDateLabel }}</h4>
      <div class="preview-slots">
        <span
          v-for="(slot, idx) in previewSlots"
          :key="idx"
          class="preview-slot"
        >
          {{ slot.start_label }} – {{ slot.end_label }}
        </span>
      </div>
    </div>

    <div class="actions">
      <button type="button" class="btn-primary" @click="save" :disabled="saving">
        <span v-if="!saving">Сохранить расписание</span>
        <span v-else class="spinner-small"></span>
      </button>
      <button type="button" class="btn-secondary" @click="generatePreview">
        <Icon name="calendar" />
        Предпросмотр
      </button>
    </div>

    <div v-if="result" :class="['result-message', result.success ? 'success' : 'error']">
      {{ result.message }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import Icon from './Icon.vue'

const props = defineProps<{
  dictionaries: Array<{ id: string; name: string }>
  initialRule?: any
  resourceId: string
}>()

const emit = defineEmits<{
  saved: [ruleId: string]
}>()

const daysOfWeek = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

const form = reactive({
  name: '',
  type: 'weekly' as string,
  days: [1, 2, 3, 4, 5] as number[],
  start_time: '09:00',
  end_time: '18:00',
  slot_duration: 45,
  break_between: 15,
  start_date: new Date().toISOString().split('T')[0],
  end_date: new Date(Date.now() + 30 * 86400000).toISOString().split('T')[0]
})

const exceptionsText = ref('')
const previewSlots = ref<any[]>([])
const saving = ref(false)
const result = ref<{ success: boolean; message: string } | null>(null)

const previewDateLabel = computed(() => {
  const d = new Date()
  return d.toLocaleDateString('ru-RU', {
    weekday: 'long',
    day: 'numeric',
    month: 'long'
  })
})

watch(() => props.initialRule, (rule) => {
  if (!rule) return
  form.name = rule.name || ''
  form.start_time = rule.recurring?.start_time || '09:00'
  form.end_time = rule.recurring?.end_time || '18:00'
  form.slot_duration = rule.recurring?.slot_duration || 45
  form.break_between = rule.recurring?.break_between || 15
  form.start_date = rule.recurring?.start_date || ''
  form.end_date = rule.recurring?.end_date || ''
  form.type = rule.recurring?.type || 'weekly'
  form.days = rule.recurring?.days || [1, 2, 3, 4, 5]
  exceptionsText.value = (rule.recurring?.exceptions || []).join('\n')
}, { immediate: true })

const generatePreview = () => {
  const recurring = buildRecurring()
  const preview = generateSlotsPreview(recurring)
  previewSlots.value = preview
}

const buildRecurring = () => {
  return {
    type: form.type,
    days: form.days,
    start_time: form.start_time,
    end_time: form.end_time,
    slot_duration: form.slot_duration,
    break_between: form.break_between,
    start_date: form.start_date,
    end_date: form.end_date,
    exceptions: exceptionsText.value.split('\n').map((s: string) => s.trim()).filter(Boolean)
  }
}

const generateSlotsPreview = (recurring: any) => {
  const slots: any[] = []
  const startHour = parseInt(recurring.start_time.split(':')[0])
  const startMin = parseInt(recurring.start_time.split(':')[1])
  const endHour = parseInt(recurring.end_time.split(':')[0])
  const endMin = parseInt(recurring.end_time.split(':')[1])

  const current = new Date()
  current.setHours(startHour, startMin, 0, 0)
  const end = new Date()
  end.setHours(endHour, endMin, 0, 0)

  const duration = recurring.slot_duration * 60000
  const breakMs = recurring.break_between * 60000

  while (current.getTime() + duration <= end.getTime()) {
    const slotEnd = new Date(current.getTime() + duration)
    slots.push({
      start_label: current.toTimeString().slice(0, 5),
      end_label: slotEnd.toTimeString().slice(0, 5)
    })
    current.setTime(current.getTime() + duration + breakMs)
  }

  return slots
}

const save = async () => {
  if (!form.name.trim()) {
    result.value = { success: false, message: 'Укажите название расписания' }
    return
  }

  if (!props.resourceId) {
    result.value = { success: false, message: 'Ошибка: не выбран ресурс' }
    return
  }

  saving.value = true
  result.value = null

  try {
    const token = localStorage.getItem('token') || ''
    const url = props.initialRule
      ? `/api/schedules/${props.initialRule.id}`
      : '/api/schedules'
    const method = props.initialRule ? 'PUT' : 'POST'

    const payload: any = {
      resource_id: props.resourceId,
      name: form.name,
      recurring: buildRecurring()
    }

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(payload)
    })

    if (response.ok) {
      const data = await response.json()
      result.value = { success: true, message: 'Расписание сохранено' }
      emit('saved', data.id)
    } else {
      const errorData = await response.json()
      result.value = { success: false, message: errorData.error || 'Ошибка сохранения' }
    }
  } catch (err) {
    result.value = { success: false, message: 'Ошибка сети' }
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.schedule-builder {
  padding: 1.5rem;
  background: var(--surface);
  border-radius: var(--radius);
  border: 1px solid var(--border);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-bottom: 1rem;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text);
}

.required {
  color: #c53030;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

input[type="text"],
input[type="time"],
input[type="date"],
input[type="number"],
textarea,
select {
  width: 100%;
  padding: 0.6rem 0.75rem;
  font-size: 0.9rem;
  font-family: inherit;
  color: var(--text);
  background: var(--bg);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  transition: border-color 0.2s;
}

input:focus,
textarea:focus,
select:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(47, 79, 138, 0.1);
}

.days-grid {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.day-checkbox {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0.7rem;
  background: var(--bg);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.day-checkbox:hover {
  border-color: var(--primary);
}

.day-checkbox input {
  accent-color: var(--primary);
}

.day-checkbox input:checked + span {
  font-weight: 600;
  color: var(--primary);
}

.hint {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.preview-section {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--primary-soft);
  border-radius: var(--radius-sm);
}

.preview-section h4 {
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
  color: var(--primary);
}

.preview-slots {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.preview-slot {
  padding: 0.3rem 0.6rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  color: var(--text);
}

.actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1rem;
}

.btn-primary {
  padding: 0.6rem 1.25rem;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-hover, #243f72);
}

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.6rem 1rem;
  background: var(--surface);
  color: var(--text);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: var(--bg);
}

.spinner-small {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  display: inline-block;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.result-message {
  margin-top: 1rem;
  padding: 0.75rem;
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
}

.result-message.success {
  background: #e8f5e9;
  color: #2e7d32;
}

.result-message.error {
  background: #fdecec;
  color: #c53030;
}
</style>
