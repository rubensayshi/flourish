<template>
  <div v-if="entries.length" :class="sidebar ? '' : 'mt-10'">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-semibold text-slate-400 uppercase tracking-wide">Recent Analyses</h3>
      <button @click="handleClear" class="text-xs text-slate-500 hover:text-slate-300">Clear</button>
    </div>
    <ul class="space-y-1.5">
      <li v-for="(entry, i) in entries" :key="i">
        <router-link
          :to="`/results/${entry.code}/${entry.fightId}/${entry.player}`"
          class="flex items-center justify-between rounded-lg px-3 py-2 bg-slate-800/50
                 hover:bg-slate-800 border transition-colors"
          :class="isActive(entry) ? 'border-emerald-600/50 bg-emerald-900/10' : 'border-slate-700/50 hover:border-slate-600'"
        >
          <div class="min-w-0">
            <span class="text-slate-200 text-sm font-medium truncate block">{{ entry.title }}</span>
            <span class="text-slate-500 text-xs">{{ entry.fight }} &middot; {{ entry.player }}</span>
          </div>
          <span class="text-slate-600 text-xs ml-3 shrink-0">{{ timeAgo(entry.ts) }}</span>
        </router-link>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getReportHistory, clearReportHistory } from '../composables/useReportHistory'

const props = defineProps({
  sidebar: { type: Boolean, default: false },
  refreshKey: { type: Number, default: 0 },
})

const route = useRoute()
const entries = ref(getReportHistory())

watch(() => props.refreshKey, () => {
  entries.value = getReportHistory()
})

function isActive(entry) {
  return route.params.code === entry.code
    && Number(route.params.fightId) === entry.fightId
    && route.params.player === entry.player
}

function handleClear() {
  clearReportHistory()
  entries.value = []
}

function timeAgo(ts) {
  const diff = Date.now() - ts
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  const days = Math.floor(hrs / 24)
  return `${days}d ago`
}
</script>
