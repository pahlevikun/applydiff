<script setup lang="ts">
type FilterResult = {
  keep: boolean
  score: number
  reasons: string[]
}

type JobView = {
  job: {
    id: string
    company: string
    title: string
    url: string
    source: string
    posted_at: string
    remote: string
    region: string
    status: string
    arrangement?: string
    filter_skip_reason?: string
    risk?: number
    flags?: { code: string; severity: string; message: string }[]
  }
  filter: FilterResult
  priority?: number
}

type FilterSpec = Record<string, unknown>

type FiltersPayload = {
  spec: FilterSpec
  active_preset_id: string
  active_preset_name?: string
}

type Preset = {
  id: string
  name: string
  spec: FilterSpec
}

type ScanResult = {
  fetched: number
  deduped: number
  dead: number
  stored: number
  skipped: number
}

const jobs = ref<JobView[]>([])
const filters = ref<FiltersPayload | null>(null)
const presets = ref<Preset[]>([])
const loading = ref(false)
const scanning = ref(false)
const error = ref<string | null>(null)
const { track, actionErrorMessage } = useActivity()
const lastScan = ref<ScanResult | null>(null)
const showFilters = ref(true)
const searchEl = ref<HTMLInputElement | null>(null)
const selectedJob = ref(0)

useSearchFocus(searchEl)
useListKeys(selectedJob, computed(() => jobs.value.length))

const filterCategories = [
  { key: 'arrangement', label: 'Work arrangement' },
  { key: 'weekly_days', label: 'Weekly schedule' },
  { key: 'async_friendly', label: 'Async / flexible' },
  { key: 'timezone_overlap_hours', label: 'Timezone overlap' },
  { key: 'regions', label: 'Region' },
  { key: 'countries', label: 'Country' },
  { key: 'employment_type', label: 'Employment type' },
  { key: 'seniority', label: 'Seniority' },
  { key: 'salary_min', label: 'Compensation' },
  { key: 'visa_sponsorship', label: 'Visa' },
  { key: 'eor_friendly', label: 'EOR / contractor' },
  { key: 'archetype', label: 'Role archetype' },
  { key: 'must_have_keywords', label: 'Must-have keywords' },
  { key: 'benefits', label: 'Benefits' },
]

async function loadFilters() {
  try {
    filters.value = await $fetch<FiltersPayload>('/api/filters')
    presets.value = await $fetch<Preset[]>('/api/filters/presets')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load filters'
  }
}

async function loadJobs() {
  loading.value = true
  error.value = null
  try {
    jobs.value = await $fetch<JobView[]>('/api/jobs')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load jobs'
  } finally {
    loading.value = false
  }
}

async function applyPreset(id: string) {
  error.value = null
  const name = presets.value.find((p) => p.id === id)?.name ?? id
  try {
    await track(`Apply filter preset · ${name}`, async () => {
      await $fetch(`/api/filters/presets/${id}/apply`, { method: 'POST' })
      await loadFilters()
      await loadJobs()
    })
  } catch (e) {
    error.value = actionErrorMessage(e)
  }
}

async function runScan() {
  scanning.value = true
  error.value = null
  try {
    lastScan.value = await track('Job scan', () => $fetch<ScanResult>('/api/scan', { method: 'POST' }), {
      successDetail: (s) => `fetched ${s.fetched}, stored ${s.stored}, skipped ${s.skipped}`,
    })
    await loadJobs()
  } catch (e) {
    error.value = actionErrorMessage(e)
  } finally {
    scanning.value = false
  }
}

