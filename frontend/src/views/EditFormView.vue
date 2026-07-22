<template>
  <div class="edit-form-page">
    <!-- Кнопка открытия sidebar на мобильных -->
    <button class="sidebar-toggle" @click="sidebarOpen = !sidebarOpen">
      <Icon name="menu" /> Вопросы
    </button>

    <!-- Sidebar навигация -->
    <div class="sidebar" :class="{ open: sidebarOpen }">
      <div class="sidebar-header">
        <h3>Вопросы</h3>
        <button class="sidebar-close" @click="sidebarOpen = false">
          <Icon name="close" />
        </button>
      </div>
      <div class="sidebar-list">
        <div
          v-for="(question, index) in formData.questions"
          :key="question.id || index"
          class="sidebar-item"
          :class="{ active: activeQuestionIndex === index }"
          @click="scrollToQuestion(index)"
        >
          <span class="sidebar-item-number">{{ index + 1 }}</span>
          <span class="sidebar-item-type">{{ getTypeIcon(question.type) }}</span>
          <span class="sidebar-item-title">{{ question.title || 'Без названия' }}</span>
        </div>
      </div>
    </div>

    <!-- Основной контент -->
    <div class="main-content">
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

    <div v-else class="form-container">
      <div class="page-header">
        <h1 class="page-title">Редактирование формы</h1>
        <p class="page-subtitle">Измените данные и сохраните</p>
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
            :data-question-index="index"
            draggable="true"
            @dragstart="onDragStart($event, index)"
            @dragover.prevent="onDragOver($event, index)"
            @dragenter="onDragEnter($event, index)"
            @dragleave="onDragLeave($event, index)"
            @drop="onDrop($event, index)"
            @dragend="onDragEnd"
            :class="{ 'drag-over': dragOverIndex === index }"
          >
            <div class="question-header">
              <div class="drag-handle" title="Перетащить для смены порядка">
                <Icon name="menu" />
              </div>
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
                <option value="schedule">Календарь бронирования</option>
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
                  <option v-for="dict in dictionaries" :key="dict.id" :value="dict.id">{{ dict.name }}</option>
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

            <div v-if="question.type === 'schedule'" class="schedule-section">
              <div class="form-group">
                <label>
                  <Icon name="link" />
                  Зависит от вопроса (выбор ресурса) <span class="required">*</span>
                </label>
                <select v-model="question.depends_on">
                  <option :value="null" disabled>— выберите предыдущий вопрос с ресурсом —</option>
                  <option 
                    v-for="q in formData.questions.filter(q => q.id !== question.id && q.type === 'dictionary')" 
                    :key="q.id" 
                    :value="q.id"
                  >
                    {{ q.title }}
                  </option>
                </select>
                <span class="hint">Пользователь сначала выберет ресурс в указанном вопросе, а здесь увидит его расписание</span>
              </div>
              <div class="info-hint">
                <Icon name="calendar" />
                <span>Расписание для элементов этого справочника настраивается в разделе «Справочники» → Управление элементами</span>
              </div>
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
          <div class="responses-link-block">
            <Icon name="link" />
            <span class="link-label">Ссылка на ответы:</span>
            <input
              type="text"
              readonly
              :value="`${hostOrigin}/responses/${route.params.id}`"
              class="responses-link-input"
              @click="($event.target as HTMLInputElement).select()"
            />
          </div>
          <router-link :to="`/form/${route.params.id}/responses`" class="btn-secondary">
            <Icon name="document" />
            Посмотреть ответы
          </router-link>
          <button type="button" class="btn-secondary" @click="$router.back()">Отмена</button>
          <button type="submit" class="btn-primary" :disabled="submitting">
            <span v-if="!submitting">Сохранить изменения</span>
            <span v-else class="spinner"></span>
          </button>
        </div>

        <div class="form-card">
          <FormResult v-if="result" :result="result" />
        </div>
      </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick, onBeforeUnmount, watch } from 'vue'
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

interface Dictionary {
  id: string
  name: string
  description: string
}

const route = useRoute()
const router = useRouter()
const hostOrigin = typeof window !== 'undefined' ? window.location.origin : ''

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
const loading = ref(true)
const error = ref<string | null>(null)
const result = ref<any>(null)
const submitting = ref(false)

// Sidebar
const sidebarOpen = ref(false)
const activeQuestionIndex = ref(-1)
let intersectionObserver: IntersectionObserver | null = null

