<!-- src/App.vue -->
<template>
  <div id="app">
    <header class="header">
      <div class="header-content">
        <div class="brand">
          <div class="logo-badge">
            <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
              <rect x="8" y="20" width="32" height="22" rx="2" fill="none" stroke="currentColor" stroke-width="2"/>
              <path d="M8 20 L24 8 L40 20" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"/>
              <rect x="20" y="28" width="8" height="14" fill="none" stroke="currentColor" stroke-width="2"/>
              <line x1="14" y1="26" x2="14" y2="34" stroke="currentColor" stroke-width="2"/>
              <line x1="34" y1="26" x2="34" y2="34" stroke="currentColor" stroke-width="2"/>
            </svg>
          </div>
          <div class="brand-text">
            <span class="school-name">Школа № 123</span>
            <span class="school-sub">Электронный сервис</span>
          </div>
        </div>

        <nav class="nav">
          <router-link to="/" class="nav-link" active-class="active" exact>Главная</router-link>
          <template v-if="isAuthenticated">
            <router-link to="/create" class="nav-link" active-class="active">Создать форму</router-link>
            <router-link to="/my-forms" class="nav-link" active-class="active">Мои формы</router-link>
            <button @click="logout" class="nav-link logout-btn">Выйти</button>
          </template>
          <template v-else>
            <router-link to="/login" class="nav-link" active-class="active">Вход</router-link>
          </template>
        </nav>

        <button class="menu-toggle" @click="menuOpen = !menuOpen" aria-label="Меню">
          <span :class="{ open: menuOpen }"></span>
        </button>
      </div>

      <div class="mobile-nav" :class="{ open: menuOpen }">
        <router-link to="/" class="nav-link" @click="menuOpen = false">Главная</router-link>
        <template v-if="isAuthenticated">
          <router-link to="/create" class="nav-link" @click="menuOpen = false">Создать форму</router-link>
          <router-link to="/my-forms" class="nav-link" @click="menuOpen = false">Мои формы</router-link>
          <button @click="handleLogout" class="nav-link logout-btn">Выйти</button>
        </template>
        <template v-else>
          <router-link to="/login" class="nav-link" @click="menuOpen = false">Вход</router-link>
        </template>
      </div>
    </header>

    <main class="main">
      <router-view />
    </main>

    <footer class="footer">
      <div class="footer-content">
        <div class="footer-brand">
          <div class="logo-badge small">
            <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
              <rect x="8" y="20" width="32" height="22" rx="2" fill="none" stroke="currentColor" stroke-width="2"/>
              <path d="M8 20 L24 8 L40 20" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"/>
              <rect x="20" y="28" width="8" height="14" fill="none" stroke="currentColor" stroke-width="2"/>
            </svg>
          </div>
          <p class="footer-text">© 2026 Школа № 123. Все права защищены.</p>
        </div>
        <div class="footer-links">
          <a href="#">Контакты</a>
          <a href="#">Политика</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const menuOpen = ref(false)
const isAuthenticated = ref(false)

onMounted(() => {
  checkAuth()
})

const checkAuth = () => {
  isAuthenticated.value = !!localStorage.getItem('token')
}

const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  isAuthenticated.value = false
  window.location.reload()
}

const handleLogout = () => {
  menuOpen.value = false
  logout()
}
</script>

<style>
@import url('https://cdn.jsdelivr.net/fontsource/fonts/inter@latest/latin-400-normal.css');
@import url('https://cdn.jsdelivr.net/fontsource/fonts/inter@latest/latin-500-normal.css');
@import url('https://cdn.jsdelivr.net/fontsource/fonts/inter@latest/latin-600-normal.css');
@import url('https://cdn.jsdelivr.net/fontsource/fonts/inter@latest/latin-700-normal.css');

