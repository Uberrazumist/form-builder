<template>
  <div class="container">
    <h2>Регистрация в конструкторе форм</h2>
    <form @submit.prevent="register">
      <div class="form-group">
        <label>Email</label>
        <input type="email" v-model="email" required />
      </div>
      <div class="form-group">
        <label>Пароль</label>
        <input type="password" v-model="password" required minlength="6" />
      </div>
      <div class="form-group">
        <label>Полное имя</label>
        <input type="text" v-model="fullName" />
      </div>
      <button type="submit">Зарегистрироваться</button>
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
.container {
  max-width: 500px;
  margin: 50px auto;
  padding: 20px;
  background: white;
  border-radius: 10px;
  box-shadow: 0 0 20px rgba(0,0,0,0.1);
}
.form-group {
  margin-bottom: 15px;
}
label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}
input {
  width: 100%;
  padding: 8px;
  box-sizing: border-box;
  border: 1px solid #ddd;
  border-radius: 4px;
}
button {
  width: 100%;
  padding: 10px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
}
button:hover {
  background: #0056b3;
}
.result {
  margin-top: 20px;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 5px;
  overflow-x: auto;
}
</style>
