<template>
  <div class="schedule-builder">
    <div class="form-group">
      <label>Название расписания <span class="required">*</span></label>
      <input type="text" v-model="form.name" placeholder="Например: Расписание Иванова" />
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

    <div class="days-config-section">
      <div class="days-header">
        <h4>Расписание по дням недели</h4>
        <div class="days-presets">
          <button type="button" class="btn-preset" @click="setWorkdays([1,2,3,4,5])">Будни</button>
          <button type="button" class="btn-preset" @click="setWorkdays([1,2,3,4,5,6])">Будни + сб</button>
          <button type="button" class="btn-preset" @click="setWorkdays([1,2,3,4,5,6,7])">Все дни</button>
        </div>
      </div>

      <div v-for="config in form.days_config" :key="config.day" class="day-row">
        <div class="day-toggle">
          <label class="switch">
            <input type="checkbox" v-model="config.is_working" />
            <span class="slider"></span>
          </label>
          <span class="day-name">{{ dayNames[config.day - 1] }}</span>
        </div>

        <div v-if="config.is_working" class="day-details">
          <div class="time-inputs">
            <input type="time" v-model="config.start_time" />
            <span class="separator">—</span>
            <input type="time" v-model="config.end_time" />
          </div>
          <div class="slot-inputs">
            <div class="input-group">
              <input type="number" v-model.number="config.slot_duration" min="5" max="240" step="5" />
              <span class="input-label">мин слот</span>
            </div>
            <div class="input-group">
              <input type="number" v-model.number="config.break_between" min="0" max="120" step="5" />
              <span class="input-label">мин перерыв</span>
            </div>
          </div>
        </div>
        <div v-else class="day-off">Выходной</div>
      </div>
    </div>

    <div class="form-group">
      <label>Исключения (праздники, отпуска)</label>
      <textarea
        v-model="exceptionsText"
        placeholder="Каждая дата на новой строке: 2026-12-31"
        rows="2"
      ></textarea>
      <span class="hint">Даты, когда расписание не действует</span>
    </div>

    <div v-if="previewSlots.length > 0" class="preview-section">
      <h4>Предпросмотр на сегодня</h4>
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
        <span v-if="!saving">{{ initialRule ? 'Обновить расписание' : 'Создать расписание' }}</span>
        <span v-else class="spinner-small"></span>
      </button>
      <button type="button" class="btn-secondary" @click="generatePreview">
        <Icon name="calendar" />
        Предпросмотр
      </button>
      <button v-if="initialRule" type="button" class="btn-danger" @click="deleteSchedule" :disabled="deleting">
        <Icon name="trash" />
        Удалить
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
  dictionaries?: Array<{ id: string; name: string }>
  initialRule?: any
  resourceId: string
}>()

const emit = defineEmits<{
  saved: [ruleId: string]
  deleted: []
}>()

const dayNames = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота', 'Воскресенье']

const form = reactive({
  name: '',
  start_date: new Date().toISOString().split('T')[0],
  end_date: new Date(Date.now() + 365 * 86400000).toISOString().split('T')[0],
  days_config: [
    { day: 1, is_working: true, start_time: '09:00', end_time: '18:00', slot_duration: 45, break_between: 15 },
    { day: 2, is_working: true, start_time: '09:00', end_time: '18:00', slot_duration: 45, break_between: 15 },
    { day: 3, is_working: true, start_time: '09:00', end_time: '18:00', slot_duration: 45, break_between: 15 },
    { day: 4, is_working: true, start_time: '09:00', end_time: '18:00', slot_duration: 45, break_between: 15 },
    { day: 5, is_working: true, start_time: '09:00', end_time: '18:00', slot_duration: 45, break_between: 15 },
    { day: 6, is_working: false, start_time: '09:00', end_time: '18:00', slot_duration: 45, break_between: 15 },
    { day: 7, is_working: false, start_time: '09:00', end_time: '18:00', slot_duration: 45, break_between: 15 }
  ]
})

const exceptionsText = ref('')
const previewSlots = ref<any[]>([])
const saving = ref(false)
const deleting = ref(false)
const result = ref<{ success: boolean; message: string } | null>(null)

const previewDayLabel = computed(() => {
  const today = new Date().getDay() || 7
  return dayNames[today - 1]
})

// Загрузка существующих данных с миграцией старого формата
watch(() => props.initialRule, (rule) => {
  if (!rule) return
  form.name = rule.name || ''
  form.start_date = rule.recurring?.start_date || form.start_date
  form.end_date = rule.recurring?.end_date || form.end_date
  exceptionsText.value = (rule.recurring?.exceptions || []).join('\n')

  // Новый формат: days_config
  if (rule.recurring?.days_config && rule.recurring.days_config.length > 0) {
    form.days_config = rule.recurring.days_config
  }
  // Старый формат: flat days + global times → миграция в days_config
  else if (rule.recurring?.days && rule.recurring.days.length > 0) {
    const workingDays = rule.recurring.days
    form.days_config.forEach(config => {
      config.is_working = workingDays.includes(config.day)
      config.start_time = rule.recurring.start_time || '09:00'
      config.end_time = rule.recurring.end_time || '18:00'
      config.slot_duration = rule.recurring.slot_duration || 45
      config.break_between = rule.recurring.break_between || 15
    })
  }
}, { immediate: true })

// Установить рабочие дни (пресет)
const setWorkdays = (days: number[]) => {
  form.days_config.forEach(config => {
    config.is_working = days.includes(config.day)
  })
}

