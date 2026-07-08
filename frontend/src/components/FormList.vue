<!-- src/components/FormList.vue -->
<template>
  <div class="form-list">
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Загрузка форм...</p>
    </div>

    <div v-else-if="forms.length === 0" class="empty-state">
      <div class="empty-icon">
        <Icon name="document" />
      </div>
      <h3>Пока нет форм</h3>
      <p>Создайте свою первую форму, чтобы начать собирать данные</p>
      <router-link to="/create" class="btn-primary">
        <Icon name="plus" />
        Создать форму
      </router-link>
    </div>

    <div v-else class="forms-grid">
      <div
        v-for="form in forms"
        :key="form.ID"
        class="form-card"
      >
        <div class="form-card-header">
          <h3>{{ form.Title }}</h3>
          <div class="header-actions">
            <span v-if="form.IsPublic" class="badge-public">Публичная</span>
            <button @click="deleteForm(form.ID)" class="btn-delete" title="Удалить форму">
              <Icon name="trash" />
            </button>
          </div>
        </div>
        <p v-if="form.Description" class="form-description">{{ form.Description }}</p>
        <div class="form-meta">
          <span>{{ form.Questions?.length || 0 }} вопросов</span>
          <span>•</span>
          <span>Обновлено: {{ formatDate(form.UpdatedAt) }}</span>
        </div>
        <div class="form-actions">
          <router-link :to="`/form/${form.ID}`" class="btn-view">
            <Icon name="eye" />
            Просмотр
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Icon from './Icon.vue'

const forms = ref([])
const loading = ref(true)

onMounted(async () => {
  await fetchForms()
})

const fetchForms = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('token')
    const response = await fetch('/api/forms', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    
    if (!response.ok) {
      if (import.meta.env.DEV && response.status === 404) {
        forms.value = [
          {
            ID: 1,
            Title: 'Анкета для учеников',
            Description: 'Опрос о предпочтениях в питании',
            IsPublic: false,
            Questions: [{}, {}, {}],
            CreatedAt: new Date().toISOString(),
            UpdatedAt: new Date().toISOString()
          },
          {
            ID: 2,
            Title: 'Обратная связь',
            Description: 'Отзывы о школьных мероприятиях',
            IsPublic: true,
            Questions: [{}, {}],
            CreatedAt: new Date(Date.now() - 86400000).toISOString(),
            UpdatedAt: new Date().toISOString()
          }
        ]
        return
      }
      throw new Error('Failed to fetch forms')
    }
    
    const data = await response.json()
    forms.value = data.forms || []
  } catch (error) {
    console.error('Error fetching forms:', error)
    forms.value = []
  } finally {
    loading.value = false
  }
}

const deleteForm = async (formId) => {
  if (!confirm('Вы уверены, что хотите удалить эту форму? Это действие нельзя отменить.')) {
    return
  }

  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/forms/${formId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (!response.ok) {
      const errorData = await response.json()
      alert(errorData.error || 'Ошибка удаления формы')
      return
    }

    // Перезагрузить список форм
    await fetchForms()
  } catch (error) {
    alert('Ошибка сети. Попробуйте позже.')
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const day = String(date.getDate()).padStart(2, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const year = date.getFullYear()
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${day}.${month}.${year} ${hours}:${minutes}`
}
</script>

<style scoped>
.form-list {
  width: 100%;
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
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover {
  background: var(--primary-hover);
  transform: translateY(-1px);
}

.forms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.form-card {
  background: var(--surface);
  padding: 1.75rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  transition: all 0.3s;
  animation: fadeUp 0.4s ease both;
}

.form-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.form-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.form-card-header h3 {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.01em;
  flex: 1;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.badge-public {
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.25rem 0.6rem;
  background: #e8f5e8;
  color: #2f855a;
  border-radius: 6px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
}

.btn-delete {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: #c53030;
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.btn-delete:hover {
  background: #fdecec;
}

.btn-delete svg {
  width: 18px;
  height: 18px;
}

.form-description {
  color: var(--text-muted);
  font-size: 0.9rem;
  line-height: 1.5;
  margin-bottom: 1rem;
}

.form-meta {
  display: flex;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 1.25rem;
}

.form-actions {
  display: flex;
  gap: 0.75rem;
}

.btn-view {
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

.btn-view:hover {
  background: var(--primary);
  color: #fff;
}

.btn-view svg {
  width: 16px;
  height: 16px;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 560px) {
  .forms-grid {
    grid-template-columns: 1fr;
  }
}
</style>
