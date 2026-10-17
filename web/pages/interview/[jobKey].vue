<script setup lang="ts">
type FrameworkStep = {
  idx: number
  name: string
  minutes: number
  required_output: string
  probes: string[]
}

type Framework = {
  id: string
  name: string
  steps: FrameworkStep[]
  rubric_dimensions: string[]
}

type Story = {
  id: string
  category: string
  situation: string
  task: string
  action: string
  result: string
  learning: string
}

type Workspace = {
  application_id: string
  loop: string[]
  topics: string[]
  questions: string[]
  stories: Story[]
  categories: string[]
  empty_categories: string[]
  mock_type: string
  framework?: Framework
}

type Job = {
  id: string
  company: string
  title: string
}

const route = useRoute()

function decodeJobKey(raw: string): string {
  try {
    return decodeURIComponent(raw)
  } catch {
    return raw
  }
}

const jobKey = computed(() => decodeJobKey(String(route.params.jobKey)))

const job = ref<Job | null>(null)
const ws = ref<Workspace | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const searchEl = ref<HTMLInputElement | null>(null)
const selectedStep = ref(0)
const remainSec = ref<number[]>([])
const timerOn = ref(false)
const stepNotes = ref<string[]>([])
const rubricScores = ref<Record<string, number>>({})
const rubricComments = ref<Record<string, string>>({})
const rubricFocus = ref(0)
const storySearch = ref('')
const editingStory = ref<Story | null>(null)
const storyForm = ref({
  category: 'failure',
  situation: '',
  task: '',
  action: '',
  result: '',
  learning: '',
})

let timerHandle: ReturnType<typeof setInterval> | null = null

useSearchFocus(searchEl)

function onListKey(e: KeyboardEvent) {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
    return
  }
  const steps = ws.value?.framework?.steps?.length ?? 0
  if (e.key === 'j' && selectedStep.value < steps - 1) {
    e.preventDefault()
    selectedStep.value++
  }
  if (e.key === 'k' && selectedStep.value > 0) {
    e.preventDefault()
    selectedStep.value--
  }
  if (e.code === 'Space') {
    e.preventDefault()
    toggleTimer()
  }
}

function resetRunner(fw: Framework) {
  remainSec.value = fw.steps.map((s) => s.minutes * 60)
  stepNotes.value = fw.steps.map(() => '')
  selectedStep.value = 0
  timerOn.value = false
  rubricScores.value = {}
  rubricComments.value = {}
  rubricFocus.value = 0
}

async function loadJob() {
  const rows = await $fetch<{ job: Job }[]>('/api/jobs')
  job.value = rows.map((r) => r.job).find((j) => j.id === jobKey.value) ?? null
}

async function loadWorkspace() {
  loading.value = true
  error.value = null
  try {
    await loadJob()
    ws.value = await $fetch<Workspace>(`/api/interview/${encodeURIComponent(jobKey.value)}`)
    if (ws.value?.framework) {
      resetRunner(ws.value.framework)
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load interview prep'
  } finally {
    loading.value = false
  }
}

const { track, actionErrorMessage } = useActivity()

async function ensureKit() {
  try {
    await track('Generate interview kit', async () => {
      await $fetch(`/api/interview/${encodeURIComponent(jobKey.value)}/kit`, { method: 'POST' })
      await loadWorkspace()
    })
  } catch (e) {
    error.value = actionErrorMessage(e)
  }
}

function toggleTimer() {
  timerOn.value = !timerOn.value
  if (timerOn.value) {
    timerHandle = setInterval(() => {
      const i = selectedStep.value
      if (remainSec.value[i] > 0) {
        remainSec.value[i]--
      }
    }, 1000)
  } else if (timerHandle) {
    clearInterval(timerHandle)
    timerHandle = null
  }
}

function formatMMSS(sec: number) {
  const s = Math.max(0, sec)
  return `${String(Math.floor(s / 60)).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`
}

function setRubricScore(dim: string, score: number) {
  rubricScores.value = { ...rubricScores.value, [dim]: score }
}

async function saveSession() {
  if (!ws.value?.framework) return
  const steps = ws.value.framework.steps.map((st, i) => ({
    idx: st.idx,
    name: st.name,
    timebox_s: st.minutes * 60,
    output: '',
    note: stepNotes.value[i] ?? '',
    weak: remainSec.value[i] <= 0,
  }))
  try {
    await track('Save mock interview session', () =>
      $fetch(`/api/interview/${encodeURIComponent(jobKey.value)}/mock/save`, {
        method: 'POST',
        body: {
          type: ws.value!.mock_type,
          steps,
          scores: rubricScores.value,
          comments: rubricComments.value,
        },
      }),
    )
  } catch (e) {
    error.value = actionErrorMessage(e)
  }
}

const filteredStories = computed(() => {
  const q = storySearch.value.toLowerCase().trim()
  const stories = ws.value?.stories ?? []
  if (!q) return stories
  return stories.filter((s) =>
    [s.category, s.situation, s.action, s.result].join(' ').toLowerCase().includes(q),
  )
})

function openStory(st: Story) {
  editingStory.value = st
  storyForm.value = { ...st }
}

function newStory() {
  editingStory.value = { id: '', category: 'failure', situation: '', task: '', action: '', result: '', learning: '' }
  storyForm.value = { category: 'failure', situation: '', task: '', action: '', result: '', learning: '' }
}

async function saveStory() {
  const label = editingStory.value?.id ? 'Update STAR story' : 'Add STAR story'
  try {
    await track(label, async () => {
      if (editingStory.value?.id) {
        await $fetch(`/api/interview/stories/${encodeURIComponent(editingStory.value.id)}`, {
          method: 'PUT',
          body: storyForm.value,
        })
      } else {
        await $fetch(`/api/interview/${encodeURIComponent(jobKey.value)}/stories`, {
          method: 'POST',
          body: storyForm.value,
        })
      }
      editingStory.value = null
      await loadWorkspace()
    })
  } catch (e) {
    error.value = actionErrorMessage(e)
  }
}

onMounted(() => {
  window.addEventListener('keydown', onListKey)
  loadWorkspace()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onListKey)
  if (timerHandle) clearInterval(timerHandle)
})
</script>

