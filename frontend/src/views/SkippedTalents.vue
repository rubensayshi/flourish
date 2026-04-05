<template>
  <div class="max-w-3xl mx-auto">
    <h2 class="text-xl font-bold mb-2">Skipped Talents</h2>
    <p class="text-slate-400 text-sm mb-6">
      These talents are excluded from the analysis. They either can't be meaningfully attributed,
      are always taken (so there's no choice to evaluate), or don't directly affect healing output.
    </p>

    <div v-if="loading" class="text-slate-400">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error }}</div>
    <div v-else>
      <div v-for="category in categories" :key="category" class="mb-8">
        <h3 class="text-lg font-semibold text-emerald-400 mb-3 border-b border-slate-700 pb-1">
          {{ category }}
        </h3>
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-slate-500">
              <th class="pb-2 pr-4 font-medium">Talent</th>
              <th class="pb-2 font-medium">Reason</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="talent in grouped[category]"
              :key="talent.name"
              class="border-t border-slate-800"
            >
              <td class="py-2 pr-4 text-slate-200 whitespace-nowrap">{{ talent.name }}</td>
              <td class="py-2 text-slate-400">{{ talent.reason }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const talents = ref([])
const loading = ref(true)
const error = ref(null)

const grouped = computed(() => {
  const groups = {}
  for (const t of talents.value) {
    if (!groups[t.category]) groups[t.category] = []
    groups[t.category].push(t)
  }
  return groups
})

const categories = computed(() => Object.keys(grouped.value).sort())

onMounted(async () => {
  try {
    const res = await fetch('/api/skipped-talents')
    if (!res.ok) throw new Error('Failed to load')
    talents.value = await res.json()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>
