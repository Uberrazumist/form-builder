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
            <h1 class="page-title">{{ dictionary?.name }}</h1>
            <p v-if="dictionary?.description" class="page-subtitle">{{ dictionary.description }}</p>
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
          <div class="col-code">Код</div>
          <div class="col-parent">Родитель</div>
          <div class="col-actions"></div>
        </div>
        <div v-for="item in items" :key="item.id" class="item-card">
          <div class="col-name">
            <div class="item-label">{{ item.name }}</div>
          </div>
          <div class="col-code">
            <div v-if="item.code" class="item-code">
              <code>{{ item.code }}</code>
            </div>
            <div v-else class="item-empty">—</div>
          </div>
          <div class="col-parent">
            <div v-if="item.parent_id" class="item-parent">
              <Icon name="link" />
              <span>{{ getParentName(item.parent_id) }}</span>
            </div>
            <div v-else class="item-empty">—</div>
          </div>
          <div class="col-actions">
            <button @click="openEditModal(item)" class="btn-edit-small" title="Редактировать">
              <Icon name="edit" />
            </button>
            <button @click="deleteItem(item.id)" class="btn-danger-small" title="Удалить элемент">
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
              <input type="text" v-model="formData.name" required placeholder="Например: 9А, Иванова М.П., 14:00" />
              <span class="hint">То, что увидят пользователи в форме</span>
            </div>

            <div class="form-group">
              <label>Код</label>
              <input type="text" v-model="formData.code" placeholder="Например: 9A, ivanova, 14:00" />
              <span class="hint">Необязательно. Используется для внутренней логики</span>
            </div>

            <div v-if="allDictionaries.length > 0" class="parent-section">
              <div class="parent-header">
                <Icon name="link" />
                <h4>Привязать к родительскому элементу</h4>
              </div>
              <p class="parent-description">
                Выберите элемент из любого справочника — это позволит строить каскадные формы:
                выбранный элемент будет фильтровать варианты в следующем вопросе.
              </p>

              <div class="form-group">
                <label>Родительский элемент</label>
                <select v-model="formData.parent_id">
                  <option :value="null">— без родителя —</option>
                  <optgroup v-for="dict in allDictionaries" :key="dict.id" :label="dict.name">
                    <option v-for="parentItem in getItemsForDictionary(dict.id)" :key="parentItem.id" :value="parentItem.id">
                      {{ parentItem.name }}
                    </option>
                  </optgroup>
                </select>
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
                  <textarea v-model="formData.metadata" placeholder='Например: {"duration": 45}' rows="3"></textarea>
                  <span class="hint">Обычно оставляйте пустым. Используется для продвинутых сценариев.</span>
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

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

interface Dictionary {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

interface DictionaryItem {
  id: string
  dictionary_id: string
  parent_id: string | null
  name: string
  code: string
  metadata: any
  created_at: string
  updated_at: string
}

const route = useRoute()

const dictionary = ref<Dictionary | null>(null)
const items = ref<DictionaryItem[]>([])
const allItems = ref<DictionaryItem[]>([]) // Все элементы из всех справочников (плоский список для getParentName)
const allDictionaries = ref<Dictionary[]>([]) // Все справочники для группировки
const dictionaryItemsMap = ref<Record<string, DictionaryItem[]>>({}) // dictId -> items[]
const loading = ref(true)
const error = ref<string | null>(null)
const result = ref<any>(null)
const showModal = ref(false)
const showAdvanced = ref(false)
const isEditing = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)
const metadataError = ref('')

const formData = reactive({
  name: '',
  code: '',
  parent_id: null as string | null,
  metadata: ''
})

onMounted(async () => {
  await loadData()
})

watch(() => formData.metadata, (value) => {
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
    const dictId = String(route.params.id)
    const token = localStorage.getItem('token')
    const headers: Record<string, string> = { 'Authorization': `Bearer ${token || ''}` }

    // Загрузка справочника
    const dictResponse = await fetch(`/api/dictionaries/${dictId}`, { headers })
    if (!dictResponse.ok) {
      error.value = dictResponse.status === 404 ? 'Справочник не найден' : 'Ошибка загрузки'
      return
    }
    dictionary.value = await dictResponse.json()

    // Загрузка элементов текущего справочника
    const itemsResponse = await fetch(`/api/dictionaries/${dictId}/items`, { headers })
    if (itemsResponse.ok) {
      const data = await itemsResponse.json()
      items.value = Array.isArray(data) ? data : []
    }

    // Загрузка ВСЕХ справочников и их элементов для кросс-привязки
    await loadAllDictionariesAndItems(headers)

  } catch (err) {
    console.error('[DictionaryItems] Load error:', err)
    error.value = 'Ошибка сети'
  } finally {
    loading.value = false
  }
}

