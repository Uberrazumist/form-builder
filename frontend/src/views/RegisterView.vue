<!-- src/components/RegistrationForm.vue -->
<template>
  <div class="form-card">
    <div class="form-header">
      <h2 class="form-title">Регистрация</h2>
      <p class="form-subtitle">Создайте аккаунт, чтобы получить доступ к формам</p>
    </div>

    <form @submit.prevent="register" class="form" novalidate>
      <FormField
        v-for="field in fields"
        :key="field.id"
        v-bind="field"
        v-model="formData[field.id]"
      />

      <button type="submit" class="btn-primary" :disabled="loading">
        <span v-if="!loading">Зарегистрироваться</span>
        <span v-else class="spinner"></span>
      </button>

      <p class="form-foot">
        Уже есть аккаунт? <router-link to="/login">Войти</router-link>
      </p>
    </form>

    <FormResult v-if="result" :result="result" />
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import FormField from './FormField.vue'
import FormResult from './FormResult.vue'

const router = useRouter()

const fields = [
  { id: 'email', type: 'email', icon: 'email', label: 'Email', placeholder: 'example@school123.ru', required: true },
  { id: 'password', type: 'password', icon: 'lock', label: 'Пароль', placeholder: 'Не менее 8 символов', hint: 'Буквы и цифры', required: true, minlength: 8 },
  { id: 'fullName', type: 'text', icon: 'user', label: 'Полное имя', placeholder: 'Иван Иванов' }
]

const formData = reactive({ email: '', password: '', fullName: '' })
const result = ref(null)
const loading = ref(false)

const register = async () => {
  loading.value = true
  result.value = null
  
  try {
    const response = await fetch('/api/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: formData.email,
        password: formData.password,
        full_name: formData.fullName
      })
    })
    
    if (!response.ok) {
      // Демо-режим только в dev при 404 (бэкенд не поднят)
      if (import.meta.env.DEV && response.status === 404) {
        result.value = {
          warning: 'Демо-режим',
          message: 'Бэкенд недоступен (404)',
          user: { 
            email: formData.email, 
            full_name: formData.fullName || 'Не указано' 
          }
        }
        return
      }
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка регистрации' }
      return
    }
    
    result.value = { success: true, message: 'Регистрация успешна! Теперь вы можете войти.' }
    setTimeout(() => router.push('/login'), 1500)
  } catch (error) {
    if (import.meta.env.DEV) {
      result.value = {
        warning: 'Network error',
        message: 'Не удалось связаться с сервером',
        details: error.message
      }
    } else {
      result.value = { error: 'Ошибка сети. Попробуйте позже.' }
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.form-card {
  background: var(--surface);
  padding: 2.25rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-lg);
  animation: fadeUp 0.5s ease both;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.form-header { margin-bottom: 1.75rem; }
.form-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.01em;
  margin-bottom: 0.35rem;
}
.form-subtitle {
  color: var(--text-muted);
  font-size: 0.92rem;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 1.1rem;
}

.btn-primary {
  margin-top: 0.5rem;
  width: 100%;
  padding: 0.85rem;
  background: var(--primary);
  color: #fff;
  font-size: 0.98rem;
  font-weight: 600;
  font-family: inherit;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background 0.2s, transform 0.15s, box-shadow 0.2s;
  box-shadow: 0 4px 14px rgba(47, 79, 138, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
}
.btn-primary:hover:not(:disabled) {
  background: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(47, 79, 138, 0.32);
}
.btn-primary:active:not(:disabled) { transform: translateY(0); }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255,255,255,0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.form-foot {
  text-align: center;
  font-size: 0.88rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
}
.form-foot a {
  color: var(--primary);
  text-decoration: none;
  font-weight: 600;
}
.form-foot a:hover { text-decoration: underline; }
</style>
