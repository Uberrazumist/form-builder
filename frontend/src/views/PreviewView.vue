<!-- src/views/PreviewView.vue -->
<template>
  <div class="preview-page">
    <div class="preview-banner">
      <div class="banner-content">
        <Icon name="eye" />
        <div>
          <strong>Режим предпросмотра</strong>
          <p>Вы видите форму так, как её увидят заполняющие. Отправка ответов в этом режиме недоступна.</p>
        </div>
      </div>
      <button @click="goBack" class="btn-close-banner">
        <Icon name="close" />
        Закрыть
      </button>
    </div>

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
      <button @click="goBack" class="btn-secondary">Назад</button>
    </div>

    <div v-else-if="form" class="form-container">
      <div class="form-header">
        <h1 class="form-title">{{ form.Title }}</h1>
        <p v-if="form.Description" class="form-description">{{ form.Description }}</p>
      </div>

      <form class="form-body" @submit.prevent="showPreviewNotice">
        <div
          v-for="(question, index) in form.Questions"
          :key="question.ID"
          v-if="isQuestionVisible(question)"
          class="question-block"
        >
          <label class="question-label">
            <span class="question-number">{{ getVisibleIndex(question) }}.</span>
            {{ question.Title }}
            <span v-if="question.IsRequired" class="required">*</span>
          </label>

          <input
            v-if="question.Type === 'text'"
            type="text"
            v-model="answers[question.ID]"
            :required="question.IsRequired"
            placeholder="Введите ответ"
            class="form-input"
          />

          <textarea
            v-else-if="question.Type === 'textarea'"
            v-model="answers[question.ID]"
            :required="question.IsRequired"
            placeholder="Введите ответ"
            rows="4"
            class="form-textarea"
          ></textarea>

          <input
            v-else-if="question.Type === 'date'"
            type="date"
            v-model="answers[question.ID]"
            :required="question.IsRequired"
            class="form-input"
          />

          <select
            v-else-if="question.Type === 'dictionary'"
            v-model="answers[question.ID]"
            :required="question.IsRequired"
            :disabled="isSelectDisabled(question)"
            class="form-select"
          >
            <option value="" disabled>— выберите значение —</option>
            <option
              v-for="option in getFilteredOptions(question)"
              :key="getOptionValue(option)"
              :value="getOptionValue(option)"
            >
              {{ getOptionLabel(option) }}
            </option>
          </select>

          <div v-if="question.Type === 'dictionary' && isSelectDisabled(question)" class="locked-hint">
            <Icon name="lock" />
            <span>{{ getLockReason(question) }}</span>
          </div>

          <div v-else-if="question.Type === 'dictionary' && isQuestionLoading(question)" class="loading-hint">
            <div class="spinner-small"></div>
            <span>Загрузка вариантов...</span>
          </div>

          <div v-else-if="question.Type === 'dictionary' && getFilteredOptions(question).length === 0 && !isSelectDisabled(question)" class="empty-hint">
            <Icon name="alert" />
            <span>Нет доступных вариантов</span>
          </div>

          <div v-else-if="question.Type === 'radio'" class="options-group">
            <label
              v-for="(option, optIdx) in question.Options"
              :key="optIdx"
              class="option-label"
            >
              <input
                type="radio"
                :name="'q_' + question.ID"
                :value="option"
                v-model="answers[question.ID]"
                :required="question.IsRequired"
              />
              <span>{{ option }}</span>
            </label>
          </div>

          <div v-else-if="question.Type === 'checkbox'" class="options-group">
            <label
              v-for="(option, optIdx) in question.Options"
              :key="optIdx"
              class="option-label"
            >
              <input
                type="checkbox"
                :value="option"
                v-model="answers[question.ID]"
              />
              <span>{{ option }}</span>
            </label>
          </div>

          <select
            v-else-if="question.Type === 'select'"
            v-model="answers[question.ID]"
            :required="question.IsRequired"
            class="form-select"
          >
            <option value="" disabled>Выберите вариант</option>
            <option
              v-for="(option, optIdx) in question.Options"
              :key="optIdx"
              :value="option"
            >
              {{ option }}
            </option>
          </select>

          <div v-else-if="question.Type === 'rating'" class="rating-group">
            <div class="stars">
              <button
                v-for="star in question.RatingMax || 5"
                :key="star"
                type="button"
                @click="answers[question.ID] = star"
                class="star-btn"
                :class="{ active: answers[question.ID] >= star }"
              >
                ★
              </button>
            </div>
            <span v-if="answers[question.ID]" class="rating-value">
              {{ answers[question.ID] }} из {{ question.RatingMax || 5 }}
            </span>
          </div>
        </div>

        <div class="form-actions">
          <button type="button" @click="goBack" class="btn-secondary-large">
            Закрыть предпросмотр
          </button>
        </div>
      </form>

      <FormResult v-if="result" :result="result" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const route = useRoute()
