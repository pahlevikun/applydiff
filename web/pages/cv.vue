<script setup lang="ts">
import rawMaster from '../../cv/master_cv.json'

type Bullet = { id: string; text: string }
type Experience = {
  id: string
  company: string
  team: string
  title: string
  location: string
  start: string
  end: string | null
  bullets: Bullet[]
}
type Education = {
  id: string
  school: string
  degree: string
  year: string
  location?: string
  bullets?: Bullet[]
}
type SkillGroup = { id: string; label: string; items: string[] }
type CV = {
  person: string
  contact: {
    name: string
    headline: string
    subtitle?: string
    stack_line?: string
    email: string
    site: string
    linkedin: string
    github: string
    location: string
    timezone_line: string
  }
  summary: { id: string; text: string }
  skills: SkillGroup[]
  experience: Experience[]
  education: Education[]
  achievements: { id: string; text: string }[]
}

type CvSection = 'contact' | 'summary' | 'skills' | 'experience' | 'education' | 'achievements'

const sectionNav: { id: CvSection; label: string }[] = [
  { id: 'contact', label: 'Header & contact' },
  { id: 'summary', label: 'Professional Summary' },
  { id: 'skills', label: 'Technical Skills' },
  { id: 'experience', label: 'Work Experience' },
  { id: 'education', label: 'Education' },
  { id: 'achievements', label: 'Achievements & Certifications' },
]

function templateCV(): CV {
  const { _comment: _, ...doc } = rawMaster as CV & { _comment?: string }
  return JSON.parse(JSON.stringify(doc)) as CV
}

const cv = ref<CV>(templateCV())
const section = ref<CvSection>('contact')
const loading = ref(false)
const saving = ref(false)
const saveError = ref<string | null>(null)
const masterSaved = ref(false)
const usingFallback = ref(true)
const sourcePaths = ref<{ json: string; tex: string } | null>(null)
const cvSource = ref<'database' | 'file' | 'template'>('template')

function nextNumericSuffix(prefix: string, ids: string[]): string {
  let n = 1
  const re = new RegExp(`^${prefix}\\.(\\d+)$`)
  const used = new Set(ids.map((id) => {
    const m = id.match(re)
    return m ? Number(m[1]) : 0
  }))
  while (used.has(n)) n++
  return `${prefix}.${n}`
}

function nextBulletId(parentId: string, bullets: Bullet[]): string {
  let n = 1
  const re = new RegExp(`^${parentId}\\.b(\\d+)$`)
  for (const b of bullets) {
    const m = b.id.match(re)
    if (m) n = Math.max(n, Number(m[1]) + 1)
  }
  return `${parentId}.b${n}`
}

function syncPersonSlug() {
  const slug = cv.value.contact.name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  cv.value.person = slug || cv.value.person || 'person'
}

function skillText(sg: SkillGroup): string {
  return sg.items.join('\n')
}

function setSkillText(sg: SkillGroup, raw: string) {
  const line = raw.trim()
  sg.items = line ? [line] : []
}

function endDisplay(end: string | null): string {
  return end ?? ''
}

function setEnd(exp: Experience, raw: string) {
  const v = raw.trim()
  exp.end = v ? v : null
}

function addExperience() {
  const id = nextNumericSuffix('exp', cv.value.experience.map((e) => e.id))
  cv.value.experience.unshift({
    id,
    company: '',
    team: '',
    title: '',
    location: '',
    start: '',
    end: null,
    bullets: [{ id: `${id}.b1`, text: '' }],
  })
}

function removeExperience(idx: number) {
  cv.value.experience.splice(idx, 1)
}

function addExpBullet(exp: Experience) {
  exp.bullets.push({ id: nextBulletId(exp.id, exp.bullets), text: '' })
}

function removeExpBullet(exp: Experience, idx: number) {
  exp.bullets.splice(idx, 1)
}

function addSkillGroup() {
  let n = 1
  while (cv.value.skills.some((s) => s.id === `skills.cat${n}`)) n++
  cv.value.skills.push({ id: `skills.cat${n}`, label: 'New category', items: [''] })
}

