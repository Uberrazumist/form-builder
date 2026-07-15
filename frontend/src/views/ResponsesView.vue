<template>
  <div class="responses-page">
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Загрузка ответов...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <div class="error-icon">
        <Icon name="error" />
      </div>
      <h2>Ошибка</h2>
      <p>{{ error }}</p>
      <router-link to="/" class="btn-secondary">На главную</router-link>
    </div>

    <div v-else class="responses-container">
      <div class="page-header">
        <h1 class="page-title">Ответы на форму</h1>
        <p class="page-subtitle">{{ form?.title || 'Форма' }}</p>
      </div>

      <div v-if="responses.length === 0" class="empty-state">
        <div class="empty-icon">
          <Icon name="document" />
        </div>
        <h3>Пока нет ответов</h3>
        <p>Когда пользователи заполнят форму, их ответы появятся здесь</p>
      </div>

      <div v-else class="table-container">
        <div class="table-header">
          <span class="response-count">Всего ответов: {{ responses.length }}</span>
          <button @click="exportCSV" class="btn-secondary-small">
            <Icon name="download" />
            Экспорт CSV
          </button>
        </div>
        <div class="table-wrapper">
          <table class="responses-table">
            <thead>
              <tr>
                <th class="col-date">Дата</th>
                <th
                  v-for="question in form?.questions || []"
                  :key="question.id"
                >
                  {{ question.title }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="response in responses" :key="response.id">
                <td class="col-date">{{ formatDate(response.created_at) }}</td>
                <td
                  v-for="question in form?.questions || []"
                  :key="question.id"
                >
                  {{ formatAnswer(response.answers?.[question.id], question.type) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'

interface Question {
  id: string
  type: string
  title: string
  is_required: boolean
  dictionary_id: string | null
  is_booking: boolean
  depends_on: string | null
  options: any[]
  rating_max: number
  order_index: number
}

interface Form {
  id: string
  title: string
  description: string
  is_public: boolean
  questions: Question[]
}

interface Response {
  id: string
  form_id: string
  user_id: string | null
  answers: Record<string, any>
  created_at: string
}

const route = useRoute()
const router = useRouter()

const form = ref<Form | null>(null)
const responses = ref<Response[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (!token) {
    router.push('/login')
    return
  }
  await loadData()
})

const loadData = async () => {
  loading.value = true
  error.value = null

  try {
    const formId = String(route.params.id)
    const token = localStorage.getItem('token')
    const headers: Record<string, string> = { 'Authorization': `Bearer ${token || ''}` }

    // Загрузка формы
    const formResponse = await fetch(`/api/forms/${formId}`, { headers })
    if (!formResponse.ok) {
      if (formResponse.status === 404) error.value = 'Форма не найдена'
      else if (formResponse.status === 403) error.value = 'У вас нет доступа к этой форме'
      else error.value = 'Не удалось загрузить форму'
      return
    }
    form.value = await formResponse.json()

    // Загрузка ответов
    const responsesResponse = await fetch(`/api/forms/${formId}/responses`, { headers })
    if (!responsesResponse.ok) {
      error.value = 'Не удалось загрузить ответы'
      return
    }
    const data = await responsesResponse.json()
    responses.value = Array.isArray(data) ? data : (data.responses || [])
  } catch (err) {
    console.error('[Responses] Load error:', err)
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return ''
  try {
    const date = new Date(dateStr)
    return date.toLocaleString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return dateStr
  }
}

const formatAnswer = (answer: any, type: string): string => {
  if (answer === null || answer === undefined || answer === '') return '—'
  if (Array.isArray(answer)) return answer.join(', ')
  return String(answer)
}

const exportCSV = () => {
  if (!form.value || responses.value.length === 0) return

  const headers = ['Дата', ...(form.value.questions || []).map(q => q.title)]
  const rows = responses.value.map(r => {
    const row = [formatDate(r.created_at)]
    for (const q of form.value?.questions || []) {
      row.push(formatAnswer(r.answers?.[q.id], q.type))
    }
    return row
  })

  const csvContent = [
    headers.join(','),
    ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
  ].join('\n')

  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `responses_${form.value.id}_${Date.now()}.csv`
  link.click()
}
</script>

<style scoped>
.responses-page { width: 100%; max-width: 1200px; margin: 0 auto; }
.loading-state, .error-state { text-align: center; padding: 4rem 2rem; }
.spinner { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon { width: 64px; height: 64px; background: #fdecec; color: #c53030; border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.error-icon svg { width: 32px; height: 32px; }
.error-state h2 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.error-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.btn-secondary { display: inline-block; padding: 0.75rem 1.5rem; background: var(--surface); color: var(--text); border: 1.5px solid var(--border); border-radius: var(--radius-sm); text-decoration: none; font-weight: 600; transition: all 0.2s; }
.btn-secondary:hover { background: var(--bg); border-color: #cfd6e3; }
.responses-container { animation: fadeUp 0.5s ease both; }
.page-header { margin-bottom: 2rem; }
.page-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; margin-bottom: 0.5rem; }
.page-subtitle { font-size: 1.05rem; color: var(--text-muted); }
.empty-state { text-align: center; padding: 4rem 2rem; background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); }
.empty-icon { width: 64px; height: 64px; background: var(--primary-soft); color: var(--primary); border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.empty-icon svg { width: 32px; height: 32px; }
.empty-state h3 { font-size: 1.25rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.empty-state p { color: var(--text-muted); }
.table-container { background: var(--surface); border-radius: var(--radius); border: 1px solid var(--border); overflow: hidden; }
.table-header { display: flex; justify-content: space-between; align-items: center; padding: 1.25rem 1.5rem; border-bottom: 1px solid var(--border); }
.response-count { font-size: 0.95rem; font-weight: 600; color: var(--text); }
.btn-secondary-small { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.5rem 1rem; background: var(--bg); color: var(--text); border: 1px solid var(--border); border-radius: var(--radius-sm); font-size: 0.85rem; font-weight: 500; cursor: pointer; transition: all 0.2s; }
.btn-secondary-small:hover { background: var(--surface); border-color: #cfd6e3; }
.btn-secondary-small svg { width: 16px; height: 16px; }
.table-wrapper { overflow-x: auto; }
.responses-table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
.responses-table th, .responses-table td { padding: 1rem 1.25rem; text-align: left; border-bottom: 1px solid var(--border); }
.responses-table th { background: var(--bg); font-weight: 600; color: var(--text); white-space: nowrap; }
.responses-table td { color: var(--text); }
.responses-table tbody tr:hover { background: var(--bg); }
.col-date { width: 180px; white-space: nowrap; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 720px) {
  .page-title { font-size: 1.5rem; }
  .table-header { flex-direction: column; gap: 1rem; align-items: flex-start; }
  .responses-table th, .responses-table td { padding: 0.75rem 1rem; font-size: 0.85rem; }
}
</style>
