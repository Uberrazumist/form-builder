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
        :key="form.id"
        class="form-card"
      >
        <div class="form-card-header">
          <h3>{{ form.title }}</h3>
          <span v-if="form.is_public" class="badge-public">Публичная</span>
        </div>
        <p v-if="form.description" class="form-description">{{ form.description }}</p>
        <div class="form-meta">
          <span>{{ form.questions?.length || 0 }} вопросов</span>
          <span>•</span>
          <span>{{ formatDate(form.created_at) }}</span>
        </div>
        <div class="form-actions">
          <button class="btn-primary-small" @click="goToForm(form.id)">Просмотр</button>
          <!-- В будущем можно добавить другие кнопки -->
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Icon from './Icon.vue'

const router = useRouter()
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
        // Демо-режим
        forms.value = [
          {
            id: 1,
            title: 'Анкета для учеников',
            description: 'Опрос о предпочтениях в питании',
            is_public: false,
            questions: [{}, {}, {}],
            created_at: new Date().toISOString()
          },
          {
            id: 2,
            title: 'Обратная связь',
            description: 'Отзывы о школьных мероприятиях',
            is_public: true,
            questions: [{}, {}],
            created_at: new Date(Date.now() - 86400000).toISOString()
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

const goToForm = (formId) => {
  router.push(`/form/${formId}`)
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>

<style scoped>
/* Стили остаются как были, добавим btn-primary-small */
.btn-primary-small {
  padding: 0.4rem 1rem;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: var(--radius-sm);
  font-weight: 600;
  cursor: pointer;
  transition: 0.2s;
}
.btn-primary-small:hover {
  background: var(--primary-hover);
}
/* остальные стили без изменений */
</style>
