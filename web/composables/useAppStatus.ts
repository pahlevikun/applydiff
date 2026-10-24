export type AppStatusPayload = {
  last_scan_at?: string
  daemon_seen_at?: string
}

const STALE_MS = 24 * 60 * 60 * 1000

export function useAppStatus() {
  const status = ref<AppStatusPayload | null>(null)
  const loading = ref(true)
  const offline = ref(false)
  const error = ref<string | null>(null)

  const lastScanAt = computed(() => {
    const raw = status.value?.last_scan_at
    if (!raw) return null
    const d = new Date(raw)
    return Number.isNaN(d.getTime()) ? null : d
  })

  const isStale = computed(() => {
    if (!lastScanAt.value) return false
    return Date.now() - lastScanAt.value.getTime() > STALE_MS
  })

  async function refresh() {
    loading.value = true
    error.value = null
    offline.value = false
    try {
      status.value = await $fetch<AppStatusPayload>('/api/status')
    } catch (e) {
      offline.value = true
      status.value = null
      error.value = e instanceof Error ? e.message : 'API unreachable'
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void refresh()
  })

  return { status, loading, offline, error, lastScanAt, isStale, refresh }
}
