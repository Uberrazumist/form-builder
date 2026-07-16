<template>
  <div class="preview-page">
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
      <div class="preview-banner">
        <Icon name="eye" />
        <span>Это предпросмотр формы. Отправка ответов отключена.</span>
      </div>

      <div class="form-header">
        <h1 class="form-title">{{ form.title }}</h1>
        <p v-if="form.description" class="form-description">{{ form.description }}</p>
        
        <div class="progress-section">
          <div class="progress-info">
            <span class="step-counter">Шаг {{ currentVisibleStepNumber }} из {{ visibleQuestionsCount }}</span>
            <span class="progress-percent">{{ progressPercent }}%</span>
          </div>
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: `${progressPercent}%` }"></div>
          </div>
        </div>
      </div>

      <form @submit.prevent="showPreviewNotice" class="form-body" novalidate>
        <template v-for="(question, index) in form.questions" :key="question.id">
          <div
            v-if="index === currentStep && isQuestionVisible(question)"
            class="question-block"
          >
            <label :for="question.id" class="question-label">
              <span class="question-number">{{ currentVisibleStepNumber }}.</span>
              {{ question.title }}
              <span v-if="question.is_required" class="required">*</span>
            </label>

            <input
              v-if="question.type === 'text'"
              :id="question.id"
              :name="question.id"
              type="text"
              v-model="answers[question.id]"
              :required="question.is_required"
              placeholder="Введите ответ"
              class="form-input"
            />

            <textarea
              v-else-if="question.type === 'textarea'"
              :id="question.id"
              :name="question.id"
              v-model="answers[question.id]"
              :required="question.is_required"
              placeholder="Введите ответ"
              rows="4"
              class="form-textarea"
            ></textarea>

            <input
              v-else-if="question.type === 'date'"
              :id="question.id"
              :name="question.id"
              type="date"
              v-model="answers[question.id]"
              :required="question.is_required"
              class="form-input"
            />

            <select
              v-else-if="question.type === 'dictionary'"
              :id="question.id"
              :name="question.id"
              v-model="answers[question.id]"
              :required="question.is_required"
              :disabled="isSelectDisabled(question)"
              class="form-select"
            >
              <option value="" disabled>— выберите значение —</option>
              <option
                v-for="item in getFilteredOptions(question)"
                :key="getOptionValue(item)"
                :value="getOptionValue(item)"
              >
                {{ getOptionLabel(item) }}
              </option>
            </select>

            <div v-if="question.type === 'dictionary' && isSelectDisabled(question)" class="locked-hint">
              <Icon name="lock" />
              <span>{{ getLockReason(question) }}</span>
            </div>

            <div v-else-if="question.type === 'dictionary' && getFilteredOptions(question).length === 0 && !isSelectDisabled(question)" class="empty-hint">
              <Icon name="alert" />
              <span>Нет доступных вариантов</span>
            </div>

            <div v-else-if="question.type === 'radio'" class="options-group">
              <label v-for="(option, optIdx) in question.options" :key="optIdx" class="option-label">
                <input 
                  type="radio" 
                  :id="question.id + '_' + optIdx" 
                  :name="question.id" 
                  :value="option" 
                  v-model="answers[question.id]" 
                  :required="question.is_required" 
                />
                <span>{{ option }}</span>
              </label>
            </div>

            <div v-else-if="question.type === 'checkbox'" class="options-group">
              <label v-for="(option, optIdx) in question.options" :key="optIdx" class="option-label">
                <input 
                  type="checkbox" 
                  :id="question.id + '_' + optIdx" 
                  :name="question.id" 
                  :value="option" 
                  v-model="answers[question.id]" 
                />
                <span>{{ option }}</span>
              </label>
            </div>

            <select 
              v-else-if="question.type === 'select'" 
              :id="question.id" 
              :name="question.id" 
              v-model="answers[question.id]" 
              :required="question.is_required" 
              class="form-select"
            >
              <option value="" disabled>Выберите вариант</option>
              <option v-for="(option, optIdx) in question.options" :key="optIdx" :value="option">{{ option }}</option>
            </select>

            <div v-else-if="question.type === 'rating'" class="rating-group">
              <div class="stars">
                <button v-for="star in question.rating_max || 5" :key="star" type="button" @click="answers[question.id] = star" class="star-btn" :class="{ active: answers[question.id] >= star }">★</button>
              </div>
              <span v-if="answers[question.id]" class="rating-value">{{ answers[question.id] }} из {{ question.rating_max || 5 }}</span>
            </div>

            <div v-else-if="question.type === 'schedule'">
              <p class="preview-hint">Календарь бронирования (рендерится в режиме заполнения)</p>
            </div>

            <div v-if="validationError" class="validation-error">
              <Icon name="error" />
              <span>{{ validationError }}</span>
            </div>
          </div>
        </template>

        <div class="form-navigation">
          <button 
            type="button" 
            class="btn-secondary" 
            @click="prevStep" 
            :disabled="currentVisibleStepNumber === 1"
          >
            <Icon name="arrow-left" />
            Назад
          </button>
          
          <button
            v-if="currentVisibleStepNumber < visibleQuestionsCount"
            type="button"
            class="btn-primary"
            @click="nextStep"
          >
            Далее
          </button>
          
          <button
            v-else
            type="button"
            class="btn-primary"
            @click="showPreviewNotice"
          >
            Отправить ответы
          </button>
        </div>

        <div v-if="showPreviewNoticeFlag" class="preview-notice">
          <Icon name="check" />
          <span>Это предпросмотр формы. В реальном режиме здесь произошла бы отправка ответов на сервер.</span>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch, nextTick, computed } from 'vue'
