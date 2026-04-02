const API_BASE = '/api'

export async function fetchReport(code) {
  const res = await fetch(`${API_BASE}/report/${code}`)
  if (!res.ok) throw new Error(`Report not found (${res.status})`)
  return res.json()
}

export async function fetchAnalysis(code, fightId, player) {
  const res = await fetch(`${API_BASE}/analyze/${code}/${fightId}/${encodeURIComponent(player)}`)
  if (!res.ok) {
    if (res.status === 429) throw new Error('Rate limit exceeded. Try again in a minute.')
    throw new Error(`Analysis failed (${res.status})`)
  }
  return res.json()
}
