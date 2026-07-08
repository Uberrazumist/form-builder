<!-- src/views/MyFormsView.vue -->
<template>
  <div class="my-forms-page">
    <div class="page-header">
      <div class="header-top">
        <div>
          <h1 class="page-title">Мои формы</h1>
          <p class="page-subtitle">Управляйте созданными формами</p>
        </div>
        <router-link to="/create" class="btn-primary">
          <Icon name="plus" />
          Создать форму
        </router-link>
      </div>
    </div>

    <FormList />
  </div>
</template>

<script setup>
import Icon from '../components/Icon.vue'
import FormList from '../components/FormList.vue'
</script>

<style scoped>
.my-forms-page {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  animation: fadeUp 0.5s ease both;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.page-header {
  margin-bottom: 2rem;
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.02em;
  margin-bottom: 0.5rem;
}

.page-subtitle {
  color: var(--text-muted);
  font-size: 1rem;
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
  box-shadow: 0 4px 14px rgba(47, 79, 138, 0.25);
}

.btn-primary:hover {
  background: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(47, 79, 138, 0.32);
}

.btn-primary svg {
  width: 16px;
  height: 16px;
}

@media (max-width: 560px) {
  .page-title {
    font-size: 1.5rem;
  }
  .header-top {
    flex-direction: column;
  }
  .btn-primary {
    width: 100%;
    justify-content: center;
  }
}
</style>
