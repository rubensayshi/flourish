const STORAGE_KEY = 'flourish:report-history'
const MAX_ENTRIES = 10

function load() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY)) || []
  } catch {
    return []
  }
}

function save(entries) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(entries))
}

export function getReportHistory() {
  return load()
}

export function addReportEntry({ code, title, fight, fightId, player }) {
  const entries = load()
  // Remove existing entry for same code+fight+player
  const filtered = entries.filter(
    e => !(e.code === code && e.fightId === fightId && e.player === player)
  )
  filtered.unshift({ code, title, fight, fightId, player, ts: Date.now() })
  save(filtered.slice(0, MAX_ENTRIES))
}

export function removeReportEntry(index) {
  const entries = load()
  entries.splice(index, 1)
  save(entries)
}

export function clearReportHistory() {
  localStorage.removeItem(STORAGE_KEY)
}
