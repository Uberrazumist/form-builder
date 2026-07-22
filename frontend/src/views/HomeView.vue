<!-- src/views/HomeView.vue -->
<template>
  <div class="home">
    <!-- HERO -->
    <section class="hero">
      <div class="hero-content">
        <div class="hero-text">
          <span class="eyebrow">Конструктор форм для школы</span>
          <h1 class="hero-title">Создавайте формы<br/>за минуты, а не часы</h1>
          <p class="hero-desc">
            Единый сервис для регистрации, опросов и сбора данных.
            Простой интерфейс, надёжная защита, удобный доступ для учеников и учителей.
          </p>

          <div class="hero-actions">
            <template v-if="isAuthenticated">
              <router-link to="/create" class="btn-primary-large">
                <Icon name="plus" />
                Создать форму
              </router-link>
              <router-link to="/my-forms" class="btn-secondary-large">
                <Icon name="document" />
                Мои формы
              </router-link>
            </template>
            <template v-else>
              <router-link to="/login" class="btn-primary-large">Войти</router-link>
              <router-link to="/register" class="btn-secondary-large">Зарегистрироваться</router-link>
            </template>
          </div>
        </div>

        <div class="hero-visual">
          <div class="visual-card">
            <div class="visual-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <rect x="3" y="3" width="18" height="18" rx="2"/>
                <path d="M9 12l2 2 4-4"/>
              </svg>
            </div>
            <h3>Быстро и просто</h3>
            <p>Интуитивный интерфейс для создания любых форм</p>
          </div>
          <div class="visual-card">
            <div class="visual-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                <path d="M2 17l10 5 10-5M2 12l10 5 10-5"/>
              </svg>
            </div>
            <h3>Гибкие настройки</h3>
            <p>6 типов вопросов, зависимости, публичный доступ</p>
          </div>
          <div class="visual-card">
            <div class="visual-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/>
                <circle cx="9" cy="7" r="4"/>
                <path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/>
              </svg>
            </div>
            <h3>Для всей школы</h3>
            <p>Ученики, учителя, администрация — все в одном месте</p>
          </div>
        </div>
      </div>
    </section>

    <!-- QUESTION TYPES GUIDE -->
    <section class="question-types">
      <div class="section-header">
        <span class="eyebrow">Инструкция</span>
        <h2 class="section-title">Типы вопросов в формах</h2>
        <p class="section-subtitle">Выбирайте подходящий тип для каждой задачи</p>
      </div>

      <div class="types-grid">
        <div class="type-card">
          <div class="type-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <line x1="17" y1="10" x2="3" y2="10"/>
              <line x1="21" y1="6" x2="3" y2="6"/>
              <line x1="21" y1="14" x2="3" y2="14"/>
              <line x1="17" y1="18" x2="3" y2="18"/>
            </svg>
          </div>
          <div class="type-content">
            <h3>Текст (одна строка)</h3>
            <p class="type-desc">Короткий ответ в одну строку</p>
            <p class="type-example">Пример: имя, фамилия, номер телефона</p>
          </div>
        </div>

        <div class="type-card">
          <div class="type-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <rect x="3" y="3" width="18" height="18" rx="2"/>
              <line x1="7" y1="8" x2="17" y2="8"/>
              <line x1="7" y1="12" x2="17" y2="12"/>
              <line x1="7" y1="16" x2="13" y2="16"/>
            </svg>
          </div>
          <div class="type-content">
            <h3>Текст (несколько строк)</h3>
            <p class="type-desc">Развёрнутый ответ в несколько строк</p>
            <p class="type-example">Пример: отзыв, предложение, комментарий</p>
          </div>
        </div>

        <div class="type-card">
          <div class="type-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="9"/>
              <circle cx="12" cy="12" r="4" fill="currentColor"/>
            </svg>
          </div>
          <div class="type-content">
            <h3>Один вариант (radio)</h3>
            <p class="type-desc">Выбор одного варианта из списка</p>
            <p class="type-example">Пример: пол, класс, любимый предмет</p>
          </div>
        </div>

        <div class="type-card">
          <div class="type-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <rect x="3" y="3" width="18" height="18" rx="2"/>
              <path d="M9 12l2 2 4-4"/>
            </svg>
          </div>
          <div class="type-content">
            <h3>Несколько вариантов (checkbox)</h3>
            <p class="type-desc">Выбор нескольких вариантов</p>
            <p class="type-example">Пример: хобби, интересы, дни недели</p>
          </div>
        </div>

        <div class="type-card">
          <div class="type-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <rect x="3" y="5" width="18" height="14" rx="2"/>
              <path d="M7 10l5 5 5-5"/>
            </svg>
          </div>
          <div class="type-content">
            <h3>Выбор из списка (select)</h3>
            <p class="type-desc">Выпадающий список с вариантами</p>
            <p class="type-example">Пример: страна, категория, статус</p>
          </div>
        </div>

        <div class="type-card">
          <div class="type-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" fill="currentColor"/>
            </svg>
          </div>
          <div class="type-content">
            <h3>Рейтинг (звёзды)</h3>
            <p class="type-desc">Оценка по шкале от 1 до 5 или 10</p>
            <p class="type-example">Пример: оценка мероприятия, качество обслуживания</p>
          </div>
        </div>
      </div>
    </section>

    <!-- DICTIONARIES GUIDE -->
    <section class="dictionaries-guide">
      <div class="section-header">
        <span class="eyebrow">Справочники</span>
        <h2 class="section-title">Как работать со справочниками</h2>
        <p class="section-subtitle">Справочники помогают не вводить одни и те же списки вручную в каждой форме</p>
      </div>

      <div class="guide-grid">
        <div class="guide-card">
          <div class="guide-icon">
            <Icon name="book" />
          </div>
          <div class="guide-content">
            <h3>Что такое справочник?</h3>
            <p>Это список, который вы создаёте один раз и используете в разных формах. Например, список классов (9А, 9Б, 10А), список учителей или список доступного времени для записи.</p>
          </div>
        </div>

        <div class="guide-card">
          <div class="guide-icon">
            <Icon name="plus" />
          </div>
          <div class="guide-content">
            <h3>Как создать справочник?</h3>
            <p>Перейдите в раздел «Справочники» в меню сверху, нажмите кнопку «Создать справочник» и дайте ему понятное название, например «Классы» или «Учителя математики».</p>
          </div>
        </div>

        <div class="guide-card">
          <div class="guide-icon">
            <Icon name="document" />
          </div>
          <div class="guide-content">
            <h3>Как добавить элементы?</h3>
            <p>Откройте созданный справочник, нажмите «Добавить элемент» и введите название — например, «9А» или «Иванова Мария Петровна». Так можно собрать весь список.</p>
          </div>
        </div>

        <div class="guide-card">
          <div class="guide-icon">
            <Icon name="edit" />
          </div>
          <div class="guide-content">
            <h3>Как использовать в форме?</h3>
            <p>В конструкторе формы добавьте вопрос, выберите тип «Выбор из справочника» и укажите нужный справочник. Варианты ответов подтянутся автоматически.</p>
          </div>
        </div>

        <div class="guide-card">
          <div class="guide-icon">
            <Icon name="link" />
          </div>
          <div class="guide-content">
            <h3>Зависимости между вопросами</h3>
            <p>Если нужно, чтобы при выборе класса показывались только учителя этого класса — в вопросе с учителями выберите «Зависит от вопроса» и укажите вопрос с классом. Список автоматически отфильтруется.</p>
          </div>
        </div>

        <div class="guide-card">
          <div class="guide-icon">
            <Icon name="calendar" />
          </div>
          <div class="guide-content">
            <h3>Проверка занятости</h3>
            <p>Если вы делаете форму для записи — например, к учителю на определённое время — включите в вопросе опцию «Проверять занятость». Тогда занятые слоты будут отмечены, и два человека не смогут записаться на одно и то же время.</p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA для неавторизованных -->
    <section v-if="!isAuthenticated" class="cta">
      <div class="cta-card">
        <h2>Готовы начать?</h2>
        <p>Зарегистрируйтесь и создайте свою первую форму за 5 минут</p>
        <div class="cta-actions">
          <router-link to="/register" class="btn-primary-large">Зарегистрироваться</router-link>
          <router-link to="/login" class="btn-secondary-large">Уже есть аккаунт? Войти</router-link>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import Icon from '../components/Icon.vue'

