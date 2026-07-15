<template>
  <div class="dictionaries-page">
    <div class="page-header">
      <h1 class="page-title">Справочники</h1>
      <button @click="openCreateModal" class="btn-primary">
        <Icon name="plus" />
        Создать справочник
      </button>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Загрузка справочников...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <div class="error-icon">
        <Icon name="error" />
      </div>
      <h2>Ошибка</h2>
      <p>{{ error }}</p>
    </div>

    <div v-else-if="dictionaries.length === 0" class="empty-state">
      <div class="empty-icon">
        <Icon name="book" />
      </div>
      <h3>Нет справочников</h3>
      <p>Создайте первый справочник для использования в формах</p>
      <button @click="openCreateModal" class="btn-primary">
        <Icon name="plus" />
        Создать справочник
      </button>
    </div>

    <div v-else class="dictionaries-grid">
      <div v-for="dict in dictionaries" :key="dict.id" class="dictionary-card">
        <div class="dict-header">
          <h3 class="dict-title">{{ dict.name }}</h3>
          <div class="dict-actions-top">
            <button @click="openEditModal(dict)" class="btn-icon" title="Редактировать">
              <Icon name="edit" />
            </button>
            <button @click="deleteDictionary(dict.id)" class="btn-icon danger" title="Удалить">
              <Icon name="trash" />
            </button>
          </div>
        </div>

        <p v-if="dict.description" class="dict-description">{{ dict.description }}</p>
        <p v-else class="dict-description empty">Без описания</p>

        <div class="dict-meta">
          <span class="meta-item">
            <Icon name="calendar" />
            {{ formatDate(dict.created_at) }}
          </span>
        </div>

        <div class="dict-footer">
          <router-link :to="`/dictionaries/${dict.id}/items`" class="btn-secondary">
            <Icon name="document" />
            Управление элементами
          </router-link>
        </div>
      </div>
    </div>

    <FormResult v-if="result" :result="result" />

    <!-- Модальное окно создания/редактирования -->
    <div v-if="showModal" class="modal-overlay" @click="closeModal">
      <div class="modal-content" @click.stop>
        <h3>{{ isEditing ? 'Редактирование справочника' : 'Новый справочник' }}</h3>
        <form @submit.prevent="saveDictionary" class="modal-form">
          <div class="form-group">
            <label>Название <span class="required">*</span></label>
            <input type="text" v-model="modalData.name" required placeholder="Например: Корпуса, Классы, Учителя" />
          </div>

          <div class="form-group">
            <label>Описание</label>
            <textarea v-model="modalData.description" placeholder="Краткое описание (необязательно)" rows="3"></textarea>
          </div>

          <div class="modal-actions">
            <button type="button" @click="closeModal" class="btn-secondary">Отмена</button>
            <button type="submit" class="btn-primary" :disabled="saving">
              <span v-if="!saving">{{ isEditing ? 'Сохранить' : 'Создать' }}</span>
              <span v-else class="spinner-small"></span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

interface Dictionary {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

const dictionaries = ref<Dictionary[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const result = ref<any>(null)
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)

const modalData = reactive({
  name: '',
  description: ''
})

onMounted(async () => {
  await loadDictionaries()
})

const loadDictionaries = async () => {
  loading.value = true
  error.value = null

  try {
    const token = localStorage.getItem('token')
    const response = await fetch('/api/dictionaries', {
      headers: { 'Authorization': `Bearer ${token || ''}` }
    })

    if (!response.ok) {
      error.value = 'Не удалось загрузить справочники'
      return
    }

    const data = await response.json()
    dictionaries.value = Array.isArray(data) ? data : (data.dictionaries || [])
  } catch (err) {
    console.error('[Dictionaries] Load error:', err)
    error.value = 'Ошибка сети. Попробуйте позже.'
  } finally {
    loading.value = false
  }
}

const openCreateModal = () => {
  isEditing.value = false
  editingId.value = null
  modalData.name = ''
  modalData.description = ''
  showModal.value = true
}

const openEditModal = (dict: Dictionary) => {
  isEditing.value = true
  editingId.value = dict.id
  modalData.name = dict.name || ''
  modalData.description = dict.description || ''
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const saveDictionary = async () => {
  if (!modalData.name.trim()) {
    result.value = { error: 'Введите название справочника' }
    return
  }

  saving.value = true
  result.value = null

  try {
    const token = localStorage.getItem('token')
    const headers = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token || ''}`
    }

    const payload = {
      name: modalData.name,
      description: modalData.description
    }

    const url = isEditing.value && editingId.value
      ? `/api/dictionaries/${editingId.value}`
      : '/api/dictionaries'
    const method = isEditing.value ? 'PUT' : 'POST'

    const response = await fetch(url, {
      method,
      headers,
      body: JSON.stringify(payload)
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка сохранения' }
      return
    }

    result.value = {
      success: true,
      message: isEditing.value ? 'Справочник обновлён' : 'Справочник создан'
    }

    closeModal()
    await loadDictionaries()
  } catch (err) {
    result.value = { error: 'Ошибка сети' }
  } finally {
    saving.value = false
  }
}

const deleteDictionary = async (dictId: string) => {
  // Железный гвард
  if (!dictId) return

  if (!confirm('Удалить справочник? Все элементы будут удалены.')) return

  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/dictionaries/${dictId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token || ''}` }
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Не удалось удалить справочник' }
      return
    }

