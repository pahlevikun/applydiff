<script setup lang="ts">
import { optionsForFilter, keywordOptions } from '~/utils/filterOptions'
import { filterTitle, filterHint } from '~/utils/filterMeta'

const { t } = useI18n()
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
  config_path?: string
}

type FilterKind = 'remote' | 'bool' | 'int' | 'keyword' | 'salary' | 'category'

const BOOL_KEYS = new Set([
  'async_friendly',
  'salary_disclosed',
  'equity',
  'visa_sponsorship',
  'relocation',
  'eor_friendly',
])
const INT_KEYS = new Set(['max_posted_days', 'timezone_overlap_hours', 'days_onsite', 'posted_within', 'max_applicants'])
const KEYWORD_KEYS = new Set(['must_have_keywords', 'nice_to_have_keywords', 'exclude_keywords'])
const SALARY_KEYS = new Set(['salary_min'])

function filterKind(key: string): FilterKind {
  if (key === 'remote_only') return 'remote'
  if (BOOL_KEYS.has(key)) return 'bool'
  if (INT_KEYS.has(key)) return 'int'
  if (KEYWORD_KEYS.has(key)) return 'keyword'
  if (SALARY_KEYS.has(key)) return 'salary'
  return 'category'
}

const filterSections: { titleKey: string; keys: string[] }[] = [
  { titleKey: 'filters.groups.work', keys: ['arrangement', 'weekly_days', 'async_friendly', 'timezone_overlap_hours'] },
  { titleKey: 'filters.groups.region', keys: ['regions', 'countries', 'exclude_regions'] },
  { titleKey: 'filters.groups.employment', keys: ['employment_type', 'exclude_employment', 'seniority', 'min_seniority'] },
  { titleKey: 'filters.groups.compensation', keys: ['salary_min', 'salary_disclosed', 'equity'] },
  { titleKey: 'filters.groups.company', keys: ['company_stage', 'company_size'] },
  { titleKey: 'filters.groups.visa', keys: ['visa_sponsorship', 'relocation', 'eor_friendly'] },
  { titleKey: 'filters.groups.role', keys: ['archetype', 'must_have_keywords', 'nice_to_have_keywords', 'exclude_keywords'] },
  { titleKey: 'filters.groups.freshness', keys: ['max_posted_days'] },
  { titleKey: 'filters.groups.benefits', keys: ['benefits'] },
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
const notice = ref('')
const toggling = ref<string | null>(null)
const saving = ref<'queries' | 'filters' | null>(null)
const queryDraft = ref<string[]>([])
const filterDraft = ref<FilterSpec>({})
const customKw = ref<Record<string, string>>({})
const { track, actionErrorMessage } = useActivity()

function syncDrafts() {
  if (!config.value) return
  queryDraft.value = [...(config.value.queries ?? [])]
  filterDraft.value = JSON.parse(JSON.stringify(config.value.filters ?? {})) as FilterSpec
}

async function load() {
  error.value = ''
  notice.value = ''
  try {
    config.value = await $fetch<ConfigView>('/api/config')
    syncDrafts()
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('errors.loadConfig')
  }
}

async function toggleSource(id: string, enabled: boolean) {
  toggling.value = id
  error.value = ''
  const name = config.value?.sources.find((s) => s.id === id)?.name ?? id
  try {
    config.value = await track(t(enabled ? 'actions.enableSource' : 'actions.disableSource', { name }), () =>
      $fetch<ConfigView>('/api/config', {
        method: 'PUT',
        body: { source_id: id, enabled },
      }),
    )
  } catch (e) {
    error.value = actionErrorMessage(e)
  } finally {
    toggling.value = null
  }
}

function addQueryRow() {
  queryDraft.value.push('')
}

function removeQueryRow(i: number) {
  queryDraft.value.splice(i, 1)
}

async function saveQueries() {
  saving.value = 'queries'
  error.value = ''
  notice.value = ''
  const queries = queryDraft.value.map((q) => q.trim()).filter(Boolean)
  try {
    config.value = await track(t('actions.saveScanQueries'), () =>
      $fetch<ConfigView>('/api/config', {
        method: 'PUT',
        body: { queries },
      }),
      {
        successDetail: () =>
          t('actions.saveScanQueriesDetail', {
            count: queries.length,
            suffix: queries.length === 1 ? 'y' : 'ies',
          }),
      },
    )
    syncDrafts()
    notice.value = t('sources.queriesSaved')
  } catch (e) {
    error.value = actionErrorMessage(e)
  } finally {
    saving.value = null
  }
}

function isFilterActive(key: string): boolean {
  if (key === 'remote_only') return filterDraft.value.remote_only === true
  const v = filterDraft.value[key]
  return v !== undefined && v !== null
}

function enableFilter(key: string) {
  const next = { ...filterDraft.value }
  if (key === 'remote_only') {
    next.remote_only = true
  } else {
    const kind = filterKind(key)
    if (kind === 'bool') next[key] = { hard: false, require: true, boost: 5 }
    else if (kind === 'int') next[key] = { hard: true, max: 14 }
    else if (kind === 'keyword') next[key] = { hard: false, values: [], all: false, boost: 3 }
    else if (kind === 'salary') next[key] = { hard: false, min: 0, currency: 'USD' }
    else next[key] = { hard: true, values: [] }
  }
  filterDraft.value = next
}

function disableFilter(key: string) {
  const next = { ...filterDraft.value }
  if (key === 'remote_only') next.remote_only = false
  else delete next[key]
  filterDraft.value = next
}

function setRemoteOnly(on: boolean) {
  filterDraft.value = { ...filterDraft.value, remote_only: on }
}

function ruleDraft(key: string): FilterRule {
  const v = filterDraft.value[key]
  if (typeof v === 'object' && v) return v
  return {}
}

function filterValueOptions(key: string): string[] {
  const selected = ruleDraft(key).values ?? []
  if (filterKind(key) === 'keyword') return keywordOptions(key, selected)
  return optionsForFilter(key, selected)
}

function addCustomKeyword(key: string) {
  const word = (customKw.value[key] ?? '').trim()
  if (!word) return
  toggleFilterValue(key, word, true)
  customKw.value = { ...customKw.value, [key]: '' }
}

function isValueChecked(key: string, value: string): boolean {
  return (ruleDraft(key).values ?? []).includes(value)
}

function toggleFilterValue(key: string, value: string, on: boolean) {
  const set = new Set(ruleDraft(key).values ?? [])
  if (on) set.add(value)
  else set.delete(value)
  patchRule(key, { values: [...set] })
}

function patchRule(key: string, patch: Partial<FilterRule>) {
  filterDraft.value = { ...filterDraft.value, [key]: { ...ruleDraft(key), ...patch } }
}

async function saveFilters() {
  saving.value = 'filters'
  error.value = ''
  notice.value = ''
  try {
    config.value = await track(t('actions.saveFilters'), () =>
      $fetch<ConfigView>('/api/config', {
        method: 'PUT',
        body: { filters: filterDraft.value },
      }),
    )
    syncDrafts()
    notice.value = t('sources.filtersSaved')
  } catch (e) {
    error.value = actionErrorMessage(e)
  } finally {
    saving.value = null
  }
}

const configPathLabel = computed(() => config.value?.config_path ?? 'config.yaml')

type SettingsSection = 'sources' | 'queries' | 'filters' | 'llm'

const navItems: { id: SettingsSection; labelKey: string }[] = [
  { id: 'sources', labelKey: 'sources.sections.sources' },
  { id: 'queries', labelKey: 'sources.sections.queries' },
  { id: 'filters', labelKey: 'sources.sections.filters' },
  { id: 'llm', labelKey: 'sources.sections.llm' },
]

const section = ref<SettingsSection>('sources')

function selectSection(id: SettingsSection) {
  section.value = id
}

onMounted(load)
</script>

<template>
  <div class="page">
    <header class="bar">
      <div class="head">
        <h1>{{ t('sources.title') }}</h1>
        <p class="lede">
          {{ t('sources.lede', { path: configPathLabel }) }}
        </p>
      </div>
      <NuxtLink to="/jobs" class="btn ghost">{{ t('sources.runScan') }}</NuxtLink>
    </header>

    <p v-if="error" class="err">{{ error }}</p>
    <p v-if="notice" class="notice">{{ notice }}</p>

    <div v-if="config" class="cfg">
      <aside class="side">
        <nav class="side-nav" aria-label="Settings sections">
          <div class="lbl">{{ t('sources.onThisPage') }}</div>
          <button
            v-for="item in navItems"
            :key="item.id"
            type="button"
            class="nav-item"
            :class="{ on: section === item.id }"
            :aria-current="section === item.id ? 'page' : undefined"
            @click="selectSection(item.id)"
          >
            {{ t(item.labelKey) }}
          </button>
        </nav>
        <div class="hr" />
        <div class="lbl">{{ t('sources.scanPolicy') }}</div>
        <div class="flow"><span class="dot-sm" /> {{ t('sources.policyNoToken') }}</div>
        <div class="flow"><span class="dot-sm" /> {{ t('sources.policyDropDead') }}</div>
        <div class="flow"><span class="dot-sm" /> {{ t('sources.policyDedupe') }}</div>
      </aside>

      <main class="cfg-main">
        <section v-show="section === 'sources'" class="panel" aria-labelledby="panel-sources">
          <h2 id="panel-sources" class="panel-title">{{ t('sources.sourcesPanel') }}</h2>
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
                  :aria-label="`${src.enabled ? 'Disable' : 'Enable'} ${src.name}`"
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

        <section v-show="section === 'queries'" class="panel" aria-labelledby="panel-queries">
          <div class="panel-head">
            <h2 id="panel-queries" class="panel-title">{{ t('sources.queriesPanel') }}</h2>
            <button type="button" class="btn primary" :disabled="saving === 'queries'" @click="saveQueries">
              {{ saving === 'queries' ? t('common.saving') : t('sources.saveQueries') }}
            </button>
          </div>
          <p class="panel-hint">{{ t('sources.queriesHint') }}</p>
          <div v-for="(q, i) in queryDraft" :key="i" class="q-edit">
            <input v-model="queryDraft[i]" type="text" class="inp" :aria-label="`Query ${i + 1}`" />
            <button type="button" class="btn ghost sm" @click="removeQueryRow(i)">Remove</button>
          </div>
          <button type="button" class="btn ghost" @click="addQueryRow">{{ t('sources.addQuery') }}</button>
        </section>

        <section v-show="section === 'filters'" class="panel" aria-labelledby="panel-filters">
          <div class="panel-head">
            <h2 id="panel-filters" class="panel-title">{{ t('sources.filtersPanel') }}</h2>
            <button type="button" class="btn primary" :disabled="saving === 'filters'" @click="saveFilters">
              {{ saving === 'filters' ? t('common.saving') : t('sources.saveFilters') }}
            </button>
          </div>
          <div class="filter-help" role="note">
            <p>{{ t('filters.help.hard') }}</p>
            <p>{{ t('filters.help.soft') }}</p>
            <p class="dim">{{ t('filters.help.andOr') }}</p>
          </div>

          <div class="filter-card">
            <div class="filter-card-head">
              <div>
                <h3 class="filter-name">{{ filterTitle('remote_only') }}</h3>
                <p class="filter-sub">{{ filterHint('remote_only') }}</p>
              </div>
              <FilterChip
                :label="t('common.enabled')"
                :checked="filterDraft.remote_only === true"
                @toggle="setRemoteOnly(filterDraft.remote_only !== true)"
              />
            </div>
          </div>

          <template v-for="group in filterSections" :key="group.title">
            <h3 class="group-title">{{ t(group.titleKey) }}</h3>
            <div v-for="key in group.keys" :key="key" class="filter-card" :class="{ on: isFilterActive(key) }">
              <div class="filter-card-head">
                <div>
                  <h3 class="filter-name">{{ filterTitle(key) }}</h3>
                  <p v-if="filterHint(key)" class="filter-sub">{{ filterHint(key) }}</p>
                  <p v-if="!isFilterActive(key)" class="filter-id mono">{{ key }}</p>
                </div>
                <FilterChip
                  :label="isFilterActive(key) ? t('common.on') : t('common.off')"
                  :checked="isFilterActive(key)"
                  @toggle="isFilterActive(key) ? disableFilter(key) : enableFilter(key)"
                />
              </div>
              <p v-if="!isFilterActive(key) && config" class="filter-summary">{{ ruleLabel(key, config.filters[key]) }}</p>

              <div v-if="isFilterActive(key) && (filterKind(key) === 'category' || filterKind(key) === 'keyword')" class="filter-card-body">
                <div class="chip-grid" role="group" :aria-label="`${filterTitle(key)} values`">
                  <FilterChip
                    v-for="opt in filterValueOptions(key)"
                    :key="opt"
                    :label="opt"
                    :checked="isValueChecked(key, opt)"
                    @toggle="toggleFilterValue(key, opt, !isValueChecked(key, opt))"
                  />
                </div>
                <div v-if="filterKind(key) === 'keyword'" class="custom-kw">
                  <input
                    v-model="customKw[key]"
                    type="text"
                    class="inp"
                    :placeholder="t('common.customKeyword')"
                    :aria-label="`Add keyword for ${key}`"
                    @keydown.enter.prevent="addCustomKeyword(key)"
                  />
                  <button type="button" class="btn ghost sm" @click="addCustomKeyword(key)">{{ t('common.addKeyword') }}</button>
                </div>
                <label v-if="filterKind(key) === 'keyword'" class="chk inline">
                  <input type="checkbox" :checked="!!ruleDraft(key).all" @change="patchRule(key, { all: ($event.target as HTMLInputElement).checked })" />
                  {{ t('common.matchAllKeywords') }}
                </label>
                <FilterRuleMode
                  :hard="!!ruleDraft(key).hard"
                  :boost="ruleDraft(key).boost"
                  show-boost
                  @update:hard="patchRule(key, { hard: $event })"
                  @update:boost="patchRule(key, { boost: $event })"
                />
              </div>

              <div v-else-if="isFilterActive(key) && filterKind(key) === 'bool'" class="filter-card-body">
                <FilterChip
                  :label="t('common.requireOnListing')"
                  :checked="!!ruleDraft(key).require"
                  @toggle="patchRule(key, { require: !ruleDraft(key).require })"
                />
                <FilterRuleMode
                  :hard="!!ruleDraft(key).hard"
                  :boost="ruleDraft(key).boost"
                  show-boost
                  @update:hard="patchRule(key, { hard: $event })"
                  @update:boost="patchRule(key, { boost: $event })"
                />
              </div>

              <div v-else-if="isFilterActive(key) && filterKind(key) === 'int'" class="filter-card-body row-fields">
                <label class="field narrow">
                  <span>Min</span>
                  <input class="inp" type="number" :value="ruleDraft(key).min ?? ''" @input="patchRule(key, { min: Number(($event.target as HTMLInputElement).value) })" />
                </label>
                <label class="field narrow">
                  <span>Max</span>
                  <input class="inp" type="number" :value="ruleDraft(key).max ?? ''" @input="patchRule(key, { max: Number(($event.target as HTMLInputElement).value) })" />
                </label>
                <FilterRuleMode :hard="!!ruleDraft(key).hard" @update:hard="patchRule(key, { hard: $event })" />
              </div>

              <div v-else-if="isFilterActive(key) && filterKind(key) === 'salary'" class="filter-card-body row-fields">
                <label class="field narrow">
                  <span>Min annual</span>
                  <input class="inp" type="number" :value="ruleDraft(key).min ?? ''" @input="patchRule(key, { min: Number(($event.target as HTMLInputElement).value) })" />
                </label>
                <label class="field narrow">
                  <span>Currency</span>
                  <input class="inp" :value="ruleDraft(key).currency ?? 'USD'" @input="patchRule(key, { currency: ($event.target as HTMLInputElement).value })" />
                </label>
                <FilterRuleMode
                  :hard="!!ruleDraft(key).hard"
                  :boost="ruleDraft(key).boost"
                  show-boost
                  @update:hard="patchRule(key, { hard: $event })"
                  @update:boost="patchRule(key, { boost: $event })"
                />
              </div>
            </div>
          </template>
        </section>

        <section v-show="section === 'llm'" class="panel" aria-labelledby="panel-llm">
          <h2 id="panel-llm" class="panel-title">{{ t('sources.llmPanel') }}</h2>
          <p class="panel-hint">{{ t('sources.llmHint') }}</p>
          <div class="llm-card">
            <div class="flow"><span class="dot-sm" /> {{ config.llm.mode === 'cli' ? t('sources.llmCli') : t('sources.llmApi') }}</div>
            <div class="flow muted">{{ t('sources.llmModel', { model: config.llm.model }) }}</div>
            <div class="flow muted">{{ t('sources.llmCheap') }}</div>
          </div>
        </section>
      </main>
    </div>
    <p v-else-if="!error" class="muted load">Loading…</p>
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  color: var(--fg);
  background: var(--bg);
}
.bar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem 1.25rem 1rem;
  border-bottom: 1px solid var(--line);
  background: var(--surface);
}
.head h1 {
  margin: 0;
  font-size: 1.35rem;
  font-weight: 600;
  letter-spacing: -0.02em;
}
.lede {
  margin: 0.35rem 0 0;
  font-size: 12.5px;
  color: var(--muted);
  max-width: 36rem;
  line-height: 1.5;
}
.dot-sm {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
  flex-shrink: 0;
}
.grow {
  flex: 1;
}
.btn.ghost {
  flex-shrink: 0;
  font-size: 12px;
  padding: 7px 12px;
  border: 1px solid var(--line2);
  border-radius: var(--r-sm);
  color: var(--fg);
  text-decoration: none;
  background: var(--bg);
}
.btn.ghost:hover {
  background: var(--surface2);
}
.cfg {
  display: flex;
  flex: 1;
  min-height: 0;
}
.side {
  width: 220px;
  flex: 0 0 auto;
  border-right: 1px solid var(--line);
  background: var(--surface);
  padding: 14px 12px;
}
.lbl {
  font-size: 10.5px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--dim);
  margin: 8px 0 6px;
  padding: 0 8px;
}
.side > .lbl:first-child {
  margin-top: 0;
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
.side-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.nav-item {
  width: 100%;
  text-align: left;
  padding: 7px 8px;
  font-size: 13px;
  color: var(--muted);
  border-radius: var(--r-sm);
  border: none;
  background: transparent;
  cursor: pointer;
  font-family: inherit;
}
.nav-item:hover {
  color: var(--fg);
  background: var(--surface2);
}
.nav-item.on {
  background: var(--surface2);
  color: var(--fg);
  font-weight: 500;
}
.nav-item:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 1px;
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
.cfg-main {
  flex: 1;
  min-width: 0;
  padding: 1rem 1.25rem 2rem;
  overflow: auto;
  max-width: 56rem;
}
.panel-title {
  margin: 0 0 0.75rem;
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: -0.01em;
}
.panel-hint {
  margin: -0.35rem 0 1rem;
  font-size: 12.5px;
  color: var(--muted);
  line-height: 1.5;
}
.flow.row {
  padding: 5px 0;
  border-bottom: 1px dashed var(--line);
}
.llm-card {
  border: 1px solid var(--line);
  border-radius: var(--r);
  background: var(--surface);
  padding: 14px 16px;
  max-width: 28rem;
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
.notice {
  margin: 0;
  padding: 8px 16px;
  font-size: 12.5px;
  color: var(--accent);
  background: var(--accent-dim);
  border-bottom: 1px solid var(--line);
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.25rem;
}
.panel-head .panel-title {
  margin: 0;
}
.btn {
  font-size: 12px;
  padding: 7px 12px;
  border-radius: var(--r-sm);
  border: 1px solid var(--line2);
  background: var(--surface);
  color: var(--fg);
  cursor: pointer;
  font-family: inherit;
}
.btn.primary {
  background: var(--accent-dim);
  border-color: var(--accent);
  color: var(--fg);
  font-weight: 500;
}
.btn:disabled {
  opacity: 0.6;
  cursor: default;
}
.btn.sm {
  padding: 5px 8px;
}
.q-edit {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
.inp {
  flex: 1;
  min-width: 0;
  padding: 7px 10px;
  border: 1px solid var(--line2);
  border-radius: var(--r-sm);
  background: var(--bg);
  color: var(--fg);
  font-family: inherit;
  font-size: 13px;
}
.inp:focus {
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}
.filter-block {
  border: 1px solid var(--line);
  border-radius: var(--r-sm);
  margin-bottom: 8px;
  background: var(--surface);
}
.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 10px;
  flex-wrap: wrap;
}
.filter-row.head {
  border-bottom: 1px solid var(--line);
}
.filter-body {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 10px;
  align-items: flex-end;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  min-width: 140px;
  font-size: 11px;
  color: var(--muted);
}
.field.narrow {
  flex: 0 0 100px;
  min-width: 80px;
}
.chk {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
}
.filter-help {
  margin: 0 0 16px;
  padding: 12px 14px;
  border-radius: var(--r-sm);
  border: 1px solid var(--line);
  background: var(--bg);
  font-size: 12.5px;
  line-height: 1.5;
  color: var(--muted);
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}
.filter-help p {
  margin: 0 0 6px;
}
.filter-help p:last-child {
  margin-bottom: 0;
}
.filter-help .dim {
  color: var(--dim);
  font-size: 12px;
}
.group-title {
  margin: 18px 0 10px;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--dim);
  font-weight: 600;
}
.filter-card {
  border: 1px solid var(--line);
  border-radius: var(--r);
  background: var(--surface);
  padding: 12px 14px;
  margin-bottom: 10px;
}
.filter-card.on {
  border-color: var(--accent-dim);
  box-shadow: 0 0 0 1px var(--amber-bg);
}
.filter-card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.filter-name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}
.filter-sub {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--dim);
  line-height: 1.4;
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}
.filter-id {
  margin: 6px 0 0;
  font-size: 10px;
  color: var(--dim);
}
.filter-summary {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--muted);
}
.filter-card-body {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid var(--line);
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.filter-card-body.row-fields {
  flex-direction: row;
  flex-wrap: wrap;
  align-items: flex-end;
}
.chip-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.custom-kw {
  display: flex;
  gap: 8px;
  align-items: center;
  max-width: 20rem;
}
.custom-kw .inp {
  flex: 1;
}
.chk.inline {
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}
.chip.dim {
  margin-left: auto;
}
</style>
