<!-- src/views/FormView.vue -->
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
          <h1 class="form-title">{{ form.Title }}</h1>
          <span class="badge" :class="form.IsPublic ? 'badge-public' : 'badge-private'">
            {{ form.IsPublic ? 'Публичная' : 'Закрытая' }}
          </span>
        </div>
        <p v-if="form.Description" class="form-description">{{ form.Description }}</p>
        <div class="form-meta">
          <span>{{ form.Questions?.length || 0 }} вопросов</span>
          <span>•</span>
          <span>{{ formatDate(form.UpdatedAt) }}</span>
        </div>
      </div>

      <div class="form-card">
        <h2 class="card-title">Вопросы</h2>
        
        <div v-if="!form.Questions || form.Questions.length === 0" class="empty-questions">
          <p>В форме пока нет вопросов</p>
        </div>

        <div v-else class="questions-list">
          <div
            v-for="(question, index) in form.Questions"
            :key="question.ID"
            class="question-item"
          >
            <div class="question-header">
              <span class="question-number">{{ index + 1 }}.</span>
              <span class="question-title">{{ question.Title }}</span>
              <span v-if="question.Required" class="required-badge">Обязательный</span>
            </div>
            <div class="question-meta">
              <span class="question-type">{{ getQuestionTypeName(question.Type) }}</span>
              <span v-if="question.DependsOn" class="depends-on">
                Зависит от вопроса #{{ getQuestionIndex(question.DependsOn) + 1 }}
              </span>
            </div>
            <div v-if="['radio', 'checkbox', 'select'].includes(question.Type)" class="question-options">
              <span class="options-label">Варианты:</span>
              <ul class="options-list">
                <li v-for="(option, optIdx) in question.Options" :key="optIdx">
                  {{ option }}
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button @click="copyLink" class="btn-secondary">
          <Icon name="link" />
          Скопировать ссылку
        </button>
        <router-link :to="`/responses/${form.ID}`" class="btn-secondary">
          <Icon name="document" />
          Ответы
        </router-link>
        <router-link :to="`/edit/${form.ID}`" class="btn-primary">
          <Icon name="edit" />
          Редактировать
        </router-link>
        <button @click="showDeleteModal = true" class="btn-danger">
          <Icon name="trash" />
          Удалить
        </button>
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

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const route = useRoute()
const router = useRouter()

const form = ref(null)
const loading = ref(true)
const error = ref(null)
const result = ref(null)
const showDeleteModal = ref(false)
const deleting = ref(false)

onMounted(async () => {
  await loadForm()
})

const loadForm = async () => {
  loading.value = true
  error.value = null

  try {
    const formId = route.params.id
    console.log('Form ID:', formId)
    
    const token = localStorage.getItem('token')

    const response = await fetch(`/api/forms/${formId}`, {
      headers: { 'Authorization': `Bearer ${token}` }
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

    form.value = await response.json()
  } catch (err) {
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const getQuestionTypeName = (type) => {
  const names = {
    text: 'Текст',
    textarea: 'Текст (многострочный)',
    radio: 'Один вариант',
    checkbox: 'Несколько вариантов',
    select: 'Выбор из списка',
    rating: 'Рейтинг',
    class_choice: 'Выбор класса',
    teacher_choice: 'Выбор учителя',
    time_choice: 'Выбор времени'
  }
  return names[type] || type
}

const getQuestionIndex = (questionId) => {
  return form.value.Questions.findIndex(q => q.ID === questionId)
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const day = String(date.getDate()).padStart(2, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const year = date.getFullYear()
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${day}.${month}.${year} ${hours}:${minutes}`
}

const copyLink = async () => {
  const link = `${window.location.origin}/fill/${form.value.ID}`
  try {
    await navigator.clipboard.writeText(link)
    result.value = { success: true, message: 'Ссылка скопирована в буфер обмена' }
    setTimeout(() => { result.value = null }, 3000)
  } catch (err) {
    result.value = { error: 'Не удалось скопировать ссылку' }
  }
}

const deleteForm = async () => {
  deleting.value = true
  result.value = null

  try {
    const formId = route.params.id
    const token = localStorage.getItem('token')

    const response = await fetch(`/api/forms/${formId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
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
.form-view-page {
  width: 100%;
  max-width: 800px;
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
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  background: var(--surface);
  color: var(--text);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.95rem;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
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

.header-top {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.form-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.02em;
  flex: 1;
}

.badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
}

.badge-public {
  background: #e8f5e8;
  color: #2f855a;
}

.badge-private {
  background: var(--primary-soft);
  color: var(--primary);
}

.form-description {
  font-size: 1.05rem;
  color: var(--text-muted);
  line-height: 1.6;
  margin-bottom: 0.75rem;
}

.form-meta {
  display: flex;
  gap: 0.5rem;
  font-size: 0.9rem;
  color: var(--text-muted);
}

.form-card {
  background: var(--surface);
  padding: 2rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
  margin-bottom: 2rem;
}

.card-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 1.5rem;
}

.empty-questions {
  text-align: center;
  padding: 2rem;
  color: var(--text-muted);
}

.questions-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.question-item {
  padding: 1.25rem;
  background: var(--bg);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}

.question-header {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.question-number {
  color: var(--primary);
  font-weight: 700;
  font-size: 1rem;
}

.question-title {
  flex: 1;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  line-height: 1.5;
}

.required-badge {
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.2rem 0.5rem;
  background: #fdecec;
  color: #c53030;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.question-meta {
  display: flex;
  gap: 1rem;
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 0.75rem;
}

.question-type {
  font-weight: 500;
}

.depends-on {
  color: var(--primary);
  font-weight: 500;
}

.question-options {
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--border);
}

.options-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 0.5rem;
  display: block;
}

.options-list {
  list-style: none;
  padding-left: 0;
  margin: 0;
}

.options-list li {
  font-size: 0.9rem;
  color: var(--text);
  padding: 0.25rem 0;
  padding-left: 1rem;
  position: relative;
}

.options-list li::before {
  content: '•';
  position: absolute;
  left: 0;
  color: var(--primary);
}

.form-actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.95rem;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 14px rgba(47, 79, 138, 0.25);
}

.btn-primary:hover {
  background: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(47, 79, 138, 0.32);
}

.btn-danger {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  background: #c53030;
  color: #fff;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-danger:hover:not(:disabled) {
  background: #9b2c2c;
  transform: translateY(-1px);
}

.btn-danger:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

/* Модальное окно */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: var(--surface);
  padding: 2rem;
  border-radius: var(--radius);
  max-width: 400px;
  width: 90%;
  box-shadow: var(--shadow-lg);
}

.modal-content h3 {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.75rem;
}

.modal-content p {
  color: var(--text-muted);
  margin-bottom: 1.5rem;
  line-height: 1.5;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 720px) {
  .form-card {
    padding: 1.5rem 1.25rem;
  }
  .form-title {
    font-size: 1.5rem;
  }
  .header-top {
    flex-direction: column;
  }
  .form-actions {
    flex-direction: column;
  }
  .btn-primary,
  .btn-secondary,
  .btn-danger {
    width: 100%;
    justify-content: center;
  }
}
</style>
