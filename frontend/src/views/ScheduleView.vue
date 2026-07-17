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

    <div class="section-title">📅 Недельное расписание</div>
    <div class="days-presets">
      <button type="button" class="btn-preset" @click="setPreset([1,2,3,4,5])">Будни</button>
      <button type="button" class="btn-preset" @click="setPreset([1,2,3,4,5,6])">Будни + Сб</button>
      <button type="button" class="btn-preset" @click="setPreset([1,2,3,4,5,6,7])">Все дни</button>
    </div>

    <div v-for="day in form.weekly_intervals" :key="day.day_of_week" class="day-row">
      <div class="day-toggle">
        <label class="switch">
          <input type="checkbox" v-model="day.is_working" @change="initIntervals(day)" />
          <span class="slider"></span>
        </label>
        <span class="day-name">{{ dayNames[day.day_of_week - 1] }}</span>
      </div>

      <div v-if="day.is_working" class="day-details">
        <!-- Визуальный таймлайн (08:00 - 20:00) -->
        <div class="timeline-preview">
          <div v-for="(int, idx) in day.intervals" :key="idx" class="timeline-bar" :style="getBarStyle(int)"></div>
        </div>
        
        <div class="intervals-list">
          <div v-for="(int, idx) in day.intervals" :key="idx" class="interval-row">
            <input type="time" v-model="int.start" @change="validateDay(day)" />
            <span>—</span>
            <input type="time" v-model="int.end" @change="validateDay(day)" />
            <button @click="removeInterval(day, idx)" class="btn-icon-delete" title="Удалить интервал">
              <Icon name="trash" />
            </button>
            <span v-if="day.errors && day.errors[idx]" class="error-text">Пересечение!</span>
          </div>
          <button type="button" class="btn-add-interval" @click="addInterval(day)" :disabled="day.intervals.length >= 3">
            <Icon name="plus" /> Добавить интервал
          </button>
        </div>
      </div>
      <div v-else class="day-off">Выходной</div>
    </div>

    <div class="section-title" style="margin-top: 1.5rem;">⚠️ Исключения (разовые дни)</div>
    <div class="exceptions-list">
      <div v-for="(exc, idx) in form.exceptions" :key="idx" class="exception-card">
        <div class="exc-header">
          <input type="date" v-model="exc.date" class="exc-date" />
          <label class="switch small">
            <input type="checkbox" v-model="exc.is_working" />
            <span class="slider"></span>
          </label>
          <span class="exc-status">{{ exc.is_working ? 'Рабочий (особый)' : 'Выходной' }}</span>
          <button @click="removeException(idx)" class="btn-icon-delete"><Icon name="trash" /></button>
        </div>
        <div v-if="exc.is_working" class="exc-intervals">
          <div v-for="(int, iIdx) in exc.intervals" :key="iIdx" class="interval-row">
            <input type="time" v-model="int.start" /> <span>—</span> <input type="time" v-model="int.end" />
            <button @click="exc.intervals.splice(iIdx, 1)" class="btn-icon-delete"><Icon name="trash" /></button>
          </div>
          <button type="button" class="btn-add-interval" @click="exc.intervals.push({start: '09:00', end: '18:00'})">
            <Icon name="plus" /> Добавить время
          </button>
        </div>
      </div>
      <button type="button" class="btn-add-exception" @click="addException">
        <Icon name="calendar" /> Добавить разовый день (например, День открытых дверей)
      </button>
    </div>

    <div class="form-group" style="margin-top: 1rem;">
      <label>Глобальные настройки слотов</label>
      <div class="slot-inputs">
        <div class="input-group">
          <input type="number" v-model.number="form.slot_duration" min="5" max="240" step="5" />
          <span class="input-label">мин на прием</span>
        </div>
        <div class="input-group">
          <input type="number" v-model.number="form.break_between" min="0" max="120" step="5" />
          <span class="input-label">мин перерыв</span>
        </div>
      </div>
    </div>

    <div class="actions">
      <button type="button" class="btn-primary" @click="save" :disabled="saving">
        <span v-if="!saving">{{ initialRule ? 'Обновить' : 'Создать' }}</span>
        <span v-else class="spinner-small"></span>
      </button>
      <button v-if="initialRule" type="button" class="btn-danger" @click="deleteSchedule" :disabled="deleting">
        <Icon name="trash" /> Удалить
      </button>
    </div>
    <div v-if="result" :class="['result-message', result.success ? 'success' : 'error']">{{ result.message }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import Icon from '../components/Icon.vue'

const props = defineProps<{ resourceId: string; initialRule?: any }>()
const emit = defineEmits<{ saved: [ruleId: string]; deleted: [] }>()

const dayNames = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота', 'Воскресенье']

const form = reactive({
  name: '',
  start_date: new Date().toISOString().split('T')[0],
  end_date: new Date(Date.now() + 365 * 86400000).toISOString().split('T')[0],
  slot_duration: 45,
  break_between: 15,
  weekly_intervals: dayNames.map((_, i) => ({
    day_of_week: i + 1,
    is_working: i < 5,
    intervals: [{ start: '09:00', end: '18:00' }],
    errors: [] as boolean[]
  })),
  exceptions: [] as { date: string; is_working: boolean; intervals: { start: string; end: string }[] }[]
})

const saving = ref(false)
const deleting = ref(false)
const result = ref<{ success: boolean; message: string } | null>(null)

// Загрузка данных с миграцией старого формата
watch(() => props.initialRule, (rule) => {
  if (!rule) return
  form.name = rule.name || ''
  form.start_date = rule.recurring?.start_date || form.start_date
  form.end_date = rule.recurring?.end_date || form.end_date
  
  if (rule.recurring) {
    form.slot_duration = rule.recurring.slot_duration || 45
    form.break_between = rule.recurring.break_between || 15
    
    if (rule.recurring.weekly_intervals) {
      form.weekly_intervals = form.weekly_intervals.map((day: any) => {
        const found = rule.recurring.weekly_intervals.find((w: any) => w.day_of_week === day.day_of_week)
        return found ? { ...day, is_working: true, intervals: found.intervals } : { ...day, is_working: false, intervals: [] }
      })
    } else if (rule.recurring.days_config) {
       // Миграция со старого days_config
       form.weekly_intervals = form.weekly_intervals.map((day: any) => {
        const found = rule.recurring.days_config.find((w: any) => w.day === day.day_of_week)
        return found && found.is_working 
          ? { ...day, is_working: true, intervals: [{ start: found.start_time, end: found.end_time }] } 
          : { ...day, is_working: false, intervals: [] }
      })
    }
    
    if (rule.recurring.exceptions) {
      form.exceptions = JSON.parse(JSON.stringify(rule.recurring.exceptions))
    }
  }
}, { immediate: true })

const setPreset = (days: number[]) => {
  form.weekly_intervals.forEach((day: any) => {
    day.is_working = days.includes(day.day_of_week)
    if (day.is_working && day.intervals.length === 0) day.intervals = [{ start: '09:00', end: '18:00' }]
  })
}

const initIntervals = (day: any) => {
  if (day.is_working && day.intervals.length === 0) day.intervals = [{ start: '09:00', end: '18:00' }]
}

const addInterval = (day: any) => {
  const last = day.intervals[day.intervals.length - 1]
  day.intervals.push({ start: last.end, end: '18:00' })
  validateDay(day)
}

const removeInterval = (day: any, idx: number) => {
  day.intervals.splice(idx, 1)
  validateDay(day)
}

const validateDay = (day: any) => {
  day.errors = new Array(day.intervals.length).fill(false)
  const sorted = [...day.intervals].sort((a: any, b: any) => a.start.localeCompare(b.start))
  for (let i = 0; i < sorted.length - 1; i++) {
    if (sorted[i].end > sorted[i+1].start) {
      const idx1 = day.intervals.indexOf(sorted[i])
      const idx2 = day.intervals.indexOf(sorted[i+1])
      day.errors[idx1] = true
      day.errors[idx2] = true
    }
  }
}

const getBarStyle = (int: { start: string; end: string }) => {
  const toMin = (t: string) => {
    const parts = t.split(':')
    const h = parseInt(parts[0] || '0', 10)
    const m = parseInt(parts[1] || '0', 10)
    return h * 60 + m
  }
  const start = toMin(int.start), end = toMin(int.end)
  const dayStart = 480 // 08:00
  const dayLen = 720   // до 20:00
  return { left: `${Math.max(0, ((start - dayStart) / dayLen) * 100)}%`, width: `${Math.max(0, ((end - start) / dayLen) * 100)}%` }
}

const addException = () => {
  const today = new Date().toISOString().split('T')[0] || ''
  form.exceptions.push({ date: today, is_working: true, intervals: [{ start: '09:00', end: '18:00' }] })
}
const removeException = (idx: number) => form.exceptions.splice(idx, 1)

const save = async () => {
  if (!form.name.trim()) { result.value = { success: false, message: 'Укажите название' }; return }
  
  const hasErrors = form.weekly_intervals.some((d: any) => d.is_working && d.errors?.some((e: boolean) => e))
  if (hasErrors) { result.value = { success: false, message: 'Исправьте пересекающиеся интервалы' }; return }

  saving.value = true
  try {
    const token = localStorage.getItem('token') || ''
    const url = props.initialRule ? `/api/schedules/${props.initialRule.id}` : '/api/schedules'
    const method = props.initialRule ? 'PUT' : 'POST'

    const payload = {
      resource_id: props.resourceId,
      name: form.name,
      recurring: {
        start_date: form.start_date,
        end_date: form.end_date,
        weekly_intervals: form.weekly_intervals.filter((d: any) => d.is_working).map((d: any) => ({ day_of_week: d.day_of_week, intervals: d.intervals })),
        exceptions: form.exceptions,
        slot_duration: form.slot_duration,
        break_between: form.break_between
      }
    }

    const res = await fetch(url, { method, headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` }, body: JSON.stringify(payload) })
    const data = await res.json()
    
    if (res.ok) {
      result.value = { success: true, message: 'Сохранено' }
      emit('saved', data.id)
    } else {
      result.value = { success: false, message: data.error || 'Ошибка' }
    }
  } catch (e) {
    result.value = { success: false, message: 'Ошибка сети' }
  } finally {
    saving.value = false
  }
}

const deleteSchedule = async () => {
  if (!confirm('Удалить расписание?')) return
  deleting.value = true
  try {
    const token = localStorage.getItem('token') || ''
    const res = await fetch(`/api/schedules/${props.initialRule.id}`, { method: 'DELETE', headers: { 'Authorization': `Bearer ${token}` } })
    if (res.ok) { result.value = { success: true, message: 'Удалено' }; emit('deleted') }
    else { result.value = { success: false, message: 'Ошибка удаления' } }
  } finally { deleting.value = false }
}
</script>

<style scoped>
.schedule-builder { padding: 1.5rem; background: var(--surface); border-radius: var(--radius); border: 1px solid var(--border); }
.form-group { display: flex; flex-direction: column; gap: 0.4rem; margin-bottom: 1rem; }
.form-group label { font-size: 0.88rem; font-weight: 600; color: var(--text); }
.required { color: #c53030; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-bottom: 1rem; }
input[type="text"], input[type="date"], input[type="time"], input[type="number"] { width: 100%; padding: 0.5rem 0.75rem; font-size: 0.9rem; font-family: inherit; color: var(--text); background: var(--bg); border: 1.5px solid var(--border); border-radius: var(--radius-sm); }
input:focus { outline: none; border-color: var(--primary); }

.section-title { font-size: 1rem; font-weight: 700; color: var(--text); margin: 1.5rem 0 0.75rem 0; display: flex; align-items: center; gap: 0.5rem; }
.days-presets { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.btn-preset { padding: 0.3rem 0.7rem; background: var(--bg); border: 1px solid var(--border); border-radius: 4px; font-size: 0.8rem; cursor: pointer; color: var(--text); }
.btn-preset:hover { background: var(--primary-soft); border-color: var(--primary); color: var(--primary); }

.day-row { display: flex; align-items: flex-start; gap: 1rem; padding: 0.75rem 0; border-bottom: 1px solid var(--border); }
.day-toggle { display: flex; align-items: center; gap: 0.75rem; min-width: 160px; margin-top: 0.5rem; }
.day-name { font-weight: 600; font-size: 0.9rem; color: var(--text); }
.day-details { flex: 1; }
.day-off { font-size: 0.85rem; color: var(--text-muted); font-style: italic; margin-top: 0.5rem; }

.timeline-preview { position: relative; height: 10px; background: #e2e8f0; border-radius: 5px; margin-bottom: 0.75rem; overflow: hidden; }
.timeline-bar { position: absolute; height: 100%; background: var(--primary); border-radius: 5px; opacity: 0.8; }

.intervals-list { display: flex; flex-direction: column; gap: 0.5rem; }
.interval-row { display: flex; align-items: center; gap: 0.5rem; }
.interval-row input { width: 110px; flex: none; }
.btn-icon-delete { width: 32px; height: 32px; border: none; background: transparent; color: #c53030; cursor: pointer; border-radius: 4px; display: flex; align-items: center; justify-content: center; }
.btn-icon-delete:hover { background: #fdecec; }
.error-text { color: #c53030; font-size: 0.75rem; font-weight: 600; margin-left: 0.5rem; }
.btn-add-interval { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.3rem 0.6rem; font-size: 0.8rem; background: var(--primary-soft); color: var(--primary); border: none; border-radius: 4px; cursor: pointer; width: fit-content; margin-top: 0.25rem; }
.btn-add-interval:disabled { opacity: 0.5; cursor: not-allowed; }

.exceptions-list { display: flex; flex-direction: column; gap: 0.75rem; }
.exception-card { padding: 1rem; background: var(--bg); border: 1px solid var(--border); border-radius: var(--radius-sm); }
.exc-header { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.75rem; }
.exc-date { width: 150px; }
.exc-status { font-size: 0.85rem; font-weight: 600; color: var(--text); flex: 1; }
.exc-intervals { padding-left: 1rem; border-left: 2px solid var(--primary-soft); }
.btn-add-exception { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.5rem 1rem; background: var(--surface); border: 1px dashed var(--border); border-radius: var(--radius-sm); color: var(--text-muted); cursor: pointer; width: 100%; justify-content: center; }
.btn-add-exception:hover { border-color: var(--primary); color: var(--primary); background: var(--primary-soft); }

.slot-inputs { display: flex; gap: 1.5rem; margin-top: 0.5rem; }
.input-group { display: flex; align-items: center; gap: 0.5rem; }
.input-group input { width: 70px; text-align: center; }
.input-label { font-size: 0.8rem; color: var(--text-muted); }

.actions { display: flex; gap: 0.75rem; margin-top: 1.5rem; }
.btn-primary, .btn-danger { padding: 0.6rem 1.25rem; border-radius: var(--radius-sm); font-size: 0.9rem; font-weight: 600; cursor: pointer; border: none; display: inline-flex; align-items: center; gap: 0.4rem; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover:not(:disabled) { background: var(--primary-hover, #243f72); }
.btn-danger { background: #fdecec; color: #c53030; border: 1.5px solid #f5c6c6; margin-left: auto; }
.btn-danger:hover:not(:disabled) { background: #c53030; color: #fff; }
.btn-primary:disabled, .btn-danger:disabled { opacity: 0.6; cursor: not-allowed; }
.spinner-small { width: 14px; height: 14px; border: 2px solid rgba(255,255,255,0.3); border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.result-message { margin-top: 1rem; padding: 0.75rem; border-radius: var(--radius-sm); font-size: 0.9rem; }
.result-message.success { background: #e8f5e9; color: #2e7d32; }
.result-message.error { background: #fdecec; color: #c53030; }

.switch { position: relative; display: inline-block; width: 40px; height: 22px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; inset: 0; background-color: #ccc; transition: .4s; border-radius: 22px; }
.slider:before { position: absolute; content: ""; height: 16px; width: 16px; left: 3px; bottom: 3px; background-color: white; transition: .4s; border-radius: 50%; }
input:checked + .slider { background-color: var(--primary); }
input:checked + .slider:before { transform: translateX(18px); }
.switch.small { width: 32px; height: 18px; }
.switch.small .slider:before { height: 12px; width: 12px; left: 3px; bottom: 3px; }
.switch.small input:checked + .slider:before { transform: translateX(14px); }
</style>