function removeSkillGroup(idx: number) {
  cv.value.skills.splice(idx, 1)
}

function addEducation() {
  const id = nextNumericSuffix('edu', cv.value.education.map((e) => e.id))
  cv.value.education.push({
    id,
    school: '',
    degree: '',
    year: '',
    location: '',
    bullets: [],
  })
}

function removeEducation(idx: number) {
  cv.value.education.splice(idx, 1)
}

function addEduBullet(ed: Education) {
  if (!ed.bullets) ed.bullets = []
  ed.bullets.push({ id: nextBulletId(ed.id, ed.bullets), text: '' })
}

function removeEduBullet(ed: Education, idx: number) {
  ed.bullets?.splice(idx, 1)
}

function addAchievement() {
  const id = nextNumericSuffix('ach', cv.value.achievements.map((a) => a.id))
  cv.value.achievements.push({ id, text: '' })
}

function removeAchievement(idx: number) {
  cv.value.achievements.splice(idx, 1)
}

type MasterCVPayload = CV & {
  _source?: 'database' | 'file'
  _master_json?: string
  _master_tex?: string
}

function applyMasterPayload(payload: MasterCVPayload) {
  const {
    _source,
    _master_json: masterJson,
    _master_tex: masterTex,
    ...doc
  } = payload
  cv.value = doc as CV
  if (masterJson && masterTex) {
    sourcePaths.value = { json: masterJson, tex: masterTex }
  }
  if (_source === 'database') {
    cvSource.value = 'database'
    masterSaved.value = true
    usingFallback.value = false
  } else if (_source === 'file') {
    cvSource.value = 'file'
    masterSaved.value = false
    usingFallback.value = false
  }
}

async function loadCV() {
  loading.value = true
  saveError.value = null
  try {
    const payload = await $fetch<MasterCVPayload>('/api/cv/master')
    applyMasterPayload(payload)
  } catch {
    cv.value = templateCV()
    masterSaved.value = false
    usingFallback.value = true
    cvSource.value = 'template'
  } finally {
    loading.value = false
  }
}

const { track, log, actionErrorMessage } = useActivity()

function loadTemplate() {
  cv.value = templateCV()
  usingFallback.value = true
  masterSaved.value = false
  cvSource.value = 'template'
  log('Reset CV from resume template', 'info')
}

async function saveMaster() {
  saving.value = true
  saveError.value = null
  syncPersonSlug()
  try {
    const res = await track('Save master CV', () =>
      $fetch<{ master_json: string; master_tex: string }>('/api/cv/master', {
        method: 'PUT',
        body: cv.value,
      }),
    )
    sourcePaths.value = { json: res.master_json, tex: res.master_tex }
    masterSaved.value = true
    usingFallback.value = false
    cvSource.value = 'database'
  } catch (e) {
    saveError.value = actionErrorMessage(e)
  } finally {
    saving.value = false
  }
}

onMounted(loadCV)
</script>

