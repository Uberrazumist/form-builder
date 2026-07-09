<!-- src/views/DictionariesView.vue -->
<template>
  <div class="dictionaries-page">
    <div class="page-header">
      <div class="header-top">
        <div>
          <h1 class="page-title">Справочники</h1>
          <p class="page-subtitle">Управление справочниками для форм</p>
        </div>
        <button @click="showCreateModal = true" class="btn-primary">
          <Icon name="plus" />
          Создать справочник
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Загрузка справочников...</p>
    </div>

    <div v-else-if="dictionaries.length === 0" class="empty-state">
      <div class="empty-icon">
        <Icon name="book" />
      </div>
      <h3>Пока нет справочников</h3>
      <p>Создайте первый справочник, чтобы использовать его в формах</p>
      <button @click="showCreateModal = true" class="btn-primary">
        <Icon name="plus" />
        Создать справочник
      </button>
    </div>

    <div v-else class="dict-grid">
      <div
        v-for="dict in dictionaries"
        :key="dict.ID"
        class="dict-card"
      >
        <div class="dict-card-header">
          <div class="dict-icon">
            <Icon name="book-open" />
          </div>
          <div class="dict-info">
            <h3>{{ dict.Name }}</h3>
            <p v-if="dict.Description" class="dict-desc">{{ dict.Description }}</p>
          </div>
        </div>
        <div class="dict-meta">
          <span>{{ dict.ItemsCount || 0 }} элементов</span>
          <span>•</span>
          <span>{{ formatDate(dict.UpdatedAt || dict.CreatedAt) }}</span>
        </div>
        <div class="dict-actions">
          <router-link :to="`/dictionaries/${dict.ID}/items`" class="btn-secondary-small">
            <Icon name="document" />
            Элементы
          </router-link>
          <button @click="deleteDictionary(dict.ID)" class="btn-danger-small">
            <Icon name="trash" />
          </button>
        </div>
      </div>
    </div>

    <FormResult v-if="result" :result="result" />

    <!-- Модальное окно создания -->
    <div v-if="showCreateModal" class="modal-overlay" @click="showCreateModal = false">
      <div class="modal-content" @click.stop>
        <h3>Новый справочник</h3>
        <form @submit.prevent="createDictionary" class="modal-form">
          <div class="form-group">
            <label>Название <span class="required">*</span></label>
            <input
              type="text"
              v-model="newDict.Name"
              required
              placeholder="Например: Классы, Учителя"
            />
          </div>
          <div class="form-group">
            <label>Описание</label>
            <textarea
              v-model="newDict.Description"
              placeholder="Краткое описание (необязательно)"
              rows="3"
            ></textarea>
          </div>
          <div class="modal-actions">
            <button type="button" @click="showCreateModal = false" class="btn-secondary">Отмена</button>
            <button type="submit" class="btn-primary" :disabled="creating">
              <span v-if="!creating">Создать</span>
              <span v-else class="spinner-small"></span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const dictionaries = ref([])
const loading = ref(true)
const result = ref(null)
const showCreateModal = ref(false)
const creating = ref(false)
const newDict = reactive({ Name: '', Description: '' })

onMounted(async () => {
  await loadDictionaries()
})

const loadDictionaries = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('token')
    const response = await fetch('/api/dictionaries', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    
    if (!response.ok) {
      result.value = { error: 'Не удалось загрузить справочники' }
      return
    }
    
    const data = await response.json()
    dictionaries.value = data.dictionaries || data || []
  } catch (err) {
    result.value = { error: 'Ошибка сети' }
  } finally {
    loading.value = false
  }
}

const createDictionary = async () => {
  if (!newDict.Name.trim()) {
    result.value = { error: 'Введите название справочника' }
    return
  }

  creating.value = true
  result.value = null

  try {
    const token = localStorage.getItem('token')
    const response = await fetch('/api/dictionaries', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        name: newDict.Name,
        description: newDict.Description
      })
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка создания справочника' }
      return
    }

    result.value = { success: true, message: 'Справочник создан' }
    showCreateModal.value = false
    newDict.Name = ''
    newDict.Description = ''
    await loadDictionaries()
  } catch (err) {
    result.value = { error: 'Ошибка сети' }
  } finally {
    creating.value = false
  }
}

const deleteDictionary = async (id) => {
  if (!confirm('Удалить справочник? Все элементы будут удалены.')) return

  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/dictionaries/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка удаления' }
      return
    }

    result.value = { success: true, message: 'Справочник удалён' }
    await loadDictionaries()
  } catch (err) {
    result.value = { error: 'Ошибка сети' }
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric'
  })
}
</script>

<style scoped>
.dictionaries-page {
  width: 100%;
  max-width: 1000px;
  margin: 0 auto;
  animation: fadeUp 0.5s ease both;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.page-header {
  margin-bottom: 2rem;
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.02em;
  margin-bottom: 0.5rem;
}

.page-subtitle {
  color: var(--text-muted);
  font-size: 1rem;
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
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 14px rgba(47, 79, 138, 0.25);
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-hover);
  transform: translateY(-1px);
}

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-primary svg {
  width: 16px;
  height: 16px;
}

.loading-state {
  text-align: center;
  padding: 3rem;
  color: var(--text-muted);
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  margin: 0 auto 1rem;
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 3rem 2rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
}

.empty-icon {
  width: 64px;
  height: 64px;
  background: var(--primary-soft);
  color: var(--primary);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.5rem;
}

.empty-icon svg {
  width: 32px;
  height: 32px;
}

.empty-state h3 {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.5rem;
}

.empty-state p {
  color: var(--text-muted);
  margin-bottom: 1.5rem;
}

.dict-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.dict-card {
  background: var(--surface);
  padding: 1.75rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  transition: all 0.3s;
  animation: fadeUp 0.4s ease both;
}

.dict-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.dict-card-header {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.dict-icon {
  width: 48px;
  height: 48px;
  background: var(--primary-soft);
  color: var(--primary);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.dict-icon svg {
  width: 24px;
  height: 24px;
}

.dict-info h3 {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.25rem;
}

.dict-desc {
  font-size: 0.88rem;
  color: var(--text-muted);
  line-height: 1.4;
}

.dict-meta {
  display: flex;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 1.25rem;
}

.dict-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-secondary-small {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  background: var(--primary-soft);
  color: var(--primary);
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary-small:hover {
  background: var(--primary);
  color: #fff;
}

.btn-secondary-small svg {
  width: 14px;
  height: 14px;
}

.btn-danger-small {
  width: 36px;
  height: 36px;
  border: 1.5px solid var(--border);
  background: var(--surface);
  color: #c53030;
  cursor: pointer;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-danger-small:hover {
  background: #fdecec;
  border-color: #c53030;
}

.btn-danger-small svg {
  width: 16px;
  height: 16px;
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
  max-width: 500px;
  width: 90%;
  box-shadow: var(--shadow-lg);
}

.modal-content h3 {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 1.25rem;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-group label {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text);
}

.required {
  color: #c53030;
}

input[type="text"],
textarea {
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

input:focus,
textarea:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--surface);
  box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1);
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 0.5rem;
}

.btn-secondary {
  padding: 0.65rem 1.25rem;
  background: var(--surface);
  color: var(--text);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: var(--bg);
  border-color: #cfd6e3;
}

@media (max-width: 560px) {
  .dict-grid {
    grid-template-columns: 1fr;
  }
  .header-top {
    flex-direction: column;
  }
  .btn-primary {
    width: 100%;
    justify-content: center;
  }
}
</style>
