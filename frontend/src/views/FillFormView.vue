<!-- src/views/FillFormView.vue -->
<template>
  <div class="fill-form-page">
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
        <h1 class="form-title">{{ form.Title }}</h1>
        <p v-if="form.Description" class="form-description">{{ form.Description }}</p>
      </div>

      <form @submit.prevent="submitResponses" class="form-body" novalidate>
        <div
          v-for="(question, index) in form.Questions"
          :key="question.ID"
          class="question-block"
        >
          <label class="question-label">
            <span class="question-number">{{ index + 1 }}.</span>
            {{ question.Title }}
            <span v-if="question.Required" class="required">*</span>
          </label>

          <!-- Text -->
          <input
            v-if="question.Type === 'text'"
            type="text"
            v-model="answers[question.ID]"
            :required="question.Required"
            placeholder="Введите ответ"
            class="form-input"
          />

          <!-- Textarea -->
          <textarea
            v-else-if="question.Type === 'textarea'"
            v-model="answers[question.ID]"
            :required="question.Required"
            placeholder="Введите ответ"
            rows="4"
            class="form-textarea"
          ></textarea>

          <!-- Radio -->
          <div v-else-if="question.Type === 'radio'" class="options-group">
            <label
              v-for="(option, optIdx) in question.Options"
              :key="optIdx"
              class="option-label"
              :class="{ disabled: isOptionBusy(question, option) }"
            >
              <input
                type="radio"
                :name="'q_' + question.ID"
                :value="option.Value !== undefined ? option.Value : option"
                v-model="answers[question.ID]"
                :required="question.Required"
                :disabled="isOptionBusy(question, option)"
                @change="onAnswerChange(question)"
              />
              <span class="option-text">
                {{ option.Value !== undefined ? option.Label || option.Value : option }}
                <span v-if="isOptionBusy(question, option)" class="busy-badge">Занято</span>
                <span v-else-if="isCheckingBusy(question, option)" class="checking-badge">Проверка...</span>
              </span>
            </label>
            <div v-if="question.IsBooking && !dictionaryItems[question.ID]" class="loading-hint">
              <div class="spinner-small"></div>
              <span>Загрузка вариантов...</span>
            </div>
          </div>

          <!-- Checkbox -->
          <div v-else-if="question.Type === 'checkbox'" class="options-group">
            <label
              v-for="(option, optIdx) in question.Options"
              :key="optIdx"
              class="option-label"
              :class="{ disabled: isOptionBusy(question, option) }"
            >
              <input
                type="checkbox"
                :value="option.Value !== undefined ? option.Value : option"
                v-model="answers[question.ID]"
                :disabled="isOptionBusy(question, option) && !answers[question.ID]?.includes(option.Value !== undefined ? option.Value : option)"
                @change="onAnswerChange(question)"
              />
              <span class="option-text">
                {{ option.Value !== undefined ? option.Label || option.Value : option }}
                <span v-if="isOptionBusy(question, option)" class="busy-badge">Занято</span>
              </span>
            </label>
          </div>

          <!-- Select -->
          <select
            v-else-if="question.Type === 'select'"
            v-model="answers[question.ID]"
            :required="question.Required"
            class="form-select"
            @change="onAnswerChange(question)"
          >
            <option value="" disabled>Выберите вариант</option>
            <template v-if="question.Type === 'select' && isDictionaryQuestion(question) && dictionaryItems[question.ID]">
              <option
                v-for="item in dictionaryItems[question.ID]"
                :key="item.ID"
                :value="item.Value || item.ID"
                :disabled="question.IsBooking && bookingStatus[question.ID]?.[item.ID] === 'busy'"
              >
                {{ item.Label || item.Value }}
                <template v-if="question.IsBooking && bookingStatus[question.ID]?.[item.ID] === 'busy'"> (занято)</template>
              </option>
            </template>
            <template v-else>
              <option
                v-for="(option, optIdx) in question.Options"
                :key="optIdx"
                :value="option"
              >
                {{ option }}
              </option>
            </template>
          </select>

          <!-- Dictionary (отображается как radio/select с динамической загрузкой) -->
          <div v-else-if="question.Type === 'dictionary'" class="dictionary-options">
            <div v-if="!dictionaryItems[question.ID]" class="loading-hint">
              <div class="spinner-small"></div>
              <span>Загрузка вариантов из справочника...</span>
            </div>
            <div v-else-if="dictionaryItems[question.ID].length === 0" class="empty-hint">
              <Icon name="alert" />
              <span>Нет доступных вариантов. {{ question.DependsOn ? 'Сначала выберите значение в предыдущем вопросе.' : '' }}</span>
            </div>
            <div v-else class="options-group">
              <label
                v-for="item in dictionaryItems[question.ID]"
                :key="item.ID"
                class="option-label"
                :class="{ disabled: question.IsBooking && bookingStatus[question.ID]?.[item.ID] === 'busy' }"
              >
                <input
                  type="radio"
                  :name="'q_' + question.ID"
                  :value="item.Value || item.ID"
                  v-model="answers[question.ID]"
                  :required="question.Required"
                  :disabled="question.IsBooking && bookingStatus[question.ID]?.[item.ID] === 'busy'"
                  @change="onAnswerChange(question)"
                />
                <span class="option-text">
                  {{ item.Label || item.Value }}
                  <span v-if="question.IsBooking && bookingStatus[question.ID]?.[item.ID] === 'busy'" class="busy-badge">Занято</span>
                  <span v-else-if="question.IsBooking && checkingItems[question.ID + '_' + item.ID]" class="checking-badge">Проверка...</span>
                </span>
              </label>
            </div>
          </div>

          <!-- Rating -->
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
          <button type="submit" class="btn-primary" :disabled="submitting">
            <span v-if="!submitting">Отправить ответы</span>
            <span v-else class="spinner"></span>
          </button>
        </div>

        <FormResult v-if="result" :result="result" />
      </form>
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
const dictionaryItems = reactive({})
const bookingStatus = reactive({}) // { questionId: { itemId: 'free'|'busy' } }
const checkingItems = reactive({}) // { "questionId_itemId": true }
const loading = ref(true)
const error = ref(null)
const result = ref(null)
const submitting = ref(false)