import { useRoute } from 'vue-router'
import Icon from '../components/Icon.vue'

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
}

interface DictionaryItem {
  id: string
  name: string
  value?: string
  parent_id: string | null
}

const route = useRoute()
const hostOrigin = typeof window !== 'undefined' ? window.location.origin : ''

const form = ref<Form | null>(null)
const answers = reactive<Record<string, any>>({})
const dictionaryItemsCache = reactive<Record<string, DictionaryItem[]>>({})
const loading = ref(true)
const availableSlots = reactive<Record<string, any[]>>({})
const error = ref<string | null>(null)
const currentStep = ref(0)
const validationError = ref('')
const showPreviewNoticeFlag = ref(false)

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

    const data = await response.json()
    
    form.value = {
      id: data.id,
      title: data.title || '',
      description: data.description || '',
      is_public: Boolean(data.is_public),
      questions: (data.questions || []).map((q: any) => ({
        id: q.id,
        type: q.type || 'text',
        title: q.title || '',
        is_required: Boolean(q.is_required),
        dictionary_id: q.dictionary_id ?? null,
        is_booking: Boolean(q.is_booking),
        depends_on: q.depends_on ?? null,
        options: Array.isArray(q.options) ? q.options : [],
        rating_max: Number(q.rating_max) || 5,
        order_index: Number(q.order_index) || 0
      }))
    }

    for (const q of form.value.questions) {
      answers[q.id] = q.type === 'checkbox' ? [] : ''
    }

    for (const q of form.value.questions) {
      if (q.type === 'dictionary' && q.dictionary_id) {
        await loadDictionaryItems(q.dictionary_id)
      }
    }

    // ГЕНЕРАЦИЯ ФЕЙКОВЫХ СЛОТОВ ДЛЯ БРОНИРОВАНИЯ
    generateFakeSlots()

    currentStep.value = 0

  } catch (err) {
    console.error('[Preview] Load error:', err)
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const loadDictionaryItems = async (dictionaryId: string) => {
  if (dictionaryItemsCache[dictionaryId]) return

  try {
    const token = localStorage.getItem('token')
    const headers: Record<string, string> = {}
    if (token) headers['Authorization'] = `Bearer ${token}`

    const response = await fetch(`/api/dictionaries/${dictionaryId}/items`, { headers })
    if (response.ok) {
      const data = await response.json()
      dictionaryItemsCache[dictionaryId] = Array.isArray(data.items) ? data.items : (Array.isArray(data) ? data : [])
    }
  } catch (err) {
    console.error(`[Preview] Failed to load dictionary ${dictionaryId}:`, err)
    dictionaryItemsCache[dictionaryId] = []
  }
}

// ГЕНЕРАЦИЯ ФЕЙКОВЫХ СЛОТОВ ДЛЯ РЕЖИМА ПРЕДПРОСМОТРА
const generateFakeSlots = () => {
  if (!form.value) return

  const fakeSlots = [
    { id: 'fake-slot-1', name: '10:00', value: '10:00', parent_id: null },
    { id: 'fake-slot-2', name: '11:00', value: '11:00', parent_id: null },
    { id: 'fake-slot-3', name: '12:00', value: '12:00', parent_id: null },
    { id: 'fake-slot-4', name: '14:00', value: '14:00', parent_id: null },
    { id: 'fake-slot-5', name: '15:00', value: '15:00', parent_id: null }
  ]

  for (const q of form.value.questions) {
    if (q.is_booking && q.type === 'dictionary') {
      availableSlots[q.id] = fakeSlots
    }
  }
}

const isQuestionVisible = (question: Question): boolean => {
  if (!question) return false
  if (!question.depends_on) return true

  const parentAnswer = answers[question.depends_on]
  if (!parentAnswer) return false
  // Если ответ — объект (например, данные из календаря), считаем заполненным
  if (typeof parentAnswer === 'object' && parentAnswer !== null) return true
  return typeof parentAnswer === 'string' && parentAnswer.trim() !== ''
}

const isSelectDisabled = (question: Question): boolean => {
  if (!question || question.type !== 'dictionary') return false

  if (question.depends_on && !answers[question.depends_on]) return true

  if (question.is_booking) {
    const dateQuestion = form.value?.questions.find(q => q.type === 'date')
    if (dateQuestion && !answers[dateQuestion.id]) return true
  }

  return false
}

const getLockReason = (question: Question): string => {
  if (question.is_booking) {
    const dateQuestion = form.value?.questions.find(q => q.type === 'date')
    if (dateQuestion && !answers[dateQuestion.id]) {
      return 'Сначала выберите дату'
    }
  }

  if (question.depends_on) {
    const parentQuestion = form.value?.questions.find(q => q.id === question.depends_on)
    if (parentQuestion && !answers[parentQuestion.id]) {
      return `Сначала выберите: "${parentQuestion.title}"`
    }
  }

  return 'Поле заблокировано'
}

// СТРОГАЯ ФИЛЬТРАЦИЯ ПО parent_id (синхронизирована с FillFormView)
const getFilteredOptions = (question: Question): any[] => {
  if (!question || question.type !== 'dictionary' || !question.dictionary_id) {
    return question.options || []
  }

  const allItems = dictionaryItemsCache[question.dictionary_id] || []

  // Для вопросов бронирования возвращаем фейковые слоты
  if (question.is_booking) {
    return availableSlots[question.id] || []
  }

  // Если нет зависимости — возвращаем все элементы
  if (!question.depends_on) return allItems

  const parentAnswer = answers[question.depends_on]
  if (!parentAnswer || String(parentAnswer).trim() === '') return []

  // СТРОГАЯ ФИЛЬТРАЦИЯ: элемент остаётся только если его parent_id совпадает с ответом родителя
  return allItems.filter((item: DictionaryItem) => {
    return String(item.parent_id) === String(parentAnswer)
  })
}

const getOptionValue = (option: any): string => {
  if (typeof option === 'object' && option !== null) {
    return String(option.value || option.id || '')
  }
  return String(option)
}

const getOptionLabel = (option: any): string => {
  if (typeof option === 'object' && option !== null) {
    return String(option.name || option.label || option.value || '')
  }
  return String(option)
}

watch(answers, (newVal, oldVal) => {
  if (!form.value || !oldVal) return

  for (const q of form.value.questions) {
    if (q.depends_on && newVal[q.depends_on] !== oldVal[q.depends_on]) {
      const newParentValue = newVal[q.depends_on]
      const oldParentValue = oldVal[q.depends_on]

      if (oldParentValue && !newParentValue) {
        answers[q.id] = q.type === 'checkbox' ? [] : ''
      }
    }
  }

  nextTick(() => {
    if (!form.value) return
    const currentQ = form.value.questions[currentStep.value]
    if (currentQ && !isQuestionVisible(currentQ)) {
      for (let i = currentStep.value; i >= 0; i--) {
        const q = form.value.questions[i]
        if (q && isQuestionVisible(q)) {
          currentStep.value = i
          return
        }
      }
      currentStep.value = 0
    }
  })
}, { deep: true })

const visibleQuestionsCount = computed(() => {
  if (!form.value?.questions) return 0
  return form.value.questions.filter(q => q && isQuestionVisible(q)).length
})

const currentVisibleStepNumber = computed(() => {
  if (!form.value?.questions) return 1
  let visibleCount = 0
  for (let i = 0; i <= currentStep.value; i++) {
    const q = form.value.questions[i]
    if (q && isQuestionVisible(q)) visibleCount++
  }
  return Math.max(1, visibleCount)
})

const progressPercent = computed(() => {
  const total = visibleQuestionsCount.value
  return total === 0 ? 0 : Math.round((currentVisibleStepNumber.value / total) * 100)
})

const prevStep = () => {
  validationError.value = ''
  showPreviewNoticeFlag.value = false
  if (!form.value) return
  let prevIndex = currentStep.value - 1

  while (prevIndex >= 0) {
    const q = form.value.questions[prevIndex]
    if (q && isQuestionVisible(q)) break
    prevIndex--
  }
  currentStep.value = Math.max(prevIndex, 0)
}

const nextStep = () => {
  validationError.value = ''
  showPreviewNoticeFlag.value = false
  if (!form.value) return
  const currentQ = form.value.questions[currentStep.value]

  if (currentQ && isQuestionVisible(currentQ) && currentQ.is_required) {
    const answer = answers[currentQ.id]
    if (!answer || (Array.isArray(answer) && answer.length === 0) || String(answer).trim() === '') {
      validationError.value = 'Заполните обязательные поля'
      return
    }
  }

  if (currentVisibleStepNumber.value >= visibleQuestionsCount.value) {
    showPreviewNotice()
    return
  }

  let nextIndex = currentStep.value + 1
  const totalQuestions = form.value.questions.length

  while (nextIndex < totalQuestions) {
    const q = form.value.questions[nextIndex]
    if (q && isQuestionVisible(q)) break
    nextIndex++
  }

  currentStep.value = nextIndex < totalQuestions ? nextIndex : totalQuestions - 1
}

const showPreviewNotice = () => {
  validationError.value = ''
  if (!form.value) return

  for (const q of form.value.questions) {
    if (!isQuestionVisible(q)) continue
    if (q.is_required) {
      const answer = answers[q.id]
      if (!answer || (Array.isArray(answer) && answer.length === 0) || String(answer).trim() === '') {
        validationError.value = 'Заполните обязательные поля'
        return
      }
    }
  }

  validationError.value = ''
  showPreviewNoticeFlag.value = true
}
</script>

<style scoped>
.preview-page { width: 100%; max-width: 700px; margin: 0 auto; }
.loading-state, .error-state { text-align: center; padding: 4rem 2rem; }
.spinner { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon { width: 64px; height: 64px; background: #fdecec; color: #c53030; border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.error-icon svg { width: 32px; height: 32px; }
.error-state h2 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.error-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.btn-secondary { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.5rem; background: var(--surface); color: var(--text); border: 1.5px solid var(--border); border-radius: var(--radius-sm); text-decoration: none; font-weight: 600; transition: all 0.2s; cursor: pointer; }
.btn-secondary:hover:not(:disabled) { background: var(--bg); border-color: var(--text-muted); }
.btn-secondary:disabled { opacity: 0.5; cursor: not-allowed; color: var(--text-muted); }
.preview-banner { display: flex; align-items: center; gap: 0.6rem; padding: 1rem 1.25rem; background: var(--primary-soft); border: 1px solid var(--primary); border-radius: var(--radius-sm); color: var(--primary); font-size: 0.9rem; font-weight: 500; margin-bottom: 1.5rem; }
.preview-banner svg { width: 20px; height: 20px; flex-shrink: 0; }
.form-container { animation: fadeUp 0.5s ease both; }
.form-header { margin-bottom: 2rem; }
.form-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; margin-bottom: 0.75rem; }
.form-description { font-size: 1.05rem; color: var(--text-muted); line-height: 1.6; margin-bottom: 1.5rem; }
.progress-section { margin-top: 1.5rem; }
.progress-info { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; }
.step-counter { font-size: 0.9rem; font-weight: 600; color: var(--text); }
.progress-percent { font-size: 0.85rem; font-weight: 600; color: var(--primary); }
.progress-bar { width: 100%; height: 8px; background: var(--border); border-radius: var(--radius-sm); overflow: hidden; }
.progress-fill { height: 100%; background: var(--primary); border-radius: var(--radius-sm); transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.form-body { background: var(--surface); padding: 2.5rem; border-radius: var(--radius); border: 1px solid var(--border); box-shadow: var(--shadow-sm); }
.question-block { margin-bottom: 2rem; animation: fadeUp 0.4s ease both; }
.question-label { display: block; font-size: 1rem; font-weight: 600; color: var(--text); margin-bottom: 0.75rem; line-height: 1.5; }
.question-number { color: var(--primary); font-weight: 700; margin-right: 0.25rem; }
.required { color: #c53030; margin-left: 0.25rem; }
.form-input, .form-textarea, .form-select { width: 100%; padding: 0.75rem 0.95rem; font-size: 0.95rem; font-family: inherit; color: var(--text); background: var(--bg); border: 1.5px solid var(--border); border-radius: var(--radius-sm); transition: all 0.2s; resize: vertical; }
.form-input:disabled, .form-select:disabled { opacity: 0.6; cursor: not-allowed; background: #f5f5f5; }
.form-input::placeholder, .form-textarea::placeholder { color: #a6afbf; }
.form-input:hover:not(:disabled), .form-textarea:hover:not(:disabled), .form-select:hover:not(:disabled) { border-color: #cfd6e3; }
.form-input:focus, .form-textarea:focus, .form-select:focus { outline: none; border-color: var(--primary); background: var(--surface); box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1); }
.options-group { display: flex; flex-direction: column; gap: 0.75rem; }
.option-label { display: flex; align-items: center; gap: 0.6rem; cursor: pointer; font-size: 0.95rem; color: var(--text); padding: 0.6rem 0.85rem; background: var(--bg); border: 1.5px solid var(--border); border-radius: var(--radius-sm); transition: all 0.2s; }
.option-label:hover { border-color: var(--primary); background: var(--surface); }
.option-label input[type="radio"], .option-label input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; accent-color: var(--primary); }
.empty-hint, .locked-hint { display: flex; align-items: center; gap: 0.6rem; padding: 0.85rem 1rem; border-radius: var(--radius-sm); font-size: 0.9rem; margin-top: 0.75rem; }
.locked-hint { background: #f5f5f5; color: #666; border: 1px dashed var(--border); }
.preview-hint { font-size: 0.85rem; color: var(--text-muted); font-style: italic; margin-top: 0.5rem; }
.empty-hint svg, .locked-hint svg { width: 18px; height: 18px; flex-shrink: 0; }
.rating-group { display: flex; flex-direction: column; gap: 0.75rem; }
.stars { display: flex; gap: 0.5rem; }
.star-btn { width: 48px; height: 48px; border: 2px solid var(--border); background: var(--surface); color: var(--border); font-size: 1.75rem; cursor: pointer; border-radius: var(--radius-sm); transition: all 0.2s; }
.star-btn:hover { border-color: var(--primary); transform: scale(1.05); }
.star-btn.active { background: var(--primary); border-color: var(--primary); color: #fff; }
.rating-value { font-size: 0.9rem; color: var(--text-muted); font-weight: 600; }
.validation-error { display: flex; align-items: center; gap: 0.6rem; padding: 0.85rem 1rem; background: #fdecec; border: 1px solid #f5c6c6; border-radius: var(--radius-sm); color: #c53030; font-size: 0.9rem; margin-top: 1rem; }
.validation-error svg { width: 18px; height: 18px; flex-shrink: 0; }
.form-navigation { display: flex; gap: 1rem; justify-content: space-between; margin-top: 2.5rem; padding-top: 1.5rem; border-top: 1px solid var(--border); }
.btn-primary { display: inline-flex; align-items: center; justify-content: center; gap: 0.5rem; padding: 0.85rem 2rem; background: var(--primary); color: #ffffff; font-size: 1rem; font-weight: 600; font-family: inherit; border: none; border-radius: var(--radius-sm); cursor: pointer; transition: all 0.2s; box-shadow: var(--shadow-sm); min-width: 160px; }
.btn-primary:hover:not(:disabled) { background: var(--primary-hover, #243f72); transform: translateY(-1px); box-shadow: var(--shadow-md); }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; transform: none; }
.preview-notice { display: flex; align-items: center; gap: 0.6rem; padding: 1rem 1.25rem; background: #e8f5e9; border: 1px solid #4caf50; border-radius: var(--radius-sm); color: #2e7d32; font-size: 0.9rem; font-weight: 500; margin-top: 1.5rem; animation: fadeUp 0.3s ease both; }
.preview-notice svg { width: 20px; height: 20px; flex-shrink: 0; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 720px) {
  .form-body { padding: 1.5rem 1.25rem; }
  .form-title { font-size: 1.5rem; }
  .star-btn { width: 40px; height: 40px; font-size: 1.5rem; }
  .form-navigation { flex-direction: column-reverse; }
  .btn-primary, .btn-secondary { width: 100%; justify-content: center; }
}
</style>