<template>
  <main class="page">
    <header class="bar">
      <NuxtLink to="/jobs" class="back">← Jobs</NuxtLink>
      <div v-if="job" class="headline">
        <h1>Interview prep</h1>
        <p class="sub">{{ job.title }} · {{ job.company }}</p>
      </div>
      <div class="actions">
        <button type="button" class="btn ghost" @click="ensureKit">Generate kit</button>
        <button type="button" class="btn" @click="saveSession">Save session</button>
      </div>
    </header>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="loading" class="meta">Loading…</p>

    <div v-if="ws && !loading" class="grid">
      <aside class="kit panel">
        <h2 class="lbl">Kit</h2>
        <h3 class="subhead">Loop</h3>
        <ul class="list">
          <li v-for="(stage, i) in ws.loop" :key="'l' + i">
            <span class="mono">{{ formatMMSS((ws.framework?.steps[i]?.minutes ?? 0) * 60) }}</span>
            {{ stage }}
          </li>
          <li v-if="!ws.loop.length" class="dim">No kit yet — generate kit.</li>
        </ul>
        <h3 class="subhead">Topics</h3>
        <ul class="list topics">
          <li v-for="t in ws.topics" :key="t">{{ t }}</li>
        </ul>
      </aside>

      <section class="runner panel">
        <h2 class="lbl">Mock runner</h2>
        <div v-if="ws.framework" class="runner-body">
          <div class="steps">
            <button
              v-for="(st, i) in ws.framework.steps"
              :key="st.idx"
              type="button"
              :class="['step', { on: i === selectedStep, weak: remainSec[i] <= 0 }]"
              @click="selectedStep = i"
            >
              <span class="mono">{{ formatMMSS(remainSec[i]) }}</span>
              {{ st.name }}
            </button>
          </div>
          <div v-if="ws.framework.steps[selectedStep]" class="detail">
            <div class="detail-head">
              <strong>{{ ws.framework.steps[selectedStep].name }}</strong>
              <button type="button" class="btn ghost sm" @click="toggleTimer">
                {{ timerOn ? 'Pause' : 'Start' }} timer
              </button>
            </div>
            <p class="dim">Required: {{ ws.framework.steps[selectedStep].required_output }}</p>
            <ul class="probes">
              <li v-for="p in ws.framework.steps[selectedStep].probes" :key="p">{{ p }}</li>
            </ul>
            <label class="lbl">Notes</label>
            <textarea
              v-model="stepNotes[selectedStep]"
              class="editor mono"
              rows="6"
              placeholder="Step notes…"
            />
          </div>
        </div>
      </section>

      <aside class="side panel">
        <h2 class="lbl">Rubric</h2>
        <div v-if="ws.framework" class="rubric">
          <div v-for="(dim, i) in ws.framework.rubric_dimensions" :key="dim" class="rubric-row">
            <span :class="['dim-label', { focus: i === rubricFocus }]">{{ dim }}</span>
            <div class="score-btns">
              <button
                v-for="n in 5"
                :key="n"
                type="button"
                :class="['score', { on: rubricScores[dim] === n }]"
                @click="setRubricScore(dim, n)"
              >
                {{ n }}
              </button>
            </div>
            <input
              v-model="rubricComments[dim]"
              class="comment"
              type="text"
              placeholder="Comment"
            />
          </div>
        </div>

        <h2 class="lbl">Story bank</h2>
        <p v-if="ws.empty_categories.length" class="warn">
          Empty categories: {{ ws.empty_categories.join(', ') }}
        </p>
        <input ref="searchEl" v-model="storySearch" class="search" type="search" placeholder="Search stories (/)…" />
        <ul class="stories">
          <li v-for="st in filteredStories" :key="st.id">
            <button type="button" class="story-link" @click="openStory(st)">
              <span class="tag">{{ st.category }}</span>
              {{ st.situation.slice(0, 60) }}…
            </button>
          </li>
        </ul>
        <button type="button" class="btn ghost sm" @click="newStory">Add story</button>

        <div v-if="editingStory" class="story-form">
          <select v-model="storyForm.category" class="search">
            <option v-for="c in ws.categories" :key="c" :value="c">{{ c }}</option>
          </select>
          <textarea v-model="storyForm.situation" class="editor mono" rows="2" placeholder="Situation" />
          <textarea v-model="storyForm.task" class="editor mono" rows="2" placeholder="Task" />
          <textarea v-model="storyForm.action" class="editor mono" rows="2" placeholder="Action" />
          <textarea v-model="storyForm.result" class="editor mono" rows="2" placeholder="Result" />
          <textarea v-model="storyForm.learning" class="editor mono" rows="2" placeholder="Learning" />
          <button type="button" class="btn sm" @click="saveStory">Save story</button>
        </div>
      </aside>
    </div>
  </main>
