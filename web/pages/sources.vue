<script setup lang="ts">
type SourceScan = {
  at: string
  fetched: number
  stored?: number
  skipped?: number
  error?: string
}

type SourceView = {
  id: string
  name: string
  hint: string
  enabled: boolean
  last_scan?: SourceScan
}

type FilterRule = {
  hard?: boolean
  values?: string[]
  require?: boolean
  min?: number
  max?: number
  boost?: number
  currency?: string
  all?: boolean
}

type FilterSpec = Record<string, FilterRule | boolean>

type ConfigView = {
  sources: SourceView[]
  queries: string[]
  filters: FilterSpec
  llm: { mode: string; model: string }
}

const filterSections: { title: string; keys: string[] }[] = [
  { title: 'Work arrangement', keys: ['arrangement', 'weekly_days', 'async_friendly', 'timezone_overlap_hours'] },
  { title: 'Region & geography', keys: ['regions', 'countries', 'exclude_regions'] },
  { title: 'Employment & seniority', keys: ['employment_type', 'exclude_employment', 'seniority', 'min_seniority'] },
  { title: 'Compensation', keys: ['salary_min', 'salary_disclosed', 'equity'] },
  { title: 'Company', keys: ['company_stage', 'company_size'] },
  { title: 'Visa & logistics', keys: ['visa_sponsorship', 'relocation', 'eor_friendly'] },
  { title: 'Role & stack', keys: ['archetype', 'must_have_keywords', 'nice_to_have_keywords', 'exclude_keywords'] },
  { title: 'Freshness', keys: ['max_posted_days'] },
  { title: 'Benefits', keys: ['benefits'] },
]

function ruleLabel(key: string, rule: FilterRule | boolean | undefined): string {
  if (rule === undefined) return '—'
  if (typeof rule === 'boolean') return rule ? 'on' : 'off'
  if (rule.values?.length) {
    return `${rule.hard ? 'hard' : 'soft'} · ${rule.values.join(', ')}`
  }
  if (rule.require) return `${rule.hard ? 'hard' : 'soft'} · required (boost ${rule.boost ?? 5})`
  if (rule.min) return `${rule.hard ? 'hard' : 'soft'} · min ${rule.min}${rule.currency ? ' ' + rule.currency : ''}`
  if (rule.max) return `${rule.hard ? 'hard' : 'soft'} · max ${rule.max} days`
  return 'configured'
}

const config = ref<ConfigView | null>(null)
const error = ref('')
const toggling = ref<string | null>(null)

const defaultQueries = [
  'senior backend engineer remote',
  'senior software engineer remote',
  'tech lead remote',
  'platform engineer remote',
  'staff engineer remote',
  'senior mobile engineer remote',
]

async function load() {
  error.value = ''
  try {
    config.value = await $fetch<ConfigView>('/api/config')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load config'
  }
}

async function toggleSource(id: string, enabled: boolean) {
  toggling.value = id
  error.value = ''
  try {
    config.value = await $fetch<ConfigView>('/api/config', {
      method: 'PUT',
      body: { source_id: id, enabled },
    })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to update source'
  } finally {
    toggling.value = null
  }
}

const queries = computed(() => {
  const q = config.value?.queries?.filter(Boolean) ?? []
  return q.length ? q : defaultQueries
})

onMounted(load)
</script>

<template>
  <div class="wrap">
    <header class="topbar">
      <NuxtLink to="/jobs" class="brand"><span class="dot" /> ApplyDiff</NuxtLink>
      <nav class="nav">
        <NuxtLink to="/jobs">Jobs</NuxtLink>
        <NuxtLink to="/applications">Applications</NuxtLink>
        <NuxtLink to="/cv">My CV</NuxtLink>
        <NuxtLink to="/sources" class="on">Sources</NuxtLink>
      </nav>
      <div class="grow" />
      <NuxtLink to="/jobs" class="btn ghost">Run scan now</NuxtLink>
    </header>

    <p v-if="error" class="err">{{ error }}</p>

    <div v-if="config" class="cfg">
      <aside class="side">
        <div class="lbl">Settings</div>
        <div class="f on">Sources</div>
        <div class="f">Queries</div>
        <div class="f">Filters</div>
        <div class="f">LLM runner</div>
        <div class="hr" />
        <div class="lbl">Scan policy</div>
        <div class="flow"><span class="dot-sm" /> no-token scan first</div>
        <div class="flow"><span class="dot-sm" /> drop dead postings</div>
        <div class="flow"><span class="dot-sm" /> dedupe URL + id</div>
      </aside>

      <section class="cfglo">
        <div class="lbl">Sources · enabled</div>
        <div class="grid">
          <div v-for="src in config.sources" :key="src.id" class="card">
            <div class="h">
              <span class="n">{{ src.name }}</span>
              <span class="id mono">{{ src.id }}</span>
              <button
                type="button"
                class="sw"
                :class="{ off: !src.enabled }"
                :disabled="toggling === src.id"
                aria-label="Toggle source"
                @click="toggleSource(src.id, !src.enabled)"
              />
            </div>
            <div class="m">
              <span>{{ src.hint }}</span>
              <span v-if="src.last_scan" class="scan">
                · last scan
                <template v-if="src.last_scan.error">error</template>
                <template v-else>fetched {{ src.last_scan.fetched }}</template>
              </span>
            </div>
          </div>
        </div>
      </section>

      <section class="cfglo narrow">
        <div class="lbl">Shared queries · remote</div>
        <div v-for="q in queries" :key="q" class="qrow">
          <span class="mono">{{ q }}</span>
        </div>
        <div class="hr" />
        <div class="lbl">Filter taxonomy</div>
        <template v-for="section in filterSections" :key="section.title">
          <div class="lbl sub">{{ section.title }}</div>
          <div v-for="key in section.keys" :key="key" class="flow">
            <span class="mono">{{ key }}</span>
            <span class="grow" />
            <span class="chip">{{ ruleLabel(key, config.filters[key]) }}</span>
          </div>
        </template>
        <div class="flow">
          remote_only legacy <span class="grow" /><span class="chip acc">{{ config.filters.remote_only === true ? 'on' : 'off' }}</span>
        </div>
        <div class="hr" />
        <div class="lbl">LLM runner</div>
        <div class="flow"><span class="dot-sm" /> {{ config.llm.mode === 'cli' ? 'CLI agent (Claude/Codex)' : 'direct API fallback' }}</div>
        <div class="flow muted">model · {{ config.llm.model }}</div>
        <div class="flow muted">cheapest model for scan/eval</div>
      </section>
    </div>
    <p v-else-if="!error" class="muted load">Loading…</p>
  </div>
