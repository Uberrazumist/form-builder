<template>
  <div class="create-form-page">
    <div class="page-header">
      <h1 class="page-title">Создание формы</h1>
      <p class="page-subtitle">Заполните данные и добавьте вопросы</p>
    </div>

    <form @submit.prevent="submitForm" class="form-builder" novalidate>
      <div class="form-card">
        <h2 class="card-title">Основная информация</h2>

        <div class="form-group">
          <label for="formTitle">
            <Icon name="edit" />
            Заголовок формы <span class="required">*</span>
          </label>
          <input id="formTitle" type="text" v-model="formData.title" required placeholder="Например: Запись к мастеру" />
        </div>

        <div class="form-group">
          <label for="formDescription">
            <Icon name="document" />
            Описание
          </label>
          <textarea id="formDescription" v-model="formData.description" placeholder="Краткое описание формы (необязательно)" rows="3"></textarea>
        </div>

        <div class="checkbox-group">
          <label class="checkbox-label">
            <input type="checkbox" v-model="formData.is_public" />
            <span class="checkbox-text">
              Публичная форма
              <span class="hint">Доступна по ссылке без авторизации</span>
            </span>
          </label>
        </div>
      </div>

      <div class="form-card">
        <div class="card-header">
          <h2 class="card-title">Вопросы</h2>
          <button type="button" class="btn-add" @click="addQuestion">
            <Icon name="plus" />
            Добавить вопрос
          </button>
        </div>

        <div v-if="formData.questions.length === 0" class="empty-state">
          <p>Пока нет вопросов. Нажмите "Добавить вопрос", чтобы начать.</p>
        </div>

        <div
          v-for="(question, index) in formData.questions"
          :key="question.id || index"
          class="question-card"
        >
          <div class="question-header">
            <span class="question-number">Вопрос {{ index + 1 }}</span>
            <button type="button" class="btn-remove" @click="removeQuestion(index)">
              <Icon name="trash" />
            </button>
          </div>

          <div class="form-group">
            <label>Тип вопроса</label>
            <select v-model="question.type">
              <option value="text">Текст (одна строка)</option>
              <option value="textarea">Текст (несколько строк)</option>
              <option value="radio">Один вариант (radio)</option>
              <option value="checkbox">Несколько вариантов (checkbox)</option>
              <option value="select">Выбор из списка (select)</option>
              <option value="rating">Рейтинг (звёзды)</option>
              <option value="date">Дата</option>
              <option value="dictionary">Выбор из справочника</option>
            </select>
          </div>

          <div class="form-group">
            <label>Текст вопроса <span class="required">*</span></label>
            <input type="text" v-model="question.title" required placeholder="Введите вопрос" />
          </div>

          <div v-if="['radio', 'checkbox', 'select'].includes(question.type)" class="options-section">
            <label>Варианты ответов</label>
            <div class="options-list">
              <div v-for="(option, optIndex) in question.options" :key="optIndex" class="option-item">
                <input type="text" v-model="question.options[optIndex]" :placeholder="`Вариант ${optIndex + 1}`" />
                <button type="button" class="btn-remove-small" @click="removeOption(question, optIndex)">
                  <Icon name="close" />
                </button>
              </div>
            </div>
            <button type="button" class="btn-add-small" @click="addOption(question)">
              <Icon name="plus" />
              Добавить вариант
            </button>
          </div>

          <div v-if="question.type === 'date'" class="info-hint">
            <Icon name="calendar" />
            <span>Пользователь сможет выбрать дату через стандартный календарь браузера</span>
          </div>

          <div v-if="question.type === 'dictionary'" class="dictionary-section">
            <div class="form-group">
              <label>
                <Icon name="book" />
                Выберите справочник <span class="required">*</span>
              </label>
              <select v-model="question.dictionary_id">
                <option :value="null" disabled>— выберите справочник —</option>
                <option v-for="dict in dictionaries" :key="dict.id" :value="dict.id">
                  {{ dict.name }}
                </option>
              </select>
              <span class="hint" v-if="dictionaries.length === 0">
                Сначала создайте справочник в разделе «Справочники»
              </span>
            </div>

            <div class="info-hint">
              <Icon name="link" />
              <span>Связи между справочниками настраиваются автоматически на основе метаданных элементов.</span>
            </div>

            <div class="checkbox-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="question.is_booking" />
                <span class="checkbox-text">
                  Проверять занятость
                  <span class="hint">Включите, если это запись. Занятые варианты будут отмечены.</span>
                </span>
              </label>
            </div>
          </div>

          <div v-if="question.type === 'rating'" class="form-group">
            <label>Максимальный рейтинг</label>
            <select v-model="question.rating_max">
              <option :value="5">5 звёзд</option>
              <option :value="10">10 звёзд</option>
            </select>
          </div>

          <div class="checkbox-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="question.is_required" />
              <span class="checkbox-text">Обязательный вопрос</span>
            </label>
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button type="button" class="btn-secondary" @click="$router.back()">Отмена</button>
        <button type="submit" class="btn-primary" :disabled="loading">
          <span v-if="!loading">Создать форму</span>
          <span v-else class="spinner"></span>
        </button>
      </div>

      <FormResult v-if="result" :result="result" />
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
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

