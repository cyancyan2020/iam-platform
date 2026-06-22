import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { loginApi, type LoginParams, getProfileApi } from '@/api'
import { getPermissionListApi, type PermissionItem } from '@/api/system'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const username = ref<string>(localStorage.getItem('username') || '')
  const roleId = ref<number>(0)
  const permissions = ref<PermissionItem[]>([])

  const isLoggedIn = computed(() => !!token.value)

  // 根据权限树计算出侧边栏菜单项
  const sidebarMenus = computed(() => {
    return permissions.value
      .filter(p => p.method === 'MENU')
      .sort((a, b) => a.id - b.id)
  })

  async function login(params: LoginParams) {
    const res = await loginApi(params)
    const data = res.data.data
    if (data?.token) {
      token.value = data.token
      username.value = params.username
      localStorage.setItem('token', data.token)
      localStorage.setItem('username', params.username)
      await fetchProfile()
      await fetchPermissions()
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

  async function fetchPermissions() {
    try {
      const res = await getPermissionListApi()
      if (res.data.data) {
        permissions.value = res.data.data
      }
    } catch {
      // ignore permission fetch error
    }
  }

  function logout() {
    token.value = ''
    username.value = ''
    roleId.value = 0
    permissions.value = []
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    window.location.href = '/login'
  }

  return { token, username, roleId, permissions, sidebarMenus, isLoggedIn, login, fetchProfile, fetchPermissions, logout }
})
