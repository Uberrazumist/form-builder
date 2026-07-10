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
          <div class="col-links">Связи</div>
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
          <div class="col-links">
            <div v-if="getLinkedNames(item).length > 0" class="item-links">
              <Icon name="link" />
              <span>{{ getLinkedNames(item).join(', ') }}</span>
            </div>
            <div v-else class="item-value-empty">—</div>
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

            <div v-if="availableDictionaries.length > 0" class="links-section">
              <div class="links-header">
                <Icon name="link" />
                <h4>Связать с элементами другого справочника</h4>
              </div>
              <p class="links-description">
                Выберите справочник и отметьте элементы, с которыми связан текущий элемент.
                Это позволит автоматически фильтровать варианты в формах на основе предыдущих ответов.
              </p>

              <div class="form-group">
                <label>Связанный справочник</label>
                <select v-model="formData.linkedDictionaryId" @change="onLinkedDictionaryChange">
                  <option :value="null">— не выбран —</option>
                  <option
                    v-for="dict in availableDictionaries"
                    :key="dict.ID"
                    :value="dict.ID"
                  >
                    {{ dict.Name }}
                  </option>
                </select>
              </div>

              <div v-if="formData.linkedDictionaryId" class="form-group">
                <label>Элементы для связи</label>
                <div v-if="loadingLinkedItems" class="links-loading">
                  <div class="spinner-small"></div>
                  <span>Загрузка элементов...</span>
                </div>
                <div v-else-if="linkedItems.length === 0" class="links-empty">
                  В этом справочнике нет элементов
                </div>
                <div v-else class="multi-select-dropdown">
                  <label
                    v-for="linkedItem in linkedItems"
                    :key="linkedItem.ID"
                    class="multi-select-option"
                  >
                    <input
                      type="checkbox"
                      :value="linkedItem.ID"
                      v-model="formData.linked_ids"
                    />
                    <span>{{ linkedItem.Name }}</span>
                  </label>
                </div>
                <span v-if="formData.linked_ids.length > 0" class="hint">
                  Выбрано элементов: {{ formData.linked_ids.length }}
                </span>
              </div>
            </div>

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
const linkedItems = ref([])
const linkedItemsCache = reactive({})
const loading = ref(true)
const loadingLinkedItems = ref(false)
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
  linkedDictionaryId: null,
  linked_ids: [],
  Metadata: ''
})

const availableDictionaries = computed(() =>
  allDictionaries.value.filter(d => d.ID !== dictionary.value?.ID)
)

onMounted(async () => {
  await Promise.all([loadData(), loadAllDictionaries()])
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

const loadLinkedItems = async (dictionaryId) => {
  if (!dictionaryId) {
    linkedItems.value = []
    return
  }

  if (linkedItemsCache[dictionaryId]) {
    linkedItems.value = linkedItemsCache[dictionaryId]
    return
  }

  loadingLinkedItems.value = true
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/dictionaries/${dictionaryId}/items`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (response.ok) {
      const data = await response.json()
      const loadedItems = data.items || data || []
      linkedItemsCache[dictionaryId] = loadedItems
      linkedItems.value = loadedItems
    }
  } catch (err) {
    console.error('[DictionaryItems] Failed to load linked items:', err)
    linkedItems.value = []
  } finally {
    loadingLinkedItems.value = false
  }
}

const onLinkedDictionaryChange = () => {
  formData.linked_ids = []
  loadLinkedItems(formData.linkedDictionaryId)
}

const getLinkedNames = (item) => {
  if (!item.Metadata?.linked_ids || !Array.isArray(item.Metadata.linked_ids)) {
    return []
  }

  const names = []
  for (const linkedId of item.Metadata.linked_ids) {
    for (const dictId in linkedItemsCache) {
      const found = linkedItemsCache[dictId].find(i => i.ID === linkedId)
      if (found) {
        names.push(found.Name)
        break
      }
    }
  }

  return names.length > 0 ? names : [`Связано элементов: ${item.Metadata.linked_ids.length}`]
}

const resetForm = () => {
  formData.Name = ''
  formData.Value = ''
  formData.linkedDictionaryId = null
  formData.linked_ids = []
  formData.Metadata = ''
  linkedItems.value = []
  metadataError.value = ''
  showAdvanced.value = false
  isEditing.value = false
  editingId.value = null
}

const openCreateModal = () => {
  resetForm()
  showModal.value = true
}

const openEditModal = (item) => {
  resetForm()
  isEditing.value = true
  editingId.value = item.ID
  formData.Name = item.Name || ''
  formData.Value = item.Value || ''

  if (item.Metadata && typeof item.Metadata === 'object') {
    if (Array.isArray(item.Metadata.linked_ids) && item.Metadata.linked_ids.length > 0) {
      formData.linked_ids = [...item.Metadata.linked_ids]

      for (const dictId in linkedItemsCache) {
        const hasMatch = item.Metadata.linked_ids.some(id =>
          linkedItemsCache[dictId].some(i => i.ID === id)
        )
        if (hasMatch) {
          formData.linkedDictionaryId = dictId
          linkedItems.value = linkedItemsCache[dictId]
          break
        }
      }
    }

    const otherKeys = Object.keys(item.Metadata).filter(k => k !== 'linked_ids')
    if (otherKeys.length > 0) {
      const otherMetadata = {}
      otherKeys.forEach(k => { otherMetadata[k] = item.Metadata[k] })
      formData.Metadata = JSON.stringify(otherMetadata, null, 2)
      showAdvanced.value = true
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

    let metadata = {}

    if (formData.linked_ids.length > 0) {
      metadata.linked_ids = [...formData.linked_ids]
    }

    if (formData.Metadata?.trim()) {
      try {
        const parsed = JSON.parse(formData.Metadata)
        metadata = { ...metadata, ...parsed }
      } catch (e) {
        // уже проверено в watch
      }
    }

    const payload = {
      Name: formData.Name,
      Value: formData.Value || undefined,
      Metadata: Object.keys(metadata).length > 0 ? metadata : undefined
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
.col-links {
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

.item-value-empty {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.item-links {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.88rem;
  color: var(--text);
  flex-wrap: wrap;
}

.item-links svg {
  width: 14px;
  height: 14px;
  color: var(--primary);
  flex-shrink: 0;
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

.links-section {
  padding: 1.25rem;
  background: color-mix(in srgb, var(--primary-soft) 40%, var(--surface));
  border: 1px dashed var(--primary-soft);
  border-radius: var(--radius-sm);
  margin-bottom: 1rem;
}

.links-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.links-header svg {
  width: 18px;
  height: 18px;
  color: var(--primary);
}

.links-header h4 {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text);
  margin: 0;
}

.links-description {
  font-size: 0.85rem;
  color: var(--text-muted);
  line-height: 1.5;
  margin-bottom: 1rem;
}

.links-loading,
.links-empty {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.85rem 1rem;
  background: var(--surface);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  font-size: 0.88rem;
}

.multi-select-dropdown {
  background: var(--surface);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  max-height: 240px;
  overflow-y: auto;
  padding: 0.5rem;
}

.multi-select-option {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.5rem 0.6rem;
  cursor: pointer;
  border-radius: 6px;
  transition: background 0.15s;
  font-size: 0.92rem;
}

.multi-select-option:hover {
  background: var(--bg);
}

.multi-select-option input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--primary);
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
