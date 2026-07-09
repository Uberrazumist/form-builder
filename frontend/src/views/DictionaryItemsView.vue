<!-- src/views/DictionaryItemsView.vue -->
<template>
  <div class="items-page">
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Загрузка...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <div class="error-icon">
        <Icon name="error" />
      </div>
      <h2>Ошибка</h2>
      <p>{{ error }}</p>
      <router-link to="/dictionaries" class="btn-secondary">
        <Icon name="arrow-left" />
        К справочникам
      </router-link>
    </div>

    <div v-else class="items-container">
      <div class="page-header">
        <div class="header-top">
          <div>
            <router-link to="/dictionaries" class="back-link">
              <Icon name="arrow-left" />
              Все справочники
            </router-link>
            <h1 class="page-title">{{ dictionary?.Name }}</h1>
            <p v-if="dictionary?.Description" class="page-subtitle">{{ dictionary.Description }}</p>
          </div>
          <button @click="showCreateModal = true" class="btn-primary">
            <Icon name="plus" />
            Добавить элемент
          </button>
        </div>
      </div>

      <div v-if="items.length === 0" class="empty-state">
        <div class="empty-icon">
          <Icon name="document" />
        </div>
        <h3>Пока нет элементов</h3>
        <p>Добавьте первый элемент в справочник</p>
        <button @click="showCreateModal = true" class="btn-primary">
          <Icon name="plus" />
          Добавить элемент
        </button>
      </div>

      <div v-else class="items-list">
        <div
          v-for="item in items"
          :key="item.ID"
          class="item-card"
        >
          <div class="item-main">
            <div class="item-label">{{ item.Label || item.Value }}</div>
            <div v-if="item.Value && item.Label !== item.Value" class="item-value">
              Значение: <code>{{ item.Value }}</code>
            </div>
            <div v-if="item.ParentID" class="item-parent">
              Входит в: {{ getParentLabel(item.ParentID) }}
            </div>
          </div>
          <button @click="deleteItem(item.ID)" class="btn-danger-small" title="Удалить элемент">
            <Icon name="trash" />
          </button>
        </div>
      </div>

      <FormResult v-if="result" :result="result" />

      <!-- Модальное окно создания элемента -->
      <div v-if="showCreateModal" class="modal-overlay" @click="showCreateModal = false">
        <div class="modal-content" @click.stop>
          <h3>Новый элемент</h3>
          <form @submit.prevent="createItem" class="modal-form">
            <div class="form-group">
              <label>Название <span class="required">*</span></label>
              <input
                type="text"
                v-model="newItem.Label"
                required
                placeholder="Например: 9А, Иванова М.П."
              />
              <span class="hint">То, что увидят пользователи в форме</span>
            </div>
            <div class="form-group">
              <label>Краткое обозначение</label>
              <input
                type="text"
                v-model="newItem.Value"
                placeholder="Например: 9A (латиницей)"
              />
              <span class="hint">Необязательно. Используется для связей между справочниками</span>
            </div>
            <div v-if="parentItems.length > 0" class="form-group">
              <label>Входит в группу</label>
              <select v-model="newItem.ParentID">
                <option :value="null">Нет (самостоятельный элемент)</option>
                <option
                  v-for="parent in parentItems"
                  :key="parent.ID"
                  :value="parent.ID"
                >
                  {{ parent.Label || parent.Value }}
                </option>
              </select>
              <span class="hint">Например, учитель может входить в группу «Математика»</span>
            </div>
            <div class="modal-actions">
              <button type="button" @click="showCreateModal = false" class="btn-secondary">Отмена</button>
              <button type="submit" class="btn-primary" :disabled="creating">
                <span v-if="!creating">Добавить</span>
                <span v-else class="spinner-small"></span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const route = useRoute()

const dictionary = ref(null)
const items = ref([])
const loading = ref(true)
const error = ref(null)
const result = ref(null)
const showCreateModal = ref(false)
const creating = ref(false)
const newItem = reactive({ Label: '', Value: '', ParentID: null })

