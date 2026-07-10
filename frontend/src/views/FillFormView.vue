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

          <!-- Date -->
          <input
            v-else-if="question.Type === 'date'"
            type="date"
            v-model="answers[question.ID]"
            :required="question.Required"
            class="form-input"
          />

          <!-- Radio -->
          <div v-else-if="question.Type === 'radio'" class="options-group">
            <label
              v-for="(option, optIdx) in getOptionsForQuestion(question)"
              :key="optIdx"
              class="option-label"
              :class="{ disabled: isOptionBusy(question, option) }"
            >
              <input
                type="radio"
                :name="'q_' + question.ID"
                :value="getOptionValue(option)"
                v-model="answers[question.ID]"
                :required="question.Required"
                :disabled="isOptionBusy(question, option)"
              />
              <span class="option-text">
                {{ getOptionLabel(option) }}
                <span v-if="isOptionBusy(question, option)" class="busy-badge">Занято</span>
              </span>
            </label>
            <div v-if="getQuestionRole(question) === 'teacher' && filteredTeachers.length === 0 && answers[questionIds.class]" class="empty-hint">
              <Icon name="alert" />
              <span>Для выбранного класса нет учителей</span>
            </div>
            <div v-if="getQuestionRole(question) === 'time' && isTimeLocked" class="locked-hint">
              <Icon name="lock" />
              <span>Сначала выберите учителя и дату</span>
            </div>
            <div v-if="getQuestionRole(question) === 'time' && !isTimeLocked && loadingSlots" class="loading-hint">
              <div class="spinner-small"></div>
              <span>Загрузка доступного времени...</span>
            </div>
            <div v-if="getQuestionRole(question) === 'time' && !isTimeLocked && !loadingSlots && availableSlots.length === 0" class="empty-hint">
              <Icon name="alert" />
              <span>На эту дату свободных слотов нет</span>
            </div>
          </div>

          <!-- Checkbox -->
          <div v-else-if="question.Type === 'checkbox'" class="options-group">
            <label
              v-for="(option, optIdx) in getOptionsForQuestion(question)"
              :key="optIdx"
              class="option-label"
            >
              <input
                type="checkbox"
                :value="getOptionValue(option)"
                v-model="answers[question.ID]"
              />
              <span>{{ getOptionLabel(option) }}</span>
            </label>
          </div>

          <!-- Select -->
          <select
            v-else-if="question.Type === 'select'"
            v-model="answers[question.ID]"
            :required="question.Required"
            :disabled="isTimeLocked && getQuestionRole(question) === 'time'"
            class="form-select"
          >
            <option value="" disabled>Выберите вариант</option>
            <option
              v-for="(option, optIdx) in getOptionsForQuestion(question)"
              :key="optIdx"
              :value="getOptionValue(option)"
            >
              {{ getOptionLabel(option) }}
            </option>
          </select>

          <!-- Dictionary (общий случай) -->
          <div v-else-if="question.Type === 'dictionary'" class="dictionary-options">
            <div v-if="getQuestionRole(question) === 'time' && isTimeLocked" class="locked-hint">
              <Icon name="lock" />
              <span>Сначала выберите учителя и дату</span>
            </div>
            <div v-else-if="getQuestionRole(question) === 'time' && loadingSlots" class="loading-hint">
              <div class="spinner-small"></div>
              <span>Загрузка доступного времени...</span>
            </div>
            <div v-else-if="getQuestionRole(question) === 'time' && !loadingSlots && availableSlots.length === 0" class="empty-hint">
              <Icon name="alert" />
              <span>На эту дату свободных слотов нет</span>
            </div>
            <div v-else-if="getQuestionRole(question) === 'teacher' && filteredTeachers.length === 0 && answers[questionIds.class]" class="empty-hint">
              <Icon name="alert" />
              <span>Для выбранного класса нет учителей</span>
            </div>
            <div v-else class="options-group">
              <label
                v-for="option in getOptionsForQuestion(question)"
                :key="getOptionValue(option)"
                class="option-label"
                :class="{ disabled: isOptionBusy(question, option) }"
              >
                <input
                  type="radio"
                  :name="'q_' + question.ID"
                  :value="getOptionValue(option)"
                  v-model="answers[question.ID]"
                  :required="question.Required"
                  :disabled="isOptionBusy(question, option) || (isTimeLocked && getQuestionRole(question) === 'time')"
                />
                <span class="option-text">
                  {{ getOptionLabel(option) }}
                  <span v-if="isOptionBusy(question, option)" class="busy-badge">Занято</span>
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
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const route = useRoute()
const router = useRouter()

