<script setup lang="ts">
const props = defineProps<{ jobId: string }>()
const jobId = computed(() => props.jobId)
const apiJobId = computed(() => encodeJobIdForApi(jobId.value))

type Hunk = {
  id: string
  field_id: string
  section: string
  before: string
  after: string
  rationale: string
  evidence: string[]
  tags?: string[]
  status: string
}

type Proposal = {
  id: string
  job_id: string
  hunks: Hunk[]
}

type HunkReview = {
  index: number
  hunk: Hunk
  decision?: string
  text?: string
  stale: boolean
  decided: boolean
}

type HunkReviewSummary = {
  job_id: string
  total: number
  decided: number
  hunks: HunkReview[]
}

type CvDoc = {
  contact: { name: string; headline: string; email: string; location?: string }
  summary: { id: string; text: string }
  skills: { id: string; label: string; items: string[] }[]
  experience: { id: string; company: string; title: string; start: string; bullets: { id: string; text: string }[] }[]
}

type JobView = { job: { id: string; company: string; title: string } }

const proposal = ref<Proposal | null>(null)
const review = ref<HunkReviewSummary | null>(null)
const showHelp = ref(false)
const editText = ref('')
const editOpen = ref(false)
const master = ref<CvDoc | null>(null)
const jobInfo = ref<{ company: string; title: string } | null>(null)
const jobMissing = ref(false)
const selectedHunkId = ref<string | null>(null)
const error = ref('')
const loading = ref(false)
const { tailorJob, track, actionErrorMessage } = useActivity()

const jobHeadline = computed(() => {
  if (jobInfo.value) return `${jobInfo.value.title} · ${jobInfo.value.company}`
  if (jobId.value.includes('\0')) {
    const parts = jobId.value.split('\0')
    if (parts.length >= 3) return `${parts[2]} · ${parts[1]}`
  }
  return jobId.value.length > 72 ? `${jobId.value.slice(0, 72)}…` : jobId.value
})

const hunkByField = computed(() => {
  const m = new Map<string, Hunk>()
  for (const h of proposal.value?.hunks ?? []) {
    m.set(h.field_id, h)
  }
  return m
})

const selectedHunk = computed(() =>
  proposal.value?.hunks.find((h) => h.id === selectedHunkId.value) ?? null
)

const decidedCount = computed(() => review.value?.decided ?? 0)
const totalHunks = computed(() => review.value?.total ?? proposal.value?.hunks.length ?? 0)
const buildReady = computed(() => totalHunks.value > 0 && decidedCount.value >= totalHunks.value)

async function resolveJob() {
  jobMissing.value = false
  jobInfo.value = null
  try {
    const jobs = await $fetch<JobView[]>('/api/jobs')
    const row = jobs.find((r) => r.job.id === jobId.value)
    if (row) {
      jobInfo.value = { company: row.job.company, title: row.job.title }
    } else {
      jobMissing.value = true
    }
  } catch {
    jobMissing.value = true
  }
}

async function loadAll() {
  error.value = ''
  loading.value = true
  try {
    await resolveJob()
    const [propRes, masterRes] = await Promise.all([
      fetch(`/api/tailor/${apiJobId.value}`),
      fetch('/api/cv/master'),
    ])
    if (masterRes.ok) {
      master.value = await masterRes.json()
    }
    if (propRes.status === 404) {
      proposal.value = null
      return
    }
    if (!propRes.ok) {
      throw new Error(await propRes.text())
    }
    proposal.value = await propRes.json()
    await loadReview()
    if (!selectedHunkId.value && proposal.value?.hunks.length) {
      selectedHunkId.value = proposal.value.hunks[0].id
      hunkIndex.value = 0
    }
  } catch (e: unknown) {
    error.value = actionErrorMessage(e)
  } finally {
    loading.value = false
  }
}

async function generate() {
  loading.value = true
  error.value = ''
  try {
    await tailorJob(jobId.value)
    await loadAll()
  } catch (e: unknown) {
    error.value = actionErrorMessage(e)
  } finally {
    loading.value = false
  }
}

