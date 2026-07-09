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
            >
              <input
                type="radio"
                :name="'q_' + question.ID"
                :value="option.Value !== undefined ? option.Value : option"
                v-model="answers[question.ID]"
                :required="question.Required"
              />
              <span>{{ option.Value !== undefined ? option.Label || option.Value : option }}</span>
            </label>
          </div>

          <!-- Checkbox -->
          <div v-else-if="question.Type === 'checkbox'" class="options-group">
            <label
              v-for="(option, optIdx) in question.Options"
              :key="optIdx"
              class="option-label"
            >
              <input
                type="checkbox"
                :value="option.Value !== undefined ? option.Value : option"
                v-model="answers[question.ID]"
              />
              <span>{{ option.Value !== undefined ? option.Label || option.Value : option }}</span>
            </label>
          </div>

          <!-- Select -->
          <select
            v-else-if="question.Type === 'select'"
            v-model="answers[question.ID]"
            :required="question.Required"
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

          <!-- Dictionary -->
          <div v-else-if="question.Type === 'dictionary'" class="dictionary-options">
            <div v-if="!dictionaryItems[question.ID]" class="loading-hint">
              <div class="spinner-small"></div>
              <span>Загрузка вариантов...</span>
            </div>
            <div v-else-if="dictionaryItems[question.ID].length === 0" class="empty-hint">
              <Icon name="alert" />
              <span>Нет доступных вариантов</span>
            </div>
            <div v-else class="options-group">
              <label
                v-for="item in dictionaryItems[question.ID]"
                :key="item.ID"
                class="option-label"
              >
                <input
                  type="radio"
                  :name="'q_' + question.ID"
                  :value="item.Value || item.ID"
                  v-model="answers[question.ID]"
                  :required="question.Required"
                />
                <span>{{ item.Label || item.Value }}</span>
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
const dictionaryItems = reactive({})
const loading = ref(true)
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
    }
  } catch (err) {
    console.error('[Preview] Failed to load dictionary items:', err)
    dictionaryItems[question.ID] = []
  }
}

watch(
  () => JSON.stringify(answers),
  async () => {
    if (!form.value) return
    for (const q of form.value.Questions) {
      if (q.DependsOn && q.Type === 'dictionary') {
        await loadDictionaryItems(q)
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
