export type ActivityLevel = 'info' | 'ok' | 'err' | 'run'

export type ActivityLine = {
  id: string
  at: number
  level: ActivityLevel
  text: string
}

export type ActiveRunStatus = 'running' | 'success' | 'failed' | 'timeout'

export type ActiveRun = {
  id: string
  label: string
  kind: string
  startedAt: number
  status: ActiveRunStatus
  phase: string
  progress: number
  timeoutMs: number
  error?: string
  detail?: string
}

type RunRow = {
  id: string
  kind: string
  job_key: string
  status: string
  error?: string
  result_json?: string
  created_at?: string
  updated_at?: string
}

import { msg } from '~/i18n'

const TAILOR_TIMEOUT_MS = 12 * 60 * 1000
const DEFAULT_TIMEOUT_MS = 30 * 60 * 1000
const MAX_LINES = 150

function tailorPhase(elapsedMs: number): string {
  if (elapsedMs < 2500) return msg('activity.phases.tailorStart')
  if (elapsedMs < 8000) return msg('activity.phases.tailorLoad')
  if (elapsedMs < 120_000) return msg('activity.phases.tailorLlm')
  return msg('activity.phases.tailorSlow')
}

function tailorProgress(elapsedMs: number, serverStatus: string): number {
  if (serverStatus === 'done') return 100
  if (serverStatus === 'failed') return 100
  const expected = 150_000
  const p = Math.min(92, Math.round((elapsedMs / expected) * 92))
  return Math.max(4, p)
}

function summaryFromResult(kind: string, resultJson?: string): string | undefined {
  if (!resultJson) return undefined
  try {
    const data = JSON.parse(resultJson) as {
      hunks?: unknown[]
      fetched?: number
      stored?: number
      total?: number
    }
    if (kind === 'tailor' && Array.isArray(data.hunks)) {
      const n = data.hunks.length
      return n ? `${n} proposed change${n === 1 ? '' : 's'}` : 'No hunks returned'
    }
    if (kind === 'scan' && typeof data.fetched === 'number') {
      return `fetched ${data.fetched}, stored ${data.stored ?? 0}`
    }
    if (kind === 'match' && typeof data.total === 'number') {
      return `score ${data.total}/25`
    }
  } catch {
    /* ignore */
  }
  return undefined
}

function formatDuration(ms: number): string {
  const s = Math.floor(ms / 1000)
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  const rem = s % 60
  return rem ? `${m}m ${rem}s` : `${m}m`
}

export function actionErrorMessage(e: unknown): string {
  if (e && typeof e === 'object') {
    const data = (e as { data?: unknown }).data
    if (typeof data === 'string' && data.trim() && !data.trim().startsWith('{') && data.length < 180) {
      return data.trim()
    }
  }
  if (e instanceof Error) {
    const m = e.message.trim()
    if (!m || m.startsWith('[') || m.includes('fetch failed')) return 'Action failed'
    if (m.length > 160) return `${m.slice(0, 160)}…`
    return m
  }
  return 'Action failed'
}

function jobKeyShort(key: string): string {
  if (!key) return ''
  if (key.includes('\0')) {
    const p = key.split('\0')
    if (p.length >= 3) return `${p[2]} @ ${p[1]}`
  }
  if (key.startsWith('http')) {
    try {
      return new URL(key).hostname.replace(/^www\./, '')
    } catch {
      return 'job link'
    }
  }
  return key.length > 48 ? `${key.slice(0, 48)}…` : key
}

function runLabel(kind: string, jobKey: string): string {
  const short = jobKeyShort(jobKey)
  switch (kind) {
    case 'tailor':
      return short ? `Tailor CV · ${short}` : 'Tailor CV'
    case 'match':
      return short ? `Match · ${short}` : 'Match'
    case 'scan':
      return 'Job scan'
    case 'build':
    case 'export':
      return short ? `Export application · ${short}` : 'Export application'
    default:
      return short ? `${kind} · ${short}` : kind
  }
}

function runToLine(run: RunRow): ActivityLine {
  const label = runLabel(run.kind, run.job_key)
  const at = Date.parse(run.updated_at || run.created_at || '') || Date.now()
  if (run.status === 'running') {
    return { id: `run-${run.id}`, at, level: 'run', text: `${label} running` }
  }
  if (run.status === 'done') {
    const detail = summaryFromResult(run.kind, run.result_json)
    return { id: `run-${run.id}`, at, level: 'ok', text: `✓ ${label}${detail ? ` — ${detail}` : ''}` }
  }
  const err = (run.error || 'failed').slice(0, 120)
  return { id: `run-${run.id}`, at, level: 'err', text: `✗ ${label}: ${err}` }
}

function mergeLines(server: ActivityLine[], session: ActivityLine[]): ActivityLine[] {
  const byId = new Map<string, ActivityLine>()
  for (const line of server) byId.set(line.id, line)
  for (const line of session) byId.set(line.id, line)
  return [...byId.values()].sort((a, b) => b.at - a.at).slice(0, MAX_LINES)
}

