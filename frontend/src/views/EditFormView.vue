<template>
  <div class="edit-form-page">
    <div class="page-header">
      <h1 class="page-title">Редактирование формы</h1>
    </div>

    <form @submit.prevent="submitForm" class="form-builder" novalidate>
      <div class="form-card">
        <h2 class="card-title">Основная информация</h2>
        <div class="form-group">
          <label for="formTitle">Заголовок формы <span class="required">*</span></label>
          <input id="formTitle" type="text" v-model="formData.title" required placeholder="Введите заголовок" />
        </div>
        <div class="form-group">
          <label for="formDescription">Описание</label>
          <textarea id="formDescription" v-model="formData.description" placeholder="Описание (необязательно)" rows="3"></textarea>
        </div>
        <div class="checkbox-group">
          <label class="checkbox-label">
            <input type="checkbox" v-model="formData.is_public" />
            <span class="checkbox-text">Публичная форма</span>
          </label>
        </div>
      </div>

      <div class="form-card">
        <div class="card-header">
          <h2 class="card-title">Вопросы</h2>
          <button type="button" class="btn-add" @click="addQuestion">+ Добавить вопрос</button>
        </div>
        <div v-if="formData.questions.length === 0" class="empty-state">Вопросов пока нет</div>
        <div v-for="(q, idx) in formData.questions" :key="q.id" class="question-card">
          <div class="question-header">
            <span class="question-number">Вопрос {{ idx + 1 }}</span>
            <button type="button" class="btn-remove" @click="removeQuestion(idx)">✕</button>
          </div>
          <div class="form-group">
            <label>Тип вопроса</label>
            <select v-model="q.type">
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
            <input type="text" v-model="q.title" required placeholder="Введите вопрос" />
          </div>
          <div v-if="['radio', 'checkbox', 'select'].includes(q.type)" class="options-section">
            <label>Варианты ответов</label>
            <div v-for="(opt, oi) in q.options" :key="oi" class="option-item">
              <input type="text" v-model="q.options[oi]" :placeholder="`Вариант ${oi+1}`" />
              <button type="button" class="btn-remove-small" @click="removeOption(q, oi)">✕</button>
            </div>
            <button type="button" class="btn-add-small" @click="addOption(q)">+ Добавить вариант</button>
          </div>
          <div v-if="q.type === 'rating'" class="form-group">
            <label>Максимальный рейтинг</label>
            <select v-model="q.rating_max">
              <option :value="5">5 звёзд</option>
              <option :value="10">10 звёзд</option>
            </select>
          </div>
          <div v-if="idx > 0" class="form-group">
            <label>Зависит от вопроса</label>
            <select v-model="q.depends_on">
              <option :value="null">Нет зависимости</option>
              <option v-for="(prevQ, pi) in formData.questions.slice(0, idx)" :key="prevQ.id" :value="prevQ.id">
                Вопрос {{ pi+1 }}: {{ prevQ.title || '(без текста)' }}
              </option>
            </select>
          </div>
          <div class="checkbox-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="q.is_required" />
              <span class="checkbox-text">Обязательный вопрос</span>
            </label>
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button type="button" class="btn-secondary" @click="$router.back()">Отмена</button>
        <button type="submit" class="btn-primary" :disabled="loading">
          <span v-if="!loading">Сохранить изменения</span>
          <span v-else class="spinner"></span>
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const formId = route.params.id
const loading = ref(false)

const formData = reactive({
  title: '',
  description: '',
  is_public: false,
  questions: []
})

let localIdCounter = 1

const loadForm = async () => {
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/forms/${formId}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (!response.ok) throw new Error('Не удалось загрузить форму')
    const data = await response.json()
    formData.title = data.title
    formData.description = data.description || ''
    formData.is_public = data.is_public
    formData.questions = data.questions.map(q => ({
      id: q.id,
      type: q.type,
      title: q.title,
      is_required: q.is_required || false,
      options: q.options || [],
      rating_max: q.rating_max || 5,
      depends_on: q.depends_on || null,
      order_index: q.order_index || 0
    }))
    // Устанавливаем счётчик для новых вопросов (если будут добавляться)
    localIdCounter = formData.questions.length + 1
  } catch (err) {
    alert('Ошибка загрузки формы')
    router.push('/')
  }
}

const addQuestion = () => {
  formData.questions.push({
    id: `new_${localIdCounter++}`, // временный ID для новых вопросов (на бэкенде он заменится)
    type: 'text',
    title: '',
    is_required: false,
    options: [],
    rating_max: 5,
    depends_on: null,
    order_index: formData.questions.length
  })
}

