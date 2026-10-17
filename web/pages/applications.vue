<script setup lang="ts">
type ApplicationRow = {
  job_id: string
  company: string
  title: string
  url: string
  status: string
  follow_up_note: string
  channel: string
  referrer: string
}

type PipelineJob = {
  job_id: string
  company: string
  title: string
  url: string
  status: string
  job_feed_status: string
  has_proposal: boolean
}

type PipelineView = {
  matched: PipelineJob[]
  shortlisted: PipelineJob[]
  tailoring: PipelineJob[]
  ready: PipelineJob[]
  with_proposal: PipelineJob[]
}

type ExportResult = {
  dir: string
  cv_tex: string
  pdf_compiled: boolean
  compile_note?: string
}

type Tab = 'all' | 'matched' | 'shortlisted' | 'tailoring' | 'ready' | 'with_proposal'

const tab = ref<Tab>('all')
const rows = ref<ApplicationRow[]>([])
const pipeline = ref<PipelineView | null>(null)
const error = ref('')
const exporting = ref<string | null>(null)
const lastExport = ref<ExportResult | null>(null)
const { track, actionErrorMessage } = useActivity()

const tabs: { id: Tab; label: string; hint: string }[] = [
  { id: 'all', label: 'All', hint: 'Every role with an application record (not the whole Jobs feed)' },
  { id: 'matched', label: 'Matched', hint: 'Scored on Match (application status matched)' },
  { id: 'shortlisted', label: 'Shortlisted', hint: 'You clicked Shortlist on Match' },
  { id: 'tailoring', label: 'Tailoring', hint: 'Tailor in progress' },
  { id: 'ready', label: 'Ready', hint: 'Tailored / ready to apply' },
  { id: 'with_proposal', label: 'With CV diff', hint: 'Has a tailoring proposal on Diff' },
]

const activeHint = computed(() => tabs.find((t) => t.id === tab.value)?.hint ?? '')

const pipelineRows = computed((): PipelineJob[] => {
  if (!pipeline.value) return []
  switch (tab.value) {
    case 'matched':
      return pipeline.value.matched
    case 'shortlisted':
      return pipeline.value.shortlisted
    case 'tailoring':
      return pipeline.value.tailoring
    case 'ready':
      return pipeline.value.ready
    case 'with_proposal':
      return pipeline.value.with_proposal
    default:
      return []
  }
})

