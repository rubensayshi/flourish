import { reactive, watch } from 'vue'

const STORAGE_KEY = 'flourish-settings'

const DEFAULTS = {
  baseStacks: 3,
}

function load() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? { ...DEFAULTS, ...JSON.parse(raw) } : { ...DEFAULTS }
  } catch {
    return { ...DEFAULTS }
  }
}

export const settings = reactive(load())

watch(settings, (val) => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(val))
})

export const SETTINGS_META = {
  baseStacks: { label: 'Base HoT stacks', min: 1, max: 5, step: 1, description: 'Average HoTs on a target before talent-added stacks. Used by Harmonious Blooming and Symbiotic Bloom Mastery to estimate marginal mastery gain. A fixed estimate avoids the complexity of tracking per-target HoT counts on every heal.' },
}
