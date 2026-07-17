<template>
  <div class="schedule-analytics">
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Загрузка аналитики...</p>
    </div>

    <div v-else class="analytics-container">
      <h2 class="page-title">📊 Аналитика расписаний</h2>

      <!-- Сводка по ресурсам -->
      <div class="stats-grid">
        <div v-for="item in analyticsData" :key="item.id" class="stat-card">
          <div class="stat-header">
            <h3>{{ item.name || 'Неизвестный ресурс' }}</h3>
            <span class="stat-id">{{ item.id }}</span>
          </div>
          
          <div class="stat-bars">
            <div class="bar-row">
              <span class="bar-label">Занято:</span>
              <div class="bar-track">
                <div class="bar-filled booked" :style="{ width: `${getOccupancyPercent(item)}%` }"></div>
              </div>
              <span class="bar-value">{{ item.booked_slots_week }}/{{ item.total_slots_week }}</span>
            </div>
            
            <div class="bar-row">
              <span class="bar-label">Загруженность:</span>
              <div class="bar-track">
                <div class="bar-filled" :class="getOccupancyClass(item)" :style="{ width: `${getOccupancyPercent(item)}%` }"></div>
              </div>
              <span class="bar-value">{{ Math.round(getOccupancyPercent(item)) }}%</span>
            </div>
          </div>

          <div class="stat-details">
            <div class="detail-row">
              <span>Всего слотов на неделю:</span>
              <strong>{{ item.total_slots_week }}</strong>
            </div>
            <div class="detail-row">
              <span>Забронировано:</span>
              <strong>{{ item.booked_slots_week }}</strong>
            </div>
            <div class="detail-row">
              <span>Свободно:</span>
              <strong>{{ item.total_slots_week - item.booked_slots_week }}</strong>
            </div>
          </div>
        </div>

        <div v-if="analyticsData.length === 0" class="empty-state">
          <p>Нет расписаний для отображения</p>
        </div>
      </div>

      <!-- Список всех расписаний -->
      <div class="schedules-list-section">
        <h3 class="section-title">📋 Все расписания</h3>
        
        <div v-if="scheduleRules.length === 0" class="empty-state">
          <p>Расписания не найдены</p>
        </div>

        <div v-else class="schedules-list">
          <div v-for="rule in scheduleRules" :key="rule.id" class="schedule-card">
            <div class="schedule-header">
              <h4>{{ rule.name }}</h4>
              <span class="schedule-status" :class="rule.is_deleted ? 'deleted' : 'active'">
                {{ rule.is_deleted ? 'Удалено' : 'Активно' }}
              </span>
            </div>
            
            <div class="schedule-meta">
              <span>📅 Создано: {{ formatDate(rule.created_at) }}</span>
              <span>🔄 Обновлено: {{ formatDate(rule.updated_at) }}</span>
            </div>

            <!-- Предпросмотр слотов для сегодня -->
            <div v-if="!rule.is_deleted" class="schedule-preview">
              <div class="preview-header">
                <span>Предпросмотр на сегодня:</span>
                <button @click="loadSlotsForRule(rule)" class="btn-small">
                  <Icon name="refresh" /> Обновить
                </button>
              </div>
              <div v-if="loadingSlots[rule.id]" class="preview-loading">
                <div class="mini-spinner"></div>
              </div>
              <div v-else-if="slotsForRule[rule.id] && slotsForRule[rule.id]!.length > 0" class="preview-slots">
                <span v-for="(slot, idx) in (slotsForRule[rule.id] || [])" :key="idx" class="preview-slot">
                  {{ slot.start_label }}–{{ slot.end_label }}
                </span>
              </div>
              <div v-else class="preview-empty">
                Нет слотов на сегодня (проверьте расписание)
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Icon from '../components/Icon.vue'

interface ResourceAnalytics {
  id: string
  name: string
  booked_slots_week: number
  total_slots_week: number
  occupancy_percent: number
}

interface ScheduleRule {
  id: string
  resource_id: string
  name: string
  is_deleted: boolean
  created_at: string
  updated_at: string
}

const loading = ref(true)
const loadingSlots = ref<Record<string, boolean>>({})
const analyticsData = ref<ResourceAnalytics[]>([])
const scheduleRules = ref<ScheduleRule[]>([])
const slotsForRule = ref<Record<string, any[]>>({})

const loadData = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('token') || ''
    
    // Загрузка аналитики
    const analyticsRes = await fetch('/api/schedules/analytics', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (analyticsRes.ok) {
      const data = await analyticsRes.json()
      analyticsData.value = data.resources || []
    }

    // Загрузка всех расписаний
    const rulesRes = await fetch('/api/schedules', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (rulesRes.ok) {
      scheduleRules.value = await rulesRes.json()
      
      // Загружаем предпросмотр слотов для каждого активного расписания
      for (const rule of scheduleRules.value) {
        if (!rule.is_deleted) {
          loadSlotsForRule(rule)
        }
      }
    }
  } catch (e) {
    console.error('Ошибка загрузки аналитики:', e)
  } finally {
    loading.value = false
  }
}

