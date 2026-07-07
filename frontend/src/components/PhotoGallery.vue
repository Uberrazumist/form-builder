<template>
  <section class="gallery">
    <div class="section-head">
      <span class="eyebrow">Наша школа</span>
      <h2 class="section-title">Современное пространство для учёбы</h2>
    </div>

    <div class="gallery-grid">
      <PhotoCard v-for="photo in photos" :key="photo.title" v-bind="photo" />
    </div>
  </section>
</template>

<script setup>
import PhotoCard from './PhotoCard.vue'

const photos = [
  {
    src: 'https://image.qwenlm.ai/public_source/cdde0bbb-c2a4-4377-8b79-9c7140718fb2/1ea7018ed-80f6-4c60-b3d1-47131c5552a6.png',
    alt: 'Фасад школы',
    tag: 'Главный корпус',
    title: 'Фасад школы',
    large: true
  },
  {
    src: 'https://image.qwenlm.ai/public_source/cdde0bbb-c2a4-4377-8b79-9c7140718fb2/1ff6dbd27-5bfc-4473-a9f4-e2b072d9ab44.png',
    alt: 'Библиотека',
    tag: 'Учёба',
    title: 'Библиотека'
  },
  {
    src: 'https://image.qwenlm.ai/public_source/cdde0bbb-c2a4-4377-8b79-9c7140718fb2/11e3e74cd-8611-4f05-a62a-28cbeff653b3.png',
    alt: 'Компьютерный класс',
    tag: 'Технологии',
    title: 'Компьютерный класс'
  },
  {
    src: 'https://image.qwenlm.ai/public_source/cdde0bbb-c2a4-4377-8b79-9c7140718fb2/11f5f57d5-ad1a-4ac3-90fb-dc1291054aa6.png',
    alt: 'Спортивный зал',
    tag: 'Активная жизнь',
    title: 'Спортивный зал'
  }
]
</script>

<style scoped>
.gallery {
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}
.section-head {
  text-align: center;
  max-width: 560px;
  margin: 0 auto;
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
  margin-bottom: 1rem;
}
.section-title {
  font-size: clamp(1.4rem, 2.5vw, 1.9rem);
  font-weight: 700;
  color: var(--text);
}
.gallery-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}
@media (max-width: 960px) {
  .gallery-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 560px) {
  .gallery-grid { grid-template-columns: 1fr; }
}
</style>
