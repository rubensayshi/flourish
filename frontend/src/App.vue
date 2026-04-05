<template>
  <div class="min-h-screen bg-slate-900 text-slate-100">
    <header class="border-b border-slate-700 px-6 py-4 flex items-center justify-between">
      <router-link to="/" class="text-xl font-bold text-emerald-400 hover:text-emerald-300 flex items-center gap-2">
        <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none">
          <path d="M12 2C6.5 6 3 11 3 16c0 3 2 5 4.5 5.5C8 19 9.5 16 12 14c2.5 2 4 5 4.5 7.5C19 21 21 19 21 16c0-5-3.5-10-9-14z" fill="currentColor" opacity="0.9"/>
          <path d="M12 8c-1.5 2.5-2.5 5-2.5 7.5 0 1.5.5 3 2.5 4 2-1 2.5-2.5 2.5-4 0-2.5-1-5-2.5-7.5z" fill="currentColor" opacity="0.4"/>
        </svg>
        Flourish
      </router-link>
      <div class="flex items-center gap-4">
        <form v-if="route.path !== '/'" @submit.prevent="goToReport" class="flex gap-2">
          <input
            v-model="headerInput"
            type="text"
            placeholder="Paste report URL or code..."
            class="w-64 rounded bg-slate-800 border border-slate-600 px-3 py-1.5 text-sm
                   text-slate-100 placeholder-slate-500
                   focus:outline-none focus:border-emerald-500"
          />
          <button
            type="submit"
            :disabled="!headerCode"
            class="rounded bg-emerald-600 px-3 py-1.5 text-sm font-medium text-white
                   hover:bg-emerald-500 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Go
          </button>
        </form>
        <a href="https://github.com/rubensayshi/flourish" target="_blank" rel="noopener noreferrer" class="text-slate-400 hover:text-emerald-400" title="GitHub">
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
          </svg>
        </a>
        <router-link to="/settings" class="text-slate-400 hover:text-emerald-400" title="Settings">
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/>
          </svg>
        </router-link>
        <button
          v-if="!auth.isAuthenticated()"
          @click="auth.login()"
          class="text-sm px-3 py-1.5 rounded bg-emerald-600 hover:bg-emerald-500 text-white transition-colors"
        >
          Login with WarcraftLogs
        </button>
        <div v-else class="flex items-center gap-2">
          <span class="text-sm text-emerald-400">Logged in</span>
          <button
            @click="auth.logout()"
            class="text-sm px-3 py-1.5 rounded bg-slate-700 hover:bg-slate-600 text-slate-300 transition-colors"
          >
            Logout
          </button>
        </div>
      </div>
    </header>
    <main class="px-4 py-8">
      <router-view />
    </main>
    <WelcomeModal />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from './composables/useAuth.js'
import WelcomeModal from './components/WelcomeModal.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuth()

const headerInput = ref('')
const headerCode = computed(() => {
  const text = headerInput.value.trim()
  const match = text.match(/warcraftlogs\.com\/reports\/([A-Za-z0-9]+)/)
  if (match) return match[1]
  if (/^[A-Za-z0-9]{10,20}$/.test(text)) return text
  return null
})

function goToReport() {
  if (headerCode.value) {
    router.push(`/analyze/${headerCode.value}`)
    headerInput.value = ''
  }
}

onMounted(() => {
  const result = auth.handleCallback()
  if (result && !result.success) {
    console.error('WCL OAuth error:', result.error)
  }
})
</script>