const loadAllDictionariesAndItems = async (headers: Record<string, string>) => {
  try {
    const dictsResp = await fetch('/api/dictionaries', { headers })
    if (!dictsResp.ok) return

    const dicts = await dictsResp.json()
    allDictionaries.value = Array.isArray(dicts) ? dicts : (dicts.dictionaries || [])

    // Загружаем элементы каждого справочника и группируем
    dictionaryItemsMap.value = {}
    allItems.value = []
    for (const dict of allDictionaries.value) {
      dictionaryItemsMap.value[dict.id] = []
      try {
        const resp = await fetch(`/api/dictionaries/${dict.id}/items`, { headers })
        if (resp.ok) {
          const data = await resp.json()
          const dictItems = Array.isArray(data) ? data : (data.items || [])
          dictionaryItemsMap.value[dict.id] = dictItems
          allItems.value.push(...dictItems)
        }
      } catch (e) {
        // пропускаем ошибки загрузки отдельных справочников
      }
    }
  } catch (e) {
    console.error('[DictionaryItems] Failed to load all dictionaries:', e)
  }
}

const getItemsForDictionary = (dictId: string): DictionaryItem[] => {
  return dictionaryItemsMap.value[dictId] || []
}

const getParentName = (parentId: string): string => {
  const parent = allItems.value.find(i => i.id === parentId)
  return parent?.name || 'Родитель не найден'
}

const resetForm = () => {
  formData.name = ''
  formData.code = ''
  formData.parent_id = null
  formData.metadata = ''
  metadataError.value = ''
  showAdvanced.value = false
  isEditing.value = false
  editingId.value = null
}

const openCreateModal = () => {
  resetForm()
  showModal.value = true
}

const openEditModal = (item: DictionaryItem) => {
  resetForm()
  isEditing.value = true
  editingId.value = item.id
  formData.name = item.name || ''
  formData.code = item.code || ''
  formData.parent_id = item.parent_id ?? null

  if (item.metadata && typeof item.metadata === 'object') {
    formData.metadata = JSON.stringify(item.metadata, null, 2)
    showAdvanced.value = true
  }

  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  resetForm()
}

