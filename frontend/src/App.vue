<template>
  <div id="app">
    <header class="header">
      <div class="header-content">
        <div class="logo">
          <span class="school-name">Школа № 123</span>
        </div>
        <h1 class="service-title">Конструктор форм</h1>
        <nav class="nav">
          <router-link to="/" class="nav-link" active-class="active" exact>Главная</router-link>
          <router-link to="/login" class="nav-link" active-class="active">Вход</router-link>
        </nav>
      </div>
    </header>
    <main class="main">
      <router-view />
    </main>
    <footer class="footer">
      <p>&copy; 2026 Школа № 123. Все права защищены.</p>
    </footer>
  </div>
</template>

<script setup>
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  background: var(--bg);
  color: var(--text);
  min-height: 100vh;
}
#app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}
.header {
  background: #2c3e50;
  color: white;
  padding: 1rem 2rem;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
}
.school-name {
  font-size: 1.2rem;
  font-weight: bold;
  background: #3498db;
  padding: 0.3rem 0.8rem;
  border-radius: 4px;
}
.service-title {
  font-size: 1.5rem;
  font-weight: 300;
  letter-spacing: 1px;
}
.nav {
  display: flex;
  gap: 1rem;
}
.nav-link {
  color: white;
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  transition: background 0.2s;
}
.nav-link:hover {
  background: #34495e;
}
.nav-link.active {
  background: #34495e;
}
.main {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 2rem;
}
.footer {
  background: #2c3e50;
  color: #bbb;
  text-align: center;
  padding: 1rem;
  font-size: 0.9rem;
}
</style>
