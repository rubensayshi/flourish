<template>
  <div class="max-w-3xl mx-auto">
    <h2 class="text-xl font-bold mb-6">Settings</h2>

    <div class="max-w-lg">
      <div v-for="(meta, key) in SETTINGS_META" :key="key" class="mb-6">
        <label class="block font-semibold mb-1">
          {{ meta.label }}
          <span class="ml-2 text-emerald-400 font-mono">{{ settings[key] }}</span>
        </label>
        <p class="text-sm text-slate-400 mb-2">{{ meta.description }}</p>
        <div class="flex items-center gap-3">
          <span class="text-xs text-slate-500">{{ meta.min }}</span>
          <input
            type="range"
            :min="meta.min"
            :max="meta.max"
            :step="meta.step"
            v-model.number="settings[key]"
            class="w-full accent-emerald-500"
          />
          <span class="text-xs text-slate-500">{{ meta.max }}</span>
        </div>
      </div>

      <p class="text-xs text-slate-500 mt-8">
        Settings are saved in your browser and applied to future analyses.
      </p>
    </div>

    <!-- Skipped Talents -->
    <div class="mt-12 border-t border-slate-700 pt-8">
      <h2 class="text-xl font-bold mb-2">Skipped Talents</h2>
      <p class="text-slate-400 text-sm mb-6">
        These talents are excluded from the analysis. They either can't be meaningfully attributed,
        are always taken (so there's no choice to evaluate), or don't directly affect healing output.
      </p>

      <div v-if="skippedLoading" class="text-slate-400">Loading...</div>
      <div v-else-if="skippedError" class="text-red-400">{{ skippedError }}</div>
      <div v-else>
        <div v-for="category in skippedCategories" :key="category" class="mb-8">
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
                v-for="talent in skippedGrouped[category]"
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { settings, SETTINGS_META } from '../composables/useSettings'

const skippedTalents = ref([])
const skippedLoading = ref(true)
const skippedError = ref(null)

const skippedGrouped = computed(() => {
  const groups = {}
  for (const t of skippedTalents.value) {
    if (!groups[t.category]) groups[t.category] = []
    groups[t.category].push(t)
  }
  return groups
})

const skippedCategories = computed(() => Object.keys(skippedGrouped.value).sort())

onMounted(async () => {
  try {
    const res = await fetch('/api/skipped-talents')
    if (!res.ok) throw new Error('Failed to load')
    skippedTalents.value = await res.json()
  } catch (e) {
    skippedError.value = e.message
  } finally {
    skippedLoading.value = false
  }
})
</script>
