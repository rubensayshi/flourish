<template>
  <div class="mt-16 text-center">
    <h1 class="text-3xl font-bold text-emerald-400 mb-2">
      Resto Druid Talent Analyzer
    </h1>
    <p class="text-slate-400 mb-8">
      See how much healing each talent contributes in your log.
    </p>
    <div class="max-w-xl mx-auto">
      <ReportInput @submit="goToReport" />
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import ReportInput from '../components/ReportInput.vue'

const router = useRouter()

function goToReport(code) {
  router.push(`/analyze/${code}`)
}
</script>
