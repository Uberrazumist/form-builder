<!-- src/views/LoginView.vue -->
<template>
  <div class="form-card">
    <div class="form-header">
      <h2 class="form-title">Вход</h2>
      <p class="form-subtitle">Введите свои данные, чтобы войти</p>
    </div>

    <form @submit.prevent="login" class="form" novalidate>
      <FormField
        id="email"
        type="email"
        icon="email"
        label="Email"
        placeholder="example@mail.com"
        required
        v-model="email"
      />
      <FormField
        id="password"
        type="password"
        icon="lock"
        label="Пароль"
        placeholder="Введите пароль"
        required
        v-model="password"
      />

      <button type="submit" class="btn-primary" :disabled="loading">
        <span v-if="!loading">Войти</span>
        <span v-else class="spinner"></span>
      </button>

      <p class="form-foot">
        Нет аккаунта? <router-link to="/register">Зарегистрироваться</router-link>
      </p>
    </form>

    <FormResult v-if="result" :result="result" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import FormField from '../components/FormField.vue'
import FormResult from '../components/FormResult.vue'

const email = ref('')
const password = ref('')
const result = ref(null)
const loading = ref(false)

const login = async () => {
  loading.value = true
  result.value = null

  try {
    const response = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: email.value, password: password.value })
    })

    if (!response.ok) {
      if (import.meta.env.DEV && response.status === 404) {
        result.value = {
          warning: 'Демо-режим',
          message: 'Бэкенд недоступен (404)'
        }
        return
      }
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Ошибка входа' }
      return
    }

    const data = await response.json()
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
    result.value = { success: true, message: 'Вход выполнен успешно' }
    
    // Перезагрузка страницы для обновления навигации
    setTimeout(() => window.location.reload(), 1000)
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
  max-width: 450px;
  width: 100%;
  margin: 0 auto;
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
}
.form-foot a {
  color: var(--primary);
  text-decoration: none;
  font-weight: 600;
}
.form-foot a:hover { text-decoration: underline; }
</style>
