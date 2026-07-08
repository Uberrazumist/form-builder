<!-- src/views/VerifyEmailView.vue -->
<template>
  <div class="verify-page">
    <div class="form-card">
      <div class="form-header">
        <div class="icon-badge">
          <Icon name="check" />
        </div>
        <h2 class="form-title">Подтверждение email</h2>
        <p class="form-subtitle">Введите код, отправленный на ваш email</p>
      </div>

      <form @submit.prevent="verify" class="form" novalidate>
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
            placeholder="example@1367.ru"
          />
        </div>

        <div class="form-group">
          <label for="code">
            <Icon name="lock" />
            Код подтверждения
          </label>
          <input
            id="code"
            type="text"
            maxlength="8"
            v-model="code"
            required
            placeholder="Введите 8-символьный код"
            class="code-input"
            autocomplete="one-time-code"
          />
          <span class="hint">8-значный код из письма (буквы и цифры)</span>
        </div>

        <button type="submit" class="btn-primary" :disabled="loading">
          <span v-if="!loading">Подтвердить</span>
          <span v-else class="spinner"></span>
        </button>

        <p class="spam-hint">
          💡 Если письмо не пришло, проверьте папку «Спам».
        </p>

        <p class="form-foot">
          Не получили код?
          <button type="button" class="link-btn" @click="resendCode" :disabled="resending">
            {{ resending ? 'Отправляем...' : 'Отправить повторно' }}
          </button>
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
const result = ref(null)
const loading = ref(false)
const resending = ref(false)

onMounted(() => {
  const queryEmail = route.query.email
  if (queryEmail) {
    email.value = queryEmail
  } else {
    const lastEmail = localStorage.getItem('lastRegisteredEmail')
    if (lastEmail) {
      email.value = lastEmail
    }
  }
})

const verify = async () => {
  if (!email.value || !code.value) {
    result.value = { error: 'Заполните все поля' }
    return
  }

  if (code.value.length !== 8) {
    result.value = { error: 'Код должен состоять из 8 символов' }
    return
  }

  loading.value = true
  result.value = null

  try {
    const response = await fetch('/api/verify-email', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: email.value, code: code.value })
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Неверный код подтверждения' }
      return
    }

    localStorage.removeItem('lastRegisteredEmail')
    
    result.value = { success: true, message: 'Email успешно подтверждён!' }
    setTimeout(() => {
      router.push({ path: '/login', query: { success: 'verified' } })
    }, 1200)
  } catch (error) {
    console.error('[Verify] Error:', error)
    result.value = { error: 'Ошибка сети. Попробуйте позже.' }
  } finally {
    loading.value = false
  }
}

const resendCode = async () => {
  if (!email.value) {
    result.value = { error: 'Сначала введите email' }
    return
  }

  resending.value = true
  result.value = null

  try {
    const response = await fetch('/api/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        email: email.value, 
        resend: true 
      })
    })

    if (!response.ok) {
      const errorData = await response.json()
      result.value = { error: errorData.error || 'Не удалось отправить код' }
      return
    }

    result.value = { 
      success: true, 
      message: 'Новый код отправлен на ваш email. Проверьте также папку «Спам».' 
    }
  } catch (error) {
    result.value = { error: 'Ошибка сети. Попробуйте позже.' }
  } finally {
    resending.value = false
  }
}
</script>

<style scoped>
.verify-page {
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
input[type="text"] {
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
  font-size: 1.25rem !important;
  font-weight: 600;
  letter-spacing: 0.3em;
  text-align: center;
  font-family: 'SF Mono', Menlo, monospace !important;
  text-transform: uppercase;
}

.hint {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.spam-hint {
  text-align: center;
  font-size: 0.85rem;
  color: var(--text-muted);
  background: var(--primary-soft);
  padding: 0.65rem 1rem;
  border-radius: var(--radius-sm);
  margin-top: -0.25rem;
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

.link-btn {
  background: none;
  border: none;
  color: var(--primary);
  font-weight: 600;
  cursor: pointer;
  padding: 0;
  font-size: inherit;
  font-family: inherit;
}

.link-btn:hover:not(:disabled) {
  text-decoration: underline;
}

.link-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
