<template>
  <div class="hero-text">
    <span class="eyebrow">Конструктор форм</span>
    <h1 class="hero-title">Создавайте формы<br/>для школы легко</h1>
    <p class="hero-desc">
      Единый сервис для регистрации, опросов и сбора данных.
      Простой интерфейс, надёжная защита, удобный доступ для учеников и учителей.
    </p>
    <div class="hero-stats">
      <StatItem v-for="(stat, idx) in stats" :key="stat.label" v-bind="stat" :is-last="idx === stats.length - 1" />
    </div>
  </div>
</template>

<script setup>
import StatItem from './StatItem.vue'

const stats = [
  { num: '123', label: 'учеников' },
  { num: '24', label: 'учителя' },
  { num: '12', label: 'форм' }
]
</script>

<style scoped>
.hero-text {
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
}
.eyebrow {
  display: inline-block;
  font-size: 0.78rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--primary);
  background: var(--primary-soft);
  padding: 0.35rem 0.8rem;
  border-radius: 999px;
  width: fit-content;
}
.hero-title {
  font-size: clamp(2rem, 4vw, 2.8rem);
  font-weight: 800;
  line-height: 1.15;
  color: var(--text);
}
.hero-desc {
  font-size: 1.05rem;
  line-height: 1.6;
  color: var(--text-muted);
  max-width: 480px;
}
.hero-stats {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  margin-top: 0.5rem;
}
</style>
