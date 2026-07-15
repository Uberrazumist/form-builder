<template>
  <div class="form-view-page">
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Загрузка формы...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <div class="error-icon">
        <Icon name="error" />
      </div>
      <h2>Ошибка</h2>
      <p>{{ error }}</p>
      <router-link to="/" class="btn-secondary">На главную</router-link>
    </div>

    <div v-else-if="form" class="form-container">
      <div class="form-header">
        <div class="header-top">
          <h1 class="form-title">{{ form.title }}</h1>
          <span class="badge" :class="form.is_public ? 'badge-public' : 'badge-private'">
            {{ form.is_public ? 'Публичная' : 'Закрытая' }}
          </span>
        </div>
        <p v-if="form.description" class="form-description">{{ form.description }}</p>
        <div class="form-meta">
          <span>{{ form.questions?.length || 0 }} вопросов</span>
          <span>•</span>
          <span>{{ formatDate(form.updated_at) }}</span>
        </div>
      </div>

      <div class="form-card">
        <h2 class="card-title">Вопросы</h2>
        <div v-if="!form.questions || form.questions.length === 0" class="empty-questions">
          <p>В форме пока нет вопросов</p>
        </div>
        <div v-else class="questions-list">
          <div
            v-for="(question, index) in form.questions"
            :key="question.id"
            class="question-item"
          >
            <div class="question-header">
              <span class="question-number">{{ index + 1 }}.</span>
              <span class="question-title">{{ question.title }}</span>
              <span v-if="question.is_required" class="required-badge">Обязательный</span>
            </div>
            <div class="question-meta">
              <span class="question-type">{{ getQuestionTypeName(question.type) }}</span>
              <span v-if="question.depends_on" class="depends-on">
                Зависит от вопроса #{{ getQuestionIndex(question.depends_on) + 1 }}
              </span>
            </div>
            <div v-if="['radio', 'checkbox', 'select'].includes(question.type)" class="question-options">
              <span class="options-label">Варианты:</span>
              <ul class="options-list">
                <li v-for="(option, optIdx) in question.options" :key="optIdx">
                  {{ option }}
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <!-- Блок ссылок -->
      <div class="links-card">
        <h2 class="card-title">Ссылки для работы с формой</h2>
        <p class="links-hint">Нажмите на кнопку, чтобы скопировать ссылку и поделиться ей</p>
        <div class="links-grid">
          <div class="link-item">
            <div class="link-info">
              <div class="link-icon fill">
                <Icon name="edit" />
              </div>
              <div>
                <h3>Ссылка для заполнения</h3>
                <p>Отправьте её тем, кто должен заполнить форму</p>
              </div>
            </div>
            <button @click="copyLink('fill')" class="btn-copy">
              <Icon name="link" />
              Скопировать
            </button>
          </div>
          <div class="link-item">
            <div class="link-info">
              <div class="link-icon preview">
                <Icon name="eye" />
              </div>
              <div>
                <h3>Предпросмотр</h3>
                <p>Посмотрите, как форма выглядит для заполняющего</p>
              </div>
            </div>
            <router-link :to="`/preview/${form.id}`" class="btn-copy">
              <Icon name="eye" />
              Открыть
            </router-link>
          </div>
          <div class="link-item">
            <div class="link-info">
              <div class="link-icon responses">
                <Icon name="document" />
              </div>
              <div>
                <h3>Ссылка на ответы</h3>
                <p>Здесь вы увидите все полученные ответы</p>
              </div>
            </div>
            <button @click="copyLink('responses')" class="btn-copy">
              <Icon name="link" />
              Скопировать
            </button>
          </div>
          <div class="link-item">
            <div class="link-info">
              <div class="link-icon edit">
                <Icon name="edit" />
              </div>
              <div>
                <h3>Ссылка для редактирования</h3>
                <p>Для быстрого доступа к редактированию формы</p>
              </div>
            </div>
            <button @click="copyLink('edit')" class="btn-copy">
              <Icon name="link" />
              Скопировать
            </button>
          </div>
        </div>
      </div>

      <div class="form-card">
        <h2 class="card-title">Действия</h2>
        <div class="form-actions">
          <router-link :to="`/edit/${form.id}`" class="btn-primary">
            <Icon name="edit" />
            Редактировать
          </router-link>
          <button @click="showDeleteModal = true" class="btn-danger">
            <Icon name="trash" />
            Удалить форму
          </button>
        </div>
      </div>

      <FormResult v-if="result" :result="result" />
    </div>

    <!-- Модальное окно подтверждения удаления -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="showDeleteModal = false">
      <div class="modal-content" @click.stop>
        <h3>Удалить форму?</h3>
        <p>Это действие нельзя отменить. Все ответы будут удалены.</p>
        <div class="modal-actions">
          <button @click="showDeleteModal = false" class="btn-secondary">Отмена</button>
          <button @click="deleteForm" class="btn-danger" :disabled="deleting">
            <span v-if="!deleting">Удалить</span>
            <span v-else class="spinner-small"></span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

