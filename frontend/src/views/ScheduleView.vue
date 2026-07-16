<template>
  <div class="schedule-page">
    <div class="page-header">
      <h1 class="page-title">Управление расписанием</h1>
      <div class="header-actions">
        <select v-model="viewMode" class="view-select">
          <option value="analytics">📊 Аналитика загруженности</option>
          <option value="calendar">📅 Сводный календарь</option>
        </select>
      </div>
    </div>

    <!-- ВКЛАДКА 1: АНАЛИТИКА -->
    <div v-if="viewMode === 'analytics'" class="analytics-section">
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>Загрузка данных...</p>
      </div>
      <div v-else-if="analyticsData.length === 0" class="empty-state">
        <Icon name="calendar" />
        <p>Пока нет настроенных расписаний. Создайте первое в разделе "Справочники".</p>
      </div>
      <div v-else class="stats-grid">
        <div v-for="res in analyticsData" :key="res.id" class="stat-card">
          <div class="stat-header">
            <h3>{{ res.name || 'Неизвестный ресурс' }}</h3>
            <span :class="['badge', res.occupancy_percent > 80 ? 'badge-high' : 'badge-normal']">
              {{ res.occupancy_percent }}%
            </span>
          </div>
          <div class="progress-bar-bg">
            <div class="progress-bar-fill" :style="{ width: res.occupancy_percent + '%' }"></div>
          </div>
          <p class="stat-detail">
            Занято: {{ res.booked_slots_week }} из {{ res.total_slots_week }} слотов на эту неделю
          </p>
          <button @click="openBuilder(res.id, res.name)" class="btn-edit-schedule">
            <Icon name="edit" /> Настроить правила
          </button>
        </div>
      </div>
    </div>

    <!-- ВКЛАДКА 2: СВОДНЫЙ КАЛЕНДАРЬ -->
    <div v-else class="calendar-section">
      <div class="calendar-filters">
        <select v-model="selectedResourceId" class="filter-select">
          <option :value="null">Все ресурсы</option>
          <option v-for="res in analyticsData" :key="res.id" :value="res.id">
            {{ res.name }}
          </option>
        </select>
        <div class="week-nav">
          <button class="nav-btn" @click="changeWeek(-1)">← Пред. неделя</button>
          <span class="current-week-label">{{ currentWeekLabel }}</span>
          <button class="nav-btn" @click="changeWeek(1)">След. неделя →</button>
        </div>
      </div>

      <div class="timetable-wrapper">
        <table class="timetable">
          <thead>
            <tr>
              <th class="col-resource">Ресурс</th>
              <th v-for="(day, idx) in weekDays" :key="idx">{{ day.label }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="res in filteredAnalytics" :key="res.id">
              <td class="col-resource">
                <strong>{{ res.name }}</strong>
                <button @click="openBuilder(res.id, res.name)" class="btn-tiny-edit">
                  <Icon name="edit" />
                </button>
              </td>
              <td v-for="(day, idx) in weekDays" :key="idx" class="day-cell">
                <div v-if="res.booked_slots_week > 0" class="booking-indicator">
                  <span class="booking-count">{{ res.booked_slots_week }} записей</span>
                </div>
                <div v-else class="no-booking">Свободно</div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Модальное окно с ScheduleBuilder -->
    <div v-if="showBuilderModal" class="modal-overlay" @click="showBuilderModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>Расписание: {{ currentEditingResourceName }}</h3>
          <button @click="showBuilderModal = false" class="btn-close">
            <Icon name="close" />
          </button>
        </div>
        <ScheduleBuilder 
          :resource-id="currentEditingResourceId" 
          @saved="onScheduleSaved"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import Icon from '../components/Icon.vue'
import ScheduleBuilder from '../components/ScheduleBuilder.vue'

const viewMode = ref('analytics')
const analyticsData = ref<any[]>([])
const loading = ref(false)
const showBuilderModal = ref(false)
const currentEditingResourceId = ref('')
const currentEditingResourceName = ref('')
const selectedResourceId = ref<string | null>(null)
const weekOffset = ref(0)

const weekDays = computed(() => {
  const days = []
  const baseDate = new Date()
  baseDate.setDate(baseDate.getDate() + (weekOffset.value * 7))
  const currentDay = baseDate.getDay()
  const diff = baseDate.getDate() - currentDay + (currentDay === 0 ? -6 : 1) // Понедельник
  
  for (let i = 0; i < 7; i++) {
    const d = new Date(baseDate)
    d.setDate(diff + i)
    days.push({
      date: d,
      label: d.toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric' })
    })
  }
  return days
})

const currentWeekLabel = computed(() => {
  const start = weekDays.value[0].date
  const end = weekDays.value[6].date
  return `${start.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })} — ${end.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })}`
})

const filteredAnalytics = computed(() => {
  if (!selectedResourceId.value) return analyticsData.value
  return analyticsData.value.filter(r => r.id === selectedResourceId.value)
})

onMounted(async () => {
  await loadAnalytics()
})

const loadAnalytics = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('token') || ''
    const response = await fetch('/api/schedules/analytics', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (response.ok) {
      const data = await response.json()
      analyticsData.value = data.resources || []
    }
  } catch (err) {
    console.error('[Schedule] Load analytics failed:', err)
  } finally {
    loading.value = false
  }
}

const changeWeek = (delta: number) => {
  weekOffset.value += delta
}

