k<!-- src/views/DictionaryItemsView.vue -->
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
          <button @click="openCreateModal" class="btn-primary">
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
        <button @click="openCreateModal" class="btn-primary">
          <Icon name="plus" />
          Добавить элемент
        </button>
      </div>

      <div v-else class="items-list">
        <div class="list-header">
          <div class="col-name">Название</div>
          <div class="col-value">Значение</div>
          <div class="col-parent">Связан с</div>
          <div class="col-actions"></div>
        </div>
        <div
          v-for="item in items"
          :key="item.ID"
          class="item-card"
        >
          <div class="col-name">
            <div class="item-label">{{ item.Name }}</div>
          </div>
          <div class="col-value">
            <div v-if="item.Value" class="item-value">
              <code>{{ item.Value }}</code>
            </div>
            <div v-else class="item-value-empty">—</div>
          </div>
          <div class="col-parent">
            <div v-if="item.ParentID" class="item-parent">
              <Icon name="link" />
              <span>
                <strong>{{ getParentDictionaryName(item.ParentID) }}</strong>
                <span v-if="getParentItemName(item.ParentID)"> → {{ getParentItemName(item.ParentID) }}</span>
              </span>
            </div>
            <div v-else class="item-parent-empty">—</div>
          </div>
          <div class="col-actions">
            <button @click="openEditModal(item)" class="btn-edit-small" title="Редактировать">
              <Icon name="edit" />
            </button>
            <button @click="deleteItem(item.ID)" class="btn-danger-small" title="Удалить элемент">
              <Icon name="trash" />
            </button>
          </div>
        </div>
      </div>

      <FormResult v-if="result" :result="result" />

      <!-- Модальное окно создания/редактирования -->
      <div v-if="showModal" class="modal-overlay" @click="closeModal">
        <div class="modal-content" @click.stop>
          <h3>{{ isEditing ? 'Редактирование элемента' : 'Новый элемент' }}</h3>
          <form @submit.prevent="saveItem" class="modal-form">
            <div class="form-group">
              <label>Название <span class="required">*</span></label>
              <input
                type="text"
                v-model="formData.Name"
                required
                placeholder="Например: 9А, Иванова М.П., 14:00"
              />
              <span class="hint">То, что увидят пользователи в форме</span>
            </div>

            <div class="form-group">
              <label>Краткое обозначение</label>
              <input
                type="text"
                v-model="formData.Value"
                placeholder="Например: 9A, ivanova, 14:00"
              />
              <span class="hint">Необязательно. Используется для связей между справочниками</span>
            </div>

            <!-- Универсальная привязка к другому справочнику -->
            <div class="form-group">
              <label>Привязать к элементу</label>
              <select v-model="formData.parentDictionaryId" @change="onParentDictionaryChange">
                <option :value="null">Нет (корневой элемент)</option>
                <option
                  v-for="dict in availableDictionaries"
                  :key="dict.ID"
                  :value="dict.ID"
                >
                  {{ dict.Name }}
                </option>
              </select>
              <span class="hint">
                Если этот элемент связан с элементом другого справочника, выберите его здесь
              </span>
            </div>

            <div v-if="formData.parentDictionaryId" class="form-group">
              <label>Элемент в справочнике «{{ getParentDictionaryName(formData.parentDictionaryId) }}»</label>
              <select v-model="formData.parentItemId" :disabled="loadingParentItems">
                <option :value="null">— выберите элемент —</option>
                <option
                  v-for="parentItem in parentItems"
                  :key="parentItem.ID"
                  :value="parentItem.ID"
                >
                  {{ parentItem.Name }}
                </option>
              </select>
              <span v-if="loadingParentItems" class="hint">Загрузка элементов...</span>
              <span v-else-if="parentItems.length === 0" class="hint">В этом справочнике нет элементов</span>
            </div>

            <!-- Дополнительные настройки (скрыты по умолчанию) -->
            <div class="advanced-section">
              <button type="button" class="advanced-toggle" @click="showAdvanced = !showAdvanced">
                <Icon :name="showAdvanced ? 'close' : 'plus'" />
                {{ showAdvanced ? 'Скрыть' : 'Показать' }} дополнительные настройки
              </button>
              <div v-if="showAdvanced" class="advanced-content">
                <div class="form-group">
                  <label>Дополнительные свойства (JSON)</label>
                  <textarea
                    v-model="formData.Metadata"
                    placeholder='Например: {"duration": 45}'
                    rows="3"
                  ></textarea>
                  <span class="hint">
                    Обычно оставляйте пустым. Используется для продвинутых сценариев.
                  </span>
                  <span v-if="metadataError" class="error-hint">{{ metadataError }}</span>
                </div>
              </div>
            </div>

            <div class="modal-actions">
              <button type="button" @click="closeModal" class="btn-secondary">Отмена</button>
              <button type="submit" class="btn-primary" :disabled="saving">
                <span v-if="!saving">{{ isEditing ? 'Сохранить' : 'Добавить' }}</span>
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
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const route = useRoute()

