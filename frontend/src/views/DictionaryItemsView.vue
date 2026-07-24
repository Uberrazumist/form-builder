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
          <div class="col-parent">Родительские связи</div>
          <div class="col-actions"></div>
        </div>
        <div v-for="item in items" :key="item.id" class="item-card">
          <div class="col-name">
            <div class="item-label">{{ item.name }}</div>
          </div>
          <div class="col-code">
            <div v-if="item.code" class="item-code-badge">{{ item.code }}</div>
            <div v-else class="item-empty">—</div>
          </div>
          <div class="col-parent">
            <div v-if="getParentNames(item.id).length > 0" class="parent-tags">
              <span v-for="pName in getParentNames(item.id)" :key="pName" class="parent-tag">
                <Icon name="link" />
                {{ pName }}
              </span>
            </div>
            <div v-else class="item-empty">Нет связей</div>
          </div>
          <div class="col-actions">
            <button @click="openScheduleModal(item)" class="btn-schedule-small" title="Настроить расписание">
              <Icon name="calendar" />
            </button>
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
              <span class="hint">Необязательно. Для внутренней логики</span>
            </div>

            <div v-if="availableParentDictionaries.length > 0" class="parent-section">
              <div class="parent-header">
                <Icon name="link" />
                <h4>Привязка к другим справочникам</h4>
              </div>
              <p class="parent-description">
                Выберите элементы из других справочников, чтобы построить каскад (например, привязать "Класс" к "Корпусу").
                Можно выбрать несколько вариантов.
              </p>

              <div class="form-group">
                <label>Справочник-родитель</label>
                <select v-model="selectedParentDictionaryId" @change="onParentDictionaryChange" class="form-select">
                  <option :value="null">— выберите справочник —</option>
                  <option v-for="dict in availableParentDictionaries" :key="dict.id" :value="dict.id">
                    {{ dict.name }}
                  </option>
                </select>
              </div>

              <div v-if="selectedParentDictionaryId" class="parent-items-section">
                <div class="parent-items-header">
                  <label class="section-label">Доступные элементы для привязки:</label>
                  <button type="button" class="btn-clear-links" @click="clearAllLinks">
                    <Icon name="close" />
                    Сбросить все связи
                  </button>
                </div>
                
                <div v-if="parentDictionaryItems.length === 0" class="empty-parent-hint">
                  В этом справочнике пока нет элементов
                </div>

                <div v-else class="parent-items-list">
                  <label v-for="pItem in parentDictionaryItems" :key="pItem.id" class="parent-item-option">
                    <input
                      type="checkbox"
                      :value="pItem.id"
                      v-model="formData.linked_ids"
                      class="custom-checkbox"
                    />
                    <span class="item-name">{{ pItem.name }}</span>
                    <span v-if="pItem.code" class="item-code-badge-small">{{ pItem.code }}</span>
                  </label>
                </div>
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
                  <textarea v-model="formData.metadataRaw" placeholder='Например: {"duration": 45}' rows="3" class="form-textarea"></textarea>
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

    <!-- Модальное окно расписания -->
    <div v-if="showScheduleModal" class="modal-overlay" @click="showScheduleModal = false">
      <div class="modal-content modal-large" @click.stop>
        <h3>Расписание: {{ currentScheduleResourceName }}</h3>
        <ScheduleBuilder
          :resource-id="currentScheduleResourceId"
          :resource-name="currentScheduleResourceName"
          :initial-rule="currentScheduleInitialRule"
          @saved="onScheduleSaved"
        />
        <button @click="showScheduleModal = false" class="btn-close-modal">Закрыть</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch, computed, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'
import ScheduleBuilder from '../components/ScheduleBuilder.vue'

interface Dictionary {
  id: string
  name: string
  description: string
}

interface DictionaryItem {
  id: string
  dictionary_id: string
  parent_id: string | null
  name: string
  code: string
  metadata: any
}

const route = useRoute()

const dictionary = ref<Dictionary | null>(null)
const items = ref<DictionaryItem[]>([])
const allDictionaries = ref<Dictionary[]>([])
const dictionaryItemsMap = ref<Record<string, DictionaryItem[]>>({})
const loading = ref(true)
const error = ref<string | null>(null)
const result = ref<any>(null)
const showModal = ref(false)
const showAdvanced = ref(false)
const isEditing = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)
const metadataError = ref('')

// Модальное окно расписания
const showScheduleModal = ref(false)
const currentScheduleResourceId = ref('')
const currentScheduleResourceName = ref('')
const currentScheduleInitialRule = ref<any>(null)