interface Dictionary {
  id: string
  name: string
  description: string
}

const router = useRouter()

const formData = reactive<{
  title: string
  description: string
  is_public: boolean
  questions: Question[]
}>({
  title: '',
  description: '',
  is_public: false,
  questions: []
})

const dictionaries = ref<Dictionary[]>([])
const result = ref<any>(null)
const loading = ref(false)

let questionIdCounter = 1

onMounted(async () => {
  await loadDictionaries()
})

const loadDictionaries = async () => {
  try {
    const token = localStorage.getItem('token')
    const response = await fetch('/api/dictionaries', {
      headers: { 'Authorization': `Bearer ${token || ''}` }
    })
    if (response.ok) {
      const data = await response.json()
      dictionaries.value = Array.isArray(data) ? data : (data.dictionaries || [])
    }
  } catch (err) {
    console.error('[CreateForm] Failed to load dictionaries:', err)
  }
}

const addQuestion = () => {
  formData.questions.push({
    id: `temp_${questionIdCounter++}`,
    type: 'text',
    title: '',
    is_required: false,
    dictionary_id: null,
    is_booking: false,
    depends_on: null,
    options: [],
    rating_max: 5,
    order_index: formData.questions.length
  })
}

const removeQuestion = (index: number) => {
  formData.questions.splice(index, 1)
}

const addOption = (question: Question) => {
  if (!Array.isArray(question.options)) question.options = []
  question.options.push('')
}

const removeOption = (question: Question, optIndex: number) => {
  if (Array.isArray(question.options)) {
    question.options.splice(optIndex, 1)
  }
}