const router = useRouter()

const form = ref(null)
const answers = reactive({})
const dictionaryItemsCache = reactive({})
const loading = ref(true)
const loadingSlots = reactive({})
const availableSlots = reactive({})
const error = ref(null)
const result = ref(null)

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

const showPreviewNotice = () => {
  result.value = {
    warning: 'Режим предпросмотра',
    message: 'Это только предпросмотр. Для отправки ответов используйте ссылку для заполнения.'
  }
  setTimeout(() => { result.value = null }, 4000)
}

onMounted(async () => {
  await loadForm()
})

const loadForm = async () => {
  loading.value = true
  error.value = null

  try {
    const formId = route.params.id
    const token = localStorage.getItem('token')

    const headers = {}
    if (token) headers['Authorization'] = `Bearer ${token}`

    const response = await fetch(`/api/forms/${formId}`, { headers })

    if (!response.ok) {
      if (response.status === 404) error.value = 'Форма не найдена'
      else if (response.status === 403) error.value = 'У вас нет доступа к этой форме'
      else error.value = 'Не удалось загрузить форму'
      return
    }

    const data = await response.json()
    form.value = data

    form.value.Questions.forEach(q => {
      if (q.Type === 'checkbox') {
        answers[q.ID] = []
      } else {
        answers[q.ID] = ''
      }
    })

    for (const q of form.value.Questions) {
      if (q.Type === 'dictionary' && q.DictionaryID) {
        await loadDictionaryItems(q.DictionaryID)
      }
    }
  } catch (err) {
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const loadDictionaryItems = async (dictionaryId) => {
  if (dictionaryItemsCache[dictionaryId]) return

  try {
    const token = localStorage.getItem('token')
    const headers = token ? { 'Authorization': `Bearer ${token}` } : {}

    const response = await fetch(`/api/dictionaries/${dictionaryId}/items`, { headers })
    if (response.ok) {
      const data = await response.json()
      dictionaryItemsCache[dictionaryId] = data.items || data || []
    }
  } catch (err) {
    console.error('[Preview] Failed to load dictionary items:', err)
    dictionaryItemsCache[dictionaryId] = []
  }
}

const findParentQuestion = (question) => {
  if (question.Type !== 'dictionary' || !question.DictionaryID) return null

  const items = dictionaryItemsCache[question.DictionaryID] || []

  const itemWithLinks = items.find(item =>
    item.Metadata?.linked_ids &&
    Array.isArray(item.Metadata.linked_ids) &&
    item.Metadata.linked_ids.length > 0
  )

  if (!itemWithLinks) return null

  const linkedId = itemWithLinks.Metadata.linked_ids[0]

  let parentDictId = null
  for (const dictId in dictionaryItemsCache) {
    if (dictId === question.DictionaryID) continue
    const found = dictionaryItemsCache[dictId].find(i => i.ID === linkedId)
    if (found) {
      parentDictId = dictId
      break
    }
  }

  if (!parentDictId) return null

  return form.value.Questions.find(q =>
    q.Type === 'dictionary' && q.DictionaryID === parentDictId
  ) || null
}

const isQuestionVisible = (question) => {
  if (question.Type !== 'dictionary') return true

  const parentQuestion = findParentQuestion(question)
  if (!parentQuestion) return true

  if (!isQuestionVisible(parentQuestion)) return false

  const parentAnswer = answers[parentQuestion.ID]
  return !!parentAnswer
}

const getVisibleIndex = (question) => {
  let visibleCount = 0
  for (const q of form.value.Questions) {
    if (isQuestionVisible(q)) {
      visibleCount++
      if (q.ID === question.ID) return visibleCount
    }
  }
  return 0
}

const isSelectDisabled = (question) => {
  if (question.Type !== 'dictionary') return false

  const parentQuestion = findParentQuestion(question)
  
  if (parentQuestion && isSelectDisabled(parentQuestion)) return true

  if (question.IsBooking) {
    const dateQuestion = form.value.Questions.find(q => q.Type === 'date')
    if (dateQuestion && !answers[dateQuestion.ID]) return true

    if (parentQuestion && !answers[parentQuestion.ID]) return true

    return false
  }

  if (parentQuestion && !answers[parentQuestion.ID]) return true

  return false
}

const getLockReason = (question) => {
  if (question.IsBooking) {
    const dateQuestion = form.value.Questions.find(q => q.Type === 'date')
    if (dateQuestion && !answers[dateQuestion.ID]) {
      return 'Сначала выберите дату'
    }
  }

  const parentQuestion = findParentQuestion(question)
  if (parentQuestion && !answers[parentQuestion.ID]) {
    return `Сначала выберите: "${parentQuestion.Title}"`
  }

  return 'Поле заблокировано'
}

const isQuestionLoading = (question) => {
  if (question.Type !== 'dictionary' || !question.DictionaryID) return false
  return loadingSlots[question.ID] || false
}

const getFilteredOptions = (question) => {
  if (question.Type !== 'dictionary' || !question.DictionaryID) {
    return question.Options || []
  }

  const allItems = dictionaryItemsCache[question.DictionaryID] || []

  if (question.IsBooking) {
    return availableSlots[question.ID] || []
  }

  const parentQuestion = findParentQuestion(question)

  if (parentQuestion) {
    const parentAnswer = answers[parentQuestion.ID]
    if (!parentAnswer) return []

    return allItems.filter(item => {
      if (!item.Metadata?.linked_ids || !Array.isArray(item.Metadata.linked_ids)) {
        return false
      }
      return item.Metadata.linked_ids.includes(parentAnswer)
    })
  }

  return allItems
}

const getOptionValue = (option) => {
  if (typeof option === 'object' && option !== null) {
    return option.Value || option.ID
  }
  return option
}

const getOptionLabel = (option) => {
  if (typeof option === 'object' && option !== null) {
    return option.Name || option.Label || option.Value || ''
  }
  return option
}

const loadAvailableSlots = async (question) => {
  if (!question.IsBooking) return

  const dateQuestion = form.value.Questions.find(q => q.Type === 'date')
  if (!dateQuestion || !answers[dateQuestion.ID]) {
    availableSlots[question.ID] = []
    return
  }

  const parentQuestion = findParentQuestion(question)
  if (parentQuestion && !answers[parentQuestion.ID]) {
    availableSlots[question.ID] = []
    return
  }

  loadingSlots[question.ID] = true

  try {
    const token = localStorage.getItem('token')
    const dateValue = answers[dateQuestion.ID]

    const params = new URLSearchParams({
      date: dateValue
    })

    if (parentQuestion && answers[parentQuestion.ID]) {
      params.append('teacher_id', answers[parentQuestion.ID])
    }

    const url = `/api/bookings/available?${params.toString()}`

    const response = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (response.ok) {
      const data = await response.json()
      availableSlots[question.ID] = data.slots || data || []

      if (data.busy_slots) {
        availableSlots[question.ID + '_busy'] = data.busy_slots
      }
    } else {
      availableSlots[question.ID] = []
    }
  } catch (err) {
    console.error('[Preview] Failed to load available slots:', err)
    availableSlots[question.ID] = []
  } finally {
    loadingSlots[question.ID] = false
  }
}

watch(
  () => JSON.stringify(answers),
  async (newVal, oldVal) => {
    if (!form.value || newVal === oldVal) return

    const newAnswers = JSON.parse(newVal)
    const oldAnswers = oldVal ? JSON.parse(oldVal) : {}

    for (const q of form.value.Questions) {
      if (q.Type === 'dictionary') {
        const parentQuestion = findParentQuestion(q)
        
        if (parentQuestion) {
          const oldParentValue = oldAnswers[parentQuestion.ID]
          const newParentValue = newAnswers[parentQuestion.ID]
          
          if (oldParentValue !== newParentValue && oldParentValue !== '' && oldParentValue !== null) {
            if (answers[q.ID] !== '') {
              answers[q.ID] = ''
            }

            if (q.IsBooking) {
              await loadAvailableSlots(q)
            }
          }
        }
      }

      if (q.Type === 'date' && newAnswers[q.ID] !== oldAnswers[q.ID]) {
        for (const resourceQ of form.value.Questions) {
          if (resourceQ.IsBooking && resourceQ.Type === 'dictionary') {
            if (answers[resourceQ.ID] !== '') {
              answers[resourceQ.ID] = ''
            }
            await loadAvailableSlots(resourceQ)
          }
        }
      }
    }
  },
  { deep: true }
)
</script>

<style scoped>
.preview-page {
  width: 100%;
  max-width: 700px;
  margin: 0 auto;
}

.preview-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, #fff4e0 0%, #ffe8b8 100%);
  border: 1px solid #f5c87a;
  border-radius: var(--radius);
  margin-bottom: 2rem;
  animation: fadeUp 0.4s ease both;
}

.banner-content {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  flex: 1;
}

.banner-content > svg {
  width: 24px;
  height: 24px;
  color: #b7791f;
  flex-shrink: 0;
  margin-top: 2px;
}

.banner-content strong {
  display: block;
  font-size: 0.95rem;
  font-weight: 700;
  color: #8a5a00;
  margin-bottom: 0.2rem;
}

.banner-content p {
  font-size: 0.85rem;
  color: #8a5a00;
  line-height: 1.4;
}

.btn-close-banner {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.5rem 0.9rem;
  background: rgba(255,255,255,0.6);
  color: #8a5a00;
  border: 1px solid #f5c87a;
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  flex-shrink: 0;
}

.btn-close-banner:hover {
  background: #fff;
}

.btn-close-banner svg {
  width: 14px;
  height: 14px;
}

.loading-state,
.error-state {
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

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-icon {
  width: 64px;
  height: 64px;
  background: #fdecec;
  color: #c53030;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.5rem;
}

.error-icon svg {
  width: 32px;
  height: 32px;
}

.error-state h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.5rem;
}

.error-state p {
  color: var(--text-muted);
  margin-bottom: 1.5rem;
}

.btn-secondary {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  background: var(--surface);
  color: var(--text);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  text-decoration: none;
  font-weight: 600;
  transition: all 0.2s;
  cursor: pointer;
}

.btn-secondary:hover {
  background: var(--bg);
  border-color: #cfd6e3;
}

.form-container {
  animation: fadeUp 0.5s ease both;
}

.form-header {
  margin-bottom: 2rem;
}

.form-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.02em;
  margin-bottom: 0.75rem;
}