<template>
  <main class="cv-page">
    <header class="top">
      <div class="head">
        <h1>CV editor</h1>
      </div>
      <span class="chip mono">{{ cv.person }}</span>
      <button type="button" class="btn ghost" :disabled="loading" @click="loadTemplate">Reset template</button>
      <button type="button" class="save-btn" :disabled="saving || loading" @click="saveMaster">
        {{ saving ? 'Saving…' : 'Save master' }}
      </button>
    </header>

    <p v-if="cvSource === 'template'" class="banner">
      Loaded bundled template — save to sync database, <span class="mono">cv/master_cv.json</span>, and
      <span class="mono">cv/master_cv.tex</span>.
    </p>
    <p v-else-if="cvSource === 'file'" class="banner">
      Loaded from <span class="mono">{{ sourcePaths?.json ?? 'cv/master_cv.json' }}</span> — save to sync database and TeX.
    </p>
    <p v-if="saveError" class="banner error">{{ saveError }}</p>
    <p v-if="masterSaved" class="banner ok">
      Master CV saved
      <template v-if="sourcePaths">
        (<span class="mono">{{ sourcePaths.json }}</span>,
        <span class="mono">{{ sourcePaths.tex }}</span>, database).
      </template>
      <template v-else>(database).</template>
    </p>

    <div v-if="loading" class="loading">Loading…</div>
    <div v-else class="cv-layout">
      <nav class="side" aria-label="Resume sections">
        <button
          v-for="item in sectionNav"
          :key="item.id"
          type="button"
          class="nav-item"
          :class="{ on: section === item.id }"
          @click="section = item.id"
        >
          {{ item.label }}
        </button>
      </nav>

      <form class="form" @submit.prevent="saveMaster">
        <!-- Header & contact (resume lines 19–28) -->
        <section v-show="section === 'contact'" class="panel">
          <h2 class="sec-title">Header & contact</h2>
          <label class="field">
            <span>Name</span>
            <input v-model="cv.contact.name" type="text" class="inp" autocomplete="name" />
          </label>
          <label class="field">
            <span>Title @ company</span>
            <input v-model="cv.contact.headline" type="text" class="inp" />
          </label>
          <label class="field">
            <span>Role tagline</span>
            <input v-model="cv.contact.subtitle" type="text" class="inp" />
          </label>
          <label class="field">
            <span>Stack line</span>
            <input v-model="cv.contact.stack_line" type="text" class="inp" />
          </label>
          <div class="row2">
            <label class="field">
              <span>Email</span>
              <input v-model="cv.contact.email" type="email" class="inp" />
            </label>
            <label class="field">
              <span>Site</span>
              <input v-model="cv.contact.site" type="text" class="inp" />
            </label>
          </div>
          <div class="row2">
            <label class="field">
              <span>GitHub</span>
              <input v-model="cv.contact.github" type="text" class="inp" />
            </label>
            <label class="field">
              <span>LinkedIn</span>
              <input v-model="cv.contact.linkedin" type="text" class="inp" />
            </label>
          </div>
          <label class="field">
            <span>Location</span>
            <input v-model="cv.contact.location" type="text" class="inp" />
          </label>
        </section>

        <section v-show="section === 'summary'" class="panel">
          <h2 class="sec-title">Professional Summary</h2>
          <label class="field">
            <span>Summary paragraph</span>
            <textarea v-model="cv.summary.text" class="inp tall" rows="14" />
            <span class="hint mono">field id: {{ cv.summary.id }}</span>
          </label>
        </section>

        <section v-show="section === 'skills'" class="panel">
          <div class="panel-head">
            <h2 class="sec-title">Technical Skills</h2>
            <button type="button" class="btn sm ghost" @click="addSkillGroup">Add category</button>
          </div>
          <p class="hint">One block per category, as in the LaTeX resume item list.</p>
          <div v-for="(sg, si) in cv.skills" :key="sg.id" class="skill-card">
            <div class="card-head">
              <label class="field grow">
                <span>Category label</span>
                <input v-model="sg.label" type="text" class="inp" />
              </label>
              <button type="button" class="btn sm danger" @click="removeSkillGroup(si)">Remove</button>
            </div>
            <label class="field">
              <span>Skills text</span>
              <textarea
                class="inp"
                rows="3"
                :value="skillText(sg)"
                @input="setSkillText(sg, ($event.target as HTMLTextAreaElement).value)"
              />
              <span class="hint mono">{{ sg.id }}</span>
            </label>
          </div>
        </section>

        <section v-show="section === 'experience'" class="panel">
          <div class="panel-head">
            <h2 class="sec-title">Work Experience</h2>
            <button type="button" class="btn sm ghost" @click="addExperience">Add role</button>
          </div>
          <article v-for="(exp, ei) in cv.experience" :key="exp.id" class="role-card">
            <div class="card-head">
              <label class="field grow">
                <span>Job title</span>
                <input v-model="exp.title" type="text" class="inp" />
              </label>
              <button type="button" class="btn sm danger" @click="removeExperience(ei)">Remove role</button>
            </div>
            <div class="row2">
              <label class="field">
                <span>Company</span>
                <input v-model="exp.company" type="text" class="inp" />
              </label>
              <label class="field">
                <span>Team / platform</span>
                <input v-model="exp.team" type="text" class="inp" />
              </label>
            </div>
            <div class="row3">
              <label class="field">
                <span>Start (YYYY-MM)</span>
                <input v-model="exp.start" type="text" class="inp" />
              </label>
              <label class="field">
                <span>End (blank = Present)</span>
                <input
                  type="text"
                  class="inp"
                  :value="endDisplay(exp.end)"
                  @input="setEnd(exp, ($event.target as HTMLInputElement).value)"
                />
              </label>
              <label class="field">
                <span>Location</span>
                <input v-model="exp.location" type="text" class="inp" />
              </label>
            </div>
            <p class="subhead">Bullets</p>
            <div v-for="(b, bi) in exp.bullets" :key="b.id" class="bullet-row">
              <textarea v-model="b.text" class="inp" rows="3" :aria-label="`Bullet ${bi + 1}`" />
              <button type="button" class="btn sm danger" @click="removeExpBullet(exp, bi)">Remove</button>
              <span class="hint mono">{{ b.id }}</span>
            </div>
            <button type="button" class="btn sm ghost" @click="addExpBullet(exp)">Add bullet</button>
          </article>
        </section>

        <section v-show="section === 'education'" class="panel">
          <div class="panel-head">
            <h2 class="sec-title">Education</h2>
            <button type="button" class="btn sm ghost" @click="addEducation">Add program</button>
          </div>
          <article v-for="(ed, edi) in cv.education" :key="ed.id" class="role-card">
            <div class="card-head">
              <span class="hint mono">{{ ed.id }}</span>
              <button type="button" class="btn sm danger" @click="removeEducation(edi)">Remove program</button>
            </div>
            <div class="row2">
              <label class="field">
                <span>Degree</span>
                <input v-model="ed.degree" type="text" class="inp" />
              </label>
              <label class="field">
                <span>School</span>
                <input v-model="ed.school" type="text" class="inp" />
              </label>
            </div>
            <div class="row2">
              <label class="field">
                <span>Years</span>
                <input v-model="ed.year" type="text" class="inp" />
              </label>
              <label class="field">
                <span>Location</span>
                <input v-model="ed.location" type="text" class="inp" />
              </label>
            </div>
            <p class="subhead">Details</p>
            <div v-for="(b, bi) in ed.bullets ?? []" :key="b.id" class="bullet-row">
              <textarea v-model="b.text" class="inp" rows="2" />
              <button type="button" class="btn sm danger" @click="removeEduBullet(ed, bi)">Remove</button>
            </div>
            <button type="button" class="btn sm ghost" @click="addEduBullet(ed)">Add detail</button>
          </article>
        </section>

        <section v-show="section === 'achievements'" class="panel">
          <div class="panel-head">
            <h2 class="sec-title">Achievements &amp; Certifications</h2>
            <button type="button" class="btn sm ghost" @click="addAchievement">Add entry</button>
          </div>
          <div v-for="(ach, ai) in cv.achievements" :key="ach.id" class="bullet-row">
            <textarea v-model="ach.text" class="inp" rows="2" :aria-label="`Achievement ${ai + 1}`" />
            <button type="button" class="btn sm danger" @click="removeAchievement(ai)">Remove</button>
            <span class="hint mono">{{ ach.id }}</span>
          </div>
        </section>

        <footer class="form-foot">
          <button type="submit" class="save-btn wide" :disabled="saving">
            {{ saving ? 'Saving…' : 'Save master CV' }}
          </button>
        </footer>
      </form>
    </div>
  </main>
