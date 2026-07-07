<template>
  <div class="photo-card" :class="{ large }">
    <img class="photo-img" :src="src" :alt="alt" />
    <div class="photo-caption">
      <span class="photo-tag">{{ tag }}</span>
      <h3>{{ title }}</h3>
    </div>
  </div>
</template>

<script setup>
defineProps({
  src: String,
  alt: String,
  tag: String,
  title: String,
  large: Boolean
})
</script>

<style scoped>
.photo-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  transition: transform 0.3s, box-shadow 0.3s;
  cursor: pointer;
}
.photo-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}
.photo-card.large {
  grid-column: span 2;
  grid-row: span 2;
}
.photo-img {
  width: 100%;
  object-fit: cover;
  transition: transform 0.5s;
}
.photo-card.large .photo-img {
  aspect-ratio: 16 / 12;
}
.photo-card:not(.large) .photo-img {
  aspect-ratio: 16 / 10;
}
.photo-card:hover .photo-img {
  transform: scale(1.03);
}
.photo-caption {
  padding: 1rem 1.15rem 1.15rem;
}
.photo-tag {
  display: inline-block;
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--primary);
  margin-bottom: 0.3rem;
}
.photo-caption h3 {
  font-size: 1.02rem;
  font-weight: 600;
  color: var(--text);
}
.photo-card.large .photo-caption h3 {
  font-size: 1.25rem;
}
@media (max-width: 960px) {
  .photo-card.large {
    grid-column: span 2;
    grid-row: span 1;
  }
  .photo-card.large .photo-img {
    aspect-ratio: 16 / 9;
  }
}
@media (max-width: 560px) {
  .photo-card.large {
    grid-column: span 1;
  }
}
</style>