export function useActivity() {
  const open = useState('activity-open', () => true)
  const serverLines = useState<ActivityLine[]>('activity-server-lines', () => [])
  const sessionLines = useState<ActivityLine[]>('activity-session-lines', () => [])
  const historyLoaded = useState('activity-history-loaded', () => false)
  const activeRunId = useState<string | null>('activity-run-id', () => null)
  const activeRun = useState<ActiveRun | null>('activity-active-run', () => null)
  const tick = useState('activity-tick', () => 0)

  const lines = computed(() => mergeLines(serverLines.value, sessionLines.value))

  let tickTimer: ReturnType<typeof setInterval> | null = null

  function startTick() {
    if (tickTimer) return
    tickTimer = setInterval(() => {
      tick.value = Date.now()
      const run = activeRun.value
      if (!run || run.status !== 'running') return
      const elapsed = Date.now() - run.startedAt
      if (elapsed >= run.timeoutMs) return
      if (run.kind === 'tailor') {
        activeRun.value = {
          ...run,
          phase: tailorPhase(elapsed),
          progress: tailorProgress(elapsed, 'running'),
        }
      }
    }, 1000)
  }

  function stopTick() {
    if (tickTimer) {
      clearInterval(tickTimer)
      tickTimer = null
    }
  }

  function log(text: string, level: ActivityLevel = 'info') {
    open.value = true
    sessionLines.value = [
      { id: `sess-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`, at: Date.now(), level, text },
      ...sessionLines.value,
    ].slice(0, MAX_LINES)
  }

  async function refreshHistory() {
    try {
      const runs = await $fetch<RunRow[]>('/api/runs')
      serverLines.value = runs.map(runToLine)
      historyLoaded.value = true
    } catch {
      /* keep prior server lines */
    }
  }

  function clear() {
    sessionLines.value = []
    serverLines.value = []
    if (activeRun.value?.status !== 'running') {
      activeRun.value = null
      activeRunId.value = null
      stopTick()
    }
  }

  async function track<T>(label: string, fn: () => Promise<T>, opts?: { successDetail?: (result: T) => string | undefined }) {
    log(`${label}…`, 'run')
    try {
      const result = await fn()
      const detail = opts?.successDetail?.(result)
      log(`✓ ${label}${detail ? ` — ${detail}` : ''}`, 'ok')
      await refreshHistory()
      return result
    } catch (e) {
      log(`✗ ${label}: ${actionErrorMessage(e)}`, 'err')
      throw e
    }
  }

  function setActiveOutcome(
    status: ActiveRunStatus,
    label: string,
    opts: { error?: string; detail?: string; elapsedMs: number },
  ) {
    const base = activeRun.value
    if (!base) return
    activeRun.value = {
      ...base,
      status,
      progress: 100,
      phase:
        status === 'success'
          ? `Completed in ${formatDuration(opts.elapsedMs)}`
          : status === 'timeout'
            ? `Timed out after ${formatDuration(opts.elapsedMs)}`
            : `Failed after ${formatDuration(opts.elapsedMs)}`,
      error: opts.error,
      detail: opts.detail,
    }
    activeRunId.value = null
    stopTick()
    void refreshHistory()
  }

  async function pollRun(runId: string, label: string, kind = 'task', timeoutMs = DEFAULT_TIMEOUT_MS) {
    activeRunId.value = runId
    open.value = true
    const startedAt = Date.now()
    activeRun.value = {
      id: runId,
      label,
      kind,
      startedAt,
      status: 'running',
      phase: kind === 'tailor' ? tailorPhase(0) : msg('activity.phases.running'),
      progress: 4,
      timeoutMs,
    }
    startTick()
    log(`${label} started`, 'run')

    while (Date.now() - startedAt < timeoutMs) {
      await new Promise((r) => setTimeout(r, 1200))
      const elapsed = Date.now() - startedAt
      try {
        const run = await $fetch<RunRow>(`/api/runs/${runId}`)
        if (run.status === 'running') {
          if (activeRun.value?.id === runId) {
            activeRun.value = {
              ...activeRun.value,
              phase: kind === 'tailor' ? tailorPhase(elapsed) : msg('activity.phases.running'),
              progress: kind === 'tailor' ? tailorProgress(elapsed, 'running') : Math.min(90, Math.round(elapsed / 1000)),
            }
          }
          continue
        }
        if (run.status === 'done') {
          const detail = summaryFromResult(run.kind, run.result_json)
          setActiveOutcome('success', label, { detail, elapsedMs: elapsed })
          return run
        }
        setActiveOutcome('failed', label, { error: run.error || 'unknown error', elapsedMs: elapsed })
        throw new Error(run.error || 'run failed')
      } catch (e) {
        if (e instanceof Error && (e.message.includes('run failed') || e.message.includes('unknown error'))) {
          throw e
        }
      }
    }

    const elapsed = Date.now() - startedAt
    setActiveOutcome('timeout', label, {
      error: 'Client stopped waiting; server may still finish',
      elapsedMs: elapsed,
    })
    throw new Error('run timed out')
  }

  async function tailorJob(jobId: string, company?: string) {
    const label = company ? `Tailor CV · ${company}` : `Tailor CV · ${jobKeyShort(jobId)}`
    const run = await $fetch<RunRow>(`/api/tailor/${encodeURIComponent(jobId)}?async=1`, { method: 'POST' })
    await pollRun(run.id, label, 'tailor', TAILOR_TIMEOUT_MS)
    return run
  }

  const elapsedMs = computed(() => {
    const run = activeRun.value
    if (!run) return 0
    return Math.max(0, (tick.value || Date.now()) - run.startedAt)
  })

  return {
    open,
    lines,
    activeRunId,
    activeRun,
    elapsedMs,
    log,
    clear,
    track,
    refreshHistory,
    historyLoaded,
    pollRun,
    tailorJob,
    formatDuration,
    actionErrorMessage,
  }
}
