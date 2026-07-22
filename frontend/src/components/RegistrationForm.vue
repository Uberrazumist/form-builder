<!-- src/components/RegistrationForm.vue -->
<template>
  <div class="form-card">
    <div class="form-header">
      <h2 class="form-title">Регистрация</h2>
      <p class="form-subtitle">Создайте аккаунт, чтобы получить доступ к формам</p>
    </div>

    <form @submit.prevent="register" class="form" novalidate>
      <div class="form-group">
        <label for="email">
          <Icon name="email" />
          Email
        </label>
        <input
          id="email"
          type="email"
          v-model="formData.email"
          required
          placeholder="example@1367.ru"
        />
      </div>

      <div class="form-group">
        <label for="fullName">
          <Icon name="user" />
          Полное имя
        </label>
        <input
          id="fullName"
          type="text"
          v-model="formData.fullName"
          placeholder="Иван Иванов"
        />
      </div>

      <div class="form-group">
        <label for="password">
          <Icon name="lock" />
          Пароль
        </label>
        <div class="password-wrapper">
          <input
            id="password"
            :type="showPassword ? 'text' : 'password'"
            v-model="formData.password"
            required
            minlength="8"
            placeholder="Не менее 8 символов"
          />
          <button
            type="button"
            class="toggle-password"
            @click="showPassword = !showPassword"
            :aria-label="showPassword ? 'Скрыть пароль' : 'Показать пароль'"
          >
            <Icon :name="showPassword ? 'eye-off' : 'eye'" />
          </button>
        </div>
        <span class="hint">Буквы и цифры</span>
      </div>

      <div class="form-group">
        <label for="confirmPassword">
          <Icon name="lock" />
          Повторите пароль
        </label>
        <div class="password-wrapper">
          <input
            id="confirmPassword"
            :type="showConfirmPassword ? 'text' : 'password'"
            v-model="confirmPassword"
            required
            placeholder="Повторите пароль"
          />
          <button
            type="button"
            class="toggle-password"
            @click="showConfirmPassword = !showConfirmPassword"
            :aria-label="showConfirmPassword ? 'Скрыть пароль' : 'Показать пароль'"
          >
            <Icon :name="showConfirmPassword ? 'eye-off' : 'eye'" />
          </button>
        </div>
      </div>

      <div class="domain-notice">
        <Icon name="lock" />
        <span>Только для сотрудников школы (email <strong>@1367.ru</strong>)</span>
      </div>

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

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import Icon from './Icon.vue'
import FormResult from './FormResult.vue'

const router = useRouter()

const formData = reactive({ 
  email: '', 
  password: '', 
  fullName: '' 
})
const confirmPassword = ref('')
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const result = ref<{ 
  error?: string
  warning?: string
  success?: boolean
  message?: string
  user?: Record<string, any>
  details?: string
} | null>(null)
const loading = ref(false)

const register = async () => {
  console.log('[Register] Submitting:', { email: formData.email, fullName: formData.fullName })
  
  // Валидация домена email
  if (!formData.email.endsWith('@1367.ru')) {
    result.value = { 
      error: 'Регистрация доступна только для сотрудников школы. Email должен заканчиваться на @1367.ru' 
    }
    return
  }
  
  // Валидация длины пароля
  if (formData.password.length < 8) {
    result.value = { error: 'Пароль должен содержать минимум 8 символов' }
    return
  }
  
  // Проверка совпадения паролей
  if (formData.password !== confirmPassword.value) {
    result.value = { error: 'Пароли не совпадают' }
    return
  }
  
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
    
    localStorage.setItem('lastRegisteredEmail', formData.email)
    
    result.value = { 
      success: true, 
      message: 'Регистрация успешна! Перенаправляем на подтверждение email...' 
    }
    
    setTimeout(() => {
      router.push({ 
        path: '/verify', 
        query: { email: formData.email } 
      })
    }, 1500)
  } catch (error: unknown) {
    console.error('[Register] Error:', error)
    if (import.meta.env.DEV) {
      const errorMessage = error instanceof Error ? error.message : String(error)
      result.value = {
        warning: 'Network error',
        message: 'Не удалось связаться с сервером',
        details: errorMessage
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

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text);
}

.form-group label svg {
  width: 15px;
  height: 15px;
  color: var(--text-muted);
}

input[type="email"],
input[type="text"],
input[type="password"] {
  width: 100%;
  padding: 0.75rem 0.95rem;
  font-size: 0.95rem;
  font-family: inherit;
  color: var(--text);
  background: var(--bg);
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  transition: border-color 0.2s, background 0.2s, box-shadow 0.2s;
}

input::placeholder {
  color: #a6afbf;
}

input:hover {
  border-color: #cfd6e3;
}

input:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--surface);
  box-shadow: 0 0 0 4px rgba(47, 79, 138, 0.1);
}

.password-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-wrapper input {
  padding-right: 3rem;
}

.toggle-password {
  position: absolute;
  right: 0.5rem;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 6px;
  transition: color 0.2s, background 0.2s;
}

.toggle-password:hover {
  color: var(--primary);
  background: var(--primary-soft);
}

.toggle-password svg {
  width: 18px;
  height: 18px;
}

.hint {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.domain-notice {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.75rem 1rem;
  background: var(--primary-soft);
  border-radius: var(--radius-sm);
  font-size: 0.88rem;
  color: var(--primary);
  line-height: 1.4;
}

.domain-notice svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.domain-notice strong {
  font-weight: 700;
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
