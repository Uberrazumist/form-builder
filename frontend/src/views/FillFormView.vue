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

      <form @submit.prevent="submitResponses" class="form-body" novalidate>
        <div
          v-for="(question, index) in form.Questions"
          :key="question.ID"
          v-if="index === currentStep && isQuestionVisible(question)"
          class="question-block"
        >
          <label :for="question.ID" class="question-label">
            <span class="question-number">{{ currentVisibleStepNumber }}.</span>
            {{ question.Title }}
            <span v-if="question.required" class="required">*</span>
          </label>

          <input
            v-if="question.Type === 'text'"
            :id="question.ID"
            :name="question.ID"
            type="text"
            v-model="answers[question.ID]"
            :required="question.required"
            placeholder="Введите ответ"
            class="form-input"
          />

          <textarea
            v-else-if="question.Type === 'textarea'"
            :id="question.ID"
            :name="question.ID"
            v-model="answers[question.ID]"
            :required="question.required"
            placeholder="Введите ответ"
            rows="4"
            class="form-textarea"
          ></textarea>

          <input
            v-else-if="question.Type === 'date'"
            :id="question.ID"
            :name="question.ID"
            type="date"
            v-model="answers[question.ID]"
            :required="question.required"
            class="form-input"
          />

          <select
            v-else-if="question.Type === 'dictionary'"
            :id="question.ID"
            :name="question.ID"
            v-model="answers[question.ID]"
            :required="question.required"
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
            <label v-for="(option, optIdx) in question.Options" :key="optIdx" class="option-label">
              <input 
                type="radio" 
                :id="question.ID + '_' + optIdx" 
                :name="question.ID" 
                :value="option" 
                v-model="answers[question.ID]" 
                :required="question.required" 
              />
              <span>{{ option }}</span>
            </label>
          </div>

          <div v-else-if="question.Type === 'checkbox'" class="options-group">
            <label v-for="(option, optIdx) in question.Options" :key="optIdx" class="option-label">
              <input 
                type="checkbox" 
                :id="question.ID + '_' + optIdx" 
                :name="question.ID" 
                :value="option" 
                v-model="answers[question.ID]" 
              />
              <span>{{ option }}</span>
            </label>
          </div>

          <select 
            v-else-if="question.Type === 'select'" 
            :id="question.ID" 
            :name="question.ID" 
            v-model="answers[question.ID]" 
            :required="question.required" 
            class="form-select"
          >
            <option value="" disabled>Выберите вариант</option>
            <option v-for="(option, optIdx) in question.Options" :key="optIdx" :value="option">{{ option }}</option>
          </select>

          <div v-else-if="question.Type === 'rating'" class="rating-group">
            <div class="stars">
              <button v-for="star in question.RatingMax || 5" :key="star" type="button" @click="answers[question.ID] = star" class="star-btn" :class="{ active: answers[question.ID] >= star }">★</button>
            </div>
            <span v-if="answers[question.ID]" class="rating-value">{{ answers[question.ID] }} из {{ question.RatingMax || 5 }}</span>
          </div>

          <div v-if="validationError" class="validation-error">
            <Icon name="error" />
            <span>{{ validationError }}</span>
          </div>
        </div>

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
            @click="submitResponses"
            :disabled="submitting"
          >
            <span v-if="!submitting">Отправить ответы</span>
            <span v-else class="spinner-small"></span>
          </button>
        </div>

        <FormResult v-if="result" :result="result" />
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, nextTick, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const route = useRoute()
const router = useRouter()
const hostOrigin = typeof window !== 'undefined' ? window.location.origin : ''

const form = ref(null)
const answers = reactive({})
const dictionaryItemsCache = reactive({})
const loading = ref(true)
const loadingSlots = reactive({})
const availableSlots = reactive({})
const error = ref(null)
const result = ref(null)
const submitting = ref(false)
const currentStep = ref(0)
const validationError = ref('')
const isProcessingWatch = ref(false)

onMounted(async () => {
  await loadForm()
})

