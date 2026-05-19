import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { post, get } from '@/api/http'

interface UserInfo {
  id: number
  username: string
  email: string
}

interface LoginResponse {
  token: string
  user: UserInfo
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<UserInfo | null>(
    localStorage.getItem('user')
      ? JSON.parse(localStorage.getItem('user') as string)
      : null,
  )

  const isAuthenticated = computed(() => !!token.value)

  async function login(username: string, password: string): Promise<void> {
    const res = await post<LoginResponse>('/auth/login', { username, password })
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
  }

  async function fetchUser(): Promise<void> {
    try {
      const res = await get<UserInfo>('/auth/me')
      user.value = res.data
      localStorage.setItem('user', JSON.stringify(res.data))
    } catch {
      logout()
    }
  }

  function logout(): void {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return { token, user, isAuthenticated, login, fetchUser, logout }
})