:root {
  --bg: #f6f8fb;
  --surface: #ffffff;
  --text: #1a2332;
  --text-muted: #6b7689;
  --primary: #2f4f8a;
  --primary-soft: #e8eef8;
  --primary-hover: #243f72;
  --border: #e5e9f0;
  --shadow-sm: 0 1px 2px rgba(20, 30, 50, 0.04);
  --shadow-md: 0 6px 24px rgba(20, 30, 50, 0.06);
  --shadow-lg: 0 20px 50px rgba(20, 30, 50, 0.08);
  --radius: 14px;
  --radius-sm: 10px;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  background: var(--bg);
  color: var(--text);
  min-height: 100vh;
  -webkit-font-smoothing: antialiased;
  line-height: 1.5;
}

#app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.header {
  background: var(--surface);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 50;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0.9rem 1.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.logo-badge {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--primary-soft);
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: transform 0.25s ease;
}
.logo-badge:hover { transform: translateY(-1px); }
.logo-badge svg { width: 26px; height: 26px; }
.logo-badge.small { width: 36px; height: 36px; }
.logo-badge.small svg { width: 20px; height: 20px; }

.brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}
.school-name {
  font-weight: 700;
  font-size: 1rem;
  color: var(--text);
  letter-spacing: -0.01em;
}
.school-sub {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin-top: 2px;
}

.nav {
  display: flex;
  gap: 0.25rem;
  align-items: center;
}
.nav-link {
  color: var(--text-muted);
  text-decoration: none;
  padding: 0.55rem 1rem;
  border-radius: 8px;
  font-size: 0.92rem;
  font-weight: 500;
  transition: all 0.2s ease;
  background: none;
  border: none;
  cursor: pointer;
  font-family: inherit;
}
.nav-link:hover {
  color: var(--text);
  background: var(--primary-soft);
}
.nav-link.active {
  color: var(--primary);
  background: var(--primary-soft);
}
.logout-btn {
  color: #c53030;
}
.logout-btn:hover {
  background: #fdecec;
  color: #c53030;
}

.menu-toggle {
  display: none;
  width: 40px;
  height: 40px;
  border: none;
  background: transparent;
  cursor: pointer;
  position: relative;
}
.menu-toggle span,
.menu-toggle span::before,
.menu-toggle span::after {
  content: '';
  position: absolute;
  left: 10px;
  width: 20px;
  height: 2px;
  background: var(--text);
  border-radius: 2px;
  transition: transform 0.25s ease, top 0.25s ease;
}
.menu-toggle span { top: 19px; }
.menu-toggle span::before { top: -6px; }
.menu-toggle span::after { top: 6px; }
.menu-toggle span.open { background: transparent; }
.menu-toggle span.open::before { top: 0; transform: rotate(45deg); }
.menu-toggle span.open::after { top: 0; transform: rotate(-45deg); }

.mobile-nav {
  display: none;
  flex-direction: column;
  padding: 0 1.5rem;
  gap: 0.25rem;
  max-height: 0;
  overflow: hidden;
  transition: max-height 0.3s ease, padding 0.3s ease;
}
.mobile-nav.open {
  max-height: 300px;
  padding-bottom: 1rem;
}

.main {
  flex: 1;
  display: flex;
  justify-content: center;
  padding: 2.5rem 1.5rem;
}

.footer {
  background: var(--surface);
  border-top: 1px solid var(--border);
  padding: 1.5rem;
  margin-top: auto;
}
.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}
.footer-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.footer-text {
  color: var(--text-muted);
  font-size: 0.88rem;
}
.footer-links {
  display: flex;
  gap: 1.25rem;
}
.footer-links a {
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.88rem;
  transition: color 0.2s;
}
.footer-links a:hover { color: var(--primary); }

@media (max-width: 720px) {
  .nav { display: none; }
  .menu-toggle { display: block; }
  .mobile-nav { display: flex; }
  .school-sub { display: none; }
}

@media (max-width: 560px) {
  .main { padding: 1.5rem 1rem; }
  .footer-content { justify-content: center; text-align: center; }
}
</style>