onMounted(async () => {
  await Promise.all([loadForm(), loadDictionaries()])
  
  nextTick(() => {
    setupIntersectionObserver()
  })
})

// Перезапускаем observer при изменении вопросов
watch(() => formData.questions.length, () => {
  nextTick(() => {
    setupIntersectionObserver()
  })
})

onBeforeUnmount(() => {
  if (intersectionObserver) {
    intersectionObserver.disconnect()
  }
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
    console.error('[EditForm] Failed to load dictionaries:', err)
  }
}

const loadForm = async () => {
  loading.value = true
  error.value = null

  try {
    const formId = String(route.params.id)
    const token = localStorage.getItem('token')

    const response = await fetch(`/api/forms/${formId}`, {
      headers: { 'Authorization': `Bearer ${token || ''}` }
    })

    if (!response.ok) {
      if (response.status === 404) {
        error.value = 'Форма не найдена'
      } else if (response.status === 403) {
        error.value = 'У вас нет доступа к этой форме'
      } else {
        error.value = 'Не удалось загрузить форму'
      }
      return
    }

    const data = await response.json()

    // Честный маппинг в snake_case
    formData.title = data.title || ''
    formData.description = data.description || ''
    formData.is_public = Boolean(data.is_public)

    const rawQuestions = Array.isArray(data.questions) ? data.questions : []
    formData.questions = rawQuestions
      .sort((a: any, b: any) => (Number(a.order_index) || 0) - (Number(b.order_index) || 0))
      .map((q: any, idx: number) => ({
        id: q.id || '',
        type: q.type || 'text',
        title: q.title || '',
        is_required: Boolean(q.is_required),
        dictionary_id: q.dictionary_id ?? null,
        is_booking: Boolean(q.is_booking),
        depends_on: q.depends_on ?? null,
        options: Array.isArray(q.options) ? q.options : [],
        rating_max: Number(q.rating_max) || 5,
        order_index: Number(q.order_index) || idx
      }))

  } catch (err) {
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const addQuestion = () => {
  formData.questions.push({
    id: '',
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
  // Пересчитываем order_index после удаления
  formData.questions.forEach((q, i) => { q.order_index = i + 1 })
}

// ==========================================
// Drag-and-Drop для перестановки вопросов
// ==========================================
const dragIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)
const dragGhostEl = ref<HTMLElement | null>(null)

const onDragStart = (event: DragEvent, index: number) => {
  dragIndex.value = index
  
  // Визуальная обратная связь: полупрозрачный клон
  const target = event.currentTarget as HTMLElement
  target.style.opacity = '0.4'
  
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', String(index))
  }
}

const onDragOver = (event: DragEvent, index: number) => {
  event.preventDefault()
  event.dataTransfer!.dropEffect = 'move'
  dragOverIndex.value = index
}

const onDragEnter = (event: DragEvent, index: number) => {
  event.preventDefault()
  dragOverIndex.value = index
}

const onDragLeave = (_event: DragEvent, _index: number) => {
  dragOverIndex.value = null
}

const onDrop = (event: DragEvent, dropIndex: number) => {
  event.preventDefault()
  if (dragIndex.value === null || dragIndex.value === dropIndex) return

  // Перемещаем вопрос
  const removed = formData.questions.splice(dragIndex.value, 1)
  if (removed.length === 0) return
  formData.questions.splice(dropIndex, 0, removed[0]!)

  // Пересчитываем order_index
  formData.questions.forEach((q, i) => { q.order_index = i + 1 })

  dragIndex.value = null
  dragOverIndex.value = null
  
  // Сброс opacity
  const questionCards = document.querySelectorAll('.question-card')
  questionCards.forEach(card => {
    if (card instanceof HTMLElement) card.style.opacity = '1'
  })
}

const onDragEnd = () => {
  dragIndex.value = null
  dragOverIndex.value = null
  
  // Сброс opacity
  const questionCards = document.querySelectorAll('.question-card')
  questionCards.forEach(card => {
    if (card instanceof HTMLElement) card.style.opacity = '1'
  })
}

// ==========================================
// Sidebar навигация
// ==========================================
const getTypeIcon = (type: string): string => {
  const icons: Record<string, string> = {
    text: '📝',
    textarea: '📄',
    radio: '🔘',
    checkbox: '☑️',
    select: '📋',
    rating: '⭐',
    date: '📅',
    dictionary: '📚',
    schedule: '🗓️'
  }
  return icons[type] || '❓'
}

const scrollToQuestion = (index: number) => {
  const el = document.querySelector(`[data-question-index="${index}"]`) as HTMLElement
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    activeQuestionIndex.value = index
  }
}

// Intersection Observer для подсветки активного вопроса
const setupIntersectionObserver = () => {
  if (intersectionObserver) {
    intersectionObserver.disconnect()
  }

  const questionCards = document.querySelectorAll('.question-card')
  intersectionObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const idx = Number(entry.target.getAttribute('data-question-index') || '-1')
        if (idx >= 0) {
          activeQuestionIndex.value = idx
        }
      }
    })
  }, {
    root: null,
    rootMargin: '-20% 0px -60% 0px',
    threshold: 0
  })

  questionCards.forEach(card => intersectionObserver!.observe(card))
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
    if (q?.type === 'schedule' && !q.depends_on) {
      result.value = { error: `Вопрос "${q.title}" (Календарь) должен быть привязан к вопросу выбора ресурса (Зависит от)` }
      return
    }
  }

  submitting.value = true
  result.value = null

  try {
    const formId = String(route.params.id)
    const token = localStorage.getItem('token')

    // Payload в строгом snake_case
    const payload = {
      title: formData.title,
      description: formData.description,
      is_public: formData.is_public,
      questions: formData.questions.map((q, idx) => ({
        id: q.id || undefined,
        type: q.type,
        title: q.title,
        order_index: idx,
        is_required: q.is_required,
        dictionary_id: q.type === 'dictionary' ? q.dictionary_id : null,
        is_booking: q.type === 'dictionary' ? q.is_booking : false,
        depends_on: (q.type === 'schedule' || q.type === 'dictionary') ? q.depends_on : null,
        options: q.options || [],
        rating_max: Number(q.rating_max) || 5
      }))
    }

    const response = await fetch(`/api/forms/${formId}`, {
      method: 'PUT',
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
          message: 'Бэкенд недоступен (404)'
        }
        setTimeout(() => router.push(`/form/${formId}`), 1500)
        return
      }
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка сохранения формы' }
      return
    }

    result.value = { success: true, message: 'Форма успешно обновлена' }
    setTimeout(() => router.push(`/form/${formId}`), 1000)
  } catch (err) {
    if (import.meta.env.DEV) {
      result.value = {
        warning: 'Network error',
        message: 'Не удалось связаться с сервером'
      }
    } else {
      result.value = { error: 'Ошибка сети. Попробуйте позже.' }
    }
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.edit-form-page { width: 100%; max-width: 800px; margin: 0 auto; }
.loading-state, .error-state { text-align: center; padding: 4rem 2rem; }
.spinner { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon { width: 64px; height: 64px; background: #fdecec; color: #c53030; border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.error-icon svg { width: 32px; height: 32px; }
.error-state h2 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.error-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.btn-secondary { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.5rem; background: var(--surface); color: var(--text); border: 1.5px solid var(--border); border-radius: var(--radius-sm); font-size: 0.95rem; font-weight: 600; text-decoration: none; cursor: pointer; transition: all 0.2s; }
.btn-secondary:hover { background: var(--bg); border-color: #cfd6e3; }
.form-container { animation: fadeUp 0.5s ease both; }
.page-header { margin-bottom: 2rem; }
.page-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; margin-bottom: 0.5rem; }
.page-subtitle { color: var(--text-muted); font-size: 1rem; }
.form-builder { display: flex; flex-direction: column; gap: 1.5rem; }
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
.question-card { background: var(--bg); padding: 1.5rem; border-radius: var(--radius-sm); border: 1px solid var(--border); margin-bottom: 1rem; animation: fadeUp 0.3s ease both; transition: all 0.2s; }
.question-card.drag-over { border-color: var(--primary); background: var(--primary-soft); box-shadow: 0 0 0 3px rgba(47, 79, 138, 0.1); }
.question-card:active { cursor: grabbing; }
.drag-handle { display: flex; align-items: center; justify-content: center; width: 28px; height: 28px; cursor: grab; color: var(--text-muted); border-radius: 4px; transition: all 0.2s; margin-right: 0.5rem; }
.drag-handle:hover { color: var(--primary); background: var(--primary-soft); }
.drag-handle:active { cursor: grabbing; }
.drag-handle svg { width: 18px; height: 18px; }
.question-number { font-size: 0.85rem; font-weight: 600; color: var(--primary); text-transform: uppercase; letter-spacing: 0.05em; }
.btn-remove { width: 32px; height: 32px; border: none; background: transparent; color: #c53030; cursor: pointer; border-radius: 6px; display: flex; align-items: center; justify-content: center; transition: background 0.2s; }
.btn-remove:hover { background: #fdecec; }
.btn-remove svg { width: 18px; height: 18px; }
.options-section, .dictionary-section { margin-bottom: 1.25rem; padding: 1rem; background: var(--surface); border-radius: var(--radius-sm); border: 1px dashed var(--border); }
.dictionary-section { border-color: var(--primary-soft); background: color-mix(in srgb, var(--primary-soft) 40%, var(--surface)); }
.schedule-section { border-color: var(--primary-soft); background: color-mix(in srgb, var(--primary-soft) 40%, var(--surface)); margin-bottom: 1.25rem; padding: 1rem; border-radius: var(--radius-sm); border: 1px dashed var(--primary-soft); }
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
.form-actions { display: flex; gap: 1rem; justify-content: flex-end; flex-wrap: wrap; align-items: center; }

/* ==========================================
   Sidebar Navigation
   ========================================== */
.sidebar {
  position: fixed;
  top: 0;
  left: -280px;
  width: 280px;
  height: 100vh;
  background: var(--surface);
  border-right: 1px solid var(--border);
  z-index: 100;
  transition: left 0.3s ease;
  overflow-y: auto;
  box-shadow: var(--shadow-md);
}

.sidebar.open {
  left: 0;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border);
  background: var(--bg);
  position: sticky;
  top: 0;
  z-index: 1;
}

.sidebar-header h3 {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text);
  margin: 0;
}

.sidebar-close {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.sidebar-close:hover {
  background: var(--primary-soft);
  color: var(--primary);
}

.sidebar-list {
  padding: 0.75rem 0;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.6rem 1.25rem;
  cursor: pointer;
  transition: all 0.2s;
  border-left: 3px solid transparent;
}

.sidebar-item:hover {
  background: var(--primary-soft);
}

.sidebar-item.active {
  background: var(--primary-soft);
  border-left-color: var(--primary);
}

.sidebar-item-number {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-muted);
  min-width: 20px;
}

.sidebar-item-type {
  font-size: 1rem;
}

.sidebar-item-title {
  font-size: 0.85rem;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.sidebar-toggle {
  display: none;
  position: fixed;
  top: 1rem;
  left: 1rem;
  z-index: 101;
  padding: 0.6rem 1rem;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  box-shadow: var(--shadow-sm);
}

.main-content {
  width: 100%;
}

@media (max-width: 1024px) {
  .sidebar-toggle {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }
  
  .sidebar-toggle svg {
    width: 16px;
    height: 16px;
  }
}
.responses-link-block { display: flex; align-items: center; gap: 0.5rem; padding: 0.6rem 1rem; background: var(--primary-soft); border-radius: var(--radius-sm); flex: 1; min-width: 280px; }
.responses-link-block svg { width: 16px; height: 16px; color: var(--primary); flex-shrink: 0; }
.link-label { font-size: 0.85rem; font-weight: 600; color: var(--primary); white-space: nowrap; }
.responses-link-input { flex: 1; padding: 0.4rem 0.6rem; font-size: 0.85rem; font-family: inherit; color: var(--text); background: var(--surface); border: 1px solid var(--border); border-radius: 6px; cursor: pointer; }
.responses-link-input:focus { outline: none; border-color: var(--primary); }
.btn-primary { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.85rem 2rem; background: var(--primary); color: #fff; font-size: 0.98rem; font-weight: 600; font-family: inherit; border: none; border-radius: var(--radius-sm); cursor: pointer; transition: all 0.2s; box-shadow: 0 4px 14px rgba(47, 79, 138, 0.25); min-width: 180px; justify-content: center; }
.btn-primary:hover:not(:disabled) { background: var(--primary-hover, #243f72); transform: translateY(-1px); box-shadow: 0 6px 18px rgba(47, 79, 138, 0.32); }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 720px) {
  .form-card { padding: 1.5rem 1.25rem; }
  .question-card { padding: 1.25rem 1rem; }
  .form-actions { flex-direction: column-reverse; }
  .btn-primary, .btn-secondary { width: 100%; justify-content: center; }
  .responses-link-block { min-width: 100%; }
}
</style>
