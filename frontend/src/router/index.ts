// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import CreateFormView from '../views/CreateFormView.vue'
import FillFormView from '../views/FillFormView.vue'
import ResponsesView from '../views/ResponsesView.vue'

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
      path: '/create',
      name: 'create',
      component: CreateFormView
    },
    {
      path: '/fill/:id',
      name: 'fill',
      component: FillFormView
    },
    {
      path: '/responses/:id',
      name: 'responses',
      component: ResponsesView
    }
  ]
})

export default router
