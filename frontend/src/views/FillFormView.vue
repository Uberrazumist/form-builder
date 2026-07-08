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
        <h1 class="form-title">{{ form.title }}</h1>
        <p v-if="form.description" class="form-description">{{ form.description }}</p>
      </div>

      <form @submit.prevent="submitResponses" class="form-body" novalidate>
        <div
          v-for="(question, index) in form.questions"
          :key="question.id"
          class="question-block"
        >
          <label class="question-label">
            <span class="question-number">{{ index + 1 }}.</span>
            {{ question.title }}
            <span v-if="question.required" class="required">*</span>
          </label>

          <!-- Text -->
          <input
            v-if="question.type === 'text'"
            type="text"
            v-model="answers[question.id]"
            :required="question.required"
            placeholder="Введите ответ"
            class="form-input"
          />

          <!-- Textarea -->
          <textarea
            v-else-if="question.type === 'textarea'"
            v-model="answers[question.id]"
            :required="question.required"
            placeholder="Введите ответ"
            rows="4"
            class="form-textarea"
          ></textarea>

          <!-- Radio -->
          <div v-else-if="question.type === 'radio'" class="options-group">
            <label
              v-for="(option, optIdx) in question.options"
              :key="optIdx"
              class="option-label"
            >
              <input
                type="radio"
                :name="'q_' + question.id"
                :value="option"
                v-model="answers[question.id]"
                :required="question.required"
              />
              <span>{{ option }}</span>
            </label>
          </div>

          <!-- Checkbox -->
          <div v-else-if="question.type === 'checkbox'" class="options-group">
            <label
              v-for="(option, optIdx) in question.options"
              :key="optIdx"
              class="option-label"
            >
              <input
                type="checkbox"
                :value="option"
                v-model="answers[question.id]"
              />
              <span>{{ option }}</span>
            </label>
          </div>

          <!-- Select -->
          <select
            v-else-if="question.type === 'select'"
            v-model="answers[question.id]"
            :required="question.required"
            class="form-select"
          >
            <option value="" disabled>Выберите вариант</option>
            <option
              v-for="(option, optIdx) in question.options"
              :key="optIdx"
              :value="option"
            >
              {{ option }}
            </option>
          </select>

          <!-- Rating -->
          <div v-else-if="question.type === 'rating'" class="rating-group">
            <div class="stars">
              <button
                v-for="star in question.rating_max || 5"
                :key="star"
                type="button"
                @click="answers[question.id] = star"
                class="star-btn"
                :class="{ active: answers[question.id] >= star }"
              >
                ★
              </button>
            </div>
            <span v-if="answers[question.id]" class="rating-value">
              {{ answers[question.id] }} из {{ question.rating_max || 5 }}
            </span>
          </div>

          <!-- Class/Teacher/Time choice (пока как текст) -->
          <input
            v-else-if="['class_choice', 'teacher_choice', 'time_choice'].includes(question.type)"
            type="text"
            v-model="answers[question.id]"
            :required="question.required"
            :placeholder="getPlaceholder(question.type)"
            class="form-input"
          />
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
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const route = useRoute()
const router = useRouter()

const form = ref(null)
const answers = reactive({})
const loading = ref(true)
const error = ref(null)
const result = ref(null)
const submitting = ref(false)

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
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    
    const response = await fetch(`/api/forms/${formId}`, { headers })
    
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
    form.value = data
    
    // Инициализация ответов
    form.value.questions.forEach(q => {
      if (q.type === 'checkbox') {
        answers[q.id] = []
      } else {
        answers[q.id] = ''
      }
    })
  } catch (err) {
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const getPlaceholder = (type) => {
  const placeholders = {
    class_choice: 'Например: 9А',
    teacher_choice: 'Введите имя учителя',
    time_choice: 'Например: 14:30'
  }
  return placeholders[type] || 'Введите значение'
}

const submitResponses = async () => {
  // Валидация обязательных полей
  for (const q of form.value.questions) {
    if (q.required) {
      const answer = answers[q.id]
      if (!answer || (Array.isArray(answer) && answer.length === 0)) {
        result.value = { error: `Пожалуйста, ответьте на вопрос: "${q.title}"` }
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
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

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

    const data = await response.json()
    result.value = { success: true, message: 'Спасибо! Ваши ответы успешно отправлены.' }
    
    // Очистка формы
    form.value.questions.forEach(q => {
      if (q.type === 'checkbox') {
        answers[q.id] = []
      } else {
        answers[q.id] = ''
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
