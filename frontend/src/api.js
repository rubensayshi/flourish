const API_BASE = '/api'

function authHeaders() {
  const token = localStorage.getItem('flourish:wcl-token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export async function fetchReport(code) {
  const res = await fetch(`${API_BASE}/report/${code}`, {
    headers: authHeaders(),
  })
  if (!res.ok) {
    if (res.status === 403) {
      const data = await res.json()
      throw new Error(data.detail || 'Login required')
    }
    throw new Error(`Report not found (${res.status})`)
  }
  return res.json()
}

export async function fetchAnalysis(code, fightId, player, settings = {}) {
  const params = new URLSearchParams()

  const qs = params.toString()
  const url = `${API_BASE}/analyze/${code}/${fightId}/${encodeURIComponent(player)}${qs ? '?' + qs : ''}`
  const res = await fetch(url, {
    headers: authHeaders(),
  })
  if (!res.ok) {
    if (res.status === 403) {
      const data = await res.json()
      throw new Error(data.detail || 'Login required')
    }
    if (res.status === 429) throw new Error('Rate limit exceeded. Try again in a minute.')
    throw new Error(`Analysis failed (${res.status})`)
  }
  return res.json()
}
