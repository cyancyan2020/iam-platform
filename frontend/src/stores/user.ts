import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { loginApi, type LoginParams, getProfileApi } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const username = ref<string>(localStorage.getItem('username') || '')
  const roleId = ref<number>(0)

  const isLoggedIn = computed(() => !!token.value)

  async function login(params: LoginParams) {
    const res = await loginApi(params)
    const data = res.data.data
    if (data?.token) {
      token.value = data.token
      username.value = params.username
      localStorage.setItem('token', data.token)
      localStorage.setItem('username', params.username)
      await fetchProfile()
    }
  }

  async function fetchProfile() {
    try {
      const res = await getProfileApi()
      if (res.data.data) {
        roleId.value = (res.data.data as any).role_id || 0
      }
    } catch {
      // ignore profile fetch error
    }
  }

  function logout() {
    token.value = ''
    username.value = ''
    roleId.value = 0
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    window.location.href = '/login'
  }

  return { token, username, roleId, isLoggedIn, login, fetchProfile, logout }
})
