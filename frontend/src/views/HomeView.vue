<template>
  <div class="form-card">
    <h2 class="form-title">Регистрация</h2>
    <form @submit.prevent="register" class="form">
      <div class="form-group">
        <label for="email">Email</label>
        <input id="email" type="email" v-model="email" required placeholder="example@mail.com" />
      </div>
      <div class="form-group">
        <label for="password">Пароль</label>
        <input id="password" type="password" v-model="password" required minlength="8" placeholder="Не менее 8 символов, буквы и цифры" />
      </div>
      <div class="form-group">
        <label for="fullName">Полное имя</label>
        <input id="fullName" type="text" v-model="fullName" placeholder="Иван Иванов" />
      </div>
      <button type="submit" class="btn-primary">Зарегистрироваться</button>
    </form>
    <div v-if="result" class="result">
      <pre>{{ JSON.stringify(result, null, 2) }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const email = ref('')
const password = ref('')
const fullName = ref('')
const result = ref(null)

const register = async () => {
  try {
    const response = await fetch('/api/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: email.value,
        password: password.value,
        full_name: fullName.value
      })
    })
    const data = await response.json()
    result.value = data
  } catch (error) {
    result.value = { error: error.message }
  }
}
</script>

<style scoped>
.form-card {
  max-width: 500px;
  width: 100%;
  background: white;
  padding: 2rem 2.5rem;
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.1);
}
.form-title {
  text-align: center;
  margin-bottom: 1.5rem;
  font-size: 1.8rem;
  color: #2c3e50;
}
.form-group {
  margin-bottom: 1.2rem;
}
label {
  display: block;
  font-weight: 600;
  margin-bottom: 0.4rem;
  color: #333;
}
input {
  width: 100%;
  padding: 0.6rem 0.8rem;
  font-size: 1rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  transition: border 0.2s;
}
input:focus {
  border-color: #3498db;
  outline: none;
}
.btn-primary {
  width: 100%;
  padding: 0.7rem;
  background: #3498db;
  color: white;
  font-size: 1.1rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}
.btn-primary:hover {
  background: #2980b9;
}
.result {
  margin-top: 1.5rem;
  padding: 0.8rem;
  background: #f8f9fa;
  border-radius: 6px;
  font-size: 0.9rem;
  overflow-x: auto;
  max-height: 200px;
  overflow-y: auto;
}
</style>