const parentItems = computed(() => items.value.filter(i => !i.ParentID))

onMounted(async () => {
  await loadData()
})

const loadData = async () => {
  loading.value = true
  error.value = null

  try {
    const dictId = route.params.id
    const token = localStorage.getItem('token')
    const headers = { 'Authorization': `Bearer ${token}` }

    const dictResponse = await fetch(`/api/dictionaries/${dictId}`, { headers })
    if (!dictResponse.ok) {
      error.value = dictResponse.status === 404 ? 'Справочник не найден' : 'Ошибка загрузки'
      return
    }
    dictionary.value = await dictResponse.json()

    const itemsResponse = await fetch(`/api/dictionaries/${dictId}/items`, { headers })
    if (itemsResponse.ok) {
      const data = await itemsResponse.json()
      items.value = data.items || data || []
    }
  } catch (err) {
    error.value = 'Ошибка сети'
  } finally {
    loading.value = false
  }
}

const createItem = async () => {
  if (!newItem.Label.trim()) {
    result.value = { error: 'Введите название элемента' }
    return
  }

  creating.value = true
  result.value = null

  try {
    const dictId = route.params.id
    const token = localStorage.getItem('token')
    
    const payload = {
      Name: newItem.Label,
      Value: newItem.Value || undefined,
      ParentID: newItem.ParentID || undefined
    }

    const response = await fetch(`/api/dictionaries/${dictId}/items`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(payload)
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Не удалось добавить элемент' }
      return
    }

    result.value = { success: true, message: 'Элемент добавлен' }
    showCreateModal.value = false
    newItem.Label = ''
    newItem.Value = ''
    newItem.ParentID = null
    await loadData()
  } catch (err) {
    result.value = { error: 'Ошибка сети' }
  } finally {
    creating.value = false
  }
}

const deleteItem = async (id) => {
  if (!confirm('Удалить элемент?')) return

  try {
    const dictId = route.params.id
    const token = localStorage.getItem('token')
    
    const response = await fetch(`/api/dictionaries/${dictId}/items/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Не удалось удалить элемент' }
      return
    }

    result.value = { success: true, message: 'Элемент удалён' }
    await loadData()
  } catch (err) {
    result.value = { error: 'Ошибка сети' }
  }
}

const getParentLabel = (parentId) => {
  const parent = items.value.find(i => i.ID === parentId)
  return parent ? (parent.Label || parent.Value) : '—'
}
</script>

<style scoped>
.items-page {
  width: 100%;
  max-width: 900px;
  margin: 0 auto;
  animation: fadeUp 0.5s ease both;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
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
  border: 2px solid rgba(255,255,255,0.35);
  border-top-color: #fff;
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

.items-container {
  animation: fadeUp 0.5s ease both;
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

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.88rem;
  font-weight: 500;
  margin-bottom: 0.75rem;
  transition: color 0.2s;
}

.back-link:hover {
  color: var(--primary);
}

.back-link svg {
  width: 14px;
  height: 14px;
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

.items-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.item-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  transition: all 0.2s;
  animation: fadeUp 0.3s ease both;
}

.item-card:hover {
  border-color: var(--primary-soft);
  box-shadow: var(--shadow-sm);
}

.item-main {
  flex: 1;
  min-width: 0;
}

.item-label {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.25rem;
}

.item-value {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 0.25rem;
}

.item-value code {
  background: var(--bg);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-size: 0.82rem;
  color: var(--primary);
}

.item-parent {
  font-size: 0.82rem;
  color: var(--text-muted);
  font-style: italic;
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
  flex-shrink: 0;
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
textarea,
select {
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
textarea:focus,
select:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--surface);
  box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1);
}

.hint {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 0.5rem;
}

@media (max-width: 560px) {
  .header-top {
    flex-direction: column;
  }
  .btn-primary {
    width: 100%;
    justify-content: center;
  }
}
</style>
