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
      component: CreateFormView
    },
    {
      path: '/form/:id',
      name: 'form',
      component: FormView
    },
    {
      path: '/edit/:id',
      name: 'edit',
      component: EditFormView
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
      component: ResponsesView
    },
    {
      path: '/my-forms',
      name: 'my-forms',
      component: MyFormsView
    },
    {
      path: '/dictionaries',
      name: 'dictionaries',
      component: DictionariesView
    },
    {
      path: '/dictionaries/:id/items',
      name: 'dictionary-items',
      component: DictionaryItemsView
    },
    {
      path: '/schedule',
      name: 'schedule',
      component: ScheduleView
    }
  ]
})

export default router