function formatPosted(iso: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso.slice(0, 10)
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

function categorySummary(key: string): string {
  const spec = filters.value?.spec
  if (!spec || !spec[key]) return '—'
  const rule = spec[key] as { hard?: boolean; values?: string[]; require?: boolean; min?: number }
  if (rule.values?.length) {
    return `${rule.hard ? 'hard' : 'soft'} · ${rule.values.join(', ')}`
  }
  if (rule.require) return `${rule.hard ? 'hard' : 'soft'} · required`
  if (rule.min) return `min ${rule.min}`
  return 'configured'
}

onMounted(async () => {
  await loadFilters()
  await loadJobs()
})
</script>

<template>
  <main class="page">
    <header class="bar">
      <h1>Jobs</h1>
      <input
        ref="searchEl"
        class="search"
        type="search"
        placeholder="Search (/)"
        aria-label="Search jobs"
      />
      <div class="actions">
        <button class="ghost" type="button" @click="showFilters = !showFilters">
          {{ showFilters ? 'Hide filters' : 'Filters' }}
        </button>
        <button class="scan" type="button" :disabled="scanning" @click="runScan">
          {{ scanning ? 'Scanning…' : 'Scan' }}
        </button>
      </div>
    </header>

    <aside v-if="showFilters && filters" class="filter-panel">
      <div class="panel-head">
        <span class="panel-title">Filter preset</span>
        <select
          class="preset-select"
          :value="filters.active_preset_id || ''"
          @change="applyPreset(($event.target as HTMLSelectElement).value)"
        >
          <option value="">Default (config.yaml)</option>
          <option v-for="p in presets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </div>
      <p v-if="filters.active_preset_name" class="active-name">Active: {{ filters.active_preset_name }}</p>
      <ul class="cat-list">
        <li v-for="cat in filterCategories" :key="cat.key">
          <span class="cat-label">{{ cat.label }}</span>
          <span class="cat-val">{{ categorySummary(cat.key) }}</span>
        </li>
      </ul>
      <NuxtLink to="/sources" class="edit-link">Edit full filter set →</NuxtLink>
    </aside>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="lastScan" class="meta">
      Last scan: fetched {{ lastScan.fetched }}, stored {{ lastScan.stored }}, skipped {{ lastScan.skipped }}, dead
      {{ lastScan.dead }}
    </p>
    <p v-if="loading && !jobs.length" class="meta">Loading…</p>

    <ul v-if="jobs.length" class="feed">
      <li v-for="row in jobs" :key="row.job.id" class="row" :class="{ skipped: !row.filter.keep }">
        <div class="main">
          <div class="title-row">
            <a class="title" :href="row.job.url" target="_blank" rel="noopener noreferrer">{{ row.job.title }}</a>
            <span v-if="row.filter.score > 0" class="score">+{{ row.filter.score }}</span>
            <NuxtLink class="job-action" :to="jobRoute('match', row.job.id)">Match</NuxtLink>
            <NuxtLink class="job-action" :to="jobRoute('diff', row.job.id)">Diff</NuxtLink>
          </div>
          <div class="sub">
            <span>{{ row.job.company }}</span>
            <span class="dot">·</span>
            <span class="source">{{ row.job.source }}</span>
            <span v-if="row.job.arrangement === 'remote' || row.job.remote === 'yes'" class="badge remote">Remote</span>
            <span v-if="row.job.risk" :class="['badge', row.job.flags?.some((f) => f.severity === 'hard') ? 'risk-hard' : 'risk-soft']">
              Risk {{ row.job.risk }}
            </span>
            <span v-if="row.job.status === 'skipped'" class="badge skip">Skipped</span>
            <span class="posted">{{ formatPosted(row.job.posted_at) }}</span>
          </div>
          <p v-if="row.filter.reasons.length" class="reasons">
            {{ row.filter.keep ? row.filter.reasons[0] : row.filter.reasons.find((r) => r.startsWith('skipped')) || row.filter.reasons[0] }}
          </p>
        </div>
      </li>
    </ul>
    <p v-else-if="!loading" class="meta">No jobs yet. Run a scan to fetch listings.</p>
  </main>
</template>

<style scoped>
.page {
  max-width: 56rem;
  margin: 0 auto;
  padding: 1.5rem 1rem 3rem;
  color: var(--fg);
  background: var(--bg);
  min-height: 100vh;
}
.bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}
.actions {
  display: flex;
  gap: 0.5rem;
}
h1 {
  font-size: 1.35rem;
  margin: 0;
  font-weight: 600;
}
.scan {
  background: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: 7px;
  padding: 0.5rem 1rem;
  font-weight: 600;
  cursor: pointer;
}
.ghost {
  background: transparent;
  color: var(--muted);
  border: 1px solid var(--line);
  border-radius: 7px;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
}
.scan:disabled {
  opacity: 0.6;
  cursor: wait;
}
.filter-panel {
  border: 1px solid var(--line);
  border-radius: 10px;
  background: var(--surface);
  padding: 1rem;
  margin-bottom: 1rem;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.panel-title {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--muted);
}
.preset-select {
  background: var(--surface2);
  color: var(--fg);
  border: 1px solid var(--line);
  border-radius: 7px;
  padding: 0.35rem 0.5rem;
  font-size: 0.85rem;
}
.active-name {
  font-size: 0.8rem;
  color: var(--accent);
  margin: 0.5rem 0 0;
}
.cat-list {
  list-style: none;
  margin: 0.75rem 0 0;
  padding: 0;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.35rem 1rem;
  font-size: 0.78rem;
}
.cat-label {
  color: var(--muted);
}
.cat-val {
  display: block;
  color: var(--fg);
  font-family: ui-monospace, monospace;
  font-size: 0.72rem;
}
.edit-link {
  display: inline-block;
  margin-top: 0.75rem;
  font-size: 0.8rem;
  color: var(--accent);
  text-decoration: none;
}
.error {
  color: var(--del);
}
.meta {
  color: var(--muted);
  font-size: 0.9rem;
}
.feed {
  list-style: none;
  margin: 1rem 0 0;
  padding: 0;
  border: 1px solid var(--line);
  border-radius: 10px;
  overflow: hidden;
}
.row {
  border-bottom: 1px solid var(--line);
  padding: 0.85rem 1rem;
  background: var(--surface);
}
.row.skipped {
  opacity: 0.72;
  background: var(--bg);
}
.row:last-child {
  border-bottom: none;
}
.title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.title {
  color: var(--fg);
  font-weight: 600;
  text-decoration: none;
}
.score {
  font-size: 0.7rem;
  color: var(--amber);
  font-variant-numeric: tabular-nums;
}
.job-action {
  font-size: 0.75rem;
  color: var(--accent);
  text-decoration: none;
  border: 1px solid var(--line);
  padding: 0.15rem 0.5rem;
  border-radius: 6px;
}
.job-action:hover {
  background: var(--surface2);
}
.title:hover {
  color: var(--accent);
}
.sub {
  margin-top: 0.35rem;
  font-size: 0.85rem;
  color: var(--muted);
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
}
.reasons {
  margin: 0.4rem 0 0;
  font-size: 0.75rem;
  color: var(--dim);
  font-family: ui-monospace, monospace;
}
.dot {
  opacity: 0.5;
}
.source {
  text-transform: capitalize;
}
.badge.remote {
  background: var(--add-bg);
  color: var(--add);
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
}
.badge.skip {
  background: var(--amber-bg);
  color: var(--amber);
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
}
.badge.risk-soft {
  background: var(--amber-bg);
  color: var(--amber);
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
}
.badge.risk-hard {
  background: var(--del-bg);
  color: var(--del);
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
}
.posted {
  margin-left: auto;
  font-variant-numeric: tabular-nums;
}
</style>
