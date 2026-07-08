<!-- src/views/RegisterView.vue -->
<template>
  <div class="register-page">
    <RegistrationForm />
  </div>
</template>

<script setup>
import RegistrationForm from '../components/RegistrationForm.vue'
</script>

<style scoped>
.register-page {
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
}
</style>
