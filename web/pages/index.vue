<script setup lang="ts">
import type { InboxCardModel } from '~/components/InboxQueueCard.vue'

type InboxItem = {
  job_key: string
  title: string
  company: string
  state: string
  reason: string
  updated_at: string
}

type InboxResponse = {
  data: InboxItem[]
  meta: { counts: Record<string, number> }
}

type JobRow = {
  job: {
    id: string
    region: string
    risk?: number
    flags?: { severity: string }[]
    arrangement?: string
    remote?: string
    posted_at: string
  }
}

type ConfigView = {
  sources: { id: string; name: string; last_scan?: { error?: string } }[]
}

const GROUP_ORDER = ['needs-decision', 'in-progress', 'sent', 'interview'] as const

const { t } = useI18n()
const { offline, refresh: refreshStatus } = useAppStatus()
const { track, actionErrorMessage } = useActivity()

const loading = ref(true)
const error = ref<string | null>(null)
const inbox = ref<InboxResponse | null>(null)
const jobsByKey = ref<Map<string, JobRow['job']>>(new Map())
const config = ref<ConfigView | null>(null)
const failedSources = ref<{ name: string; error: string }[]>([])

const showFilters = ref(false)
const filterPreset = ref('')
const filterVerdict = ref('')
const filterRisk = ref('')
const filterArrangement = ref('')

const presets = ref<{ id: string; name: string }[]>([])
const selectedIndex = ref(0)

const flatItems = computed(() => {
  const cards: InboxCardModel[] = []
  for (const state of GROUP_ORDER) {
    for (const row of grouped.value[state] ?? []) {
      cards.push(row)
    }
  }
  return cards
})

function verdictForJob(job?: JobRow['job']): string {
  if (!job?.posted_at) return 'unknown'
  const d = new Date(job.posted_at)
  if (Number.isNaN(d.getTime())) return 'unknown'
  const days = (Date.now() - d.getTime()) / 86400000
  return days <= 14 ? 'fresh' : 'stale'
}

function enrich(item: InboxItem): InboxCardModel {
  const job = jobsByKey.value.get(item.job_key)
  return {
    ...item,
    region: job?.region,
    risk: job?.risk,
    riskHard: job?.flags?.some((f) => f.severity === 'hard'),
    verdict:
      verdictForJob(job) === 'fresh' ? t('inbox.verdict.fresh') : t('inbox.verdict.stale'),
  }
}

const grouped = computed(() => {
  const out: Record<string, InboxCardModel[]> = {
    'needs-decision': [],
    'in-progress': [],
    sent: [],
    interview: [],
  }
  for (const row of inbox.value?.data ?? []) {
    const card = enrich(row)
    const job = jobsByKey.value.get(row.job_key)
    if (filterVerdict.value === 'fresh' && verdictForJob(job) !== 'fresh') continue
    if (filterVerdict.value === 'stale' && verdictForJob(job) !== 'stale') continue
    if (filterRisk.value === 'low' && (job?.risk ?? 0) > 3) continue
    if (filterRisk.value === 'high' && (job?.risk ?? 0) <= 3) continue
    const arr = job?.arrangement || (job?.remote === 'yes' ? 'remote' : '')
    if (filterArrangement.value && filterArrangement.value !== arr) continue
    if (out[row.state]) out[row.state].push(card)
  }
  return out
})

const filtersActivePreset = ref('')

const activeFilterSummary = computed(() => {
  const parts: string[] = []
  if (filterPreset.value) {
    const name = presets.value.find((p) => p.id === filterPreset.value)?.name ?? filterPreset.value
    parts.push(`${t('inbox.filterLabels.preset')}: ${name}`)
  }
  if (filterVerdict.value) parts.push(`${t('inbox.filterLabels.verdict')}: ${filterVerdict.value}`)
  if (filterRisk.value) parts.push(`${t('inbox.filterLabels.risk')}: ${filterRisk.value}`)
  if (filterArrangement.value) {
    parts.push(`${t('inbox.filterLabels.arrangement')}: ${filterArrangement.value}`)
  }
  return parts[0] ?? ''
})

const hasActiveFilters = computed(
  () => !!(filterPreset.value || filterVerdict.value || filterRisk.value || filterArrangement.value),
)

