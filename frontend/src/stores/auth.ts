import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/lib/api'
import type { JwtPayload, LoginResponse } from '@/lib/types'

function parseJwt(token: string): JwtPayload | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const role = ref<'admin' | 'profile' | null>(null)
  const profileId = ref<number | null>(null)
  const profileType = ref<'plus' | 'standard' | 'min' | null>(null)

  function initFromToken() {
    if (token.value) {
      const payload = parseJwt(token.value)
      if (payload && payload.exp * 1000 > Date.now()) {
        role.value = payload.role
        profileId.value = payload.profile_id ?? null
        profileType.value = payload.profile_type ?? null
      } else {
        clearAuth()
      }
    }
  }

  function clearAuth() {
    token.value = null
    role.value = null
    profileId.value = null
    profileType.value = null
    localStorage.removeItem('token')
  }

  async function login(loginValue: string, password: string) {
    const { data } = await api.post<LoginResponse>('/auth/login', {
      login: loginValue,
      password,
    })
    token.value = data.token
    localStorage.setItem('token', data.token)
    initFromToken()
  }

  function logout() {
    clearAuth()
  }

  const isAuthenticated = computed(() => !!token.value && !!role.value)
  const isAdmin = computed(() => role.value === 'admin')
  const canCreateChildren = computed(
    () => profileType.value === 'plus' || profileType.value === 'standard'
  )
  const displayRole = computed(() => {
    if (role.value === 'admin') return 'Администратор'
    if (profileType.value === 'plus') return 'Profile+'
    if (profileType.value === 'standard') return 'Profile'
    if (profileType.value === 'min') return 'ProfileMin'
    return 'Пользователь'
  })

  initFromToken()

  return {
    token,
    role,
    profileId,
    profileType,
    login,
    logout,
    isAuthenticated,
    isAdmin,
    canCreateChildren,
    displayRole,
  }
})