async function load() {
  error.value = ''
  try {
    rows.value = (await $fetch<ApplicationRow[]>('/api/applications')) ?? []
    pipeline.value = (await $fetch<PipelineView>('/api/pipeline')) ?? {
      matched: [],
      shortlisted: [],
      tailoring: [],
      ready: [],
      with_proposal: [],
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load applications'
  }
}

async function runExport(jobId: string) {
  exporting.value = jobId
  error.value = ''
  const row = rows.value.find((r) => r.job_id === jobId)
  const label = row ? `Export · ${row.title}` : 'Export application'
  try {
    lastExport.value = await track(label, () => $fetch<ExportResult>(`/api/export/${encodeURIComponent(jobId)}`, { method: 'POST' }), {
      successDetail: (r) => (r.pdf_compiled ? 'PDF ready' : 'LaTeX written'),
    })
    await load()
  } catch (e) {
    error.value = actionErrorMessage(e)
  } finally {
    exporting.value = null
  }
}

onMounted(load)
</script>

<template>
  <div class="page">
    <header class="bar">
      <h1>Applications</h1>
      <p class="lede">
        <strong>Skip</strong> closes the application and marks the job discarded.
        <strong>Shortlist</strong> keeps the role in your pipeline.
        <strong>Tailor</strong> runs CV tailoring and opens Diff when done — watch the Activity panel below.
      </p>
    </header>

    <nav class="tabs" aria-label="Pipeline filters">
      <button
        v-for="t in tabs"
        :key="t.id"
        type="button"
        class="tab"
        :class="{ active: tab === t.id }"
        @click="tab = t.id"
      >
        {{ t.label }}
        <span v-if="t.id !== 'all' && pipeline" class="count">
          {{
            t.id === 'matched'
              ? pipeline.matched.length
              : t.id === 'shortlisted'
              ? pipeline.shortlisted.length
              : t.id === 'tailoring'
                ? pipeline.tailoring.length
                : t.id === 'ready'
                  ? pipeline.ready.length
                  : t.id === 'with_proposal'
                    ? pipeline.with_proposal.length
                    : ''
          }}
        </span>
      </button>
    </nav>
    <p class="hint">{{ activeHint }}</p>

    <p v-if="error" class="err">{{ error }}</p>
    <p v-if="lastExport?.compile_note" class="note">{{ lastExport.compile_note }}</p>

    <ul v-if="tab === 'all'" class="list">
      <li v-for="row in rows" :key="row.job_id" class="item">
        <div class="main">
          <NuxtLink :to="jobRoute('diff', row.job_id)" class="title">{{ row.title }}</NuxtLink>
          <span class="muted">{{ row.company }}</span>
          <span class="chip">{{ row.status }}</span>
          <span v-if="row.channel" class="chip channel">{{ row.channel }}</span>
          <span v-if="row.referrer" class="muted">via {{ row.referrer }}</span>
        </div>
        <div class="actions">
          <NuxtLink :to="jobRoute('match', row.job_id)" class="btn ghost">Match</NuxtLink>
          <NuxtLink :to="jobRoute('diff', row.job_id)" class="btn ghost">Diff</NuxtLink>
          <button
            type="button"
            class="btn"
            :disabled="exporting === row.job_id"
            @click="runExport(row.job_id)"
          >
            {{ exporting === row.job_id ? 'Exporting…' : 'Export' }}
          </button>
        </div>
        <p v-if="row.follow_up_note" class="follow">{{ row.follow_up_note }}</p>
      </li>
    </ul>

    <ul v-else class="list">
      <li v-for="row in pipelineRows" :key="row.job_id" class="item">
        <div class="main">
          <NuxtLink :to="jobRoute('diff', row.job_id)" class="title">{{ row.title }}</NuxtLink>
          <span class="muted">{{ row.company }}</span>
          <span class="chip">{{ row.status }}</span>
          <span v-if="row.has_proposal" class="chip diff">diff</span>
        </div>
        <div class="actions">
          <NuxtLink :to="jobRoute('match', row.job_id)" class="btn ghost">Match</NuxtLink>
          <NuxtLink :to="jobRoute('diff', row.job_id)" class="btn ghost">Diff</NuxtLink>
        </div>
      </li>
    </ul>

    <p v-if="tab === 'all' && !rows.length && !error" class="muted empty">
      No applications yet — from Jobs open Match, then Shortlist or Tailor (or Export) to add a role here.
    </p>
    <p v-else-if="tab !== 'all' && !pipelineRows.length && !error" class="muted empty">Nothing in this stage yet.</p>
  </div>
</template>

<style scoped>
.page {
  color: var(--fg);
  background: var(--bg);
  min-height: 100%;
}
.bar {
  padding: 1.25rem 1.25rem 0.75rem;
  border-bottom: 1px solid var(--line);
  background: var(--surface);
}
.bar h1 {
  margin: 0;
  font-size: 1.35rem;
  font-weight: 600;
  letter-spacing: -0.02em;
}
.lede {
  margin: 0.5rem 0 0;
  font-size: 13px;
  color: var(--muted);
  line-height: 1.5;
  max-width: 42rem;
}
.tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 0.75rem 1.25rem 0;
}
.tab {
  border: 1px solid var(--line);
  background: var(--surface);
  color: var(--muted);
  border-radius: var(--r-sm, 7px);
  padding: 6px 12px;
  font-size: 12px;
  cursor: pointer;
}
.tab.active {
  color: var(--fg);
  border-color: var(--accent);
  background: var(--accent-dim);
}
.count {
  margin-left: 6px;
  opacity: 0.75;
}
.hint {
  margin: 0;
  padding: 0.35rem 1.25rem 0.5rem;
  font-size: 12px;
  color: var(--dim);
}
.list {
  list-style: none;
  margin: 0;
  padding: 1rem 1.25rem 2rem;
  max-width: 56rem;
}
.item {
  border: 1px solid var(--line, var(--line));
  border-radius: var(--r-sm, 7px);
  padding: 12px 14px;
  margin-bottom: 10px;
  background: var(--surface, var(--surface));
}
.main {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: baseline;
}
.title {
  font-weight: 600;
  color: var(--fg);
  text-decoration: none;
}
.title:hover {
  text-decoration: underline;
}
.muted {
  color: var(--muted);
  font-size: 13px;
}
.chip {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--surface2);
  color: var(--muted);
  text-transform: lowercase;
}
.chip.diff {
  background: var(--accent-dim);
  color: var(--accent);
}
.actions {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}
.btn {
  font-size: 12px;
  padding: 6px 12px;
  border-radius: var(--r-sm, 7px);
  border: 1px solid var(--line);
  background: var(--fg);
  color: var(--bg);
  cursor: pointer;
}
.btn.ghost {
  background: transparent;
  color: var(--fg);
  text-decoration: none;
  display: inline-flex;
  align-items: center;
}
.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.follow {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--muted);
}
.err {
  color: #c0392b;
  padding: 0 1.25rem;
}
.note {
  padding: 0 1.25rem;
  font-size: 12px;
  color: var(--muted);
}
.empty {
  padding: 1rem 1.25rem;
}
</style>