const loadForm = async () => {
  loading.value = true
  error.value = null

  try {
    const formId = String(route.params.id)
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

    const rawForm = await response.json()
    
    form.value = {
      ID: rawForm.ID || rawForm.id,
      Title: rawForm.Title || rawForm.title || '',
      Description: rawForm.Description || rawForm.description || '',
      IsPublic: rawForm.IsPublic !== undefined ? rawForm.IsPublic : (rawForm.is_public !== undefined ? rawForm.is_public : true),
      Questions: rawForm.Questions || rawForm.questions || []
    }

    // 1. Схема Нормализации при ПОЛУЧЕНИИ данных
    for (const q of form.value.Questions) {
      if (!q) continue
      q.ID = q.ID || q.id
      q.Type = q.Type || q.type || 'text'
      q.Title = q.Title || q.title || ''
      q.DictionaryID = q.DictionaryID || q.dictionary_id || null
      q.IsBooking = q.IsBooking !== undefined ? q.IsBooking : (q.is_booking !== undefined ? q.is_booking : false)
      q.required = q.is_required || q.IsRequired || q.Required || false
      q.Options = q.Options || q.options || []
      q.RatingMax = q.RatingMax || q.rating_max || 5
    }

    form.value.Questions.forEach(q => {
      if (!q) return
      if (q.Type === 'checkbox') {
        answers[q.ID] = []
      } else {
        answers[q.ID] = ''
      }
    })

    for (const q of form.value.Questions) {
      if (!q) continue
      if (q.Type === 'dictionary' && q.DictionaryID) {
        await loadDictionaryItems(q.DictionaryID)
      }
    }

    currentStep.value = 0

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
    console.error('[FillForm] Failed to load dictionary items:', err)
    dictionaryItemsCache[dictionaryId] = []
  }
}

const findParentQuestion = (question) => {
  if (!question || question.Type !== 'dictionary' || !question.DictionaryID) return null

  const items = dictionaryItemsCache[question.DictionaryID]
  if (!items || items.length === 0) return null

  let sampleLinkedId = null
  for (const item of items) {
    if (!item) continue
    const firstLinkedId = item?.Metadata?.linked_ids?.[0]
    if (firstLinkedId) {
      sampleLinkedId = firstLinkedId
      break
    }
  }

  if (!sampleLinkedId) return null

  let parentDictId = null
  for (const dictId in dictionaryItemsCache) {
    if (dictId === question.DictionaryID) continue
    const dictItems = dictionaryItemsCache[dictId] || []
    const found = dictItems.find(i => i?.ID === sampleLinkedId)
    if (found) {
      parentDictId = dictId
      break
    }
  }

  if (!parentDictId) return null

  return form.value?.Questions?.find(q => q?.Type === 'dictionary' && q?.DictionaryID === parentDictId) || null
}

const isQuestionVisible = (q) => {
  if (!q) return false
  if (q.Type !== 'dictionary') return true

  const firstDict = form.value?.Questions?.find(item => item?.Type === 'dictionary')
  if (firstDict && q.ID === firstDict.ID) return true

  const parentQuestion = findParentQuestion(q)
  if (!parentQuestion) return true

  const parentAnswer = answers[parentQuestion.ID]
  return !!parentAnswer && parentAnswer !== ''
}

const isSelectDisabled = (question) => {
  if (!question || question.Type !== 'dictionary') return false

  const parentQuestion = findParentQuestion(question)
  if (parentQuestion && isSelectDisabled(parentQuestion)) return true

  if (question.IsBooking) {
    const dateQuestion = form.value.Questions.find(q => q?.Type === 'date')
    if (dateQuestion && !answers[dateQuestion.ID]) return true
    if (parentQuestion && !answers[parentQuestion.ID]) return true
    return false
  }

  if (parentQuestion && !answers[parentQuestion.ID]) return true
  return false
}

