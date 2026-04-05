import { reactive, watch } from 'vue'

const STORAGE_KEY = 'flourish-settings'

const DEFAULTS = {}

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

export const SETTINGS_META = {}