const openBuilder = (id: string, name: string) => {
  currentEditingResourceId.value = id
  currentEditingResourceName.value = name
  showBuilderModal.value = true
}

const onScheduleSaved = () => {
  showBuilderModal.value = false
  loadAnalytics()
}
</script>

<style scoped>
.schedule-page { width: 100%; max-width: 1200px; margin: 0 auto; padding: 2rem; font-family: inherit; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; flex-wrap: wrap; gap: 1rem; }
.page-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; }
.view-select, .filter-select { padding: 0.6rem 1rem; border: 1.5px solid var(--border); border-radius: var(--radius-sm); font-family: inherit; background: var(--surface); color: var(--text); font-size: 0.9rem; }
.view-select:focus, .filter-select:focus { outline: none; border-color: var(--primary); }

.analytics-section { animation: fadeUp 0.4s ease both; }
.loading-state, .empty-state { text-align: center; padding: 4rem 2rem; background: var(--surface); border-radius: var(--radius); border: 1px solid var(--border); color: var(--text-muted); }
.spinner { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
@keyframes spin { to { transform: rotate(360deg); } }
.empty-state svg { width: 48px; height: 48px; margin-bottom: 1rem; opacity: 0.5; }

.stats-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 1.5rem; }
.stat-card { background: var(--surface); padding: 1.5rem; border-radius: var(--radius); border: 1px solid var(--border); box-shadow: var(--shadow-sm); transition: transform 0.2s; }
.stat-card:hover { transform: translateY(-2px); box-shadow: var(--shadow-md); }
.stat-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
.stat-header h3 { font-size: 1.1rem; font-weight: 700; color: var(--text); margin: 0; }
.badge { padding: 0.25rem 0.6rem; border-radius: 6px; font-size: 0.85rem; font-weight: 700; }
.badge-high { background: #fdecec; color: #c53030; }
.badge-normal { background: #e8f5e9; color: #2e7d32; }
.progress-bar-bg { width: 100%; height: 8px; background: var(--bg); border-radius: 4px; overflow: hidden; margin-bottom: 0.75rem; }
.progress-bar-fill { height: 100%; background: var(--primary); transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1); }
.stat-detail { font-size: 0.85rem; color: var(--text-muted); margin-bottom: 1.25rem; }
.btn-edit-schedule { width: 100%; padding: 0.6rem; background: var(--primary-soft); color: var(--primary); border: none; border-radius: var(--radius-sm); cursor: pointer; font-weight: 600; display: flex; align-items: center; justify-content: center; gap: 0.5rem; transition: all 0.2s; font-family: inherit; }
.btn-edit-schedule:hover { background: var(--primary); color: #fff; }
.btn-edit-schedule svg { width: 16px; height: 16px; }

.calendar-section { animation: fadeUp 0.4s ease both; }
.calendar-filters { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; flex-wrap: wrap; gap: 1rem; }
.week-nav { display: flex; align-items: center; gap: 1rem; }
.nav-btn { padding: 0.5rem 1rem; background: var(--surface); border: 1.5px solid var(--border); border-radius: var(--radius-sm); cursor: pointer; font-weight: 600; color: var(--text); transition: all 0.2s; font-family: inherit; }
.nav-btn:hover { background: var(--primary-soft); border-color: var(--primary); color: var(--primary); }
.current-week-label { font-weight: 700; color: var(--text); font-size: 1rem; min-width: 180px; text-align: center; }

.timetable-wrapper { overflow-x: auto; background: var(--surface); border-radius: var(--radius); border: 1px solid var(--border); }
.timetable { width: 100%; border-collapse: collapse; min-width: 800px; }
.timetable th, .timetable td { padding: 1rem; text-align: left; border-bottom: 1px solid var(--border); }
.timetable th { background: var(--bg); font-weight: 600; color: var(--text-muted); font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.05em; }
.col-resource { width: 250px; font-weight: 600; color: var(--text); display: flex; align-items: center; justify-content: space-between; }
.btn-tiny-edit { width: 28px; height: 28px; border: none; background: transparent; color: var(--text-muted); cursor: pointer; border-radius: 4px; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.btn-tiny-edit:hover { background: var(--primary-soft); color: var(--primary); }
.btn-tiny-edit svg { width: 14px; height: 14px; }
.day-cell { text-align: center; }
.booking-indicator { background: var(--primary-soft); color: var(--primary); padding: 0.4rem 0.8rem; border-radius: 6px; font-size: 0.85rem; font-weight: 600; display: inline-block; }
.no-booking { color: var(--text-muted); font-size: 0.85rem; font-style: italic; }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; backdrop-filter: blur(2px); animation: fadeIn 0.2s ease; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
.modal-content { background: var(--surface); padding: 2rem; border-radius: var(--radius); width: 90%; max-width: 600px; max-height: 90vh; overflow-y: auto; box-shadow: var(--shadow-lg); }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
.modal-header h3 { font-size: 1.25rem; font-weight: 700; color: var(--text); margin: 0; }
.btn-close { width: 32px; height: 32px; border: none; background: var(--bg); color: var(--text-muted); border-radius: 50%; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.btn-close:hover { background: #fdecec; color: #c53030; }
.btn-close svg { width: 16px; height: 16px; }

@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 720px) {
  .stats-grid { grid-template-columns: 1fr; }
  .calendar-filters { flex-direction: column; align-items: stretch; }
  .week-nav { justify-content: space-between; }
}
</style>