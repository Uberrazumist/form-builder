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
          <div class="col-meta">Дополнительно</div>
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
          <div class="col-meta">
            <div v-if="isTeachersDictionary && getClassNamesForTeacher(item).length > 0" class="item-classes">
              <Icon name="book" />
              <span>{{ getClassNamesForTeacher(item).join(', ') }}</span>
            </div>
            <div v-else-if="item.Metadata && hasOtherMetadata(item)" class="item-metadata-hint">
              Есть доп. свойства
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

            <!-- Мультивыбор классов для учителя -->
            <div v-if="isTeachersDictionary && classesList.length > 0" class="form-group">
              <label>Классы, которые ведёт учитель</label>
              <div class="multi-select">
                <button
                  type="button"
                  class="multi-select-toggle"
                  @click="showClassesDropdown = !showClassesDropdown"
                >
                  <span v-if="formData.class_ids.length === 0" class="multi-select-placeholder">
                    Выберите классы...
                  </span>
                  <span v-else class="multi-select-value">
                    Выбрано классов: {{ formData.class_ids.length }}
                  </span>
                  <Icon :name="showClassesDropdown ? 'eye-off' : 'eye'" />
                </button>
                <div v-if="showClassesDropdown" class="multi-select-dropdown">
                  <label
                    v-for="cls in classesList"
                    :key="cls.ID"
                    class="multi-select-option"
                  >
                    <input
                      type="checkbox"
                      :value="cls.ID"
                      v-model="formData.class_ids"
                    />
                    <span>{{ cls.Name }}</span>
                  </label>
                  <div v-if="classesList.length === 0" class="multi-select-empty">
                    Нет доступных классов
                  </div>
                </div>
              </div>
              <span class="hint">
                Отметьте все классы, в которых ведёт занятия этот учитель. 
                Это позволит фильтровать учителей по классу в формах.
              </span>
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
import { ref, reactive, computed, onMounted, watch, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const route = useRoute()

const dictionary = ref(null)
const items = ref([])
const classesList = ref([])
const loading = ref(true)
const error = ref(null)
const result = ref(null)
const showModal = ref(false)
const showAdvanced = ref(false)
const showClassesDropdown = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const saving = ref(false)
const metadataError = ref('')

const formData = reactive({
  Name: '',
  Value: '',
  Metadata: '',
  class_ids: []
})

const isTeachersDictionary = computed(() => {
  if (!dictionary.value?.Name) return false
  const name = dictionary.value.Name.toLowerCase()
  return (
    name.includes('учитель') ||
    name.includes('учителя') ||
    name.includes('учителей') ||
    name.includes('teacher')
  )
})

onMounted(async () => {
  await loadData()
  if (isTeachersDictionary.value) {
    await loadClassesDictionary()
  }
})

onUnmounted(() => {
  // Закрываем dropdown при размонтировании
  showClassesDropdown.value = false
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

const loadClassesDictionary = async () => {
  try {
    const token = localStorage.getItem('token')
    const headers = { 'Authorization': `Bearer ${token}` }
    
    const dictsResponse = await fetch('/api/dictionaries', { headers })
    if (!dictsResponse.ok) return
    
    const dictsData = await dictsResponse.json()
    const dicts = dictsData.dictionaries || dictsData || []
    
    const classesDict = dicts.find(d => {
      const name = (d.Name || '').toLowerCase()
      return name.includes('класс') || name.includes('class')
    })
    
    if (!classesDict) {
      console.log('[DictionaryItems] Classes dictionary not found')
      return
    }
    
    const itemsResponse = await fetch(`/api/dictionaries/${classesDict.ID}/items`, { headers })
    if (itemsResponse.ok) {
      const data = await itemsResponse.json()
      classesList.value = data.items || data || []
    }
  } catch (err) {
    console.error('[DictionaryItems] Failed to load classes:', err)
  }
}

const getClassNamesForTeacher = (teacher) => {
  if (!teacher.Metadata?.class_ids || !Array.isArray(teacher.Metadata.class_ids)) {
    return []
  }
  return teacher.Metadata.class_ids
    .map(classId => {
      const cls = classesList.value.find(c => c.ID === classId)
      return cls ? cls.Name : null
    })
    .filter(Boolean)
}

const hasOtherMetadata = (item) => {
  if (!item.Metadata || typeof item.Metadata !== 'object') return false
  const keys = Object.keys(item.Metadata)
  return keys.length > 0
}

const resetForm = () => {
  formData.Name = ''
  formData.Value = ''
  formData.Metadata = ''
  formData.class_ids = []
  metadataError.value = ''
  showAdvanced.value = false
  showClassesDropdown.value = false
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
  
  // Восстанавливаем метаданные
  if (item.Metadata && typeof item.Metadata === 'object') {
    // Восстанавливаем class_ids для учителя
    if (Array.isArray(item.Metadata.class_ids)) {
      formData.class_ids = [...item.Metadata.class_ids]
    }
    
    // Остальные метаданные показываем в JSON
    const otherKeys = Object.keys(item.Metadata).filter(k => k !== 'class_ids')
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
    
    // Собираем метаданные
    let metadata = {}
    
    // Добавляем class_ids для учителя
    if (isTeachersDictionary.value && formData.class_ids.length > 0) {
      metadata.class_ids = [...formData.class_ids]
    }
    
    // Добавляем дополнительные свойства из JSON
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
.col-meta {
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

.item-classes {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.88rem;
  color: var(--text);
  flex-wrap: wrap;
}

.item-classes svg {
  width: 14px;
  height: 14px;
  color: var(--primary);
  flex-shrink: 0;
}

.item-metadata-hint {
  font-size: 0.85rem;
  color: var(--text-muted);
  font-style: italic;
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

/* Мультивыбор */
.multi-select {
  position: relative;
}

.multi-select-toggle {
  width: 100%;
  padding: 0.75rem 0.95rem;
  font-size: 0.95rem;
  font-family: inherit;
  color: var(--text);
  background: var(--bg);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  text-align: left;
}

.multi-select-toggle:hover {
  border-color: #cfd6e3;
}

.multi-select-placeholder {
  color: #a6afbf;
}

.multi-select-value {
  font-weight: 500;
}

.multi-select-toggle svg {
  width: 16px;
  height: 16px;
  color: var(--text-muted);
}

.multi-select-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: var(--surface);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-md);
  max-height: 240px;
  overflow-y: auto;
  z-index: 10;
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

.multi-select-empty {
  padding: 0.75rem;
  text-align: center;
  color: var(--text-muted);
  font-size: 0.88rem;
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
