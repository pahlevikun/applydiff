/** Build a client route that survives URL-like job ids (slashes, null bytes). */
export function jobRoute(base: 'match' | 'diff', jobId: string) {
  return { path: `/${base}`, query: { job: encodeJobIdForQuery(jobId) } }
}

export function encodeJobIdForApi(jobId: string) {
  return encodeURIComponent(jobId)
}

function encodeJobIdForQuery(jobId: string): string {
  if (typeof globalThis.TextEncoder === 'undefined') {
    return `b64:${btoa(jobId)}`
  }
  const bytes = new TextEncoder().encode(jobId)
  let bin = ''
  for (const b of bytes) {
    bin += String.fromCharCode(b)
  }
  const b64 = btoa(bin).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
  return `b64:${b64}`
}

function decodeJobIdFromQuery(encoded: string): string {
  const pad = encoded.length % 4 === 0 ? '' : '='.repeat(4 - (encoded.length % 4))
  const b64 = encoded.replace(/-/g, '+').replace(/_/g, '/') + pad
  const bin = atob(b64)
  if (typeof globalThis.TextDecoder !== 'undefined') {
    const bytes = Uint8Array.from(bin, (c) => c.charCodeAt(0))
    return new TextDecoder().decode(bytes)
  }
  return bin
}

/** Parse `job` query without losing embedded null bytes (dedupe keys). */
export function parseJobQueryParam(raw: string): string {
  const trimmed = raw.trim()
  if (!trimmed) return ''
  if (trimmed.startsWith('b64:')) {
    try {
      return decodeJobIdFromQuery(trimmed.slice(4))
    } catch {
      return ''
    }
  }
  return trimmed
}

/** Read job id from the browser query string (fallback when the router truncates at \\0). */
export function jobIdFromLocationSearch(): string {
  if (typeof window === 'undefined') return ''
  const qs = window.location.search
  if (!qs || qs.length < 2) return ''
  const params = qs.slice(1).split('&')
  for (const pair of params) {
    const eq = pair.indexOf('=')
    if (eq < 0) continue
    const name = pair.slice(0, eq)
    if (name !== 'job') continue
    const val = pair.slice(eq + 1)
    try {
      return parseJobQueryParam(decodeURIComponent(val.replace(/\+/g, '%20')))
    } catch {
      return parseJobQueryParam(val)
    }
  }
  return ''
}

/** Job id from `?job=` (preferred) or legacy `/match/:id` path segment. */
export function useJobIdFromRoute() {
  const route = useRoute()
  const fromLocation = ref('')
  onMounted(() => {
    const loc = jobIdFromLocationSearch()
    if (loc) fromLocation.value = loc
  })
  return computed(() => {
    const q = route.query.job
    if (typeof q === 'string' && q.trim()) {
      const parsed = parseJobQueryParam(q)
      if (fromLocation.value && fromLocation.value.length > parsed.length) {
        return fromLocation.value
      }
      return parsed
    }
    if (fromLocation.value) return fromLocation.value
    const id = route.params.id
    if (id === undefined || id === null || id === '') {
      return ''
    }
    const raw = Array.isArray(id) ? id.join('/') : String(id)
    try {
      return decodeURIComponent(raw)
    } catch {
      return raw
    }
  })
}