const getLockReason = (question) => {
  if (!question) return 'Поле заблокировано'

  if (question.IsBooking) {
    const dateQuestion = form.value.Questions.find(q => q?.Type === 'date')
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
  if (!question || question.Type !== 'dictionary' || !question.DictionaryID) return false
  return loadingSlots[question.ID] || false
}

const getFilteredOptions = (question) => {
  if (!question || question.Type !== 'dictionary' || !question.DictionaryID) {
    return question?.Options || []
  }

  const allItems = dictionaryItemsCache[question.DictionaryID] || []

  if (question.IsBooking) {
    return availableSlots[question.ID] || []
  }

  const parentQuestion = findParentQuestion(question)
  if (!parentQuestion) return allItems

  const parentAnswer = answers[parentQuestion.ID]
  if (!parentAnswer) return []

  return allItems.filter(item => {
    if (!item?.Metadata?.linked_ids || !Array.isArray(item.Metadata.linked_ids)) {
      return false
    }
    return item.Metadata.linked_ids.includes(parentAnswer)
  })
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
  if (!question || !question.IsBooking) return

  const dateQuestion = form.value.Questions.find(q => q?.Type === 'date')
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

    const params = new URLSearchParams({ date: dateValue })

    if (parentQuestion && answers[parentQuestion.ID]) {
      params.append('teacher_id', answers[parentQuestion.ID])
    }

    const url = `/api/bookings/available?${params.toString()}`

    const response = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` }
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
    console.error('[FillForm] Failed to load available slots:', err)
    availableSlots[question.ID] = []
  } finally {
    loadingSlots[question.ID] = false
  }
}

watch(
  () => JSON.stringify(answers),
  async (newVal, oldVal) => {
    if (!form.value || newVal === oldVal || isProcessingWatch.value) return

    isProcessingWatch.value = true

    try {
      const newAnswers = JSON.parse(newVal)
      const oldAnswers = oldVal ? JSON.parse(oldVal) : {}
      const questionsToReset = []

      for (const q of form.value.Questions) {
        if (!q) continue

        if (q.Type === 'dictionary') {
          const parentQuestion = findParentQuestion(q)
          if (parentQuestion) {
            const oldParentValue = oldAnswers[parentQuestion.ID]
            const newParentValue = newAnswers[parentQuestion.ID]

            if (oldParentValue !== newParentValue && oldParentValue !== '' && oldParentValue !== null && oldParentValue !== undefined) {
              if (answers[q.ID] !== '' && answers[q.ID] !== null && answers[q.ID] !== undefined) {
                questionsToReset.push(q.ID)
              }
              if (q.IsBooking) {
                await loadAvailableSlots(q)
              }
            }
          }
        }

        if (q.Type === 'date' && newAnswers[q.ID] !== oldAnswers[q.ID]) {
          for (const resourceQ of form.value.Questions) {
            if (!resourceQ) continue
            if (resourceQ.IsBooking && resourceQ.Type === 'dictionary') {
              if (answers[resourceQ.ID] !== '' && answers[resourceQ.ID] !== null && answers[resourceQ.ID] !== undefined) {
                questionsToReset.push(resourceQ.ID)
              }
              await loadAvailableSlots(resourceQ)
            }
          }
        }
      }

      if (questionsToReset.length > 0) {
        await nextTick()
        for (const qId of questionsToReset) {
          if (answers[qId] !== '') {
            answers[qId] = ''
          }
        }
      }
    } finally {
      await nextTick()
      isProcessingWatch.value = false
    }
  },
  { deep: true }
)

const visibleQuestionsCount = computed(() => {
  if (!form.value?.Questions) return 0
  let count = 0
  for (const q of form.value.Questions) {
    if (!q) continue
    if (isQuestionVisible(q)) count++
  }
  return count
})

const currentVisibleStepNumber = computed(() => {
  if (!form.value?.Questions) return 1
  let visibleCount = 0
  for (let i = 0; i <= currentStep.value; i++) {
    const q = form.value.Questions[i]
    if (!q) continue
    if (isQuestionVisible(q)) visibleCount++
  }
  return Math.max(1, visibleCount)
})

const progressPercent = computed(() => {
  const total = visibleQuestionsCount.value
  if (total === 0) return 0
  return Math.round((currentVisibleStepNumber.value / total) * 100)
})

const prevStep = () => {
  validationError.value = ''
  let prevIndex = currentStep.value - 1

  while (prevIndex >= 0) {
    const q = form.value?.Questions?.[prevIndex]
    if (!q) {
      prevIndex--
      continue
    }
    if (isQuestionVisible(q)) {
      break
    }
    prevIndex--
  }

  currentStep.value = Math.max(prevIndex, 0)
}

const nextStep = () => {
  validationError.value = ''
  const currentQ = form.value?.Questions?.[currentStep.value]

  if (currentQ && isQuestionVisible(currentQ)) {
    if (currentQ.required) {
      const answer = answers[currentQ.ID]
      if (!answer || (Array.isArray(answer) && answer.length === 0) || answer === '') {
        validationError.value = `Заполните обязательные поля`
        return
      }
    }
  }

  if (currentVisibleStepNumber.value >= visibleQuestionsCount.value) {
    submitResponses()
    return
  }

  let nextIndex = currentStep.value + 1
  const totalQuestions = form.value?.Questions?.length || 0

  while (nextIndex < totalQuestions) {
    const q = form.value.Questions[nextIndex]
    if (!q) {
      nextIndex++
      continue
    }
    if (isQuestionVisible(q)) {
      break
    }
    nextIndex++
  }

  currentStep.value = nextIndex < totalQuestions ? nextIndex : totalQuestions - 1
}

const submitResponses = async () => {
  validationError.value = ''

  for (const q of form.value.Questions) {
    if (!q) continue
    if (!isQuestionVisible(q)) continue

    if (q.required) {
      const answer = answers[q.ID]
      if (!answer || (Array.isArray(answer) && answer.length === 0) || answer === '') {
        validationError.value = `Заполните обязательные поля`
        return
      }
    }
  }

  submitting.value = true
  result.value = null

  try {
    const formId = String(route.params.id)
    const token = localStorage.getItem('token')

    const headers = { 'Content-Type': 'application/json' }
    if (token) headers['Authorization'] = `Bearer ${token}`

    const payloadAnswers = {}
    for (const q of form.value.Questions) {
      if (!q) continue
      if (!isQuestionVisible(q)) continue
      
      const answer = answers[q.ID]
      if (answer !== '' && answer !== null && answer !== undefined) {
        payloadAnswers[q.ID] = answer
      }
    }

    const response = await fetch('/api/responses', {
      method: 'POST',
      headers,
      body: JSON.stringify({
        form_id: formId,
        answers: payloadAnswers
      })
    })

    if (response.status === 409) {
      result.value = {
        error: 'Извините, это время только что заняли. Пожалуйста, выберите другое время.'
      }
      for (const q of form.value.Questions) {
        if (!q) continue
        if (q.IsBooking && q.Type === 'dictionary') {
          await loadAvailableSlots(q)
        }
      }
      return
    }

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка отправки ответов' }
      return
    }

    result.value = { success: true, message: 'Спасибо! Ваши ответы успешно отправлены.' }

    form.value.Questions.forEach(q => {
      if (!q) return
      if (q.Type === 'checkbox') {
        answers[q.ID] = []
      } else {
        answers[q.ID] = ''
      }
    })

    Object.keys(availableSlots).forEach(key => {
      availableSlots[key] = []
    })

    currentStep.value = 0
  } catch (err) {
    if (import.meta.env.DEV) {
      result.value = {
        warning: 'Демо-режим',
        message: 'Ответы не отправлены (бэкенд недоступен)',
        data: { form_id: formId, answers }
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
.fill-form-page { width: 100%; max-width: 700px; margin: 0 auto; }
.loading-state, .error-state { text-align: center; padding: 4rem 2rem; }
.spinner { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
.spinner-small { width: 18px; height: 18px; border: 2px solid rgba(255, 255, 255, 0.3); border-top-color: #ffffff; border-radius: 50%; animation: spin 0.7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon { width: 64px; height: 64px; background: #fdecec; color: #c53030; border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.error-icon svg { width: 32px; height: 32px; }
.error-state h2 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.error-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.btn-secondary { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.5rem; background: var(--surface); color: var(--text); border: 1.5px solid var(--border); border-radius: var(--radius-sm); text-decoration: none; font-weight: 600; transition: all 0.2s; cursor: pointer; }
.btn-secondary:hover:not(:disabled) { background: var(--bg); border-color: var(--text-muted); }
.btn-secondary:disabled { opacity: 0.5; cursor: not-allowed; color: var(--text-muted); }
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
.loading-hint, .empty-hint, .locked-hint { display: flex; align-items: center; gap: 0.6rem; padding: 0.85rem 1rem; border-radius: var(--radius-sm); font-size: 0.9rem; margin-top: 0.75rem; }
.loading-hint { background: var(--primary-soft); color: var(--primary); }
.empty-hint { background: #fff8e1; color: #8a6d00; }
.locked-hint { background: #f5f5f5; color: #666; border: 1px dashed var(--border); }
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
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 720px) {
  .form-body { padding: 1.5rem 1.25rem; }
  .form-title { font-size: 1.5rem; }
  .star-btn { width: 40px; height: 40px; font-size: 1.5rem; }
  .form-navigation { flex-direction: column-reverse; }
  .btn-primary, .btn-secondary { width: 100%; justify-content: center; }
}
</style>