    result.value = { success: true, message: 'Справочник удалён' }
    dictionaries.value = dictionaries.value.filter(d => d.id !== dictId)
  } catch (err) {
    result.value = { error: 'Ошибка сети' }
  }
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return ''
  try {
    const date = new Date(dateStr)
    return date.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    })
  } catch {
    return dateStr
  }
}
</script>

<style scoped>
.dictionaries-page { width: 100%; max-width: 1100px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; flex-wrap: wrap; gap: 1rem; }
.page-title { font-size: 2rem; font-weight: 700; color: var(--text); letter-spacing: -0.02em; }
.loading-state, .error-state, .empty-state { text-align: center; padding: 4rem 2rem; background: var(--surface); border-radius: var(--radius); border: 1px solid var(--border); }
.spinner, .spinner-small { width: 40px; height: 40px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto 1rem; }
.spinner-small { width: 18px; height: 18px; border-width: 2px; margin: 0; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon, .empty-icon { width: 64px; height: 64px; background: var(--primary-soft); color: var(--primary); border-radius: 16px; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem; }
.error-icon svg, .empty-icon svg { width: 32px; height: 32px; }
.error-state h2, .empty-state h3 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.error-state p, .empty-state p { color: var(--text-muted); margin-bottom: 1.5rem; }
.dictionaries-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 1.5rem; }
.dictionary-card { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); padding: 1.5rem; transition: all 0.2s; animation: fadeUp 0.4s ease both; display: flex; flex-direction: column; }
.dictionary-card:hover { box-shadow: var(--shadow-md); border-color: var(--primary-soft); transform: translateY(-2px); }
.dict-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 0.75rem; margin-bottom: 0.75rem; }
.dict-title { font-size: 1.15rem; font-weight: 700; color: var(--text); line-height: 1.3; flex: 1; }
.dict-actions-top { display: flex; gap: 0.35rem; }
.btn-icon { width: 32px; height: 32px; border: 1px solid var(--border); background: var(--surface); color: var(--text-muted); cursor: pointer; border-radius: 6px; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.btn-icon:hover { background: var(--bg); color: var(--primary); border-color: var(--primary-soft); }
.btn-icon.danger:hover { background: #fdecec; color: #c53030; border-color: #f5c6c6; }
.btn-icon svg { width: 16px; height: 16px; }
.dict-description { color: var(--text-muted); font-size: 0.9rem; line-height: 1.5; margin-bottom: 1rem; min-height: 2.7rem; flex: 1; }
.dict-description.empty { font-style: italic; }
.dict-meta { display: flex; gap: 1rem; margin-bottom: 1rem; padding-top: 1rem; border-top: 1px solid var(--border); }
.meta-item { display: inline-flex; align-items: center; gap: 0.35rem; font-size: 0.85rem; color: var(--text-muted); }
.meta-item svg { width: 14px; height: 14px; }
.dict-footer { margin-top: auto; }
.btn-primary, .btn-secondary { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.6rem 1rem; font-size: 0.9rem; font-weight: 600; border-radius: var(--radius-sm); text-decoration: none; cursor: pointer; transition: all 0.2s; border: 1.5px solid transparent; font-family: inherit; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover:not(:disabled) { background: var(--primary-hover, #243f72); }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }
.btn-secondary { background: var(--surface); color: var(--text); border-color: var(--border); width: 100%; justify-content: center; }
.btn-secondary:hover { background: var(--bg); border-color: var(--text-muted); }
.btn-primary svg, .btn-secondary svg { width: 16px; height: 16px; }
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0, 0, 0, 0.5); display: flex; align-items: center; justify-content: center; z-index: 100; animation: fadeIn 0.2s ease; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
.modal-content { background: var(--surface); padding: 2rem; border-radius: var(--radius); max-width: 540px; width: 90%; box-shadow: var(--shadow-lg); }
.modal-content h3 { font-size: 1.35rem; font-weight: 700; color: var(--text); margin-bottom: 1.5rem; }
.modal-form { display: flex; flex-direction: column; gap: 1rem; }
.form-group { display: flex; flex-direction: column; gap: 0.4rem; }
.form-group label { font-size: 0.88rem; font-weight: 600; color: var(--text); }
.required { color: #c53030; }
input[type="text"], textarea { width: 100%; padding: 0.75rem 0.95rem; font-size: 0.95rem; font-family: inherit; color: var(--text); background: var(--bg); border: 1.5px solid var(--border); border-radius: var(--radius-sm); transition: border-color 0.2s; resize: vertical; }
input::placeholder, textarea::placeholder { color: #a6afbf; }
input:focus, textarea:focus { outline: none; border-color: var(--primary); background: var(--surface); box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1); }
.modal-actions { display: flex; gap: 0.75rem; justify-content: flex-end; margin-top: 0.5rem; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
@media (max-width: 720px) {
  .dictionaries-grid { grid-template-columns: 1fr; }
}
</style>