watch(flatItems, (list) => {
  if (selectedIndex.value >= list.length) {
    selectedIndex.value = Math.max(0, list.length - 1)
  }
})

useInboxKeys(selectedIndex, computed(() => flatItems.value.length), {
  onOpen() {
    const item = flatItems.value[selectedIndex.value]
    if (item) navigateTo(jobDetailPath(item.job_key))
  },
  onSkip() {
    void skipSelected()
  },
  onJumpEnd(toEnd: boolean) {
    if (!flatItems.value.length) return
    selectedIndex.value = toEnd ? flatItems.value.length - 1 : 0
  },
})

async function loadInbox() {
  loading.value = true
  error.value = null
  try {
    const [inboxRes, jobsRes, filterRes, presetRes, cfg] = await Promise.all([
      $fetch<InboxResponse>('/api/inbox'),
      $fetch<JobRow[]>('/api/jobs'),
      $fetch<{ active_preset_id: string }>('/api/filters'),
      $fetch<{ id: string; name: string }[]>('/api/filters/presets'),
      $fetch<ConfigView>('/api/config'),
    ])
    inbox.value = inboxRes
    const map = new Map<string, JobRow['job']>()
    for (const row of jobsRes) map.set(row.job.id, row.job)
    jobsByKey.value = map
    filtersActivePreset.value = filterRes.active_preset_id || ''
    presets.value = presetRes
    config.value = cfg
    failedSources.value = (cfg.sources ?? [])
      .filter((s) => s.last_scan?.error)
      .map((s) => ({ name: s.name || s.id, error: s.last_scan!.error! }))
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('errors.loadJobs')
  } finally {
    loading.value = false
  }
}

async function skipSelected() {
  const item = flatItems.value[selectedIndex.value]
  if (!item) return
  const reason = window.prompt(t('inbox.skipPrompt'), '') ?? ''
  try {
    await track(t('actions.skipJob', { who: item.title }), async () => {
      await $fetch(`/api/applications/${encodeJobIdForApi(item.job_key)}/status`, {
        method: 'POST',
        body: { status: 'closed', follow_up_note: reason },
      }).catch(async () => {
        await $fetch(`/api/jobs/${encodeJobIdForApi(item.job_key)}`, {
          method: 'PUT',
          body: { status: 'discarded' },
        })
      })
    })
    await loadInbox()
  } catch (e) {
    error.value = actionErrorMessage(e)
  }
}

async function retryScan() {
  try {
    await track(t('common.scan'), () => $fetch('/api/scan', { method: 'POST' }))
    await loadInbox()
    await refreshStatus()
  } catch (e) {
    error.value = actionErrorMessage(e)
  }
}

onMounted(() => {
  void loadInbox()
})

function groupLabel(state: string) {
  const map: Record<string, string> = {
    'needs-decision': t('inbox.groups.needsDecision'),
    'in-progress': t('inbox.groups.inProgress'),
    sent: t('inbox.groups.sent'),
    interview: t('inbox.groups.interview'),
  }
  return map[state] ?? state
}
</script>

