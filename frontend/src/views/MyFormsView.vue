<template>
  <div class="my-forms-page">
    <div class="page-header">
      <h1 class="page-title">Мои формы</h1>
      <router-link to="/create" class="btn-primary">
        <Icon name="plus" />
        Создать форму
      </router-link>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Загрузка форм...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <div class="error-icon">
        <Icon name="error" />
      </div>
      <h2>Ошибка</h2>
      <p>{{ error }}</p>
    </div>

    <div v-else-if="forms.length === 0" class="empty-state">
      <div class="empty-icon">
        <Icon name="document" />
      </div>
      <h3>У вас пока нет форм</h3>
      <p>Создайте первую форму, чтобы начать сбор ответов</p>
      <router-link to="/create" class="btn-primary">
        <Icon name="plus" />
        Создать форму
      </router-link>
    </div>

    <div v-else class="forms-grid">
      <div v-for="form in forms" :key="form.id" class="form-card">
        <div class="form-card-header">
          <h3 class="form-title">{{ form.title }}</h3>
          <span v-if="form.is_public" class="badge-public">Публичная</span>
          <span v-else class="badge-private">Приватная</span>
        </div>

        <p v-if="form.description" class="form-description">{{ form.description }}</p>
        <p v-else class="form-description empty">Без описания</p>

        <div class="form-meta">
          <span class="meta-item">
            <Icon name="calendar" />
            {{ formatDate(form.created_at) }}
          </span>
        </div>

        <div class="form-actions">
          <router-link :to="`/edit/${form.id}`" class="btn-secondary">
            <Icon name="edit" />
            Редактировать
          </router-link>
          <router-link :to="`/responses/${form.id}`" class="btn-secondary">
            <Icon name="document" />
            Ответы
          </router-link>
          <router-link :to="`/fill/${form.id}`" class="btn-secondary">
            <Icon name="eye" />
            Открыть
          </router-link>
          <button @click="deleteForm(form.id)" class="btn-danger">
            <Icon name="trash" />
          </button>
        </div>
      </div>
    </div>

    <FormResult v-if="result" :result="result" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

interface Form {
  id: string
  title: string
  description: string
  is_public: boolean
  is_published: boolean
  created_at: string
  updated_at: string
}

const forms = ref<Form[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const result = ref<any>(null)

onMounted(async () => {
  await loadForms()
})

const loadForms = async () => {
  loading.value = true
  error.value = null

  try {
    const token = localStorage.getItem('token')
    const headers: Record<string, string> = {}
    if (token) headers['Authorization'] = `Bearer ${token}`

    const response = await fetch('/api/forms', { headers })

    if (!response.ok) {
      if (response.status === 401) {
        error.value = 'Требуется авторизация'
      } else {
        error.value = 'Не удалось загрузить список форм'
      }
      return
    }

    const data = await response.json()
    forms.value = Array.isArray(data) ? data : []
  } catch (err) {
    console.error('[MyForms] Load error:', err)
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const deleteForm = async (formId: string) => {
  if (!confirm('Удалить форму? Это действие необратимо.')) return

  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/forms/${formId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token || ''}` }
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Не удалось удалить форму' }
      return
    }

    result.value = { success: true, message: 'Форма удалена' }
    forms.value = forms.value.filter(f => f.id !== formId)
  } catch (err) {
    result.value = { error: 'Ошибка сети' }
  }
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return ''
  try {
    const date = new Date(dateStr)
    return date.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    })
  } catch {
    return dateStr
  }
}
</script>

<style scoped>
.my-forms-page { width: 100%; max-width: 1100px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; flex-wrap: wrap; gap: 1rem; }
.page-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; }
.loading-state, .error-state, .empty-state { text-align: center; padding: 4rem 2rem; background: var(--surface); border-radius: var(--radius); border: 1px solid var(--border); }
.spinner { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon, .empty-icon { width: 64px; height: 64px; background: var(--primary-soft); color: var(--primary); border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.error-icon svg, .empty-icon svg { width: 32px; height: 32px; }
.error-state h2, .empty-state h3 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.error-state p, .empty-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.forms-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 1.5rem; }
.form-card { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); padding: 1.5rem; transition: all 0.2s; animation: fadeUp 0.4s ease both; }
.form-card:hover { box-shadow: var(--shadow-md); border-color: var(--primary-soft); transform: translateY(-2px); }
.form-card-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 0.75rem; margin-bottom: 0.75rem; }
.form-title { font-size: 1.15rem; font-weight: 700; color: var(--text); line-height: 1.3; flex: 1; }
.badge-public, .badge-private { font-size: 0.75rem; font-weight: 600; padding: 0.25rem 0.6rem; border-radius: 6px; white-space: nowrap; }
.badge-public { background: #e8f5e9; color: #2e7d32; }
.badge-private { background: var(--primary-soft); color: var(--primary); }
.form-description { color: var(--text-muted); font-size: 0.9rem; line-height: 1.5; margin-bottom: 1rem; min-height: 2.7rem; }
.form-description.empty { font-style: italic; }
.form-meta { display: flex; gap: 1rem; margin-bottom: 1rem; padding-top: 1rem; border-top: 1px solid var(--border); }
.meta-item { display: inline-flex; align-items: center; gap: 0.35rem; font-size: 0.85rem; color: var(--text-muted); }
.meta-item svg { width: 14px; height: 14px; }
.form-actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.btn-primary, .btn-secondary, .btn-danger { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.55rem 0.9rem; font-size: 0.88rem; font-weight: 600; border-radius: var(--radius-sm); text-decoration: none; cursor: pointer; transition: all 0.2s; border: 1.5px solid transparent; font-family: inherit; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover { background: var(--primary-hover, #243f72); }
.btn-secondary { background: var(--surface); color: var(--text); border-color: var(--border); }
.btn-secondary:hover { background: var(--bg); border-color: var(--text-muted); }
.btn-danger { background: var(--surface); color: #c53030; border-color: var(--border); }
.btn-danger:hover { background: #fdecec; border-color: #c53030; }
.btn-primary svg, .btn-secondary svg, .btn-danger svg { width: 14px; height: 14px; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 720px) {
  .forms-grid { grid-template-columns: 1fr; }
  .form-actions { flex-direction: column; }
  .btn-primary, .btn-secondary, .btn-danger { width: 100%; justify-content: center; }
}
</style>