const router = useRouter()
const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)

const logout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.home {
  width: 100%;
  max-width: 1200px;
  display: flex;
  flex-direction: column;
  gap: 4rem;
}

/* HERO */
.hero-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
  align-items: center;
}

.hero-text {
  animation: fadeUp 0.6s ease both;
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
  margin-bottom: 1.25rem;
}

.hero-title {
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 700;
  line-height: 1.15;
  letter-spacing: -0.02em;
  color: var(--text);
  margin-bottom: 1rem;
}

.hero-desc {
  font-size: 1.1rem;
  color: var(--text-muted);
  line-height: 1.6;
  margin-bottom: 2rem;
  max-width: 500px;
}

.hero-actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.btn-primary-large,
.btn-secondary-large {
  padding: 0.95rem 2rem;
  font-size: 1rem;
  font-weight: 600;
  border-radius: var(--radius-sm);
  text-decoration: none;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  border: none;
  font-family: inherit;
}

.btn-primary-large {
  background: var(--primary);
  color: #fff;
  box-shadow: 0 4px 14px rgba(47, 79, 138, 0.25);
}

.btn-primary-large:hover {
  background: var(--primary-hover);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(47, 79, 138, 0.32);
}

.btn-secondary-large {
  background: var(--surface);
  color: var(--text);
  border: 1.5px solid var(--border);
}

