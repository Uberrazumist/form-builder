<!-- src/views/ResetPasswordView.vue -->
<template>
  <div class="reset-page">
    <div class="form-card">
      <div class="form-header">
        <div class="icon-badge">
          <Icon name="lock" />
        </div>
        <h2 class="form-title">Сброс пароля</h2>
        <p class="form-subtitle">Введите код из письма и новый пароль</p>
      </div>

      <form @submit.prevent="resetPassword" class="form" novalidate>
        <div class="form-group">
          <label for="email">
            <Icon name="email" />
            Email
          </label>
          <input
            id="email"
            type="email"
            v-model="email"
            required
            placeholder="example@mail.com"
          />
        </div>

        <div class="form-group">
          <label for="code">
            <Icon name="check" />
            Код из письма
          </label>
          <input
            id="code"
            type="text"
            inputmode="numeric"
            pattern="\d{6}"
            maxlength="6"
            v-model="code"
            required
            placeholder="000000"
            class="code-input"
          />
        </div>

        <div class="form-group">
          <label for="newPassword">
            <Icon name="lock" />
            Новый пароль
          </label>
          <input
            id="newPassword"
            type="password"
            v-model="newPassword"
            required
            minlength="8"
            placeholder="Не менее 8 символов"
          />
          <span class="hint">Буквы и цифры</span>
        </div>

        <div class="form-group">
          <label for="confirmPassword">
            <Icon name="lock" />
            Подтвердите пароль
          </label>
          <input
            id="confirmPassword"
            type="password"
            v-model="confirmPassword"
            required
            placeholder="Повторите пароль"
          />
        </div>

        <button type="submit" class="btn-primary" :disabled="loading">
          <span v-if="!loading">Сбросить пароль</span>
          <span v-else class="spinner"></span>
        </button>

        <p class="form-foot">
          Вспомнили пароль? <router-link to="/login">Войти</router-link>
        </p>
      </form>

      <FormResult v-if="result" :result="result" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'
import FormResult from '../components/FormResult.vue'

const route = useRoute()
const router = useRouter()

const email = ref('')
const code = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const result = ref(null)
const loading = ref(false)

onMounted(() => {
  // Автоподстановка email из query-параметра
  const queryEmail = route.query.email
  if (queryEmail) {
    email.value = queryEmail
  }
})

const resetPassword = async () => {
  if (!email.value || !code.value || !newPassword.value || !confirmPassword.value) {
    result.value = { error: 'Заполните все поля' }
    return
  }

  if (!/^\d{6}$/.test(code.value)) {
    result.value = { error: 'Код должен состоять из 6 цифр' }
    return
  }

  if (newPassword.value.length < 8) {
    result.value = { error: 'Пароль должен содержать минимум 8 символов' }
    return
  }

  if (newPassword.value !== confirmPassword.value) {
    result.value = { error: 'Пароли не совпадают' }
    return
  }

  loading.value = true
  result.value = null

  try {
    const response = await fetch('/api/reset-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: email.value,
        code: code.value,
        new_password: newPassword.value
      })
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Не удалось сбросить пароль' }
      return
    }

    result.value = { 
      success: true, 
      message: 'Пароль успешно сброшен! Перенаправляем на вход...' 
    }
    
    setTimeout(() => {
      router.push({ 
        path: '/login', 
        query: { success: 'reset' } 
      })
    }, 1500)
  } catch (error) {
    console.error('[ResetPassword] Error:', error)
    result.value = { error: 'Ошибка сети. Попробуйте позже.' }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.reset-page {
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
  animation: fadeUp 0.5s ease both;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.form-card {
  background: var(--surface);
  padding: 2.25rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-lg);
}

.form-header {
  margin-bottom: 1.75rem;
  text-align: center;
}

.icon-badge {
  width: 64px;
  height: 64px;
  background: var(--primary-soft);
  color: var(--primary);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.25rem;
}

.icon-badge svg {
  width: 32px;
  height: 32px;
}

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

.code-input {
  font-size: 1.5rem !important;
  font-weight: 600;
  letter-spacing: 0.5em;
  text-align: center;
  font-family: 'SF Mono', Menlo, monospace !important;
}

.hint {
  font-size: 0.78rem;
  color: var(--text-muted);
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

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255,255,255,0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

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

.form-foot a:hover {
  text-decoration: underline;
}
</style>