const form = ref(null)
const answers = reactive({})
const bookingStatus = reactive({})
const loading = ref(true)
const loadingSlots = ref(false)
const error = ref(null)
const result = ref(null)
const submitting = ref(false)
const availableSlots = ref([])

// Кэши для справочников
const dictionaryItemsCache = reactive({})
const dictionaryMeta = reactive({}) // { dictId: { Name, ... } }

// ID вопросов по ролям
const questionIds = reactive({
  class: null,
  teacher: null,
  date: null,
  time: null
})

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
    
    // Инициализация ответов
    form.value.Questions.forEach(q => {
      if (q.Type === 'checkbox') {
        answers[q.ID] = []
      } else {
        answers[q.ID] = ''
      }
    })
    
    // Определяем роли вопросов и загружаем справочники
    await identifyQuestionRoles()
    
    // Загружаем элементы справочников для всех dictionary-вопросов
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

const identifyQuestionRoles = async () => {
  if (!form.value?.Questions) return
  
  const token = localStorage.getItem('token')
  const headers = token ? { 'Authorization': `Bearer ${token}` } : {}
  
  // Загружаем информацию о справочниках
  try {
    const dictsResponse = await fetch('/api/dictionaries', { headers })
    if (dictsResponse.ok) {
      const dictsData = await dictsResponse.json()
      const dicts = dictsData.dictionaries || dictsData || []
      dicts.forEach(d => {
        dictionaryMeta[d.ID] = d
      })
    }
  } catch (err) {
    console.error('[FillForm] Failed to load dictionaries meta:', err)
  }
  
  // Определяем роли
  for (const q of form.value.Questions) {
    if (q.Type === 'date') {
      questionIds.date = q.ID
    } else if (q.Type === 'dictionary' && q.DictionaryID) {
      const dictName = (dictionaryMeta[q.DictionaryID]?.Name || '').toLowerCase()
      
      if (dictName.includes('класс') || dictName.includes('class')) {
        questionIds.class = q.ID
      } else if (dictName.includes('учитель') || dictName.includes('teacher')) {
        questionIds.teacher = q.ID
      } else if (q.IsBooking || dictName.includes('время') || dictName.includes('time')) {
        questionIds.time = q.ID
      }
    }
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
    console.error('[FillForm] Failed to load dictionary items:', err)
    dictionaryItemsCache[dictionaryId] = []
  }
}

// Определяет роль вопроса: 'class', 'teacher', 'date', 'time' или 'other'
const getQuestionRole = (question) => {
  if (question.ID === questionIds.class) return 'class'
  if (question.ID === questionIds.teacher) return 'teacher'
  if (question.ID === questionIds.date) return 'date'
  if (question.ID === questionIds.time) return 'time'
  return 'other'
}

// Получает отфильтрованные варианты для вопроса
const getOptionsForQuestion = (question) => {
  const role = getQuestionRole(question)
  
  if (role === 'teacher') {
    return filteredTeachers.value
  }
  
  if (role === 'time') {
    if (isTimeLocked.value) return []
    return availableSlots.value
  }
  
  if (question.Type === 'dictionary' && question.DictionaryID) {
    return dictionaryItemsCache[question.DictionaryID] || []
  }
  
  return question.Options || []
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

// Фильтрация учителей по выбранному классу
const filteredTeachers = computed(() => {
  const teacherQuestion = form.value?.Questions.find(q => q.ID === questionIds.teacher)
  if (!teacherQuestion || !teacherQuestion.DictionaryID) return []
  
  const allTeachers = dictionaryItemsCache[teacherQuestion.DictionaryID] || []
  const selectedClassId = answers[questionIds.class]
  
  if (!selectedClassId) return allTeachers
  
  // Фильтруем: показываем только тех учителей, у которых в metadata.class_ids есть выбранный класс
  return allTeachers.filter(teacher => {
    if (!teacher.Metadata?.class_ids || !Array.isArray(teacher.Metadata.class_ids)) {
      return false
    }
    return teacher.Metadata.class_ids.includes(selectedClassId)
  })
})

// Блокировка времени
const isTimeLocked = computed(() => {
  const teacherAnswer = answers[questionIds.teacher]
  const dateAnswer = answers[questionIds.date]
  return !teacherAnswer || !dateAnswer
})

// Загрузка доступных слотов
const loadAvailableSlots = async () => {
  const teacherId = answers[questionIds.teacher]
  const dateValue = answers[questionIds.date]
  
  if (!teacherId || !dateValue) {
    availableSlots.value = []
    return
  }
  
  loadingSlots.value = true
  try {
    const token = localStorage.getItem('token')
    const url = `/api/bookings/available?teacher_id=${encodeURIComponent(teacherId)}&date=${encodeURIComponent(dateValue)}`
    
    const response = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      availableSlots.value = data.slots || data || []
    } else {
      availableSlots.value = []
    }
  } catch (err) {
    console.error('[FillForm] Failed to load available slots:', err)
    availableSlots.value = []
  } finally {
    loadingSlots.value = false
  }
}