<template>
  <div class="page">
    <header class="bar">
      <h1>{{ t('inbox.title') }}</h1>
      <button type="button" class="ghost" @click="showFilters = !showFilters">
        {{ showFilters ? t('inbox.hideFilters') : t('inbox.showFilters') }}
      </button>
    </header>

    <div v-if="!showFilters && hasActiveFilters" class="filter-collapsed">
      <span class="dim">{{ t('inbox.activeFilter') }}:</span> {{ activeFilterSummary }}
    </div>

    <section v-if="showFilters" class="filters">
      <label>
        {{ t('inbox.filterLabels.preset') }}
        <select v-model="filterPreset">
          <option value="">{{ t('jobs.presetDefault') }}</option>
          <option v-for="p in presets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </label>
      <label>
        {{ t('inbox.filterLabels.verdict') }}
        <select v-model="filterVerdict">
          <option value="">—</option>
          <option value="fresh">{{ t('inbox.verdict.fresh') }}</option>
          <option value="stale">{{ t('inbox.verdict.stale') }}</option>
        </select>
      </label>
      <label>
        {{ t('inbox.filterLabels.risk') }}
        <select v-model="filterRisk">
          <option value="">—</option>
          <option value="low">≤ 3</option>
          <option value="high">&gt; 3</option>
        </select>
      </label>
      <label>
        {{ t('inbox.filterLabels.arrangement') }}
        <select v-model="filterArrangement">
          <option value="">—</option>
          <option value="remote">remote</option>
          <option value="hybrid">hybrid</option>
          <option value="onsite">onsite</option>
        </select>
      </label>
    </section>

    <div v-if="failedSources.length" class="partial">
      <p v-for="src in failedSources" :key="src.name">
        {{ t('inbox.partialSource', { name: src.name, error: src.error }) }}
        <button type="button" class="linkish" @click="retryScan">{{ t('inbox.sourceRetry') }}</button>
      </p>
    </div>

    <p v-if="error" class="error">{{ error }}</p>

    <div v-if="loading" class="skeleton-list" aria-busy="true">
      <div v-for="n in 4" :key="n" class="sk-card" />
    </div>

    <div v-else-if="offline" class="empty">
      <p>{{ t('status.offline') }}</p>
      <code>{{ t('status.offlineCmd') }}</code>
    </div>

    <div v-else-if="!flatItems.length" class="empty">
      <h2>{{ t('inbox.emptyTitle') }}</h2>
      <p>{{ t('inbox.emptyDaemon') }}</p>
      <p>{{ t('inbox.emptyScan') }}</p>
      <NuxtLink to="/jobs">{{ t('inbox.emptyScanLink') }}</NuxtLink>
    </div>

    <template v-else>
      <section v-for="state in GROUP_ORDER" :key="state" class="group">
        <h2 v-if="grouped[state]?.length" class="group-title">{{ groupLabel(state) }}</h2>
        <ul v-if="grouped[state]?.length" class="list">
          <li
            v-for="(item, idx) in grouped[state]"
            :key="item.job_key"
            :class="{ selected: flatItems[selectedIndex]?.job_key === item.job_key }"
          >
            <InboxQueueCard
              :item="item"
              :selected="flatItems[selectedIndex]?.job_key === item.job_key"
            />
          </li>
        </ul>
      </section>
    </template>
  </div>
</template>

<style scoped>
.page {
  max-width: 42rem;
  margin: 0 auto;
  padding: 1.25rem 1rem 3rem;
  color: var(--fg);
  animation: fade-in 120ms ease;
}
.bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}
h1 {
  margin: 0;
  font-size: 1.35rem;
  font-weight: 600;
}
.ghost {
  background: transparent;
  color: var(--muted);
  border: 1px solid var(--line);
  border-radius: var(--r-sm);
  padding: 0.4rem 0.75rem;
  cursor: pointer;
}
.filters {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding: 0.75rem;
  border: 1px solid var(--line);
  border-radius: var(--r);
  background: var(--surface);
  font-size: 0.85rem;
}
.filters label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  color: var(--muted);
}
.filters select {
  background: var(--surface2);
  color: var(--fg);
  border: 1px solid var(--line);
  border-radius: var(--r-sm);
  padding: 0.35rem;
}
.filter-collapsed {
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
  color: var(--fg);
}
.dim {
  color: var(--dim);
}
.partial {
  font-size: 0.85rem;
  color: var(--muted);
  margin-bottom: 0.75rem;
  padding: 0.65rem;
  border: 1px solid var(--line);
  border-radius: var(--r);
  background: var(--surface);
}
.linkish {
  background: none;
  border: none;
  color: var(--accent);
  cursor: pointer;
  text-decoration: underline;
  margin-left: 0.35rem;
}
.error {
  color: var(--del);
}
.empty {
  text-align: center;
  padding: 2rem 1rem;
  color: var(--muted);
}
.empty h2 {
  color: var(--fg);
  font-size: 1.1rem;
}
.empty code {
  display: block;
  margin-top: 0.75rem;
  color: var(--dim);
}
.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.sk-card {
  height: 7rem;
  border-radius: var(--r);
  background: var(--surface);
  border: 1px solid var(--line);
  animation: pulse 1.2s ease-in-out infinite;
}
@keyframes pulse {
  50% {
    opacity: 0.55;
  }
}
.group-title {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--muted);
  margin: 1.25rem 0 0.5rem;
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}
@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
