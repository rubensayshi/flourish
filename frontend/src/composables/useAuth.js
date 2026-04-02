import { ref } from 'vue'

const STORAGE_KEY = 'flourish:wcl-token'

const token = ref(localStorage.getItem(STORAGE_KEY) || null)

export function useAuth() {
  function login() {
    window.location.href = '/api/auth/login'
  }

  function logout() {
    token.value = null
    localStorage.removeItem(STORAGE_KEY)
  }

  function setToken(t) {
    token.value = t
    localStorage.setItem(STORAGE_KEY, t)
  }

  function handleCallback() {
    const params = new URLSearchParams(window.location.search)
    const t = params.get('wcl_token')
    const error = params.get('auth_error')

    if (t) {
      setToken(t)
      // Clean URL
      window.history.replaceState({}, '', window.location.pathname)
      return { success: true }
    }
    if (error) {
      window.history.replaceState({}, '', window.location.pathname)
      return { success: false, error }
    }
    return null
  }

  return {
    token,
    isAuthenticated: () => !!token.value,
    login,
    logout,
    handleCallback,
  }
}
