<template>
  <div class="booking-calendar">
    <div v-if="loading" class="loading-state">
      <div class="spinner-small"></div>
      <span>Загрузка слотов...</span>
    </div>

    <div v-else-if="slots.length === 0" class="empty-state">
      <Icon name="calendar" />
      <p>Нет доступных слотов на эту дату</p>
    </div>

    <div v-else>
      <div class="date-header">
        <button class="nav-btn" @click="changeDate(-1)" :disabled="loading">←</button>
        <span class="date-label">{{ formatDate(selectedDate) }}</span>
        <button class="nav-btn" @click="changeDate(1)" :disabled="loading">→</button>
      </div>

      <div class="slots-grid">
        <button
          v-for="(slot, idx) in slots"
          :key="idx"
          type="button"
          class="slot-btn"
          :class="{ selected: isSelected(slot) }"
          :disabled="!slot.available"
          @click="selectSlot(slot)"
        >
          {{ slot.start_label }} – {{ slot.end_label }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import Icon from './Icon.vue'

const props = defineProps<{
  resourceId: string
}>()

const emit = defineEmits<{
  select: [data: { date: string; start_time: string; end_time: string }]
}>()

interface SlotData {
  start_label: string
  end_label: string
  start_time: string
  end_time: string
  available: boolean
}

const loading = ref(false)
const slots = ref<SlotData[]>([])
const selectedSlot = ref<SlotData | null>(null)
const selectedDate = ref(new Date())
let loadSlotsTimeout: ReturnType<typeof setTimeout> | null = null

// Безопасное форматирование локальной даты (избегает сдвига из-за UTC)
const formatLocalDate = (d: Date): string => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatDate = (d: Date) => {
  return d.toLocaleDateString('ru-RU', {
    weekday: 'long',
    day: 'numeric',
    month: 'long'
  })
}

const isSelected = (slot: SlotData) => {
  if (!selectedSlot.value) return false
  return selectedSlot.value.start_time === slot.start_time
}

const changeDate = (delta: number) => {
  const newDate = new Date(selectedDate.value)
  newDate.setDate(newDate.getDate() + delta)
  selectedDate.value = newDate
  selectedSlot.value = null
  
  // Debounce: если уже есть pending запрос — отменяем его
  if (loadSlotsTimeout) clearTimeout(loadSlotsTimeout)
  loadSlotsTimeout = setTimeout(() => {
    loadSlots()
  }, 200)
}

const selectSlot = (slot: SlotData) => {
  if (!slot.available) return
  selectedSlot.value = slot
  emit('select', {
    date: formatLocalDate(selectedDate.value),
    start_time: slot.start_time,
    end_time: slot.end_time
  })
}

const loadSlots = async () => {
  if (!props.resourceId) return
  loading.value = true
  try {
    const dateStr = formatLocalDate(selectedDate.value)
    const token = localStorage.getItem('token') || ''
    const response = await fetch(
      `/api/schedules/available?resource_id=${props.resourceId}&date=${dateStr}`,
      { headers: token ? { Authorization: `Bearer ${token}` } : {} }
    )
    if (response.ok) {
      const data = await response.json()
      slots.value = (data.slots || []).map((s: SlotData) => ({
        ...s,
        available: true
      }))
      selectedSlot.value = null
    } else {
      slots.value = []
    }
  } catch (err) {
    console.error('[BookingCalendar] Load slots failed:', err)
    slots.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadSlots()
})

watch(() => props.resourceId, () => {
  selectedSlot.value = null
  loadSlots()
})
</script>

<style scoped>
.booking-calendar {
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 1rem;
  background: var(--bg);
}

.loading-state {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem;
  color: var(--text-muted);
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(47, 79, 138, 0.3);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--text-muted);
}

.empty-state svg {
  width: 32px;
  height: 32px;
  margin-bottom: 0.5rem;
  opacity: 0.5;
}

.date-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.nav-btn {
  width: 32px;
  height: 32px;
  border: 1px solid var(--border);
  background: var(--surface);
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  transition: all 0.2s;
}

.nav-btn:hover:not(:disabled) {
  background: var(--primary-soft);
  border-color: var(--primary);
}

.nav-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.date-label {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text);
  text-transform: capitalize;
}

.slots-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 0.5rem;
}

.slot-btn {
  padding: 0.6rem 0.75rem;
  border: 1.5px solid var(--border);
  background: var(--surface);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text);
  cursor: pointer;
  transition: all 0.2s;
}

.slot-btn:hover:not(:disabled) {
  border-color: var(--primary);
  background: var(--primary-soft);
}

.slot-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  text-decoration: line-through;
}

.slot-btn.selected {
  background: var(--primary);
  color: #fff;
  border-color: var(--primary);
}
</style>
