<script setup lang="ts">
type Dimensions = {
  skills: number
  seniority: number
  domain: number
  logistics: number
  motivation: number
}

type Requirement = {
  phrase: string
  status: 'match' | 'partial' | 'gap'
  cv_field_id?: string
  cv_field_text?: string
}

type Legitimacy = {
  company_verified: boolean
  official_ats_url: boolean
  salary_consistent: boolean
  flags?: string[]
}

type MatchResult = {
  job_id: string
  total: number
  dimensions: Dimensions
  requirements: Requirement[]
  legitimacy: Legitimacy
  jd_text: string
  computed_at: string
}

type Job = {
  id: string
  company: string
  title: string
  url: string
}

const route = useRoute()
const jobId = computed(() => String(route.params.id))

const job = ref<Job | null>(null)
const match = ref<MatchResult | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const decision = ref<string | null>(null)

const dimensionRows = computed(() => {
  if (!match.value) return []
  const d = match.value.dimensions
  return [
    { key: 'skills', label: 'Skills', score: d.skills },
    { key: 'seniority', label: 'Seniority', score: d.seniority },
    { key: 'domain', label: 'Domain', score: d.domain },
    { key: 'logistics', label: 'Logistics', score: d.logistics },
    { key: 'motivation', label: 'Motivation', score: d.motivation },
  ]
})

async function loadJob() {
  const jobs = await $fetch<Job[]>('/api/jobs')
  job.value = jobs.find((j) => j.id === jobId.value) ?? null
}

async function runMatch() {
  loading.value = true
  error.value = null
  try {
    await loadJob()
    match.value = await $fetch<MatchResult>(`/api/match/${encodeURIComponent(jobId.value)}`, {
      method: 'POST',
    })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Match failed'
  } finally {
    loading.value = false
  }
}

async function setDecision(status: string) {
  await $fetch(`/api/jobs/${encodeURIComponent(jobId.value)}`, {
    method: 'PUT',
    body: { status },
  })
  decision.value = status
}

onMounted(() => {
  runMatch()
})
</script>

<template>
  <main class="page">
    <header class="bar">
      <NuxtLink to="/jobs" class="back">← Jobs</NuxtLink>
      <div v-if="job" class="headline">
        <h1>{{ job.title }}</h1>
        <p class="sub">{{ job.company }}</p>
      </div>
      <div v-if="match" class="score">
        <span class="score-num">{{ match.total }}</span>
        <span class="score-max">/ 25</span>
      </div>
    </header>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="loading" class="meta">Computing match…</p>

    <section v-if="match" class="legit">
      <span :class="['flag', match.legitimacy.company_verified ? 'ok' : 'warn']">
        Company {{ match.legitimacy.company_verified ? 'verified' : 'unverified' }}
      </span>
      <span :class="['flag', match.legitimacy.official_ats_url ? 'ok' : 'warn']">
        ATS URL {{ match.legitimacy.official_ats_url ? 'official' : 'unofficial' }}
      </span>
      <span :class="['flag', match.legitimacy.salary_consistent ? 'ok' : 'warn']">
        Salary {{ match.legitimacy.salary_consistent ? 'consistent' : 'mismatch' }}
      </span>
    </section>

    <section v-if="match" class="fit">
      <h2>Fit</h2>
      <div v-for="row in dimensionRows" :key="row.key" class="fit-row">
        <span class="fit-label">{{ row.label }}</span>
        <div class="fit-track">
          <div class="fit-fill" :style="{ width: `${(row.score / 5) * 100}%` }" />
        </div>
        <span class="fit-score">{{ row.score }}/5</span>
      </div>
    </section>

    <div v-if="match" class="split">
      <section class="panel">
        <h2>Job description</h2>
        <pre class="jd">{{ match.jd_text }}</pre>
      </section>
      <section class="panel">
        <h2>CV mapping</h2>
        <ul class="reqs">
          <li v-for="(req, i) in match.requirements" :key="i" :class="['req', req.status]">
            <div class="req-phrase">{{ req.phrase }}</div>
            <div class="req-meta">
              <span class="badge">{{ req.status }}</span>
              <span v-if="req.cv_field_id" class="field">{{ req.cv_field_id }}</span>
            </div>
            <p v-if="req.cv_field_text" class="cv-snippet">{{ req.cv_field_text }}</p>
          </li>
        </ul>
      </section>
    </div>

    <footer v-if="match" class="actions">
      <button type="button" class="skip" @click="setDecision('discarded')">Skip</button>
      <button type="button" class="shortlist" @click="setDecision('shortlisted')">Shortlist</button>
      <button type="button" class="tailor" @click="setDecision('tailoring')">Tailor</button>
      <span v-if="decision" class="decision-note">Status → {{ decision }}</span>
    </footer>
  </main>