.form-description {
  font-size: 1.05rem;
  color: var(--text-muted);
  line-height: 1.6;
}

.form-body {
  background: var(--surface);
  padding: 2.5rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
}

.question-block {
  margin-bottom: 2rem;
  animation: fadeUp 0.4s ease both;
}

.question-label {
  display: block;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.75rem;
  line-height: 1.5;
}

.question-number {
  color: var(--primary);
  font-weight: 700;
  margin-right: 0.25rem;
}

.required {
  color: #c53030;
  margin-left: 0.25rem;
}

.form-input,
.form-textarea,
.form-select {
  width: 100%;
  padding: 0.75rem 0.95rem;
  font-size: 0.95rem;
  font-family: inherit;
  color: var(--text);
  background: var(--bg);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  transition: all 0.2s;
  resize: vertical;
}

.form-input::placeholder,
.form-textarea::placeholder {
  color: #a6afbf;
}

.form-input:hover,
.form-textarea:hover,
.form-select:hover {
  border-color: #cfd6e3;
}

.form-input:focus,
.form-textarea:focus,
.form-select:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--surface);
  box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1);
}

.options-group {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.option-label {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  cursor: pointer;
  font-size: 0.95rem;
  color: var(--text);
  padding: 0.6rem 0.85rem;
  background: var(--bg);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  transition: all 0.2s;
}

.option-label:hover {
  border-color: var(--primary);
  background: var(--surface);
}

.option-label input[type="radio"],
.option-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: var(--primary);
}

