<template>
  <div class="home">
    <section class="hero-inner">
      <HeroSection />
      <RegistrationForm />
    </section>
  </div>
</template>

<script setup>
import HeroSection from '../components/HeroSection.vue'
import RegistrationForm from '../components/RegistrationForm.vue'
</script>

<style scoped>
.home {
  width: 100%;
  max-width: 1200px;
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 4rem;
}
.hero-inner {
  display: grid;
  grid-template-columns: 1.05fr 1fr;
  gap: 3rem;
  align-items: center;
}
@media (max-width: 960px) {
  .hero-inner {
    grid-template-columns: 1fr;
    gap: 2rem;
  }
}
</style>