const removeQuestion = (index) => {
  formData.questions.splice(index, 1)
}

const addOption = (q) => {
  q.options.push('')
}

const removeOption = (q, optIndex) => {
  q.options.splice(optIndex, 1)
}

const submitForm = async () => {
  if (!formData.title.trim()) {
    alert('Заголовок обязателен')
    return
  }
  if (formData.questions.length === 0) {
    alert('Добавьте хотя бы один вопрос')
    return
  }
  for (const q of formData.questions) {
    if (!q.title.trim()) {
      alert('Все вопросы должны иметь текст')
      return
    }
    if (['radio', 'checkbox', 'select'].includes(q.type) && q.options.length === 0) {
      alert(`У вопроса "${q.title}" должны быть варианты`)
      return
    }
  }

  loading.value = true
  try {
    // Преобразуем depends_on: если пустая строка или null, оставляем null
    const payload = {
      title: formData.title,
      description: formData.description,
      is_public: formData.is_public,
      questions: formData.questions.map(q => ({
        // Если у вопроса есть ID и он не начинается с 'new_', значит существующий – сохраняем
        ...(q.id && !String(q.id).startsWith('new_') ? { id: q.id } : {}),
        type: q.type,
        title: q.title,
        order_index: q.order_index || 0,
        is_required: q.is_required || false,
        options: q.options || [],
        depends_on: q.depends_on || null,
        depends_values: q.depends_values || []
      }))
    }

    const token = localStorage.getItem('token')
    const response = await fetch(`/api/forms/${formId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(payload)
    })
    if (!response.ok) {
      const err = await response.json()
      throw new Error(err.error || 'Ошибка сохранения')
    }
    router.push(`/form/${formId}`)
  } catch (err) {
    alert(err.message)
  } finally {
    loading.value = false
  }
}

onMounted(loadForm)
</script>

<style scoped>
/* Стили аналогичны CreateFormView – копируем основные стили */
.edit-form-page {
  max-width: 800px;
  margin: 0 auto;
}
.page-header { margin-bottom: 2rem; }
.page-title { font-size: 2rem; font-weight: 700; color: var(--text); }
.form-builder { display: flex; flex-direction: column; gap: 1.5rem; }
.form-card { background: var(--surface); padding: 2rem; border-radius: var(--radius); border: 1px solid var(--border); }
.card-title { font-size: 1.25rem; font-weight: 700; margin-bottom: 1.5rem; }
.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
.form-group { margin-bottom: 1.25rem; }
.form-group label { display: block; font-weight: 600; margin-bottom: 0.3rem; }
.required { color: #e53e3e; }
input, textarea, select { width: 100%; padding: 0.7rem; border: 1px solid var(--border); border-radius: var(--radius-sm); background: var(--bg); }
.checkbox-group { margin: 1rem 0; }
.checkbox-label { display: flex; align-items: center; gap: 0.5rem; cursor: pointer; }
.question-card { background: var(--bg); padding: 1.5rem; border-radius: var(--radius-sm); border: 1px solid var(--border); margin-bottom: 1rem; }
.question-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
.question-number { font-weight: 600; color: var(--primary); }
.btn-remove { background: none; border: none; color: #e53e3e; font-size: 1.2rem; cursor: pointer; }
.options-section { margin-bottom: 1rem; }
.option-item { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; }
.option-item input { flex: 1; }
.btn-remove-small { background: none; border: 1px solid #e53e3e; color: #e53e3e; border-radius: 4px; padding: 0 0.5rem; cursor: pointer; }
.btn-add, .btn-add-small { background: var(--primary-soft); color: var(--primary); border: none; padding: 0.5rem 1rem; border-radius: var(--radius-sm); cursor: pointer; }
.btn-add-small { padding: 0.3rem 0.8rem; }
.btn-add:hover, .btn-add-small:hover { background: var(--primary); color: #fff; }
.form-actions { display: flex; gap: 1rem; justify-content: flex-end; margin-top: 2rem; }
.btn-primary, .btn-secondary { padding: 0.8rem 2rem; border: none; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover { background: var(--primary-hover); }
.btn-secondary { background: var(--bg); color: var(--text); border: 1px solid var(--border); }
.btn-secondary:hover { background: var(--surface); }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }
.spinner { display: inline-block; width: 18px; height: 18px; border: 2px solid rgba(255,255,255,0.3); border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.empty-state { text-align: center; padding: 2rem; color: var(--text-muted); }
</style>