.loading-hint,
.empty-hint,
.locked-hint {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.85rem 1rem;
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
  margin-top: 0.75rem;
}

.loading-hint {
  background: var(--primary-soft);
  color: var(--primary);
}

.empty-hint {
  background: #fff8e1;
  color: #8a6d00;
}

.locked-hint {
  background: #f5f5f5;
  color: #666;
  border: 1px dashed var(--border);
}

.empty-hint svg,
.locked-hint svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.rating-group {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.stars {
  display: flex;
  gap: 0.5rem;
}

.star-btn {
  width: 48px;
  height: 48px;
  border: 2px solid var(--border);
  background: var(--surface);
  color: var(--border);
  font-size: 1.75rem;
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all 0.2s;
}

.star-btn:hover {
  border-color: var(--primary);
  transform: scale(1.05);
}

.star-btn.active {
  background: var(--primary);
  border-color: var(--primary);
  color: #fff;
}

.rating-value {
  font-size: 0.9rem;
  color: var(--text-muted);
  font-weight: 600;
}

.form-actions {
  margin-top: 2.5rem;
  display: flex;
  justify-content: center;
}

.btn-secondary-large {
  padding: 0.85rem 2.5rem;
  background: var(--surface);
  color: var(--text);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary-large:hover {
  background: var(--bg);
  border-color: #cfd6e3;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 720px) {
  .form-body {
    padding: 1.5rem 1.25rem;
  }
  .form-title {
    font-size: 1.5rem;
  }
  .star-btn {
    width: 40px;
    height: 40px;
    font-size: 1.5rem;
  }
  .preview-banner {
    flex-direction: column;
    align-items: stretch;
  }
  .btn-close-banner {
    align-self: flex-end;
  }
}
</style>