interface Question {
  id: string
  type: string
  title: string
  is_required: boolean
  dictionary_id: string | null
  is_booking: boolean
  depends_on: string | null
  options: string[]
  rating_max: number
  order_index: number
}

interface Form {
  id: string
  title: string
  description: string
  is_public: boolean
  questions: Question[]
  created_at: string
  updated_at: string
}

const route = useRoute()
const router = useRouter()
const hostOrigin = typeof window !== 'undefined' ? window.location.origin : ''

const form = ref<Form | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const result = ref<any>(null)
const showDeleteModal = ref(false)
const deleting = ref(false)

onMounted(async () => {
  await loadForm()
})

const loadForm = async () => {
  loading.value = true
  error.value = null

  try {
    const formId = String(route.params.id)
    const token = localStorage.getItem('token')
    const headers: Record<string, string> = {}
    if (token) headers['Authorization'] = `Bearer ${token}`

    const response = await fetch(`/api/forms/${formId}`, { headers })
    if (!response.ok) {
      if (response.status === 404) error.value = 'Форма не найдена'
      else if (response.status === 403) error.value = 'У вас нет доступа к этой форме'
      else error.value = 'Не удалось загрузить форму'
      return
    }
    form.value = await response.json()
  } catch (err) {
    console.error('[FormView] Load error:', err)
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const getQuestionTypeName = (type: string): string => {
  const names: Record<string, string> = {
    text: 'Текст (одна строка)',
    textarea: 'Текст (несколько строк)',
    radio: 'Один вариант',
    checkbox: 'Несколько вариантов',
    select: 'Выбор из списка',
    rating: 'Рейтинг',
    dictionary: 'Выбор из справочника',
    date: 'Дата'
  }
  return names[type] || type
}

const getQuestionIndex = (questionId: string): number => {
  if (!form.value?.questions) return -1
  return form.value.questions.findIndex(q => q.id === questionId)
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return ''
  try {
    const date = new Date(dateStr)
    const day = String(date.getDate()).padStart(2, '0')
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const year = date.getFullYear()
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${day}.${month}.${year} ${hours}:${minutes}`
  } catch {
    return dateStr
  }
}

const copyToClipboard = async (text: string): Promise<boolean> => {
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch (err) {
      console.warn('Clipboard API failed, trying fallback:', err)
    }
  }
  try {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.left = '-9999px'
    textarea.style.top = '-9999px'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.focus()
    textarea.select()
    const success = document.execCommand('copy')
    document.body.removeChild(textarea)
    return success
  } catch (err) {
    console.error('Fallback copy failed:', err)
    return false
  }
}

const copyLink = async (type: string) => {
  if (!form.value) return
  const formId = form.value.id
  let path = ''
  let label = ''

  switch (type) {
    case 'fill':
      path = `/fill/${formId}`
      label = 'ссылка для заполнения'
      break
    case 'responses':
      path = `/responses/${formId}`
      label = 'ссылка на ответы'
      break
    case 'edit':
      path = `/edit/${formId}`
      label = 'ссылка для редактирования'
      break
  }

  const url = `${hostOrigin}${path}`
  const success = await copyToClipboard(url)

  if (success) {
    result.value = {
      success: true,
      message: `${label.charAt(0).toUpperCase() + label.slice(1)} скопирована в буфер обмена`
    }
  } else {
    result.value = { error: `Не удалось скопировать. Скопируйте вручную: ${url}` }
  }

  setTimeout(() => { result.value = null }, 3000)
}

const deleteForm = async () => {
  deleting.value = true
  result.value = null

  try {
    const formId = String(route.params.id)
    const token = localStorage.getItem('token')

    const response = await fetch(`/api/forms/${formId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token || ''}` }
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка удаления формы' }
      return
    }

    result.value = { success: true, message: 'Форма успешно удалена' }
    setTimeout(() => router.push('/'), 1000)
  } catch (err) {
    result.value = { error: 'Ошибка сети. Попробуйте позже.' }
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.form-view-page { width: 100%; max-width: 800px; margin: 0 auto; }
.loading-state, .error-state { text-align: center; padding: 4rem 2rem; }
.spinner { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
.spinner-small { width: 16px; height: 16px; border: 2px solid rgba(255,255,255,0.35); border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon { width: 64px; height: 64px; background: #fdecec; color: #c53030; border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.error-icon svg { width: 32px; height: 32px; }
.error-state h2 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.error-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.btn-secondary { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.5rem; background: var(--surface); color: var(--text); border: 1.5px solid var(--border); border-radius: var(--radius-sm); font-size: 0.95rem; font-weight: 600; text-decoration: none; cursor: pointer; transition: all 0.2s; }
.btn-secondary:hover { background: var(--bg); border-color: #cfd6e3; }
.form-container { animation: fadeUp 0.5s ease both; display: flex; flex-direction: column; gap: 1.5rem; }
.form-header { margin-bottom: 0.5rem; }
.header-top { display: flex; align-items: flex-start; gap: 1rem; margin-bottom: 0.75rem; }
.form-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; flex: 1; }
.badge { font-size: 0.75rem; font-weight: 600; padding: 0.35rem 0.75rem; border-radius: 6px; text-transform: uppercase; letter-spacing: 0.05em; white-space: nowrap; }
.badge-public { background: #e8f5e8; color: #2f855a; }
.badge-private { background: var(--primary-soft); color: var(--primary); }
.form-description { font-size: 1.05rem; color: var(--text-muted); line-height: 1.6; margin-bottom: 0.75rem; }
.form-meta { display: flex; gap: 0.5rem; font-size: 0.9rem; color: var(--text-muted); }
.form-card { background: var(--surface); padding: 2rem; border-radius: var(--radius); border: 1px solid var(--border); box-shadow: var(--shadow-sm); }
.links-card { background: var(--surface); padding: 2rem; border-radius: var(--radius); border: 1px solid var(--border); box-shadow: var(--shadow-sm); }
.card-title { font-size: 1.25rem; font-weight: 700; color: var(--text); margin-bottom: 1rem; }
.links-hint { color: var(--text-muted); font-size: 0.9rem; margin-bottom: 1.5rem; margin-top: -0.5rem; }
.links-grid { display: flex; flex-direction: column; gap: 0.85rem; }
.link-item { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding: 1rem 1.25rem; background: var(--bg); border: 1px solid var(--border); border-radius: var(--radius-sm); transition: all 0.2s; }
.link-item:hover { border-color: var(--primary-soft); background: var(--surface); }
.link-info { display: flex; align-items: center; gap: 1rem; flex: 1; min-width: 0; }
.link-icon { flex-shrink: 0; width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.link-icon svg { width: 20px; height: 20px; }
.link-icon.fill { background: #e8f5e8; color: #2f855a; }
.link-icon.preview { background: #fff4e0; color: #b7791f; }
.link-icon.responses { background: var(--primary-soft); color: var(--primary); }
.link-icon.edit { background: #fdecec; color: #c53030; }
.link-info h3 { font-size: 0.98rem; font-weight: 700; color: var(--text); margin-bottom: 0.15rem; }
.link-info p { font-size: 0.82rem; color: var(--text-muted); line-height: 1.4; }
.btn-copy { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.55rem 1rem; background: var(--primary-soft); color: var(--primary); border: none; border-radius: var(--radius-sm); font-size: 0.85rem; font-weight: 600; text-decoration: none; cursor: pointer; transition: all 0.2s; white-space: nowrap; flex-shrink: 0; }
.btn-copy:hover { background: var(--primary); color: #fff; }
.btn-copy svg { width: 14px; height: 14px; }
.empty-questions { text-align: center; padding: 2rem; color: var(--text-muted); }
.questions-list { display: flex; flex-direction: column; gap: 1rem; }
.question-item { padding: 1.25rem; background: var(--bg); border-radius: var(--radius-sm); border: 1px solid var(--border); }
.question-header { display: flex; align-items: flex-start; gap: 0.5rem; margin-bottom: 0.5rem; }
.question-number { color: var(--primary); font-weight: 700; font-size: 1rem; }
.question-title { flex: 1; font-size: 1rem; font-weight: 600; color: var(--text); line-height: 1.5; }
.required-badge { font-size: 0.72rem; font-weight: 600; padding: 0.2rem 0.5rem; background: #fdecec; color: #c53030; border-radius: 4px; text-transform: uppercase; letter-spacing: 0.05em; }
.question-meta { display: flex; gap: 1rem; font-size: 0.85rem; color: var(--text-muted); margin-bottom: 0.75rem; }
.question-type { font-weight: 500; }
.depends-on { color: var(--primary); font-weight: 500; }
.question-options { margin-top: 0.75rem; padding-top: 0.75rem; border-top: 1px solid var(--border); }
.options-label { font-size: 0.85rem; font-weight: 600; color: var(--text-muted); margin-bottom: 0.5rem; display: block; }
.options-list { list-style: none; padding-left: 0; margin: 0; }
.options-list li { font-size: 0.9rem; color: var(--text); padding: 0.25rem 0; padding-left: 1rem; position: relative; }
.options-list li::before { content: '•'; position: absolute; left: 0; color: var(--primary); }
.form-actions { display: flex; gap: 1rem; flex-wrap: wrap; }
.btn-primary { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.5rem; background: var(--primary); color: #fff; border: none; border-radius: var(--radius-sm); font-size: 0.95rem; font-weight: 600; text-decoration: none; cursor: pointer; transition: all 0.2s; box-shadow: 0 4px 14px rgba(47, 79, 138, 0.25); }
.btn-primary:hover { background: var(--primary-hover); transform: translateY(-1px); box-shadow: 0 6px 18px rgba(47, 79, 138, 0.32); }
.btn-danger { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.5rem; background: #c53030; color: #fff; border: none; border-radius: var(--radius-sm); font-size: 0.95rem; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.btn-danger:hover:not(:disabled) { background: #9b2c2c; transform: translateY(-1px); }
.btn-danger:disabled { opacity: 0.7; cursor: not-allowed; }
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0, 0, 0, 0.5); display: flex; align-items: center; justify-content: center; z-index: 100; animation: fadeIn 0.2s ease; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
.modal-content { background: var(--surface); padding: 2rem; border-radius: var(--radius); max-width: 400px; width: 90%; box-shadow: var(--shadow-lg); }
.modal-content h3 { font-size: 1.25rem; font-weight: 700; color: var(--text); margin-bottom: 0.75rem; }
.modal-content p { color: var(--text-muted); margin-bottom: 1.5rem; line-height: 1.5; }
.modal-actions { display: flex; gap: 1rem; justify-content: flex-end; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 720px) {
  .form-card, .links-card { padding: 1.5rem 1.25rem; }
  .form-title { font-size: 1.5rem; }
  .header-top { flex-direction: column; }
  .link-item { flex-direction: column; align-items: stretch; }
  .btn-copy { justify-content: center; }
  .form-actions { flex-direction: column; }
  .btn-primary, .btn-danger { width: 100%; justify-content: center; }
}
</style>