.btn-secondary-large:hover {
  background: var(--bg);
  border-color: #cfd6e3;
}

.btn-primary-large svg,
.btn-secondary-large svg {
  width: 18px;
  height: 18px;
}

.hero-visual {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  animation: fadeUp 0.6s ease 0.1s both;
}

.visual-card {
  background: var(--surface);
  padding: 1.5rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-sm);
  transition: all 0.3s;
}

.visual-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.visual-card:first-child {
  grid-column: span 2;
}

.visual-icon {
  width: 48px;
  height: 48px;
  background: var(--primary-soft);
  color: var(--primary);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.visual-icon svg {
  width: 24px;
  height: 24px;
}

.visual-card h3 {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.5rem;
}

.visual-card p {
  font-size: 0.9rem;
  color: var(--text-muted);
  line-height: 1.5;
}

/* QUESTION TYPES */
.section-header {
  text-align: center;
  margin-bottom: 2.5rem;
}

.section-title {
  font-size: clamp(1.5rem, 3vw, 2rem);
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.01em;
  margin-bottom: 0.5rem;
}

.section-subtitle {
  font-size: 1rem;
  color: var(--text-muted);
  max-width: 600px;
  margin: 0 auto;
}

.types-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;
}

.type-card {
  display: flex;
  gap: 1.25rem;
  padding: 1.5rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  transition: all 0.3s;
}

.type-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-md);
  border-color: var(--primary-soft);
}

.type-icon {
  flex-shrink: 0;
  width: 52px;
  height: 52px;
  background: var(--primary-soft);
  color: var(--primary);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.type-icon svg {
  width: 26px;
  height: 26px;
}

.type-content {
  flex: 1;
  min-width: 0;
}

.type-content h3 {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.35rem;
  letter-spacing: -0.01em;
}

.type-desc {
  font-size: 0.92rem;
  color: var(--text);
  margin-bottom: 0.35rem;
  line-height: 1.5;
}

.type-example {
  font-size: 0.85rem;
  color: var(--text-muted);
  font-style: italic;
  line-height: 1.4;
}

/* DICTIONARIES GUIDE */
.guide-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;
}

.guide-card {
  display: flex;
  gap: 1.25rem;
  padding: 1.75rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  transition: all 0.3s;
}

.guide-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-md);
  border-color: var(--primary-soft);
}

.guide-icon {
  flex-shrink: 0;
  width: 52px;
  height: 52px;
  background: var(--primary-soft);
  color: var(--primary);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.guide-icon svg {
  width: 26px;
  height: 26px;
}

.guide-content {
  flex: 1;
  min-width: 0;
}

.guide-content h3 {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.5rem;
  letter-spacing: -0.01em;
}

.guide-content p {
  font-size: 0.93rem;
  color: var(--text-muted);
  line-height: 1.6;
}

/* CTA */
.cta-card {
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-hover) 100%);
  padding: 3rem 2rem;
  border-radius: var(--radius);
  text-align: center;
  color: #fff;
  animation: fadeUp 0.5s ease both;
}

.cta-card h2 {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.75rem;
}

.cta-card p {
  font-size: 1.1rem;
  opacity: 0.9;
  margin-bottom: 2rem;
}

.cta-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.cta-actions .btn-primary-large {
  background: #fff;
  color: var(--primary);
}

.cta-actions .btn-primary-large:hover {
  background: var(--bg);
}

.cta-actions .btn-secondary-large {
  background: transparent;
  color: #fff;
  border-color: rgba(255,255,255,0.3);
}

.cta-actions .btn-secondary-large:hover {
  background: rgba(255,255,255,0.1);
  border-color: rgba(255,255,255,0.5);
}

/* ANIMATIONS */
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

/* RESPONSIVE */
@media (max-width: 960px) {
  .hero-content {
    grid-template-columns: 1fr;
    gap: 2rem;
  }
  .hero-text { text-align: center; }
  .hero-desc { margin-left: auto; margin-right: auto; }
  .hero-actions { justify-content: center; }
  .types-grid { grid-template-columns: 1fr; }
  .guide-grid { grid-template-columns: 1fr; }
}

@media (max-width: 560px) {
  .hero-actions { flex-direction: column; }
  .btn-primary-large,
  .btn-secondary-large { width: 100%; justify-content: center; }
  .cta-actions { flex-direction: column; }
  .type-card,
  .guide-card {
    flex-direction: column;
    gap: 1rem;
  }
}
</style>