const selectedParentDictionaryId = ref<string | null>(null)
const parentDictionaryItems = ref<DictionaryItem[]>([])

const formData = reactive({
  name: '',
  code: '',
  linked_ids: [] as string[],
  metadataRaw: ''
})

const handleEsc = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    if (showScheduleModal.value) {
      showScheduleModal.value = false
    } else if (showModal.value) {
      closeModal()
    }
  }
}

onMounted(async () => {
  await loadData()
  window.addEventListener('keydown', handleEsc)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleEsc)
})

watch(() => formData.metadataRaw, (value) => {
  if (!value?.trim()) {
    metadataError.value = ''
    return
  }
  try {
    JSON.parse(value)
    metadataError.value = ''
  } catch (e) {
    metadataError.value = 'Неверный формат JSON'
  }
})

const loadData = async () => {
  loading.value = true
  error.value = null

  try {
    const dictId = String(route.params.id)
    const token = localStorage.getItem('token')
    const headers: Record<string, string> = { 'Authorization': `Bearer ${token || ''}` }

    const dictResponse = await fetch(`/api/dictionaries/${dictId}`, { headers })
    if (!dictResponse.ok) {
      error.value = dictResponse.status === 404 ? 'Справочник не найден' : 'Ошибка загрузки'
      return
    }
    dictionary.value = await dictResponse.json()

    const itemsResponse = await fetch(`/api/dictionaries/${dictId}/items`, { headers })
    if (itemsResponse.ok) {
      const data = await itemsResponse.json()
      items.value = Array.isArray(data) ? data : []
    }

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

    dictionaryItemsMap.value = {}
    for (const dict of allDictionaries.value) {
      try {
        const resp = await fetch(`/api/dictionaries/${dict.id}/items`, { headers })
        if (resp.ok) {
          const data = await resp.json()
          // ГАРАНТИРОВАННО МАССИВ: защита от undefined
          dictionaryItemsMap.value[dict.id] = Array.isArray(data) ? data : (data.items || [])
        } else {
          dictionaryItemsMap.value[dict.id] = []
        }
      } catch (e) {
        dictionaryItemsMap.value[dict.id] = []
      }
    }
  } catch (e) {
    console.error('[DictionaryItems] Failed to load all dictionaries:', e)
  }
}

const availableParentDictionaries = computed(() => {
  if (!dictionary.value) return allDictionaries.value
  return allDictionaries.value.filter(d => d.id !== dictionary.value!.id)
})

const onParentDictionaryChange = () => {
  const dictId = selectedParentDictionaryId.value
  parentDictionaryItems.value = dictId ? (dictionaryItemsMap.value[dictId] || []) : []
  // Сбрасываем linked_ids при смене справочника, чтобы не было "висячих" ID из другого справочника
  formData.linked_ids.splice(0)
}

const clearAllLinks = () => {
  formData.linked_ids.splice(0, formData.linked_ids.length)
  selectedParentDictionaryId.value = null
  parentDictionaryItems.value = []
}

const getParentNames = (itemId: string): string[] => {
  const names: string[] = []
  const currentItem = items.value.find(i => i.id === itemId)
  if (!currentItem) return names

  const idsToCheck = new Set<string>()
  if (currentItem.parent_id) idsToCheck.add(currentItem.parent_id)
  if (currentItem.metadata?.linked_ids && Array.isArray(currentItem.metadata.linked_ids)) {
    currentItem.metadata.linked_ids.forEach((id: string) => idsToCheck.add(id))
  }

  for (const dictId in dictionaryItemsMap.value) {
    // БЕЗОПАСНЫЙ ДОСТУП: фоллбек на пустой массив
    const dictItems = dictionaryItemsMap.value[dictId] || []
    for (const pItem of dictItems) {
      if (idsToCheck.has(pItem.id)) {
        names.push(pItem.name)
      }
    }
  }
  return names
}

