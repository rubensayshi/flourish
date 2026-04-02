<template>
  <div class="flex gap-8 justify-center">
    <!-- Main content -->
    <div class="max-w-5xl w-full">
      <LoadingSpinner v-if="loading">Loading report...</LoadingSpinner>

      <div v-else-if="error" class="text-red-400">
        {{ error }}
        <button
          v-if="isLoginError"
          @click="auth.login()"
          class="ml-3 text-sm px-3 py-1.5 rounded bg-emerald-600 hover:bg-emerald-500 text-white transition-colors"
        >
          Login with WarcraftLogs
        </button>
      </div>

      <template v-else-if="report">
        <h2 class="text-xl font-bold mb-4">{{ report.title }}</h2>

        <div class="grid grid-cols-2 gap-4 mb-6">
          <FightSelector v-model="selectedFight" :fights="report.fights" />
          <PlayerSelector v-model="selectedPlayer" :druids="report.druids" />
        </div>

        <button
          @click="runAnalysis"
          :disabled="!selectedFight || !selectedPlayer || analyzing"
          class="rounded-lg bg-emerald-600 px-6 py-2.5 font-semibold text-white
                 hover:bg-emerald-500 disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {{ analyzing ? 'Analyzing...' : 'Run Analysis' }}
        </button>

        <LoadingSpinner v-if="analyzing" class="mt-4">
          Analyzing (this may take a few seconds)...
        </LoadingSpinner>

        <ResultsTable v-if="results" :data="results" class="mt-6" />
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
import { ref, computed, onMounted } from 'vue'
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
const isLoginError = computed(() => error.value && error.value.includes('Log in'))

const report = ref(null)
const loading = ref(true)
const error = ref(null)
const selectedFight = ref(0)
const selectedPlayer = ref('')
const analyzing = ref(false)
const results = ref(null)
const historyRefresh = ref(0)

onMounted(async () => {
  try {
    report.value = await fetchReport(route.params.code)

    if (route.params.fightId && route.params.player) {
      selectedFight.value = Number(route.params.fightId)
      selectedPlayer.value = route.params.player
      await runAnalysis()
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

async function runAnalysis() {
  analyzing.value = true
  error.value = null
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
    error.value = e.message
  } finally {
    analyzing.value = false
  }
}
</script>