onMounted(async () => {
  await loadForm()
})

const isDictionaryQuestion = (q) => q.Type === 'dictionary'

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
    
    // Инициализация ответов
    form.value.Questions.forEach(q => {
      if (q.Type === 'checkbox') {
        answers[q.ID] = []
      } else {
        answers[q.ID] = ''
      }
    })
    
    // Загрузка элементов справочников для всех dictionary-вопросов
    for (const q of form.value.Questions) {
      if (q.Type === 'dictionary' && q.DictionaryID) {
        await loadDictionaryItems(q)
      }
    }
  } catch (err) {
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const loadDictionaryItems = async (question) => {
  try {
    // Определяем значение родительского вопроса, если есть зависимость
    let parentValue = null
    if (question.DependsOn) {
      parentValue = answers[question.DependsOn]
      if (!parentValue) {
        dictionaryItems[question.ID] = []
        return
      }
    }
    
    const url = new URL(`/api/dictionaries/${question.DictionaryID}/items`, window.location.origin)
    if (parentValue) {
      url.searchParams.append('parent', parentValue)
    }
    
    const response = await fetch(url.toString())
    if (response.ok) {
      const data = await response.json()
      dictionaryItems[question.ID] = data.items || data || []
      
      // Если включена проверка занятости — проверяем все элементы
      if (question.IsBooking) {
        for (const item of dictionaryItems[question.ID]) {
          await checkBooking(question, item)
        }
      }
    }
  } catch (err) {
    console.error('[FillForm] Failed to load dictionary items:', err)
    dictionaryItems[question.ID] = []
  }
}

const checkBooking = async (question, item) => {
  const key = question.ID + '_' + item.ID
  checkingItems[key] = true
  
  try {
    const url = `/api/bookings/check?item_id=${item.ID}&form_id=${route.params.id}`
    const response = await fetch(url)
    
    if (!bookingStatus[question.ID]) {
      bookingStatus[question.ID] = {}
    }
    
    if (response.ok) {
      const data = await response.json()
      bookingStatus[question.ID][item.ID] = data.is_busy ? 'busy' : 'free'
    } else {
      bookingStatus[question.ID][item.ID] = 'free'
    }
  } catch (err) {
    if (!bookingStatus[question.ID]) bookingStatus[question.ID] = {}
    bookingStatus[question.ID][item.ID] = 'free'
  } finally {
    checkingItems[key] = false
  }
}

const isOptionBusy = (question, option) => {
  if (!question.IsBooking) return false
  const itemId = option.Value !== undefined ? option.Value : option
  return bookingStatus[question.ID]?.[itemId] === 'busy'
}

const isCheckingBusy = (question, option) => {
  const itemId = option.Value !== undefined ? option.Value : option
  return checkingItems[question.ID + '_' + itemId]
}

// Watch за изменениями ответов для обновления зависимых вопросов
watch(
  () => JSON.stringify(answers),
  async () => {
    if (!form.value) return
    
    // Находим вопросы, которые зависят от изменённого
    for (const q of form.value.Questions) {
      if (q.DependsOn && q.Type === 'dictionary') {
        await loadDictionaryItems(q)
      }
    }
  },
  { deep: true }
)

const onAnswerChange = async (question) => {
  // Обновление зависимых вопросов уже обрабатывается через watch
}

const submitResponses = async () => {
  for (const q of form.value.Questions) {
    if (q.Required) {
      const answer = answers[q.ID]
      if (!answer || (Array.isArray(answer) && answer.length === 0)) {
        result.value = { error: `Пожалуйста, ответьте на вопрос: "${q.Title}"` }
        return
      }
    }
  }

  submitting.value = true
  result.value = null

  try {
    const formId = route.params.id
    const token = localStorage.getItem('token')
    
    const headers = { 'Content-Type': 'application/json' }
    if (token) headers['Authorization'] = `Bearer ${token}`

    const response = await fetch('/api/responses', {
      method: 'POST',
      headers,
      body: JSON.stringify({
        form_id: formId,
        answers: answers
      })
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка отправки ответов' }
      return
    }

    result.value = { success: true, message: 'Спасибо! Ваши ответы успешно отправлены.' }
    
    form.value.Questions.forEach(q => {
      if (q.Type === 'checkbox') {
        answers[q.ID] = []
      } else {
        answers[q.ID] = ''
      }
    })
  } catch (err) {
    if (import.meta.env.DEV) {
      result.value = {
        warning: 'Демо-режим',
        message: 'Ответы не отправлены (бэкенд недоступен)',
        data: { form_id: route.params.id, answers }
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
.fill-form-page {
  width: 100%;
  max-width: 700px;
  margin: 0 auto;
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

.option-label:hover:not(.disabled) {
  border-color: var(--primary);
  background: var(--surface);
}

.option-label.disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: #f8f8f8;
}

.option-label input[type="radio"],
.option-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: var(--primary);
}

.option-label.disabled input {
  cursor: not-allowed;
}

.option-text {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
}

.busy-badge {
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  background: #fdecec;
  color: #c53030;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.checking-badge {
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  background: var(--primary-soft);
  color: var(--primary);
  border-radius: 4px;
}

.loading-hint,
.empty-hint {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.85rem 1rem;
  background: var(--primary-soft);
  border-radius: var(--radius-sm);
  color: var(--primary);
  font-size: 0.9rem;
}

.empty-hint {
  background: #fff8e1;
  color: #8a6d00;
}

.empty-hint svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.dictionary-options {
  margin-top: 0.5rem;
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
  justify-content: flex-end;
}

.btn-primary {
  padding: 0.85rem 2.5rem;
  background: var(--primary);
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
  font-family: inherit;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 14px rgba(47, 79, 138, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 200px;
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(47, 79, 138, 0.32);
}

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
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
}
</style>