const dictionary = ref(null)
const items = ref([])
const allDictionaries = ref([])
const parentItems = ref([])
const parentDictionaryNames = reactive({})
const parentItemNames = reactive({})
const loading = ref(true)
const loadingParentItems = ref(false)
const error = ref(null)
const result = ref(null)
const showModal = ref(false)
const showAdvanced = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const saving = ref(false)
const metadataError = ref('')

const formData = reactive({
  Name: '',
  Value: '',
  parentDictionaryId: null,
  parentItemId: null,
  Metadata: ''
})

const availableDictionaries = computed(() => 
  allDictionaries.value.filter(d => d.ID !== dictionary.value?.ID)
)

onMounted(async () => {
  await loadData()
  await loadAllDictionaries()
})

watch(() => formData.Metadata, (value) => {
  if (!value?.trim()) {
    metadataError.value = ''
    return
  }
  try {
    JSON.parse(value)
    metadataError.value = ''
  } catch (e) {
    metadataError.value = 'Неверный формат JSON. Пример: {"ключ": "значение"}'
  }
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
      
      // Загружаем информацию о родительских элементах
      for (const item of items.value) {
        if (item.ParentID) {
          await loadParentInfo(item.ParentID)
        }
      }
    }
  } catch (err) {
    console.error('[DictionaryItems] Load error:', err)
    error.value = 'Ошибка сети'
  } finally {
    loading.value = false
  }
}

