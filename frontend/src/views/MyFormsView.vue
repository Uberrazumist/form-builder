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
          
          <!-- Dropdown "Поделиться" -->
          <div class="dropdown-container" @click.stop>
            <button 
              @click="toggleDropdown(form.id)" 
              class="btn-share"
              :class="{ active: openDropdownId === form.id }"
            >
              <Icon name="share" />
              Поделиться
            </button>
            
            <div v-if="openDropdownId === form.id" class="dropdown-menu">
              <button @click="copyShareLink(form.id)" class="dropdown-item">
                <Icon name="link" />
                <span>
                  <strong>Копировать ссылку</strong>
                  <small>Для заполнения формы</small>
                </span>
              </button>
              <router-link :to="`/preview/${form.id}`" class="dropdown-item" @click="closeDropdown">
                <Icon name="eye" />
                <span>
                  <strong>Предпросмотр</strong>
                  <small>Безопасное тестирование</small>
                </span>
              </router-link>
            </div>
          </div>

          <button @click="deleteForm(form.id)" class="btn-danger">
            <Icon name="trash" />
          </button>
        </div>
      </div>
    </div>

    <!-- Уведомление о копировании -->
    <div v-if="showCopyNotification" class="copy-notification">
      <Icon name="check" />
      Ссылка скопирована в буфер обмена
    </div>

    <FormResult v-if="result" :result="result" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
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
const openDropdownId = ref<string | null>(null)
const showCopyNotification = ref(false)
const hostOrigin = typeof window !== 'undefined' ? window.location.origin : ''

onMounted(async () => {
  await loadForms()
  document.addEventListener('click', closeDropdownOnOutsideClick)
})

onUnmounted(() => {
  document.removeEventListener('click', closeDropdownOnOutsideClick)
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

const toggleDropdown = (formId: string) => {
  openDropdownId.value = openDropdownId.value === formId ? null : formId
}

const closeDropdown = () => {
  openDropdownId.value = null
}

const closeDropdownOnOutsideClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.dropdown-container')) {
    closeDropdown()
  }
}

const copyShareLink = async (formId: string) => {
  const shareUrl = `${hostOrigin}/fill/${formId}`
  
  try {
    await navigator.clipboard.writeText(shareUrl)
    showCopyNotification.value = true
    setTimeout(() => {
      showCopyNotification.value = false
    }, 2000)
    closeDropdown()
  } catch (err) {
    console.error('Failed to copy:', err)
    const textarea = document.createElement('textarea')
    textarea.value = shareUrl
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    try {
      document.execCommand('copy')
      showCopyNotification.value = true
      setTimeout(() => {
        showCopyNotification.value = false
      }, 2000)
    } catch (fallbackErr) {
      result.value = { error: 'Не удалось скопировать ссылку' }
    }
    document.body.removeChild(textarea)
    closeDropdown()
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
.btn-primary, .btn-secondary, .btn-danger, .btn-share { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.55rem 0.9rem; font-size: 0.88rem; font-weight: 600; border-radius: var(--radius-sm); text-decoration: none; cursor: pointer; transition: all 0.2s; border: 1.5px solid transparent; font-family: inherit; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover { background: var(--primary-hover, #243f72); }
.btn-secondary { background: var(--surface); color: var(--text); border-color: var(--border); }
.btn-secondary:hover { background: var(--bg); border-color: var(--text-muted); }
.btn-danger { background: var(--surface); color: #c53030; border-color: var(--border); }
.btn-danger:hover { background: #fdecec; border-color: #c53030; }
.btn-share { background: var(--primary-soft); color: var(--primary); border-color: transparent; }
.btn-share:hover { background: var(--primary); color: #fff; }
.btn-share.active { background: var(--primary); color: #fff; }
.btn-primary svg, .btn-secondary svg, .btn-danger svg, .btn-share svg { width: 14px; height: 14px; }
.dropdown-container { position: relative; }
.dropdown-menu { position: absolute; bottom: 100%; left: 0; margin-bottom: 0.5rem; background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius-sm); box-shadow: var(--shadow-lg); min-width: 240px; z-index: 10; animation: dropdownFadeIn 0.2s ease; }
.dropdown-item { display: flex; align-items: flex-start; gap: 0.75rem; padding: 0.85rem 1rem; width: 100%; background: none; border: none; text-align: left; cursor: pointer; transition: all 0.15s; text-decoration: none; color: var(--text); font-family: inherit; }
.dropdown-item:hover { background: var(--bg); }
.dropdown-item:first-child { border-radius: var(--radius-sm) var(--radius-sm) 0 0; }
.dropdown-item:last-child { border-radius: 0 0 var(--radius-sm) var(--radius-sm); }
.dropdown-item svg { width: 18px; height: 18px; color: var(--primary); flex-shrink: 0; margin-top: 0.15rem; }
.dropdown-item span { display: flex; flex-direction: column; gap: 0.15rem; }
.dropdown-item strong { font-size: 0.9rem; font-weight: 600; color: var(--text); }
.dropdown-item small { font-size: 0.78rem; color: var(--text-muted); }
@keyframes dropdownFadeIn { from { opacity: 0; transform: translateY(4px); } to { opacity: 1; transform: translateY(0); } }
.copy-notification { position: fixed; bottom: 2rem; right: 2rem; background: var(--primary); color: #fff; padding: 1rem 1.5rem; border-radius: var(--radius-sm); box-shadow: var(--shadow-lg); display: flex; align-items: center; gap: 0.5rem; font-weight: 600; animation: slideUp 0.3s ease; z-index: 100; }
.copy-notification svg { width: 20px; height: 20px; }
@keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 720px) {
  .forms-grid { grid-template-columns: 1fr; }
  .form-actions { flex-direction: column; }
  .btn-primary, .btn-secondary, .btn-danger, .btn-share { width: 100%; justify-content: center; }
  .dropdown-menu { left: 0; right: 0; min-width: auto; }
  .copy-notification { left: 1rem; right: 1rem; bottom: 1rem; }
}
</style>
