<template>
  <div class="mt-16 text-center mx-auto max-w-5xl">
    <svg class="w-16 h-16 mx-auto mb-4 text-emerald-400" viewBox="0 0 24 24" fill="none">
      <path d="M12 2C6.5 6 3 11 3 16c0 3 2 5 4.5 5.5C8 19 9.5 16 12 14c2.5 2 4 5 4.5 7.5C19 21 21 19 21 16c0-5-3.5-10-9-14z" fill="currentColor" opacity="0.9"/>
      <path d="M12 8c-1.5 2.5-2.5 5-2.5 7.5 0 1.5.5 3 2.5 4 2-1 2.5-2.5 2.5-4 0-2.5-1-5-2.5-7.5z" fill="currentColor" opacity="0.4"/>
    </svg>
    <h1 class="text-4xl font-bold text-emerald-400 mb-2">
      Flourish
    </h1>
    <p class="text-slate-400 mb-1">
      Resto Druid Talent Analyzer
    </p>
    <p class="text-slate-500 text-sm mb-8">
      See how much healing each talent contributes in your WarcraftLogs report.
    </p>
    <div class="max-w-xl mx-auto">
      <ReportInput @submit="goToReport" />
      <ReportHistory />
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import ReportInput from '../components/ReportInput.vue'
import ReportHistory from '../components/ReportHistory.vue'

const router = useRouter()

function goToReport(code) {
  router.push(`/analyze/${code}`)
}
</script>