const loadAllDictionaries = async () => {
  try {
    const token = localStorage.getItem('token')
    const response = await fetch('/api/dictionaries', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (response.ok) {
      const data = await response.json()
      allDictionaries.value = data.dictionaries || data || []
    }
  } catch (err) {
    console.error('[DictionaryItems] Failed to load dictionaries:', err)
  }
}

const loadParentItems = async (dictionaryId) => {
  if (!dictionaryId) {
    parentItems.value = []
    return
  }
  
  loadingParentItems.value = true
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/dictionaries/${dictionaryId}/items`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (response.ok) {
      const data = await response.json()
      parentItems.value = data.items || data || []
    }
  } catch (err) {
    console.error('[DictionaryItems] Failed to load parent items:', err)
    parentItems.value = []
  } finally {
    loadingParentItems.value = false
  }
}

const loadParentInfo = async (parentId) => {
  try {
    const token = localStorage.getItem('token')
    const headers = { 'Authorization': `Bearer ${token}` }
    
    // Ищем элемент во всех справочниках
    for (const dict of allDictionaries.value) {
      const response = await fetch(`/api/dictionaries/${dict.ID}/items`, { headers })
      if (response.ok) {
        const data = await response.json()
        const items = data.items || data || []
        const parentItem = items.find(i => i.ID === parentId)
        if (parentItem) {
          parentDictionaryNames[parentId] = dict.Name
          parentItemNames[parentId] = parentItem.Name
          return
        }
      }
    }
  } catch (err) {
    console.error('[DictionaryItems] Failed to load parent info:', err)
  }
}

const onParentDictionaryChange = () => {
  formData.parentItemId = null
  loadParentItems(formData.parentDictionaryId)
}

const resetForm = () => {
  formData.Name = ''
  formData.Value = ''
  formData.parentDictionaryId = null
  formData.parentItemId = null
  formData.Metadata = ''
  parentItems.value = []
  metadataError.value = ''
  showAdvanced.value = false
  isEditing.value = false
  editingId.value = null
}

const openCreateModal = () => {
  resetForm()
  showModal.value = true
}

const openEditModal = async (item) => {
  resetForm()
  isEditing.value = true
  editingId.value = item.ID
  formData.Name = item.Name || ''
  formData.Value = item.Value || ''
  
  // Восстанавливаем метаданные
  if (item.Metadata && typeof item.Metadata === 'object') {
    formData.Metadata = JSON.stringify(item.Metadata, null, 2)
    showAdvanced.value = Object.keys(item.Metadata).length > 0
  }
  
  // Восстанавливаем родительский элемент
  if (item.ParentID) {
    // Находим справочник родителя
    for (const dict of allDictionaries.value) {
      const token = localStorage.getItem('token')
      const response = await fetch(`/api/dictionaries/${dict.ID}/items`, {
        headers: { 'Authorization': `Bearer ${token}` }
      })
      if (response.ok) {
        const data = await response.json()
        const items = data.items || data || []
        const parentItem = items.find(i => i.ID === item.ParentID)
        if (parentItem) {
          formData.parentDictionaryId = dict.ID
          await loadParentItems(dict.ID)
          formData.parentItemId = item.ParentID
          break
        }
      }
    }
  }
  
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  resetForm()
}

const saveItem = async () => {
  if (!formData.Name.trim()) {
    result.value = { error: 'Введите название элемента' }
    return
  }

  if (formData.Metadata && metadataError.value) {
    result.value = { error: 'Исправьте ошибку в дополнительных свойствах' }
    return
  }

  saving.value = true
  result.value = null

  try {
    const dictId = route.params.id
    const token = localStorage.getItem('token')
    
    let metadata = null
    if (formData.Metadata?.trim()) {
      try {
        metadata = JSON.parse(formData.Metadata)
      } catch (e) {
        // уже проверено в watch
      }
    }
    
    const payload = {
      Name: formData.Name,
      Value: formData.Value || undefined,
      ParentID: formData.parentItemId || undefined,
      Metadata: metadata || undefined
    }

    const url = isEditing.value 
      ? `/api/dictionaries/${dictId}/items/${editingId.value}`
      : `/api/dictionaries/${dictId}/items`
    const method = isEditing.value ? 'PUT' : 'POST'

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(payload)
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || (isEditing.value ? 'Не удалось сохранить' : 'Не удалось добавить') }
      return
    }

    result.value = { 
      success: true, 
      message: isEditing.value ? 'Элемент обновлён' : 'Элемент добавлен' 
    }
    closeModal()
    await loadData()
  } catch (err) {
    console.error('[DictionaryItems] Save error:', err)
    result.value = { error: 'Ошибка сети' }
  } finally {
    saving.value = false
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

const getParentDictionaryName = (parentId) => {
  return parentDictionaryNames[parentId] || 'Загрузка...'
}

const getParentItemName = (parentId) => {
  return parentItemNames[parentId] || ''
}
</script>

<style scoped>
.items-page {
  width: 100%;
  max-width: 1000px;
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
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
}

.list-header {
  display: grid;
  grid-template-columns: 2fr 1.5fr 2fr 90px;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: var(--bg);
  border-bottom: 1px solid var(--border);
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.item-card {
  display: grid;
  grid-template-columns: 2fr 1.5fr 2fr 90px;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border);
  transition: all 0.2s;
  animation: fadeUp 0.3s ease both;
}

.item-card:last-child {
  border-bottom: none;
}

.item-card:hover {
  background: var(--bg);
}

.col-name,
.col-value,
.col-parent {
  display: flex;
  align-items: center;
}

.col-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
}

.item-label {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
}

.item-value code {
  background: var(--bg);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.85rem;
  color: var(--primary);
  font-family: 'SF Mono', Menlo, monospace;
}

.item-value-empty,
.item-parent-empty {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.item-parent {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.9rem;
  color: var(--text);
}

.item-parent svg {
  width: 14px;
  height: 14px;
  color: var(--primary);
  flex-shrink: 0;
}

.item-parent strong {
  font-weight: 600;
  color: var(--primary);
}

.btn-edit-small,
.btn-danger-small {
  width: 36px;
  height: 36px;
  border: 1.5px solid var(--border);
  background: var(--surface);
  cursor: pointer;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}

.btn-edit-small {
  color: var(--primary);
}

.btn-edit-small:hover {
  background: var(--primary-soft);
  border-color: var(--primary);
}

.btn-danger-small {
  color: #c53030;
}

.btn-danger-small:hover {
  background: #fdecec;
  border-color: #c53030;
}

.btn-edit-small svg,
.btn-danger-small svg {
  width: 16px;
  height: 16px;
}

.advanced-section {
  border-top: 1px solid var(--border);
  padding-top: 0.75rem;
  margin-top: 0.25rem;
}

.advanced-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.75rem;
  background: transparent;
  color: var(--text-muted);
  border: 1px dashed var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.advanced-toggle:hover {
  color: var(--primary);
  border-color: var(--primary-soft);
  background: var(--primary-soft);
}

.advanced-toggle svg {
  width: 14px;
  height: 14px;
}

.advanced-content {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--bg);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}

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
  max-width: 540px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
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
  line-height: 1.4;
}

.error-hint {
  font-size: 0.78rem;
  color: #c53030;
  font-weight: 500;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 0.5rem;
}

@media (max-width: 720px) {
  .header-top {
    flex-direction: column;
  }
  .btn-primary {
    width: 100%;
    justify-content: center;
  }
  .list-header {
    display: none;
  }
  .item-card {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }
  .col-actions {
    justify-content: flex-end;
  }
}
</style>
