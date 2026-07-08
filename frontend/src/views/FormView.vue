<template>
  <div class="form-view-page">
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Загрузка...</p>
    </div>

    <div v-else-if="form">
      <div class="page-header">
        <h1 class="page-title">{{ form.title }}</h1>
        <div class="status-badge" :class="{ public: form.is_public, private: !form.is_public }">
          {{ form.is_public ? 'Публичная' : 'Закрытая' }}
        </div>
      </div>

      <p v-if="form.description" class="form-description">{{ form.description }}</p>

      <div class="form-stats">
        <span>Вопросов: {{ form.questions?.length || 0 }}</span>
        <span>•</span>
        <span>Создана: {{ formatDate(form.created_at) }}</span>
      </div>

      <div class="questions-list">
        <h3>Вопросы</h3>
        <div v-if="!form.questions || form.questions.length === 0" class="empty-message">
          Вопросов нет
        </div>
        <div v-else>
          <div v-for="(q, idx) in form.questions" :key="q.id" class="question-item">
            <span class="q-number">{{ idx + 1 }}.</span>
            <span class="q-title">{{ q.title }}</span>
            <span class="q-type">({{ q.type }})</span>
            <span v-if="q.depends_on" class="q-depends">— зависит от вопроса с ID {{ q.depends_on }}</span>
          </div>
        </div>
      </div>

      <div class="action-buttons">
        <button class="btn-primary" @click="editForm">Редактировать</button>
        <button class="btn-danger" @click="deleteForm">Удалить</button>
        <button class="btn-secondary" @click="viewResponses">Ответы</button>
        <button class="btn-secondary" @click="copyLink">Скопировать ссылку</button>
      </div>
    </div>

    <div v-else class="error-state">
      <p>Форма не найдена или доступ запрещён</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const form = ref(null)
const loading = ref(true)
const formId = route.params.id

const fetchForm = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/forms/${formId}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (!response.ok) {
      if (response.status === 404) {
        form.value = null
      } else {
        throw new Error('Ошибка загрузки')
      }
      return
    }
    const data = await response.json()
    form.value = data
  } catch (err) {
    console.error(err)
    form.value = null
  } finally {
    loading.value = false
  }
}

const editForm = () => {
  router.push(`/edit/${formId}`)
}

const deleteForm = async () => {
  if (!confirm('Удалить форму? Это действие необратимо.')) return
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/forms/${formId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (!response.ok) throw new Error('Ошибка удаления')
    router.push('/')
  } catch (err) {
    alert('Не удалось удалить форму')
  }
}

const viewResponses = () => {
  router.push(`/responses/${formId}`)
}

const copyLink = () => {
  const url = `${window.location.origin}/fill/${formId}`
  navigator.clipboard.writeText(url)
  alert('Ссылка скопирована')
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

onMounted(fetchForm)
</script>

<style scoped>
.form-view-page {
  max-width: 800px;
  margin: 0 auto;
}
.page-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.5rem;
}
.page-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  margin: 0;
}
.status-badge {
  padding: 0.2rem 0.8rem;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}
.status-badge.public {
  background: #e8f5e8;
  color: #2f855a;
}
.status-badge.private {
  background: #fef3c7;
  color: #b45309;
}
.form-description {
  color: var(--text-muted);
  font-size: 1.05rem;
  margin: 1rem 0;
}
.form-stats {
  font-size: 0.9rem;
  color: var(--text-muted);
  margin-bottom: 2rem;
}
.questions-list {
  margin: 2rem 0;
}
.questions-list h3 {
  margin-bottom: 1rem;
}
.question-item {
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border);
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.q-number {
  font-weight: 600;
  color: var(--text-muted);
}
.q-title {
  font-weight: 500;
}
.q-type {
  color: var(--text-muted);
  font-size: 0.85rem;
}
.q-depends {
  color: var(--text-muted);
  font-size: 0.85rem;
}
.empty-message {
  color: var(--text-muted);
}
.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.8rem;
  margin-top: 2rem;
}
.btn-primary, .btn-danger, .btn-secondary {
  padding: 0.7rem 1.5rem;
  border: none;
  border-radius: var(--radius-sm);
  font-weight: 600;
  cursor: pointer;
  transition: 0.2s;
}
.btn-primary {
  background: var(--primary);
  color: #fff;
}
.btn-primary:hover {
  background: var(--primary-hover);
}
.btn-danger {
  background: #e53e3e;
  color: #fff;
}
.btn-danger:hover {
  background: #c53030;
}
.btn-secondary {
  background: var(--bg);
  color: var(--text);
  border: 1px solid var(--border);
}
.btn-secondary:hover {
  background: var(--surface);
}
.loading-state, .error-state {
  text-align: center;
  padding: 3rem;
  color: var(--text-muted);
}
.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  margin: 0 auto 1rem;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
