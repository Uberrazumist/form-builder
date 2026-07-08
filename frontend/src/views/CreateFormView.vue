<!-- src/views/CreateFormView.vue (исправленный) -->
<template>
  <div class="create-form-page">
    <div class="page-header">
      <h1 class="page-title">Создание формы</h1>
      <p class="page-subtitle">Заполните данные и добавьте вопросы</p>
    </div>

    <form @submit.prevent="submitForm" class="form-builder" novalidate>
      <!-- Основная информация -->
      <div class="form-card">
        <h2 class="card-title">Основная информация</h2>

        <div class="form-group">
          <label for="formTitle">
            <Icon name="edit" />
            Заголовок формы <span class="required">*</span>
          </label>
          <input
            id="formTitle"
            type="text"
            v-model="formData.title"
            required
            placeholder="Например: Анкета для учеников"
          />
        </div>

        <div class="form-group">
          <label for="formDescription">
            <Icon name="document" />
            Описание
          </label>
          <textarea
            id="formDescription"
            v-model="formData.description"
            placeholder="Краткое описание формы (необязательно)"
            rows="3"
          ></textarea>
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

      <!-- Вопросы -->
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
          :key="question.id"
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
              <option value="class_choice">Выбор класса</option>
              <option value="teacher_choice">Выбор учителя</option>
              <option value="time_choice">Выбор времени</option>
            </select>
          </div>

          <div class="form-group">
            <label>Текст вопроса <span class="required">*</span></label>
            <input
              type="text"
              v-model="question.title"
              required
              placeholder="Введите вопрос"
            />
          </div>

          <!-- Варианты ответов для radio/checkbox/select -->
          <div
            v-if="['radio', 'checkbox', 'select'].includes(question.type)"
            class="options-section"
          >
            <label>Варианты ответов</label>
            <div class="options-list">
              <div
                v-for="(option, optIndex) in question.options"
                :key="optIndex"
                class="option-item"
              >
                <input
                  type="text"
                  v-model="question.options[optIndex]"
                  :placeholder="`Вариант ${optIndex + 1}`"
                />
                <button
                  type="button"
                  class="btn-remove-small"
                  @click="removeOption(question, optIndex)"
                >
                  <Icon name="close" />
                </button>
              </div>
            </div>
            <button
              type="button"
              class="btn-add-small"
              @click="addOption(question)"
            >
              <Icon name="plus" />
              Добавить вариант
            </button>
          </div>

          <!-- Максимальный рейтинг для rating -->
          <div v-if="question.type === 'rating'" class="form-group">
            <label>Максимальный рейтинг</label>
            <select v-model="question.rating_max">
              <option :value="5">5 звёзд</option>
              <option :value="10">10 звёзд</option>
            </select>
          </div>

          <!-- Зависимость от предыдущего вопроса -->
          <div v-if="index > 0" class="form-group">
            <label>
              <Icon name="link" />
              Зависит от вопроса
            </label>
            <select v-model="question.depends_on">
              <option :value="null">Нет зависимости</option>
              <option
                v-for="prevQ in getPreviousQuestions(index)"
                :key="prevQ.id"
                :value="prevQ.id"
              >
                Вопрос {{ prevQ.index + 1 }}: {{ prevQ.title || '(без текста)' }}
              </option>
            </select>
          </div>

          <div class="checkbox-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="question.required" />
              <span class="checkbox-text">Обязательный вопрос</span>
            </label>
          </div>
        </div>
      </div>

      <!-- Кнопки действий -->
      <div class="form-actions">
        <button type="button" class="btn-secondary" @click="$router.back()">
          Отмена
        </button>
        <button type="submit" class="btn-primary" :disabled="loading">
          <span v-if="!loading">Создать форму</span>
          <span v-else class="spinner"></span>
        </button>
      </div>

      <FormResult v-if="result" :result="result" />
    </form>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const router = useRouter()

const formData = reactive({
  title: '',
  description: '',
  is_public: false,
  questions: []
})

const result = ref(null)
const loading = ref(false)
let questionIdCounter = 1

const addQuestion = () => {
  formData.questions.push({
    id: questionIdCounter++,
    type: 'text',
    title: '',
    required: false,
    options: [],
    rating_max: 5,
    depends_on: null
  })
}

const removeQuestion = (index) => {
  formData.questions.splice(index, 1)
}

const addOption = (question) => {
  question.options.push('')
}

const removeOption = (question, optIndex) => {
  question.options.splice(optIndex, 1)
}

const getPreviousQuestions = (currentIndex) => {
  return formData.questions.slice(0, currentIndex).map((q, idx) => ({
    id: q.id,
    index: idx,
    title: q.title
  }))
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
    if (!q.title.trim()) {
      result.value = { error: 'Все вопросы должны иметь текст' }
      return
    }
    if (['radio', 'checkbox', 'select'].includes(q.type) && q.options.length === 0) {
      result.value = { error: `Вопрос "${q.title}" должен иметь хотя бы один вариант ответа` }
      return
    }
  }

  loading.value = true
  result.value = null

  // Подготавливаем payload, преобразуя depends_on в null при необходимости
  const payload = {
    title: formData.title,
    description: formData.description,
    is_public: formData.is_public,
    questions: formData.questions.map(q => ({
      ...q,
      depends_on: q.depends_on || null,
      depends_values: q.depends_values || []
    }))
  }

  try {
    const token = localStorage.getItem('token')
    const response = await fetch('/api/forms', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
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
    // Редирект на страницу просмотра созданной формы
    setTimeout(() => router.push(`/form/${data.id}`), 1000)
  } catch (error) {
    if (import.meta.env.DEV) {
      result.value = {
        warning: 'Network error',
        message: 'Не удалось связаться с сервером',
        details: error.message
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
/* (стили остаются без изменений, они уже есть) */
</style>
