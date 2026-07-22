import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<string | null>(localStorage.getItem('user'))

  const isAuthenticated = computed(() => !!token.value)

  // Синхронизация: при изменении token обновляем localStorage
  watch(token, (newToken) => {
    if (newToken) {
      localStorage.setItem('token', newToken)
    } else {
      localStorage.removeItem('token')
    }
  })

  watch(user, (newUser) => {
    if (newUser) {
      localStorage.setItem('user', newUser)
    } else {
      localStorage.removeItem('user')
    }
  })

  function setToken(newToken: string) {
    token.value = newToken
  }

  function setUser(newUser: string) {
    user.value = newUser
  }

  function logout() {
    token.value = null
    user.value = null
  }

  return { token, user, isAuthenticated, setToken, setUser, logout }
})
