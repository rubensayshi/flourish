export function parseReportUrl(text) {
  text = text.trim()
  let code = null
  let fightId = null
  let sourceId = null

  const urlMatch = text.match(/warcraftlogs\.com\/reports\/([A-Za-z0-9]+)/)
  if (urlMatch) {
    code = urlMatch[1]
    // WCL puts fight/source in either query string (?fight=) or hash fragment (#fight=)
    const fightMatch = text.match(/[?&#]fight=(\w+)/)
    if (fightMatch) fightId = fightMatch[1] === 'last' ? 'last' : Number(fightMatch[1])
    const sourceMatch = text.match(/[?&#]source=(\d+)/)
    if (sourceMatch) sourceId = Number(sourceMatch[1])
  } else if (/^[A-Za-z0-9]{10,20}$/.test(text)) {
    code = text
  }

  return { code, fightId, sourceId }
}