const resetForm = () => {
  formData.name = ''
  formData.code = ''
  formData.linked_ids.splice(0)
  formData.metadataRaw = ''
  metadataError.value = ''
  showAdvanced.value = false
  isEditing.value = false
  editingId.value = null
  selectedParentDictionaryId.value = null
  parentDictionaryItems.value = []
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
  
  if (item.metadata?.linked_ids && Array.isArray(item.metadata.linked_ids)) {
    formData.linked_ids.push(...item.metadata.linked_ids)
  } else if (item.parent_id) {
    formData.linked_ids.push(item.parent_id)
  }

  if (formData.linked_ids.length > 0) {
    const firstLinkId = formData.linked_ids[0]
    for (const dictId in dictionaryItemsMap.value) {
      // БЕЗОПАСНЫЙ ДОСТУП: фоллбек на пустой массив перед вызовом .some()
      const dictItems = dictionaryItemsMap.value[dictId] || []
      if (dictItems.some(i => i.id === firstLinkId)) {
        selectedParentDictionaryId.value = dictId
        parentDictionaryItems.value = dictItems
        break
      }
    }
  }

  if (item.metadata && typeof item.metadata === 'object') {
    const { linked_ids, ...rest } = item.metadata
    if (Object.keys(rest).length > 0) {
      formData.metadataRaw = JSON.stringify(rest, null, 2)
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
  if (!formData.name.trim()) {
    result.value = { error: 'Введите название элемента' }
    return
  }

  if (formData.metadataRaw && metadataError.value) {
    result.value = { error: 'Исправьте ошибку в JSON' }
    return
  }

  saving.value = true
  result.value = null

  try {
    const dictId = String(route.params.id)
    const token = localStorage.getItem('token')

    let parsedMetadata: any = {}
    if (formData.metadataRaw?.trim()) {
      parsedMetadata = JSON.parse(formData.metadataRaw)
    }

    const finalMetadata = {
      ...parsedMetadata,
      linked_ids: formData.linked_ids.length > 0 ? formData.linked_ids : undefined
    }
    
    Object.keys(finalMetadata).forEach(key => {
      if (finalMetadata[key] === undefined) delete finalMetadata[key]
    })

    const payload = {
      name: formData.name,
      code: formData.code || undefined,
      parent_id: formData.linked_ids.length > 0 ? formData.linked_ids[0] : null,
      metadata: Object.keys(finalMetadata).length > 0 ? finalMetadata : undefined
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
      result.value = { error: errorData.error || 'Ошибка сохранения' }
      return
    }

    result.value = { success: true, message: isEditing.value ? 'Элемент обновлён' : 'Элемент добавлен' }
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
  if (!confirm('Удалить элемент? Это действие необратимо.')) return

  try {
    const dictId = String(route.params.id)
    const token = localStorage.getItem('token')

    const response = await fetch(`/api/dictionaries/${dictId}/items/${itemId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token || ''}` }
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Не удалось удалить' }
      return
    }

    result.value = { success: true, message: 'Элемент удалён' }
    await loadData()
  } catch (err) {
    result.value = { error: 'Ошибка сети' }
  }
}

const openScheduleModal = async (item: DictionaryItem) => {
  currentScheduleResourceId.value = item.id
  currentScheduleResourceName.value = item.name
  currentScheduleInitialRule.value = null

  // Загружаем существующее расписание
  try {
    const token = localStorage.getItem('token') || ''
    const res = await fetch(`/api/schedules?resource_id=${item.id}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (res.ok) {
      const data = await res.json()
      if (Array.isArray(data) && data.length > 0) {
        currentScheduleInitialRule.value = data[0]
      }
    }
  } catch (e) {
    console.error('Failed to load schedule:', e)
  }

  showScheduleModal.value = true
}

const onScheduleSaved = () => {
  showScheduleModal.value = false
}
</script>

<style scoped>
.items-page { width: 100%; max-width: 1000px; margin: 0 auto; animation: fadeUp 0.5s ease both; font-family: inherit; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
.loading-state, .error-state { text-align: center; padding: 4rem 2rem; }
.spinner, .spinner-small { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
.spinner-small { width: 18px; height: 18px; border-width: 2px; margin: 0; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon { width: 64px; height: 64px; background: #fdecec; color: #c53030; border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.error-icon svg { width: 32px; height: 32px; }
.error-state h2 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.error-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.btn-secondary { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1.5rem; background: var(--surface); color: var(--text); border: 1.5px solid var(--border); border-radius: var(--radius-sm); font-size: 0.95rem; font-weight: 600; text-decoration: none; cursor: pointer; transition: all 0.2s; font-family: inherit; }
.btn-secondary:hover { background: var(--bg); border-color: #cfd6e3; }
.items-container { animation: fadeUp 0.5s ease both; }
.page-header { margin-bottom: 2rem; }
.header-top { display: flex; justify-content: space-between; align-items: flex-start; gap: 1rem; flex-wrap: wrap; }
.back-link { display: inline-flex; align-items: center; gap: 0.35rem; color: var(--text-muted); text-decoration: none; font-size: 0.88rem; font-weight: 500; margin-bottom: 0.75rem; transition: color 0.2s; font-family: inherit; }
.back-link:hover { color: var(--primary); }
.back-link svg { width: 14px; height: 14px; }
.page-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; margin-bottom: 0.5rem; font-family: inherit; }
.page-subtitle { color: var(--text-muted); font-size: 1rem; font-family: inherit; }
.btn-primary { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.6rem 1rem; background: var(--primary); color: #fff; border: none; border-radius: var(--radius-sm); font-size: 0.9rem; font-weight: 600; cursor: pointer; transition: all 0.2s; font-family: inherit; }
.btn-primary:hover:not(:disabled) { background: var(--primary-hover, #243f72); }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }
.btn-primary svg { width: 16px; height: 16px; }
.empty-state { text-align: center; padding: 3rem 2rem; background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); }
.empty-icon { width: 64px; height: 64px; background: var(--primary-soft); color: var(--primary); border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.empty-icon svg { width: 32px; height: 32px; }
.empty-state h3 { font-size: 1.25rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; font-family: inherit; }
.empty-state p { color: var(--text-muted); margin-bottom: 1.5rem; font-family: inherit; }
.items-list { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden; }
.list-header { display: grid; grid-template-columns: 2fr 1fr 2.5fr 90px; gap: 1rem; padding: 1rem 1.5rem; background: var(--bg); border-bottom: 1px solid var(--border); font-size: 0.85rem; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; font-family: inherit; }
.item-card { display: grid; grid-template-columns: 2fr 1fr 2.5fr 90px; gap: 1rem; padding: 1.25rem 1.5rem; border-bottom: 1px solid var(--border); transition: all 0.2s; animation: fadeUp 0.3s ease both; font-family: inherit; }
.item-card:last-child { border-bottom: none; }
.item-card:hover { background: var(--bg); }
.col-name, .col-code, .col-parent { display: flex; align-items: center; }
.col-actions { display: flex; align-items: center; justify-content: center; gap: 0.35rem; }
.item-label { font-size: 1rem; font-weight: 600; color: var(--text); font-family: inherit; }
.item-code-badge, .item-code-badge-small { background: var(--bg); padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem; color: var(--primary); font-family: 'SF Mono', Menlo, monospace; border: 1px solid var(--border); }
.item-code-badge-small { font-size: 0.75rem; padding: 0.1rem 0.4rem; margin-left: 0.5rem; }
.item-empty { color: var(--text-muted); font-size: 0.9rem; font-family: inherit; }
.parent-tags { display: flex; flex-wrap: wrap; gap: 0.4rem; }
.parent-tag { display: inline-flex; align-items: center; gap: 0.3rem; font-size: 0.85rem; color: var(--text); background: var(--primary-soft); color: var(--primary); padding: 0.25rem 0.6rem; border-radius: 6px; font-weight: 500; font-family: inherit; }
.parent-tag svg { width: 12px; height: 12px; }
.btn-edit-small, .btn-danger-small { width: 36px; height: 36px; border: 1.5px solid var(--border); background: var(--surface); cursor: pointer; border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; transition: all 0.2s; flex-shrink: 0; }
.btn-schedule-small { width: 36px; height: 36px; border: 1.5px solid var(--border); background: var(--surface); color: var(--primary); cursor: pointer; border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; transition: all 0.2s; flex-shrink: 0; }
.btn-schedule-small:hover { background: var(--primary-soft); border-color: var(--primary); }
.btn-schedule-small svg { width: 16px; height: 16px; }
.btn-edit-small { color: var(--primary); }
.btn-edit-small:hover { background: var(--primary-soft); border-color: var(--primary); }
.btn-danger-small { color: #c53030; }
.btn-danger-small:hover { background: #fdecec; border-color: #c53030; }
.btn-edit-small svg, .btn-danger-small svg { width: 16px; height: 16px; }
.parent-section { padding: 1.25rem; background: color-mix(in srgb, var(--primary-soft) 30%, var(--surface)); border: 1px dashed var(--primary-soft); border-radius: var(--radius-sm); margin-bottom: 1rem; }
.parent-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.parent-header svg { width: 18px; height: 18px; color: var(--primary); }
.parent-header h4 { font-size: 0.95rem; font-weight: 700; color: var(--text); margin: 0; font-family: inherit; }
.parent-description { font-size: 0.85rem; color: var(--text-muted); line-height: 1.5; margin-bottom: 1rem; font-family: inherit; }
.form-group { display: flex; flex-direction: column; gap: 0.4rem; margin-bottom: 1rem; }
.form-group label, .section-label { font-size: 0.88rem; font-weight: 600; color: var(--text); font-family: inherit; }
.form-select, input[type="text"], .form-textarea { width: 100%; padding: 0.75rem 0.95rem; font-size: 0.95rem; font-family: inherit; color: var(--text); background: var(--surface); border: 1.5px solid var(--border); border-radius: var(--radius-sm); transition: border-color 0.2s; }
.form-select:focus, input[type="text"]:focus, .form-textarea:focus { outline: none; border-color: var(--primary); box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1); }
.hint { font-size: 0.8rem; color: var(--text-muted); font-weight: 400; font-family: inherit; }
.error-hint { font-size: 0.8rem; color: #c53030; font-weight: 500; font-family: inherit; }
.parent-items-section { margin-top: 0.75rem; padding: 1rem; background: var(--surface); border-radius: var(--radius-sm); border: 1px solid var(--border); }
.parent-items-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; }
.btn-clear-links { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.35rem 0.65rem; background: transparent; color: var(--text-muted); border: 1px dashed var(--border); border-radius: var(--radius-sm); font-size: 0.8rem; font-weight: 500; cursor: pointer; transition: all 0.15s; font-family: inherit; }
.btn-clear-links:hover { color: #c53030; border-color: #c53030; background: #fdecec; }
.btn-clear-links svg { width: 12px; height: 12px; }
.parent-items-list { display: flex; flex-direction: column; gap: 0.25rem; max-height: 220px; overflow-y: auto; padding: 0.25rem; margin-top: 0.5rem; }
.parent-item-option { display: flex; align-items: center; gap: 0.6rem; cursor: pointer; padding: 0.5rem 0.75rem; border-radius: 6px; transition: background 0.15s; font-family: inherit; }
.parent-item-option:hover { background: var(--bg); }
.custom-checkbox { width: 16px; height: 16px; accent-color: var(--primary); cursor: pointer; flex-shrink: 0; }
.item-name { font-size: 0.92rem; color: var(--text); font-weight: 500; font-family: inherit; flex: 1; }
.empty-parent-hint { padding: 1rem; color: var(--text-muted); font-size: 0.88rem; text-align: center; font-family: inherit; }
.advanced-section { border-top: 1px solid var(--border); padding-top: 0.75rem; margin-top: 0.25rem; }
.advanced-toggle { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.5rem 0.75rem; background: transparent; color: var(--text-muted); border: 1px dashed var(--border); border-radius: var(--radius-sm); font-size: 0.85rem; font-weight: 500; cursor: pointer; transition: all 0.2s; font-family: inherit; width: 100%; justify-content: center; }
.advanced-toggle:hover { color: var(--primary); border-color: var(--primary-soft); background: var(--primary-soft); }
.advanced-toggle svg { width: 14px; height: 14px; }
.advanced-content { margin-top: 1rem; padding: 1rem; background: var(--bg); border-radius: var(--radius-sm); border: 1px solid var(--border); }
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0, 0, 0, 0.5); display: flex; align-items: center; justify-content: center; z-index: 100; animation: fadeIn 0.2s ease; backdrop-filter: blur(2px); }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
.modal-content { background: var(--surface); padding: 2rem; border-radius: var(--radius); max-width: 540px; width: 90%; max-height: 90vh; overflow-y: auto; box-shadow: var(--shadow-lg); font-family: inherit; }
.modal-content h3 { font-size: 1.35rem; font-weight: 700; color: var(--text); margin-bottom: 1.5rem; font-family: inherit; }
.modal-large { max-width: 800px; }
.btn-close-modal { margin-top: 1rem; width: 100%; padding: 0.75rem; background: var(--bg); border: 1px solid var(--border); border-radius: var(--radius-sm); cursor: pointer; font-family: inherit; font-size: 0.95rem; }
.btn-close-modal:hover { background: var(--surface); }
.modal-form { display: flex; flex-direction: column; gap: 0.5rem; }
.required { color: #c53030; }
.modal-actions { display: flex; gap: 0.75rem; justify-content: flex-end; margin-top: 1rem; }
@media (max-width: 720px) {
  .list-header, .item-card { grid-template-columns: 1fr; gap: 0.5rem; }
  .col-actions { justify-content: flex-start; }
}
</style>