</template>

<style scoped>
.wrap {
  min-height: 100vh;
  background: var(--bg, var(--bg));
  color: var(--fg, var(--fg));
}
.topbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--line, var(--line));
}
.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--accent, var(--accent));
  text-decoration: none;
  font-weight: 600;
}
.dot,
.dot-sm {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent, var(--accent));
  flex-shrink: 0;
}
.nav {
  display: flex;
  gap: 14px;
}
.nav a {
  color: var(--muted, var(--muted));
  text-decoration: none;
  font-size: 13px;
}
.nav a.on {
  color: var(--fg, var(--fg));
  font-weight: 600;
}
.grow {
  flex: 1;
}
.btn.ghost {
  font-size: 12px;
  padding: 6px 12px;
  border: 1px solid var(--line2, var(--line2));
  border-radius: var(--r-sm, 7px);
  color: var(--fg, var(--fg));
  text-decoration: none;
}
.cfg {
  display: flex;
  min-height: calc(100vh - 52px);
}
.side {
  width: 250px;
  flex: 0 0 auto;
  border-right: 1px solid var(--line, var(--line));
  background: var(--surface, var(--surface));
  padding: 12px;
}
.lbl {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--dim, var(--dim));
  margin: 10px 0 8px;
}
.lbl.sub {
  margin-top: 14px;
  color: var(--muted, var(--muted));
}
.chip {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--surface2, var(--surface2));
  color: var(--muted, var(--muted));
  max-width: 55%;
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.f {
  padding: 6px 8px;
  font-size: 13px;
  color: var(--muted, var(--muted));
  border-radius: var(--r-sm, 7px);
}
.f.on {
  background: var(--surface2, var(--surface2));
  color: var(--fg, var(--fg));
}
.hr {
  height: 1px;
  background: var(--line, var(--line));
  margin: 12px 0;
}
.flow {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--muted, var(--muted));
  padding: 4px 0;
}
.cfglo {
  flex: 1;
  min-width: 0;
  border-right: 1px solid var(--line, var(--line));
  padding: 16px;
  overflow: auto;
}
.cfglo.narrow {
  width: 340px;
  flex: 0 0 auto;
  border-right: none;
}
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 11px;
}
.card {
  border: 1px solid var(--line, var(--line));
  border-radius: var(--r, 10px);
  background: var(--surface, var(--surface));
  padding: 13px;
}
.card .h {
  display: flex;
  align-items: center;
  gap: 9px;
  margin-bottom: 6px;
}
.card .n {
  font-weight: 600;
  font-size: 13px;
}
.card .id {
  font-size: 10.5px;
  color: var(--dim, var(--dim));
}
.card .m {
  font-size: 11.5px;
  color: var(--muted, var(--muted));
}
.card .scan {
  display: block;
  margin-top: 4px;
  font-size: 10.5px;
  color: var(--dim, var(--dim));
}
.mono {
  font-family: ui-monospace, monospace;
}
.sw {
  width: 34px;
  height: 19px;
  border-radius: 999px;
  background: var(--accent, var(--accent));
  position: relative;
  margin-left: auto;
  flex: 0 0 auto;
  border: none;
  cursor: pointer;
  padding: 0;
}
.sw::after {
  content: '';
  position: absolute;
  top: 2.5px;
  right: 2.5px;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--accent-dim);
}
.sw.off {
  background: var(--line2, var(--line2));
}
.sw.off::after {
  left: 2.5px;
  right: auto;
  background: var(--dim, var(--dim));
}
.qrow {
  padding: 7px 0;
  border-bottom: 1px dashed var(--line, var(--line));
  font-size: 12.5px;
  color: var(--muted, var(--muted));
}
.qrow .mono {
  color: var(--fg, var(--fg));
  font-size: 12px;
}
.chip.acc {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--accent-dim, var(--accent-dim));
  color: var(--accent, var(--accent));
}
.err {
  color: var(--del, var(--del));
  padding: 8px 16px;
}
.muted {
  color: var(--muted, var(--muted));
}
.load {
  padding: 24px;
}
</style>