const submitForm = async () => {
  if (!formData.title.trim()) {
    result.value = { error: 'Заголовок формы обязателен' }
    return
  }

  if (formData.questions.length === 0) {
    result.value = { error: 'Добавьте хотя бы один вопрос' }
    return
  }

  for (const q of formData.questions) {
    if (!q?.title?.trim()) {
      result.value = { error: 'Все вопросы должны иметь текст' }
      return
    }
    if (['radio', 'checkbox', 'select'].includes(q?.type) && (!q.options || q.options.length === 0)) {
      result.value = { error: `Вопрос "${q.title}" должен иметь хотя бы один вариант ответа` }
      return
    }
    if (q?.type === 'dictionary' && !q.dictionary_id) {
      result.value = { error: `Вопрос "${q.title}" должен иметь выбранный справочник` }
      return
    }
  }

  loading.value = true
  result.value = null

  try {
    const token = localStorage.getItem('token')

    // Payload в строгом snake_case
    const payload = {
      title: formData.title,
      description: formData.description,
      is_public: formData.is_public,
      questions: formData.questions.map((q, idx) => ({
        type: q.type,
        title: q.title,
        order_index: idx,
        is_required: q.is_required,
        dictionary_id: q.type === 'dictionary' ? q.dictionary_id : null,
        is_booking: q.type === 'dictionary' ? q.is_booking : false,
        depends_on: q.depends_on || null,
        options: q.options || [],
        rating_max: q.rating_max || 5
      }))
    }

    const response = await fetch('/api/forms', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token || ''}`
      },
      body: JSON.stringify(payload)
    })

    if (!response.ok) {
      if (import.meta.env.DEV && response.status === 404) {
        result.value = {
          warning: 'Демо-режим',
          message: 'Бэкенд недоступен (404)',
          data: formData
        }
        setTimeout(() => router.push('/'), 1500)
        return
      }
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка создания формы' }
      return
    }

    const data = await response.json()
    result.value = { success: true, message: 'Форма успешно создана' }

    // Редирект на страницу формы (snake_case id)
    const newFormId = data.id
    if (newFormId) {
      setTimeout(() => router.push(`/form/${newFormId}`), 1000)
    } else {
      setTimeout(() => router.push('/'), 1000)
    }
  } catch (err: any) {
    if (import.meta.env.DEV) {
      result.value = {
        warning: 'Network error',
        message: 'Не удалось связаться с сервером',
        details: err.message
      }
    } else {
      result.value = { error: 'Ошибка сети. Попробуйте позже.' }
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.create-form-page { width: 100%; max-width: 800px; margin: 0 auto; }
.page-header { margin-bottom: 2rem; }
.page-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; margin-bottom: 0.5rem; }
.page-subtitle { color: var(--text-muted); font-size: 1rem; }
.form-builder { display: flex; flex-direction: column; gap: 1.5rem; animation: fadeUp 0.5s ease both; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
.form-card { background: var(--surface); padding: 2rem; border-radius: var(--radius); border: 1px solid var(--border); box-shadow: var(--shadow-sm); }
.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
.card-title { font-size: 1.25rem; font-weight: 700; color: var(--text); letter-spacing: -0.01em; }
.form-group { display: flex; flex-direction: column; gap: 0.4rem; margin-bottom: 1.25rem; }
.form-group label { display: flex; align-items: center; gap: 0.45rem; font-size: 0.88rem; font-weight: 600; color: var(--text); }
.form-group label svg { width: 15px; height: 15px; color: var(--text-muted); }
.required { color: #c53030; }
input[type="text"], input[type="email"], textarea, select { width: 100%; padding: 0.75rem 0.95rem; font-size: 0.95rem; font-family: inherit; color: var(--text); background: var(--bg); border: 1.5px solid var(--border); border-radius: var(--radius-sm); transition: border-color 0.2s, background 0.2s, box-shadow 0.2s; resize: vertical; }
input::placeholder, textarea::placeholder { color: #a6afbf; }
input:hover, textarea:hover, select:hover { border-color: #cfd6e3; }
input:focus, textarea:focus, select:focus { outline: none; border-color: var(--primary); background: var(--surface); box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1); }
.checkbox-group { margin-bottom: 1rem; }
.checkbox-label { display: flex; align-items: flex-start; gap: 0.6rem; cursor: pointer; font-size: 0.92rem; }
.checkbox-label input[type="checkbox"] { width: 18px; height: 18px; margin-top: 2px; cursor: pointer; accent-color: var(--primary); }
.checkbox-text { display: flex; flex-direction: column; gap: 0.15rem; }
.hint { font-size: 0.8rem; color: var(--text-muted); font-weight: 400; }
.empty-state { text-align: center; padding: 2rem; color: var(--text-muted); font-size: 0.95rem; }
.info-hint { display: flex; align-items: center; gap: 0.6rem; padding: 0.85rem 1rem; background: var(--primary-soft); border-radius: var(--radius-sm); color: var(--primary); font-size: 0.88rem; margin-bottom: 1rem; }
.info-hint svg { width: 18px; height: 18px; flex-shrink: 0; }
.question-card { background: var(--bg); padding: 1.5rem; border-radius: var(--radius-sm); border: 1px solid var(--border); margin-bottom: 1rem; animation: fadeUp 0.3s ease both; }
.question-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
.question-number { font-size: 0.85rem; font-weight: 600; color: var(--primary); text-transform: uppercase; letter-spacing: 0.05em; }
.btn-remove { width: 32px; height: 32px; border: none; background: transparent; color: #c53030; cursor: pointer; border-radius: 6px; display: flex; align-items: center; justify-content: center; transition: background 0.2s; }
.btn-remove:hover { background: #fdecec; }
.btn-remove svg { width: 18px; height: 18px; }
.options-section, .dictionary-section { margin-bottom: 1.25rem; padding: 1rem; background: var(--surface); border-radius: var(--radius-sm); border: 1px dashed var(--border); }
.dictionary-section { border-color: var(--primary-soft); background: color-mix(in srgb, var(--primary-soft) 40%, var(--surface)); }
.options-list { display: flex; flex-direction: column; gap: 0.5rem; margin-bottom: 0.75rem; }
.option-item { display: flex; gap: 0.5rem; }
.option-item input { flex: 1; }
.btn-remove-small { width: 36px; height: 36px; border: 1.5px solid var(--border); background: var(--surface); color: #c53030; cursor: pointer; border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.btn-remove-small:hover { background: #fdecec; border-color: #c53030; }
.btn-remove-small svg { width: 16px; height: 16px; }
.btn-add, .btn-add-small { display: flex; align-items: center; gap: 0.4rem; padding: 0.6rem 1rem; background: var(--primary-soft); color: var(--primary); border: none; border-radius: var(--radius-sm); font-size: 0.9rem; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.btn-add:hover, .btn-add-small:hover { background: var(--primary); color: #fff; }
.btn-add svg, .btn-add-small svg { width: 16px; height: 16px; }
.btn-add-small { padding: 0.5rem 0.85rem; font-size: 0.85rem; }
.form-actions { display: flex; gap: 1rem; justify-content: flex-end; }
.btn-primary, .btn-secondary { padding: 0.85rem 2rem; font-size: 0.98rem; font-weight: 600; font-family: inherit; border: none; border-radius: var(--radius-sm); cursor: pointer; transition: all 0.2s; display: flex; align-items: center; justify-content: center; min-width: 140px; }
.btn-primary { background: var(--primary); color: #fff; box-shadow: 0 4px 14px rgba(47, 79, 138, 0.25); }
.btn-primary:hover:not(:disabled) { background: var(--primary-hover); transform: translateY(-1px); box-shadow: 0 6px 18px rgba(47, 79, 138, 0.32); }
.btn-secondary { background: var(--surface); color: var(--text); border: 1.5px solid var(--border); }
.btn-secondary:hover { background: var(--bg); border-color: #cfd6e3; }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }
.spinner { width: 18px; height: 18px; border: 2px solid rgba(255,255,255,0.35); border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 720px) {
  .form-card { padding: 1.5rem 1.25rem; }
  .question-card { padding: 1.25rem 1rem; }
  .form-actions { flex-direction: column-reverse; }
  .btn-primary, .btn-secondary { width: 100%; }
}
</style>
