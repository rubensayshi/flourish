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

const emit = defineEmits(['submit'])
const input = ref('')

const code = computed(() => {
  const text = input.value.trim()
  const match = text.match(/warcraftlogs\.com\/reports\/([A-Za-z0-9]+)/)
  if (match) return match[1]
  if (/^[A-Za-z0-9]{10,20}$/.test(text)) return text
  return null
})

function submit() {
  if (code.value) emit('submit', code.value)
}
</script>
