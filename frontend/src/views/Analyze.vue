<template>
  <div class="flex gap-8 justify-center">
    <!-- Main content -->
    <div class="max-w-5xl w-full">
      <LoadingSpinner v-if="loading">Loading report...</LoadingSpinner>

      <div v-else-if="error && !isLoginError" class="text-red-400">{{ error }}</div>

      <template v-else>
        <!-- Login prompt (shown for both page-load and mid-analysis limit errors) -->
        <div v-if="isLoginError" class="mt-4 mb-6 rounded-lg border border-amber-500/30 bg-amber-950/50 p-6 max-w-md mx-auto text-center">
          <div class="w-10 h-10 mx-auto mb-3 rounded-full bg-amber-500/20 flex items-center justify-center">
            <svg class="w-5 h-5 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v.01M12 9v3m-7.938 5A9.968 9.968 0 0012 22a9.968 9.968 0 007.938-4M4.062 13A9.968 9.968 0 012 12C2 6.477 6.477 2 12 2s10 4.477 10 10c0 1.105-.18 2.168-.512 3.162" />
            </svg>
          </div>
          <p class="text-amber-300 font-semibold text-lg mb-2">Free analysis limit reached</p>
          <p class="text-slate-400 text-sm mb-5 leading-relaxed">
            To stay within WarcraftLogs API rate limits, we ask you to log in after 2 free analyses.
            We only use your login to analyze logs on your behalf — nothing else.
          </p>
          <button
            @click="auth.login()"
            class="px-5 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white font-semibold transition-colors"
          >
            Login with WarcraftLogs
          </button>
        </div>

        <template v-if="report">
          <h2 class="text-xl font-bold mb-4">{{ report.title }}</h2>

          <div class="flex items-end gap-4 mb-6">
            <FightSelector v-model="selectedFight" :fights="report.fights" class="flex-1" />
            <PlayerSelector v-model="selectedPlayer" :druids="report.druids" class="flex-1" />
            <button
              @click="runAnalysis"
              :disabled="!selectedFight || !selectedPlayer || analyzing || isLoginError"
              class="rounded-lg bg-emerald-600 px-6 py-2.5 font-semibold text-white whitespace-nowrap
                     hover:bg-emerald-500 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {{ analyzing ? 'Analyzing...' : 'Run Analysis' }}
            </button>
          </div>

          <div v-if="analyzeError && !isLoginError" class="mt-4 text-red-400">{{ analyzeError }}</div>

          <LoadingSpinner v-if="analyzing" class="mt-4">
            Analyzing (this may take a few seconds)...
          </LoadingSpinner>

          <ResultsTable v-if="results" :data="results" class="mt-6" />
        </template>
      </template>
    </div>

    <!-- Sidebar: recent analyses (large screens only, once results shown) -->
    <aside v-if="results" class="hidden xl:block w-72 shrink-0">
      <div class="sticky top-8">
        <ReportHistory sidebar :refreshKey="historyRefresh" />
      </div>
    </aside>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchReport, fetchAnalysis } from '../api'
import { addReportEntry } from '../composables/useReportHistory'
import { settings } from '../composables/useSettings'
import { useAuth } from '../composables/useAuth'
import FightSelector from '../components/FightSelector.vue'
import PlayerSelector from '../components/PlayerSelector.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ResultsTable from '../components/ResultsTable.vue'
import ReportHistory from '../components/ReportHistory.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuth()

const report = ref(null)
const loading = ref(true)
const error = ref(null)
const analyzeError = ref(null)
const selectedFight = ref(0)
const selectedPlayer = ref('')
const analyzing = ref(false)
const results = ref(null)
const historyRefresh = ref(0)
const isLoginError = computed(() => {
  const msg = analyzeError.value || error.value || ''
  return msg.includes('Log in')
})

async function loadFromRoute() {
  loading.value = true
  error.value = null
  analyzeError.value = null
  results.value = null
  try {
    report.value = await fetchReport(route.params.code)

    // Pre-select from /results/:code/:fightId/:player route
    if (route.params.fightId && route.params.player) {
      selectedFight.value = Number(route.params.fightId)
      selectedPlayer.value = route.params.player
      await runAnalysis()
    }
    // Pre-select from ?fight=X&source=Y query params (from WCL URL)
    else if (route.query.fight && route.query.source) {
      const fights = report.value.fights
      if (route.query.fight === 'last') {
        selectedFight.value = fights.length ? fights[fights.length - 1].id : 0
      } else {
        selectedFight.value = Number(route.query.fight)
      }
      const sourceId = Number(route.query.source)
      const druid = report.value.druids.find(d => d.id === sourceId)
      if (druid) {
        selectedPlayer.value = druid.name
        await runAnalysis()
      }
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadFromRoute)

watch(() => [route.params.code, route.params.fightId, route.params.player], (newP, oldP) => {
  if (newP[0] !== oldP[0] || newP[1] !== oldP[1] || newP[2] !== oldP[2]) {
    loadFromRoute()
  }
})

async function runAnalysis() {
  analyzing.value = true
  analyzeError.value = null
  try {
    results.value = await fetchAnalysis(
      route.params.code, selectedFight.value, selectedPlayer.value, settings
    )
    const fight = report.value.fights.find(f => f.id === selectedFight.value)
    addReportEntry({
      code: route.params.code,
      title: report.value.title,
      fight: fight ? `${fight.name} (${fight.kill ? 'Kill' : 'Wipe'})` : `Fight ${selectedFight.value}`,
      fightId: selectedFight.value,
      player: selectedPlayer.value,
    })
    historyRefresh.value++
    router.replace(`/results/${route.params.code}/${selectedFight.value}/${selectedPlayer.value}`)
  } catch (e) {
    analyzeError.value = e.message
  } finally {
    analyzing.value = false
  }
}
</script>