</template>

<style scoped>
.page {
  max-width: 72rem;
  margin: 0 auto;
  padding: 1.5rem 1rem 4rem;
  font-family: system-ui, sans-serif;
  color: var(--fg);
  background: var(--bg);
  min-height: 100vh;
}
.bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1.25rem;
}
.back {
  color: var(--muted);
  text-decoration: none;
  font-size: 0.9rem;
}
.back:hover {
  color: var(--accent);
}
.headline h1 {
  margin: 0;
  font-size: 1.35rem;
}
.sub {
  margin: 0.25rem 0 0;
  color: var(--muted);
  font-size: 0.9rem;
}
.score {
  margin-left: auto;
  font-variant-numeric: tabular-nums;
}
.score-num {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--accent);
}
.score-max {
  color: var(--dim);
}
.error {
  color: var(--del);
}
.meta {
  color: var(--muted);
}
.legit {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}
.flag {
  font-size: 0.75rem;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  border: 1px solid var(--line);
}
.flag.ok {
  background: var(--add-bg);
  color: var(--add);
}
.flag.warn {
  background: var(--amber-bg);
  color: var(--amber);
}
.fit {
  margin-bottom: 1.25rem;
  padding: 1rem;
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: 10px;
}
.fit h2 {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: var(--muted);
  font-weight: 600;
}
.fit-row {
  display: grid;
  grid-template-columns: 6rem 1fr 2.5rem;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 0.45rem;
}
.fit-label {
  font-size: 0.85rem;
}
.fit-track {
  height: 8px;
  background: var(--surface2);
  border-radius: 4px;
  overflow: hidden;
}
.fit-fill {
  height: 100%;
  background: var(--accent);
  border-radius: 4px;
}
.fit-score {
  font-size: 0.8rem;
  color: var(--muted);
  text-align: right;
}
.split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
@media (max-width: 768px) {
  .split {
    grid-template-columns: 1fr;
  }
}
.panel {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: 10px;
  padding: 1rem;
  min-height: 12rem;
}
.panel h2 {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: var(--muted);
}
.jd {
  margin: 0;
  white-space: pre-wrap;
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 0.8rem;
  line-height: 1.45;
  color: var(--muted);
}
.reqs {
  list-style: none;
  margin: 0;
  padding: 0;
}
.req {
  padding: 0.65rem 0;
  border-bottom: 1px solid var(--line);
}
.req:last-child {
  border-bottom: none;
}
.req-phrase {
  font-size: 0.85rem;
}
.req-meta {
  margin-top: 0.35rem;
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.badge {
  font-size: 0.7rem;
  text-transform: uppercase;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}
.req.match .badge {
  background: var(--add-bg);
  color: var(--add);
}
.req.partial .badge {
  background: var(--amber-bg);
  color: var(--amber);
}
.req.gap .badge {
  background: var(--del-bg);
  color: var(--del);
}
.field {
  font-family: ui-monospace, monospace;
  font-size: 0.7rem;
  color: var(--dim);
}
.cv-snippet {
  margin: 0.35rem 0 0;
  font-size: 0.78rem;
  color: var(--muted);
}
.actions {
  position: sticky;
  bottom: 0;
  margin-top: 1.5rem;
  padding: 1rem;
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  align-items: center;
}
.actions button {
  border: none;
  border-radius: 7px;
  padding: 0.55rem 1.1rem;
  font-weight: 600;
  cursor: pointer;
}
.skip {
  background: var(--surface2);
  color: var(--muted);
}
.shortlist {
  background: var(--add-bg);
  color: var(--add);
}
.tailor {
  background: var(--accent);
  color: var(--bg);
}
.decision-note {
  margin-left: auto;
  font-size: 0.85rem;
  color: var(--muted);
}
</style>