</template>

<style scoped>
.page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px 16px 40px;
}
.bar {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}
.back {
  color: var(--muted);
  text-decoration: none;
  font-size: 13px;
  padding-top: 4px;
}
.headline h1 {
  margin: 0;
  font-size: 1.15rem;
}
.sub {
  margin: 4px 0 0;
  color: var(--muted);
  font-size: 13px;
}
.actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}
.grid {
  display: grid;
  grid-template-columns: 240px 1fr 280px;
  gap: 12px;
  align-items: start;
}
@media (max-width: 960px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
.panel {
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--r);
  padding: 14px;
}
.lbl {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--dim);
  margin: 0 0 10px;
}
.subhead {
  font-size: 12px;
  color: var(--muted);
  margin: 12px 0 6px;
}
.list {
  list-style: none;
  padding: 0;
  margin: 0;
  font-size: 13px;
}
.list li {
  margin-bottom: 6px;
}
.topics li {
  color: var(--fg);
}
.dim {
  color: var(--dim);
  font-size: 12px;
}
.mono {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--muted);
  margin-right: 6px;
}
.steps {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}
.step {
  text-align: left;
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--r-sm);
  padding: 8px 10px;
  color: var(--fg);
  cursor: pointer;
  font-size: 13px;
}
.step.on {
  border-color: var(--accent-dim);
  background: var(--surface);
}
.step.weak {
  border-style: dashed;
  color: var(--muted);
}
.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.probes {
  margin: 8px 0 12px;
  padding-left: 18px;
  color: var(--muted);
  font-size: 13px;
}
.editor {
  width: 100%;
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--r-sm);
  color: var(--fg);
  padding: 10px;
  margin-bottom: 8px;
}
.rubric-row {
  margin-bottom: 12px;
}
.dim-label {
  font-size: 12px;
  color: var(--muted);
  display: block;
  margin-bottom: 4px;
}
.dim-label.focus {
  color: var(--accent);
}
.score-btns {
  display: flex;
  gap: 4px;
  margin-bottom: 4px;
}
.score {
  width: 28px;
  height: 28px;
  border-radius: var(--r-sm);
  border: 1px solid var(--line);
  background: var(--surface2);
  color: var(--fg);
  cursor: pointer;
  font-size: 12px;
}
.score.on {
  background: var(--accent);
  color: var(--ink);
  border-color: var(--accent);
}
.comment {
  width: 100%;
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--r-sm);
  color: var(--fg);
  padding: 6px 8px;
  font-size: 12px;
}
.warn {
  color: var(--muted);
  font-size: 12px;
  border: 1px dashed var(--line);
  border-radius: var(--r-sm);
  padding: 8px;
  margin-bottom: 8px;
}
.search {
  width: 100%;
  margin-bottom: 8px;
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--r-sm);
  padding: 8px 10px;
  color: var(--fg);
  font-size: 13px;
}
.stories {
  list-style: none;
  padding: 0;
  margin: 0 0 10px;
  max-height: 200px;
  overflow: auto;
}
.story-link {
  width: 100%;
  text-align: left;
  background: none;
  border: none;
  color: var(--fg);
  padding: 6px 0;
  cursor: pointer;
  font-size: 12px;
}
.tag {
  font-size: 10px;
  color: var(--dim);
  margin-right: 6px;
}
.btn {
  background: var(--accent);
  color: var(--ink);
  border: none;
  border-radius: var(--r-sm);
  padding: 8px 14px;
  font-weight: 600;
  cursor: pointer;
  font-size: 13px;
}
.btn.ghost {
  background: var(--surface2);
  color: var(--fg);
  border: 1px solid var(--line);
}
.btn.sm {
  padding: 6px 10px;
  font-size: 12px;
}
.error {
  color: var(--del);
}
.meta {
  color: var(--muted);
}
.story-form {
  margin-top: 12px;
  border-top: 1px solid var(--line);
  padding-top: 12px;
}
</style>