const isOptionBusy = (question, option) => {
  if (!question.IsBooking) return false
  const itemId = getOptionValue(option)
  return bookingStatus[question.ID]?.[itemId] === 'busy'
}

// Каскадные watch'и
watch(
  () => answers[questionIds.class],
  (newVal, oldVal) => {
    if (newVal === oldVal) return
    // При изменении класса — сбрасываем учителя, дату и время
    if (questionIds.teacher) answers[questionIds.teacher] = ''
    if (questionIds.date) answers[questionIds.date] = ''
    if (questionIds.time) answers[questionIds.time] = ''
    availableSlots.value = []
  }
)

watch(
  () => answers[questionIds.teacher],
  (newVal, oldVal) => {
    if (newVal === oldVal) return
    // При изменении учителя — сбрасываем дату и время
    if (questionIds.date) answers[questionIds.date] = ''
    if (questionIds.time) answers[questionIds.time] = ''
    availableSlots.value = []
  }
)

watch(
  () => answers[questionIds.date],
  (newVal, oldVal) => {
    if (newVal === oldVal) return
    // При изменении даты — сбрасываем время и делаем запрос слотов
    if (questionIds.time) answers[questionIds.time] = ''
    
    if (answers[questionIds.teacher] && newVal) {
      loadAvailableSlots()
    } else {
      availableSlots.value = []
    }
  }
)

// Отдельный watch на учителя для загрузки слотов (когда уже есть дата)
watch(
  () => answers[questionIds.teacher],
  (newVal) => {
    if (newVal && answers[questionIds.date]) {
      loadAvailableSlots()
    }
  }
)

const submitResponses = async () => {
  // Валидация обязательных полей
  for (const q of form.value.Questions) {
    if (q.Required) {
      const answer = answers[q.ID]
      if (!answer || (Array.isArray(answer) && answer.length === 0)) {
        result.value = { error: `Пожалуйста, ответьте на вопрос: "${q.Title}"` }
        return
      }
    }
  }

  // Дополнительная проверка: если есть вопрос времени, он должен быть заполнен
  if (questionIds.time && !answers[questionIds.time]) {
    result.value = { error: 'Пожалуйста, выберите время' }
    return
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

    // Обработка ошибки 409 Conflict (слот уже занят)
    if (response.status === 409) {
      result.value = { 
        error: 'Извините, это время только что заняли. Пожалуйста, выберите другое время.' 
      }
      // Перезагружаем доступные слоты
      await loadAvailableSlots()
      return
    }

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка отправки ответов' }
      return
    }

    result.value = { success: true, message: 'Спасибо! Ваши ответы успешно отправлены.' }
    
    // Сброс формы
    form.value.Questions.forEach(q => {
      if (q.Type === 'checkbox') {
        answers[q.ID] = []
      } else {
        answers[q.ID] = ''
      }
    })
    availableSlots.value = []
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

.form-input:disabled,
.form-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: #f5f5f5;
}

.form-input::placeholder,
.form-textarea::placeholder {
  color: #a6afbf;
}

.form-input:hover:not(:disabled),
.form-textarea:hover:not(:disabled),
.form-select:hover:not(:disabled) {
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

.loading-hint,
.empty-hint,
.locked-hint {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.85rem 1rem;
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
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