</template>

<style scoped>
.cv-page {
  min-height: 100vh;
  background: var(--bg);
  color: var(--fg);
  display: flex;
  flex-direction: column;
}
.top {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--line);
  background: var(--surface);
  flex-wrap: wrap;
}
.head h1 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
}
.lede {
  margin: 0.25rem 0 0;
  font-size: 12px;
  color: var(--muted);
}
.save-btn {
  margin-left: auto;
  padding: 0.45rem 0.9rem;
  border-radius: 8px;
  border: 1px solid var(--accent);
  background: var(--accent);
  color: var(--bg);
  font-weight: 600;
  cursor: pointer;
}
.btn {
  font-size: 0.85rem;
  padding: 0.45rem 0.75rem;
  border-radius: 8px;
  border: 1px solid var(--line2);
  background: var(--bg);
  color: var(--fg);
  cursor: pointer;
  font-family: inherit;
}
.btn.ghost {
  background: var(--surface);
}
.btn.sm {
  font-size: 0.75rem;
  padding: 0.35rem 0.55rem;
}
.btn.danger {
  border-color: var(--del);
  color: var(--del);
}
.save-btn:disabled,
.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.banner {
  margin: 0;
  padding: 0.65rem 1.25rem;
  background: var(--surface2);
  border-bottom: 1px solid var(--line);
  color: var(--muted);
  font-size: 0.9rem;
}
.banner.ok {
  color: var(--accent);
}
.banner.error {
  color: var(--del);
}
.chip {
  font-size: 0.75rem;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: var(--surface2);
  border: 1px solid var(--line);
  color: var(--muted);
}
.mono {
  font-family: ui-monospace, monospace;
}
.loading {
  padding: 2rem;
  color: var(--muted);
}
.cv-layout {
  flex: 1;
  display: flex;
  min-height: 0;
}
.side {
  width: 220px;
  flex-shrink: 0;
  border-right: 1px solid var(--line);
  background: var(--surface);
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.nav-item {
  text-align: left;
  border: none;
  background: transparent;
  color: var(--muted);
  font-size: 0.8rem;
  padding: 0.5rem 0.6rem;
  border-radius: var(--r-sm);
  cursor: pointer;
  font-family: inherit;
}
.nav-item.on {
  background: var(--surface2);
  color: var(--fg);
  font-weight: 500;
}
.nav-item:hover {
  color: var(--fg);
}
.form {
  flex: 1;
  overflow: auto;
  padding: 1rem 1.5rem 2rem;
  max-width: 52rem;
}
.panel {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.sec-title {
  margin: 0 0 0.25rem;
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: -0.01em;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  flex: 1;
  min-width: 0;
}
.field > span:first-child {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--dim);
  font-weight: 600;
}
.inp {
  width: 100%;
  box-sizing: border-box;
  background: var(--surface);
  border: 1px solid var(--line2);
  border-radius: var(--r-sm);
  padding: 0.5rem 0.65rem;
  color: var(--fg);
  font-size: 0.85rem;
  line-height: 1.45;
  font-family: inherit;
}
.inp.tall {
  min-height: 12rem;
}
.inp:focus {
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}
.row2 {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.row3 {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.75rem;
}
.hint {
  font-size: 0.68rem;
  color: var(--dim);
}
.skill-card,
.role-card {
  border: 1px solid var(--line);
  border-radius: var(--r);
  padding: 0.85rem 1rem;
  background: var(--surface);
}
.role-head {
  margin: 0 0 0.5rem;
  font-size: 0.9rem;
  font-weight: 600;
}
.subhead {
  margin: 0.5rem 0 0;
  font-size: 0.7rem;
  text-transform: uppercase;
  color: var(--dim);
  font-weight: 600;
}
.bullet-row {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  margin-bottom: 0.5rem;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.panel-head .sec-title {
  margin: 0;
}
.card-head {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
}
.card-head .grow {
  flex: 1;
  min-width: 12rem;
}
.form-foot {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--line);
}
.save-btn.wide {
  width: 100%;
  margin-left: 0;
}
</style>