// Собрать recurring JSON для отправки на бэкенд
const buildRecurring = () => ({
  type: 'weekly',
  start_date: form.start_date,
  end_date: form.end_date,
  exceptions: exceptionsText.value.split('\n').map((s: string) => s.trim()).filter(Boolean),
  days_config: form.days_config.map(c => ({
    day: c.day,
    is_working: c.is_working,
    start_time: c.start_time,
    end_time: c.end_time,
    slot_duration: c.slot_duration,
    break_between: c.break_between
  }))
})

// Генерация предпросмотра слотов для текущего дня
const generatePreview = () => {
  const today = new Date().getDay() || 7
  const todayConfig = form.days_config.find(c => c.day === today)
  previewSlots.value = []

  if (!todayConfig || !todayConfig.is_working) return

  const startTimeParts = (todayConfig.start_time || '09:00').split(':')
  const endTimeParts = (todayConfig.end_time || '18:00').split(':')
  const startH = parseInt(startTimeParts[0], 10)
  const startM = parseInt(startTimeParts[1], 10)
  const endH = parseInt(endTimeParts[0], 10)
  const endM = parseInt(endTimeParts[1], 10)

  const current = new Date()
  current.setHours(startH, startM, 0, 0)
  const end = new Date()
  end.setHours(endH, endM, 0, 0)

  const duration = (todayConfig.slot_duration || 45) * 60000
  const breakMs = (todayConfig.break_between || 15) * 60000

  while (current.getTime() + duration <= end.getTime()) {
    const slotEnd = new Date(current.getTime() + duration)
    previewSlots.value.push({
      start_label: current.toTimeString().slice(0, 5),
      end_label: slotEnd.toTimeString().slice(0, 5)
    })
    current.setTime(current.getTime() + duration + breakMs)
  }
}

// Сохранение
const save = async () => {
  if (!form.name.trim()) {
    result.value = { success: false, message: 'Укажите название расписания' }
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

// Удаление расписания (soft-delete)
const deleteSchedule = async () => {
  if (!confirm('Вы уверены, что хотите удалить это расписание?\n\nСуществующие записи в расписании останутся в истории, но новые слоты генерироваться перестанут.')) return

  deleting.value = true
  result.value = null

  try {
    const token = localStorage.getItem('token') || ''
    const response = await fetch(`/api/schedules/${props.initialRule.id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (response.ok) {
      result.value = { success: true, message: 'Расписание удалено' }
      emit('deleted')
    } else {
      const errorData = await response.json()
      result.value = { success: false, message: errorData.error || 'Ошибка удаления' }
    }
  } catch (err) {
    result.value = { success: false, message: 'Ошибка сети' }
  } finally {
    deleting.value = false
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
  margin-bottom: 1rem;
}

input[type="text"],
input[type="date"],
input[type="time"],
input[type="number"],
textarea {
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
textarea:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(47, 79, 138, 0.1);
}

.days-config-section {
  margin: 1.5rem 0;
  padding: 1rem;
  background: var(--bg);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}

.days-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.days-header h4 {
  margin: 0;
  font-size: 0.95rem;
  color: var(--text);
}

.days-presets {
  display: flex;
  gap: 0.4rem;
}

.btn-preset {
  padding: 0.35rem 0.7rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  color: var(--text);
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.btn-preset:hover {
  background: var(--primary-soft);
  border-color: var(--primary);
  color: var(--primary);
}

.day-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border);
}

.day-row:last-child {
  border-bottom: none;
}

.day-toggle {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 160px;
}

.day-name {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text);
}

.day-details {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  flex: 1;
  flex-wrap: wrap;
}

.time-inputs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.time-inputs input {
  width: 100px;
}

.separator {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.slot-inputs {
  display: flex;
  gap: 1rem;
}

.input-group {
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.input-group input {
  width: 60px;
  text-align: center;
  padding: 0.4rem 0.5rem;
}

.input-label {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.day-off {
  font-size: 0.85rem;
  color: var(--text-muted);
  font-style: italic;
}

/* Toggle Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: .4s;
  border-radius: 22px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--primary);
}

input:checked + .slider:before {
  transform: translateX(18px);
}

.preview-section {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--primary-soft);
  border-radius: var(--radius-sm);
}

.preview-section h4 {
  margin: 0 0 0.5rem 0;
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
}

.actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.5rem;
  flex-wrap: wrap;
}

.btn-primary,
.btn-secondary,
.btn-danger {
  padding: 0.6rem 1.25rem;
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-family: inherit;
  border: none;
}

.btn-primary {
  background: var(--primary);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-hover, #243f72);
}

.btn-secondary {
  background: var(--surface);
  color: var(--text);
  border: 1.5px solid var(--border);
}

.btn-secondary:hover {
  background: var(--bg);
}

.btn-danger {
  background: #fdecec;
  color: #c53030;
  border: 1.5px solid #f5c6c6;
  margin-left: auto;
}

.btn-danger:hover:not(:disabled) {
  background: #c53030;
  color: #fff;
}

.btn-primary:disabled,
.btn-danger:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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

.btn-danger .spinner-small {
  border-color: rgba(197, 48, 48, 0.3);
  border-top-color: #c53030;
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

.hint {
  font-size: 0.8rem;
  color: var(--text-muted);
}

@media (max-width: 720px) {
  .day-row {
    flex-direction: column;
    align-items: flex-start;
  }
  .day-details {
    flex-direction: column;
    gap: 0.75rem;
  }
  .days-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