const saveItem = async () => {
  if (!formData.name.trim()) {
    result.value = { error: 'Введите название элемента' }
    return
  }

  if (formData.metadata && metadataError.value) {
    result.value = { error: 'Исправьте ошибку в дополнительных свойствах' }
    return
  }

  saving.value = true
  result.value = null

  try {
    const dictId = String(route.params.id)
    const token = localStorage.getItem('token')

    let metadata: any = null
    if (formData.metadata?.trim()) {
      try {
        metadata = JSON.parse(formData.metadata)
      } catch (e) {
        // уже проверено в watch
      }
    }

    const payload = {
      name: formData.name,
      code: formData.code || undefined,
      parent_id: formData.parent_id || null,
      metadata: metadata || undefined
    }

    const url = isEditing.value && editingId.value
      ? `/api/dictionaries/${dictId}/items/${editingId.value}`
      : `/api/dictionaries/${dictId}/items`
    const method = isEditing.value ? 'PUT' : 'POST'

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token || ''}`
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

const deleteItem = async (itemId: string) => {
  if (!itemId) return
  if (!confirm('Удалить элемент?')) return

  try {
    const dictId = String(route.params.id)
    const token = localStorage.getItem('token')

    const response = await fetch(`/api/dictionaries/${dictId}/items/${itemId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token || ''}` }
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
.items-page { width: 100%; max-width: 1000px; margin: 0 auto; animation: fadeUp 0.5s ease both; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
.loading-state, .error-state { text-align: center; padding: 4rem 2rem; }
.spinner, .spinner-small { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
.spinner-small { width: 18px; height: 18px; border-width: 2px; margin: 0; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon { width: 64px; height: 64px; background: #fdecec; color: #c53030; border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.error-icon svg { width: 32px; height: 32px; }
.error-state h2 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.error-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.btn-secondary { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.5rem; background: var(--surface); color: var(--text); border: 1.5px solid var(--border); border-radius: var(--radius-sm); font-size: 0.95rem; font-weight: 600; text-decoration: none; cursor: pointer; transition: all 0.2s; }
.btn-secondary:hover { background: var(--bg); border-color: #cfd6e3; }
.items-container { animation: fadeUp 0.5s ease both; }
.page-header { margin-bottom: 2rem; }
.header-top { display: flex; justify-content: space-between; align-items: flex-start; gap: 1rem; flex-wrap: wrap; }
.back-link { display: inline-flex; align-items: center; gap: 0.35rem; color: var(--text-muted); text-decoration: none; font-size: 0.88rem; font-weight: 500; margin-bottom: 0.75rem; transition: color 0.2s; }
.back-link:hover { color: var(--primary); }
.back-link svg { width: 14px; height: 14px; }
.page-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; margin-bottom: 0.5rem; }
.page-subtitle { color: var(--text-muted); font-size: 1rem; }
.btn-primary { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.5rem; background: var(--primary); color: #fff; border: none; border-radius: var(--radius-sm); font-size: 0.95rem; font-weight: 600; cursor: pointer; transition: all 0.2s; box-shadow: var(--shadow-sm); }
.btn-primary:hover:not(:disabled) { background: var(--primary-hover, #243f72); transform: translateY(-1px); }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }
.btn-primary svg { width: 16px; height: 16px; }
.empty-state { text-align: center; padding: 3rem 2rem; background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); }
.empty-icon { width: 64px; height: 64px; background: var(--primary-soft); color: var(--primary); border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.empty-icon svg { width: 32px; height: 32px; }
.empty-state h3 { font-size: 1.25rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.empty-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.items-list { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden; }
.list-header { display: grid; grid-template-columns: 2fr 1fr 2fr 90px; gap: 1rem; padding: 1rem 1.5rem; background: var(--bg); border-bottom: 1px solid var(--border); font-size: 0.85rem; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; }
.item-card { display: grid; grid-template-columns: 2fr 1fr 2fr 90px; gap: 1rem; padding: 1.25rem 1.5rem; border-bottom: 1px solid var(--border); transition: all 0.2s; animation: fadeUp 0.3s ease both; }
.item-card:last-child { border-bottom: none; }
.item-card:hover { background: var(--bg); }
.col-name, .col-code, .col-parent { display: flex; align-items: center; }
.col-actions { display: flex; align-items: center; justify-content: center; gap: 0.35rem; }
.item-label { font-size: 1rem; font-weight: 600; color: var(--text); }
.item-code code { background: var(--bg); padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.85rem; color: var(--primary); font-family: 'SF Mono', Menlo, monospace; }
.item-empty { color: var(--text-muted); font-size: 0.9rem; }
.item-parent { display: flex; align-items: center; gap: 0.4rem; font-size: 0.88rem; color: var(--text); flex-wrap: wrap; }
.item-parent svg { width: 14px; height: 14px; color: var(--primary); flex-shrink: 0; }
.btn-edit-small, .btn-danger-small { width: 36px; height: 36px; border: 1.5px solid var(--border); background: var(--surface); cursor: pointer; border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; transition: all 0.2s; flex-shrink: 0; }
.btn-edit-small { color: var(--primary); }
.btn-edit-small:hover { background: var(--primary-soft); border-color: var(--primary); }
.btn-danger-small { color: #c53030; }
.btn-danger-small:hover { background: #fdecec; border-color: #c53030; }
.btn-edit-small svg, .btn-danger-small svg { width: 16px; height: 16px; }
.parent-section { padding: 1.25rem; background: color-mix(in srgb, var(--primary-soft) 40%, var(--surface)); border: 1px dashed var(--primary-soft); border-radius: var(--radius-sm); margin-bottom: 1rem; }
.parent-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.parent-header svg { width: 18px; height: 18px; color: var(--primary); }
.parent-header h4 { font-size: 0.95rem; font-weight: 700; color: var(--text); margin: 0; }
.parent-description { font-size: 0.85rem; color: var(--text-muted); line-height: 1.5; margin-bottom: 1rem; }
.advanced-section { border-top: 1px solid var(--border); padding-top: 0.75rem; margin-top: 0.25rem; }
.advanced-toggle { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.5rem 0.75rem; background: transparent; color: var(--text-muted); border: 1px dashed var(--border); border-radius: var(--radius-sm); font-size: 0.85rem; font-weight: 500; cursor: pointer; transition: all 0.2s; font-family: inherit; }
.advanced-toggle:hover { color: var(--primary); border-color: var(--primary-soft); background: var(--primary-soft); }
.advanced-toggle svg { width: 14px; height: 14px; }
.advanced-content { margin-top: 1rem; padding: 1rem; background: var(--bg); border-radius: var(--radius-sm); border: 1px solid var(--border); }
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0, 0, 0, 0.5); display: flex; align-items: center; justify-content: center; z-index: 100; animation: fadeIn 0.2s ease; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
.modal-content { background: var(--surface); padding: 2rem; border-radius: var(--radius); max-width: 540px; width: 90%; max-height: 90vh; overflow-y: auto; box-shadow: var(--shadow-lg); }
.modal-content h3 { font-size: 1.35rem; font-weight: 700; color: var(--text); margin-bottom: 1.5rem; }
.modal-form { display: flex; flex-direction: column; gap: 1rem; }
.form-group { display: flex; flex-direction: column; gap: 0.4rem; }
.form-group label { font-size: 0.88rem; font-weight: 600; color: var(--text); }
.required { color: #c53030; }
input[type="text"], textarea, select { width: 100%; padding: 0.75rem 0.95rem; font-size: 0.95rem; font-family: inherit; color: var(--text); background: var(--bg); border: 1.5px solid var(--border); border-radius: var(--radius-sm); transition: border-color 0.2s; resize: vertical; }
input::placeholder, textarea::placeholder { color: #a6afbf; }
input:focus, textarea:focus, select:focus { outline: none; border-color: var(--primary); background: var(--surface); box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1); }
.hint { font-size: 0.8rem; color: var(--text-muted); font-weight: 400; }
.error-hint { font-size: 0.8rem; color: #c53030; font-weight: 500; }
.modal-actions { display: flex; gap: 0.75rem; justify-content: flex-end; margin-top: 0.5rem; }
@media (max-width: 720px) {
  .list-header, .item-card { grid-template-columns: 1fr; gap: 0.5rem; }
  .col-actions { justify-content: flex-start; }
}
</style>
