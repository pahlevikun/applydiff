/** URL segment for `/job/[key]` (job ids may contain slashes and null bytes). */
export function jobKeyToRouteParam(jobKey: string): string {
  return encodeJobIdForQuery(jobKey)
}

export function jobKeyFromRouteParam(raw: string): string {
  const trimmed = raw.trim()
  if (!trimmed) return ''
  return parseJobQueryParam(trimmed.startsWith('b64:') ? trimmed : `b64:${trimmed}`)
}

export function jobDetailPath(jobKey: string, tab?: string) {
  const key = jobKeyToRouteParam(jobKey)
  const query = tab ? { tab } : undefined
  const segment = key.startsWith('b64:') ? key.slice(4) : key
  return { path: `/job/${encodeURIComponent(segment)}`, query }
}
