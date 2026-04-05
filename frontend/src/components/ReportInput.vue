<template>
  <form @submit.prevent="submit" class="flex gap-3">
    <input
      v-model="input"
      type="text"
      placeholder="Paste WarcraftLogs URL or report code..."
      class="flex-1 rounded-lg bg-slate-800 border border-slate-600 px-4 py-3
             text-slate-100 placeholder-slate-500
             focus:outline-none focus:border-emerald-500"
    />
    <button
      type="submit"
      :disabled="!code"
      class="rounded-lg bg-emerald-600 px-6 py-3 font-semibold text-white
             hover:bg-emerald-500 disabled:opacity-40 disabled:cursor-not-allowed"
    >
      Analyze
    </button>
  </form>
  <p v-if="input && !code" class="mt-2 text-sm text-red-400">
    Could not extract a report code from this input.
  </p>
</template>

<script setup>
import { ref, computed } from 'vue'
import { parseReportUrl } from '../utils/parseReportUrl'

const emit = defineEmits(['submit'])
const input = ref('')

const parsed = computed(() => parseReportUrl(input.value))
const code = computed(() => parsed.value.code)

function submit() {
  if (parsed.value.code) emit('submit', parsed.value)
}
</script>
