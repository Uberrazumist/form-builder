// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import CreateFormView from '../views/CreateFormView.vue'
import FormView from '../views/FormView.vue'
import EditFormView from '../views/EditFormView.vue'
import FillFormView from '../views/FillFormView.vue'
import PreviewView from '../views/PreviewView.vue'
import ResponsesView from '../views/ResponsesView.vue'
import MyFormsView from '../views/MyFormsView.vue'
import VerifyEmailView from '../views/VerifyEmailView.vue'
import ForgotPasswordView from '../views/ForgotPasswordView.vue'
import ResetPasswordView from '../views/ResetPasswordView.vue'
import DictionariesView from '../views/DictionariesView.vue'
import DictionaryItemsView from '../views/DictionaryItemsView.vue'
import ScheduleView from '../views/ScheduleView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterView
    },
    {
      path: '/verify',
      name: 'verify',
      component: VerifyEmailView
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: ForgotPasswordView
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: ResetPasswordView
    },
    {
      path: '/create',
      name: 'create',
      component: CreateFormView,
      meta: { requiresAuth: true }
    },
    {
      path: '/form/:id',
      name: 'form',
      component: FormView
    },
    {
      path: '/edit/:id',
      name: 'edit',
      component: EditFormView,
      meta: { requiresAuth: true }
    },
    {
      path: '/fill/:id',
      name: 'fill',
      component: FillFormView
    },
    {
      path: '/preview/:id',
      name: 'preview',
      component: PreviewView
    },
    {
      path: '/responses/:id',
      name: 'responses',
      component: ResponsesView,
      meta: { requiresAuth: true }
    },
    {
      path: '/my-forms',
      name: 'my-forms',
      component: MyFormsView,
      meta: { requiresAuth: true }
    },
    {
      path: '/dictionaries',
      name: 'dictionaries',
      component: DictionariesView,
      meta: { requiresAuth: true }
    },
    {
      path: '/dictionaries/:id/items',
      name: 'dictionary-items',
      component: DictionaryItemsView,
      meta: { requiresAuth: true }
    },
    {
      path: '/schedule',
      name: 'schedule',
      component: ScheduleView,
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    return { name: 'login' }
  }
})

export default router