const loadSlotsForRule = async (rule: ScheduleRule) => {
  loadingSlots.value[rule.id] = true
  try {
    const today = new Date().toISOString().split('T')[0]
    const token = localStorage.getItem('token') || ''
    const res = await fetch(`/api/schedules/available?resource_id=${rule.resource_id}&date=${today}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (res.ok) {
      const data = await res.json()
      slotsForRule.value[rule.id] = data.slots || []
    }
  } catch (e) {
    console.error('Ошибка загрузки слотов:', e)
  } finally {
    loadingSlots.value[rule.id] = false
  }
}

const getOccupancyPercent = (item: ResourceAnalytics): number => {
  if (item.total_slots_week === 0) return 0
  return (item.booked_slots_week / item.total_slots_week) * 100
}

const getOccupancyClass = (item: ResourceAnalytics): string => {
  const percent = getOccupancyPercent(item)
  if (percent < 30) return 'low'
  if (percent < 70) return 'medium'
  return 'high'
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '—'
  try {
    return new Date(dateStr).toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    })
  } catch {
    return dateStr
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.schedule-analytics {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  animation: fadeUp 0.5s ease both;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.loading-state {
  text-align: center;
  padding: 4rem 2rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 2rem;
  font-family: inherit;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.stat-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1.5rem;
  transition: all 0.2s;
}

.stat-card:hover {
  box-shadow: var(--shadow);
}

.stat-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
  gap: 0.5rem;
}

.stat-header h3 {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
  font-family: inherit;
}

.stat-id {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-family: 'SF Mono', Menlo, monospace;
  word-break: break-all;
}

.stat-bars {
  margin-bottom: 1.5rem;
}

.bar-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.bar-label {
  font-size: 0.85rem;
  color: var(--text-muted);
  min-width: 100px;
  font-family: inherit;
}

.bar-track {
  flex: 1;
  height: 12px;
  background: #e2e8f0;
  border-radius: 6px;
  overflow: hidden;
}

.bar-filled {
  height: 100%;
  border-radius: 6px;
  transition: width 0.3s ease;
}

.bar-filled.booked {
  background: var(--primary);
}

.bar-filled.low {
  background: #22c55e;
}

.bar-filled.medium {
  background: #f59e0b;
}

.bar-filled.high {
  background: #ef4444;
}

.bar-value {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text);
  min-width: 80px;
  text-align: right;
  font-family: inherit;
}

.stat-details {
  border-top: 1px solid var(--border);
  padding-top: 1rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
  font-family: inherit;
}

.detail-row strong {
  color: var(--primary);
}

.schedules-list-section {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1.5rem;
}

.section-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 1.5rem 0;
  font-family: inherit;
}

.schedules-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.schedule-card {
  padding: 1.25rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  transition: all 0.2s;
}

.schedule-card:hover {
  border-color: var(--primary);
}

.schedule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  gap: 1rem;
}

.schedule-header h4 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
  font-family: inherit;
}

.schedule-status {
  font-size: 0.8rem;
  font-weight: 600;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-family: inherit;
}

.schedule-status.active {
  background: #e8f5e9;
  color: #2e7d32;
}

.schedule-status.deleted {
  background: #fdecec;
  color: #c53030;
}

.schedule-meta {
  display: flex;
  gap: 1.5rem;
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 1rem;
  font-family: inherit;
}

.schedule-preview {
  border-top: 1px solid var(--border);
  padding-top: 1rem;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.preview-header span {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text);
  font-family: inherit;
}

.btn-small {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.35rem 0.75rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  color: var(--text-muted);
  cursor: pointer;
  font-family: inherit;
  transition: all 0.2s;
}

.btn-small:hover {
  border-color: var(--primary);
  color: var(--primary);
}

.btn-small svg {
  width: 14px;
  height: 14px;
}

.preview-loading {
  text-align: center;
  padding: 1rem;
}

.mini-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  margin: 0 auto;
}

.preview-slots {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.preview-slot {
  padding: 0.3rem 0.6rem;
  background: var(--primary-soft);
  border: 1px solid var(--primary-soft);
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  color: var(--primary);
  font-family: inherit;
}

.preview-empty {
  font-size: 0.85rem;
  color: var(--text-muted);
  font-style: italic;
  font-family: inherit;
}

.empty-state {
  text-align: center;
  padding: 3rem 2rem;
  color: var(--text-muted);
  font-family: inherit;
}

@media (max-width: 720px) {
  .schedule-analytics {
    padding: 1rem;
  }
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .schedule-meta {
    flex-direction: column;
    gap: 0.5rem;
  }
}
</style>