async function loadReview() {
  const res = await fetch(`/api/jobs/${apiJobId.value}/hunks`)
  if (res.status === 404) {
    review.value = null
    return
  }
  if (!res.ok) {
    throw new Error(await res.text())
  }
  review.value = (await res.json()) as HunkReviewSummary
  if (review.value?.hunks.length) {
    proposal.value = {
      id: proposal.value?.id ?? '',
      job_id: review.value.job_id,
      hunks: review.value.hunks.map((r) => ({
        ...r.hunk,
        status: r.hunk.status,
        stale: r.stale,
        decision: r.decision,
      })),
    }
  }
}

async function postDecision(index: number, decision: string, text = '') {
  const res = await fetch(`/api/jobs/${apiJobId.value}/hunks/${index}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ decision, text }),
  })
  if (!res.ok) {
    throw new Error(await res.text())
  }
  review.value = (await res.json()) as HunkReviewSummary
  if (review.value?.hunks.length) {
    proposal.value = {
      id: proposal.value?.id ?? '',
      job_id: review.value.job_id,
      hunks: review.value.hunks.map((r) => r.hunk),
    }
  }
}

async function setStatus(hunk: Hunk, action: 'accept' | 'reject') {
  const idx = (proposal.value?.hunks.findIndex((h) => h.id === hunk.id) ?? -1) + 1
  if (idx < 1) return
  const verb = action === 'accept' ? 'Accept' : 'Reject'
  try {
    await track(`${verb} hunk · ${hunk.field_id}`, async () => {
      await postDecision(idx, action)
    })
  } catch (e) {
    error.value = actionErrorMessage(e)
  }
}

async function buildPdf() {
  if (!buildReady.value) return
  try {
    await track('Build PDF', async () => {
      const res = await fetch(`/api/jobs/${apiJobId.value}/build`, { method: 'POST' })
      if (!res.ok) {
        throw new Error(await res.text())
      }
      return res.json()
    })
  } catch (e) {
    error.value = actionErrorMessage(e)
  }
}

function openEdit() {
  const h = selectedHunk.value
  if (!h) return
  editText.value = h.after
  editOpen.value = true
}

async function submitEdit() {
  const idx = hunkIndex.value + 1
  editOpen.value = false
  try {
    await postDecision(idx, 'edit', editText.value)
  } catch (e) {
    error.value = actionErrorMessage(e)
  }
}

function displayText(fieldId: string, fallback: string): string {
  const h = hunkByField.value.get(fieldId)
  if (h?.status === 'accepted') return h.after
  return fallback
}

function isChanged(fieldId: string): boolean {
  const h = hunkByField.value.get(fieldId)
  return !!h && h.status !== 'rejected' && h.before !== h.after
}

function hunkClass(fieldId: string): string {
  const h = hunkByField.value.get(fieldId)
  if (!h || h.status === 'rejected') return ''
  if (h.status === 'accepted') return 'accepted'
  return 'chg add'
}

const hunkIndex = ref(0)

function moveHunk(delta: number) {
  const n = proposal.value?.hunks.length ?? 0
  if (n === 0) return
  hunkIndex.value = Math.max(0, Math.min(n - 1, hunkIndex.value + delta))
  selectedHunkId.value = proposal.value!.hunks[hunkIndex.value].id
}

async function toggleSelectedHunk() {
  const h = proposal.value?.hunks[hunkIndex.value]
  if (!h) return
  await setStatus(h, h.status === 'accepted' ? 'reject' : 'accept')
}

function onDiffKey(e: KeyboardEvent) {
  if (editOpen.value) return
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
  if (showHelp.value) {
    if (e.key === '?' || e.key === 'Escape') {
      e.preventDefault()
      showHelp.value = false
    }
    return
  }
  if (e.key === '?') {
    e.preventDefault()
    showHelp.value = true
    return
  }
  if (e.key === 'j') {
    e.preventDefault()
    moveHunk(1)
  }
  if (e.key === 'k') {
    e.preventDefault()
    moveHunk(-1)
  }
  if (e.key === 'g') {
    e.preventDefault()
    hunkIndex.value = 0
    if (proposal.value?.hunks[0]) selectedHunkId.value = proposal.value.hunks[0].id
  }
  if (e.key === 'G') {
    e.preventDefault()
    const n = proposal.value?.hunks.length ?? 0
    if (n > 0) {
      hunkIndex.value = n - 1
      selectedHunkId.value = proposal.value!.hunks[n - 1].id
    }
  }
  if (e.code === 'Space') {
    e.preventDefault()
    void toggleSelectedHunk()
  }
  if (e.key === 'a') {
    e.preventDefault()
    const h = proposal.value?.hunks[hunkIndex.value]
    if (h) void setStatus(h, 'accept')
  }
  if (e.key === 'x') {
    e.preventDefault()
    const h = proposal.value?.hunks[hunkIndex.value]
    if (h) void setStatus(h, 'reject')
  }
  if (e.key === 'e') {
    e.preventDefault()
    openEdit()
  }
  if (e.key === 'b') {
    e.preventDefault()
    void buildPdf()
  }
}

onMounted(() => {
  loadAll()
  window.addEventListener('keydown', onDiffKey)
})
onUnmounted(() => window.removeEventListener('keydown', onDiffKey))
watch(jobId, loadAll)
watch(
  () => proposal.value?.hunks.length,
  () => {
    hunkIndex.value = 0
  }
)
</script>

<template>
  <div class="page">
    <header class="topbar">
      <NuxtLink to="/jobs" class="brand">← Jobs</NuxtLink>
      <span class="muted job-title">{{ jobHeadline }}</span>
      <div class="grow" />
      <button class="btn ghost" :disabled="loading || jobMissing" @click="generate">Tailor for this JD</button>
    </header>

    <p v-if="error" class="err">{{ error }}</p>
    <p v-if="loading" class="meta">Loading diff…</p>

    <div v-if="!loading && jobMissing" class="empty">
      <p>This job is not in your database yet.</p>
      <p class="hint">
        Run a <NuxtLink to="/jobs">scan</NuxtLink> and open Diff from the job row. Job ids are
        <span class="mono">url + company + title</span> — use the link from Jobs, not only the board URL.
      </p>
    </div>

    <div v-else-if="!loading && !proposal" class="empty">
      <p>No tailoring proposal yet for this role.</p>
      <button class="btn" :disabled="jobMissing" @click="generate">Generate hunks</button>
      <p v-if="!master" class="hint">
        You also need a saved master CV on <NuxtLink to="/cv">My CV</NuxtLink> before tailoring.
      </p>
    </div>

    <div v-else-if="!loading && proposal" class="rows">
      <aside class="changes">
        <div class="lbl">Changes · {{ proposal?.hunks.length ?? 0 }}</div>
        <button
          v-for="h in proposal?.hunks ?? []"
          :key="h.id"
          type="button"
          class="ci"
          :class="{ on: h.id === selectedHunkId }"
          @click="selectedHunkId = h.id"
        >
          <div class="s mono">@@ {{ h.field_id }} @@</div>
          <div class="t">
            {{ h.after.slice(0, 48) }}…
            <span class="st" :class="h.status">{{ h.status }}</span>
          </div>
        </button>
      </aside>

      <div v-if="!master" class="sheet-missing">
        <p>Save your master CV on <NuxtLink to="/cv">My CV</NuxtLink> to preview the tailored document.</p>
        <p class="muted">Hunk list and rationale are still available in the side panels.</p>
      </div>
      <div v-else class="sheetwrap">
        <div class="sheet">
          <div class="pad">
            <div class="cvn">{{ master.contact.name }}</div>
            <div class="cvh">{{ displayText('contact.headline', master.contact.headline) }}</div>
            <div class="cvc">{{ master.contact.email }}</div>

            <div class="cvsec">Professional Summary</div>
            <div class="cvp" :class="hunkClass(master.summary.id)">
              <span v-if="isChanged(master.summary.id)" class="mk">+</span>
              {{ displayText(master.summary.id, master.summary.text) }}
              <span v-if="hunkByField.get(master.summary.id)?.status === 'pending'" class="ctl">
                <button type="button" class="ok" @click="setStatus(hunkByField.get(master.summary.id)!, 'accept')">Accept</button>
                <button type="button" class="no" @click="setStatus(hunkByField.get(master.summary.id)!, 'reject')">Reject</button>
              </span>
            </div>

            <div class="cvsec">Technical Skills</div>
            <div class="cvsk">
              <div v-for="sg in master.skills" :key="sg.id" :class="hunkClass(sg.id)">
                <span v-if="isChanged(sg.id)" class="mk">+</span>
                <b>{{ sg.label }}</b>: {{ displayText(sg.id, sg.items.join(', ')) }}
                <span v-if="hunkByField.get(sg.id)?.tags?.includes('ats_keyword')" class="ats">ATS keyword</span>
                <span v-if="hunkByField.get(sg.id)?.status === 'pending'" class="ctl">
                  <button type="button" class="ok" @click="setStatus(hunkByField.get(sg.id)!, 'accept')">Accept</button>
                  <button type="button" class="no" @click="setStatus(hunkByField.get(sg.id)!, 'reject')">Reject</button>
                </span>
              </div>
            </div>

            <div class="cvsec">Work Experience</div>
            <template v-for="exp in master.experience" :key="exp.id">
              <div class="cvrow">
                <div class="l">{{ exp.title }} <span>· {{ exp.company }}</span></div>
                <div class="r">{{ exp.start }}</div>
              </div>
              <ul class="dots">
                <li v-for="b in exp.bullets" :key="b.id" :class="hunkClass(b.id)">
                  <span v-if="isChanged(b.id)" class="mk">+</span>
                  <ins v-if="isChanged(b.id)">{{ displayText(b.id, b.text) }}</ins>
                  <template v-else>{{ b.text }}</template>
                  <span v-if="hunkByField.get(b.id)?.status === 'pending'" class="ctl">
                    <button type="button" class="ok" @click="setStatus(hunkByField.get(b.id)!, 'accept')">Accept</button>
                    <button type="button" class="no" @click="setStatus(hunkByField.get(b.id)!, 'reject')">Reject</button>
                  </span>
                </li>
              </ul>
            </template>
          </div>
        </div>
      </div>

      <aside class="rait">
        <div class="lbl">Why this edit</div>
        <template v-if="selectedHunk">
          <p class="muted">{{ selectedHunk.rationale }}</p>
          <div v-for="ev in selectedHunk.evidence" :key="ev" class="ev mono">{{ ev }}</div>
        </template>
        <p v-else class="muted">Select a change.</p>
      </aside>
    </div>

    <footer v-if="proposal" class="diff-bar">
      <span class="progress">{{ decidedCount }} of {{ totalHunks }} decided</span>
      <button type="button" class="btn" :disabled="!buildReady || loading" @click="buildPdf">Build PDF</button>
    </footer>

    <div v-if="showHelp" class="help" role="dialog" aria-label="Keyboard shortcuts">
      <p><strong>Diff keyboard</strong> (press ? to close)</p>
      <ul>
        <li>j / k — next / previous hunk</li>
        <li>g / G — first / last hunk</li>
        <li>a accept · x reject · e edit · space toggle</li>
        <li>b — build PDF when all hunks decided</li>
      </ul>
    </div>

    <div v-if="editOpen" class="edit-modal" role="dialog">
      <label class="lbl">Edit hunk text</label>
      <textarea v-model="editText" rows="5" />
      <div class="edit-actions">
        <button type="button" class="btn ghost" @click="editOpen = false">Cancel</button>
        <button type="button" class="btn" @click="submitEdit">Save</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg, var(--bg));
  color: var(--fg, var(--fg));
}
.topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--line);
  background: var(--surface);
}
.brand {
  color: var(--accent);
  text-decoration: none;
  font-weight: 600;
}
.grow {
  flex: 1;
}
.btn {
  background: var(--accent);
  color: var(--accent-dim);
  border: none;
  border-radius: 7px;
  padding: 8px 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn.ghost {
  background: transparent;
  color: var(--fg);
  border: 1px solid var(--line2);
}
.rows {
  flex: 1;
  display: flex;
  min-height: 0;
}
.changes {
  width: 264px;
  border-right: 1px solid var(--line);
  background: var(--surface);
  padding: 12px;
  overflow: auto;
}
.ci {
  display: block;
  width: 100%;
  text-align: left;
  padding: 9px 10px;
  border-radius: 7px;
  margin-bottom: 6px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--fg);
  cursor: pointer;
}
.ci.on {
  background: var(--surface2);
  border-color: var(--line);
}
.ci .s {
  font-size: 10px;
  color: var(--dim);
  font-family: ui-monospace, monospace;
}
.ci .t {
  font-size: 12.5px;
  margin-top: 3px;
}
.st {
  font-size: 10px;
  margin-left: 6px;
  text-transform: uppercase;
}
.st.pending {
  color: var(--amber);
}
.st.accepted {
  color: var(--add);
}
.sheetwrap {
  flex: 1;
  overflow: auto;
  background: var(--bg);
  display: flex;
  justify-content: center;
  padding: 16px 0;
}
.sheet {
  width: min(840px, 100%);
  background: var(--paper);
  color: var(--ink);
  box-shadow: 0 24px 70px var(--sheet-shadow);
}
.pad {
  padding: 34px 56px 60px;
}
.cvn {
  font-weight: 700;
  font-size: 25px;
}
.cvh {
  font-weight: 600;
  font-size: 12.5px;
  color: var(--ink-mut);
  margin-top: 4px;
}
.cvc {
  font-size: 11.5px;
  color: var(--ink-mut);
  margin-top: 6px;
}
.cvsec {
  font-weight: 700;
  font-size: 10.5px;
  letter-spacing: 1.1px;
  text-transform: uppercase;
  border-bottom: 1.4px solid var(--ink);
  padding-bottom: 3px;
  margin: 20px 0 9px;
}
.cvp,
.cvsk,
.dots li {
  font-size: 12.5px;
  line-height: 1.55;
  position: relative;
  padding-left: 20px;
}
.chg .mk {
  position: absolute;
  left: 0;
  top: 2px;
  font-family: ui-monospace, monospace;
  font-size: 10px;
  font-weight: 700;
  color: var(--add);
}
.chg::before {
  content: '';
  position: absolute;
  left: 12px;
  top: 1px;
  bottom: 1px;
  width: 2px;
  background: var(--accent)88;
  border-radius: 2px;
}
ins {
  background: var(--accent)20;
  border-bottom: 1px solid var(--paper-line);
  text-decoration: none;
}
.ats {
  font-size: 10px;
  color: var(--accent-dim);
  background: var(--accent)30;
  padding: 1px 5px;
  border-radius: 4px;
  margin-left: 6px;
}
.ctl {
  display: inline-flex;
  gap: 6px;
  margin-left: 8px;
}
.ctl button {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid var(--paper-line);
  cursor: pointer;
  background: var(--paper);
}
.ctl .ok {
  color: var(--add);
  border-color: var(--add);
}
.ctl .no {
  color: var(--del);
  border-color: var(--del);
}
.rait {
  width: 290px;
  border-left: 1px solid var(--line);
  background: var(--surface);
  padding: 14px;
  overflow: auto;
}
.lbl {
  font-size: 10.5px;
  text-transform: uppercase;
  letter-spacing: 0.9px;
  color: var(--dim);
  font-weight: 600;
  margin-bottom: 8px;
}
.ev {
  font-size: 11px;
  margin: 4px 0;
  font-family: ui-monospace, monospace;
}
.muted {
  color: var(--muted);
}
.err {
  color: var(--del);
  padding: 8px 16px;
}
.empty {
  padding: 40px;
  text-align: center;
}
.hint {
  margin-top: 12px;
  font-size: 13px;
  color: var(--muted);
  max-width: 28rem;
  margin-left: auto;
  margin-right: auto;
  line-height: 1.5;
}
.hint a {
  color: var(--accent);
}
.meta {
  padding: 12px 16px;
  color: var(--muted);
}
.job-title {
  font-size: 13px;
  max-width: min(48vw, 520px);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sheet-missing {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px;
  color: var(--muted);
  text-align: center;
  gap: 8px;
}
.sheet-missing a {
  color: var(--accent);
}
.mono {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 12px;
}
.dots {
  margin: 5px 0 0 15px;
}
.cvrow {
  display: flex;
  justify-content: space-between;
  margin-top: 11px;
  font-size: 12.5px;
}
.cvrow .l span {
  color: var(--ink-mut);
  font-weight: 400;
}
.accepted {
  opacity: 0.85;
}
.diff-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 1rem;
  border-top: 1px solid var(--line);
  background: var(--surface);
}
.progress {
  font-size: 0.9rem;
  color: var(--muted);
}
.help {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 40;
}
.help ul {
  margin: 0.5rem 0 0;
  padding-left: 1.2rem;
  line-height: 1.6;
}
.help p {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--r);
  padding: 1rem 1.25rem;
  max-width: 22rem;
}
.edit-modal {
  position: fixed;
  bottom: 4rem;
  left: 50%;
  transform: translateX(-50%);
  width: min(520px, 92vw);
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--r);
  padding: 1rem;
  z-index: 35;
}
.edit-modal textarea {
  width: 100%;
  margin-top: 0.5rem;
  font-family: inherit;
}
.edit-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.75rem;
}
.st.stale {
  color: var(--amber);
}
</style>
