<script setup lang="ts">
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
type SkillGroup = { id: string; label: string; items: string[] }
type CV = {
  person: string
  contact: {
    name: string
    headline: string
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
  education: { id: string; school: string; degree: string; year: string }[]
  achievements: { id: string; text: string }[]
}

type TreeNode = {
  id: string
  label: string
  section?: string
  fieldId?: string
  kind: 'section' | 'field'
}

const fallbackCV: CV = {
  person: 'farhan-yuda-pahlevi',
  contact: {
    name: 'Farhan Yuda Pahlevi',
    headline: 'Senior Software Engineer',
    email: 'farhan.y.pahlevi@gmail.com',
    site: 'pahlevikun.id',
    linkedin: 'linkedin.com/in/pahlevikun',
    github: '',
    location: 'Bogor, West Java, Indonesia',
    timezone_line: 'GMT+7 · available for EU and US Eastern overlap',
  },
  summary: {
    id: 'summary',
    text: 'Senior fullstack engineer with 8+ years building merchant and platform systems at scale.',
  },
  skills: [
    { id: 'skills.backend', label: 'Backend', items: ['Go', 'Ruby on Rails', 'TypeScript'] },
  ],
  experience: [
    {
      id: 'exp.1',
      company: 'GoTo Group',
      team: 'GoFood Merchant',
      title: 'Senior Software Engineer, Fullstack',
      location: 'Jakarta, Indonesia',
      start: '2023-03',
      end: null,
      bullets: [
        {
          id: 'exp.1.b1',
          text: 'Architected ABAC authorization across 6+ microservices, cutting policy rollout time by 40%.',
        },
      ],
    },
  ],
  education: [{ id: 'edu.1', school: 'Universitas Gunadarma', degree: 'B.S. Computer Science', year: '2018' }],
  achievements: [{ id: 'ach.1', text: 'Mentored 3 engineers to promotion readiness.' }],
}

const cv = ref<CV>(fallbackCV)
const selectedId = ref('summary')
const provenanceSource = 'fixtures/resume_ats_v2.tex'
const loading = ref(false)

const tree = computed<TreeNode[]>(() => {
  const nodes: TreeNode[] = [
    { id: 'summary', label: 'Summary', fieldId: 'summary', kind: 'field' },
    { id: 'sec-skills', label: 'Technical Skills', kind: 'section' },
  ]
  for (const sg of cv.value.skills) {
    nodes.push({ id: sg.id, label: sg.label, fieldId: sg.id, kind: 'field' })
  }
  nodes.push({ id: 'sec-exp', label: 'Work Experience', kind: 'section' })
  for (const exp of cv.value.experience) {
    const label = exp.team ? `${exp.team}` : exp.company
    for (const b of exp.bullets) {
      nodes.push({
        id: b.id,
        label: `${label} · bullet`,
        fieldId: b.id,
        kind: 'field',
      })
    }
  }
  nodes.push({ id: 'sec-edu', label: 'Education', kind: 'section' })
  for (const ed of cv.value.education) {
    nodes.push({ id: ed.id, label: ed.school, fieldId: ed.id, kind: 'field' })
  }
  nodes.push({ id: 'sec-ach', label: 'Achievements', kind: 'section' })
  for (const ach of cv.value.achievements) {
    nodes.push({ id: ach.id, label: 'Achievement', fieldId: ach.id, kind: 'field' })
  }
  return nodes
})

const selectedField = computed(() => {
  const id = selectedId.value
  if (id === 'summary') return { title: 'Summary', fieldId: 'summary', value: cv.value.summary.text }
  const skill = cv.value.skills.find((s) => s.id === id)
  if (skill) return { title: `Skills · ${skill.label}`, fieldId: skill.id, value: skill.items.join(', ') }
  for (const exp of cv.value.experience) {
    const bullet = exp.bullets.find((b) => b.id === id)
    if (bullet) {
      return {
        title: `Experience · ${exp.team || exp.company}`,
        fieldId: bullet.id,
        value: bullet.text,
      }
    }
  }
  const ed = cv.value.education.find((e) => e.id === id)
  if (ed) return { title: 'Education', fieldId: ed.id, value: `${ed.degree}, ${ed.school} (${ed.year})` }
  const ach = cv.value.achievements.find((a) => a.id === id)
  if (ach) return { title: 'Achievement', fieldId: ach.id, value: ach.text }
  return { title: 'Contact', fieldId: 'contact.name', value: cv.value.contact.name }
})

function selectNode(node: TreeNode) {
  if (node.kind === 'section') return
  selectedId.value = node.id
}

async function loadCV() {
  loading.value = true
  try {
    cv.value = await $fetch<CV>('/api/cv/master')
  } catch {
    cv.value = fallbackCV
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCV()
})
</script>

<template>
  <main class="cv-page">
    <header class="top">
      <h1>Structured CV</h1>
      <span class="chip mono">{{ cv.person }}</span>
    </header>

    <div class="cv-layout">
      <aside class="tree" aria-label="CV sections">
        <p class="tree-label">Sections</p>
        <button
          v-for="node in tree"
          :key="node.id"
          type="button"
          class="tree-item"
          :class="{ on: selectedId === node.id, section: node.kind === 'section' }"
          @click="selectNode(node)"
        >
          {{ node.label }}
        </button>
      </aside>

      <section class="editor" aria-label="Field editor">
        <p v-if="loading" class="meta">Loading…</p>
        <div v-else class="field">
          <h2>
            {{ selectedField.title }}
            <span class="field-id">field id: {{ selectedField.fieldId }}</span>
          </h2>
          <textarea class="inp" rows="6" :value="selectedField.value" readonly />
          <div class="contact-block">
            <h3>Contact</h3>
            <div class="row2">
              <input class="inp" :value="cv.contact.name" readonly />
              <input class="inp" :value="cv.contact.headline" readonly />
            </div>
            <input class="inp" :value="cv.contact.email" readonly />
          </div>
        </div>
      </section>

      <aside class="meta-panel" aria-label="Provenance">
        <p class="tree-label">Provenance</p>
        <ul class="prov">
          <li>from {{ provenanceSource }}</li>
          <li>metrics from facts.json</li>
        </ul>
        <p class="tree-label">History · this field</p>
        <div class="hist-placeholder">
          <p class="dim">Per-field history will appear here after tailoring (Phase 4).</p>
        </div>
      </aside>
    </div>
  </main>
</template>

<style scoped>
.cv-page {
  min-height: 100vh;
  background: #0b0e13;
  color: #e7eaf0;
  font-family: system-ui, sans-serif;
  display: flex;
  flex-direction: column;
}
.top {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #232a36;
  background: #11151d;
}
h1 {
  font-size: 1.2rem;
  margin: 0;
  font-weight: 600;
}
.chip {
  font-size: 0.75rem;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: #171c26;
  border: 1px solid #232a36;
  color: #8b95a5;
}
.mono {
  font-family: ui-monospace, monospace;
}
.cv-layout {
  flex: 1;
  display: flex;
  min-height: 0;
}
.tree {
  width: 250px;
  flex-shrink: 0;
  border-right: 1px solid #232a36;
  background: #11151d;
  padding: 0.75rem;
  overflow: auto;
}
.tree-label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #5c6675;
  font-weight: 600;
  margin: 0 0 0.5rem;
}
.tree-item {
  display: block;
  width: 100%;
  text-align: left;
  border: none;
  background: transparent;
  color: #8b95a5;
  font-size: 0.8rem;
  padding: 0.4rem 0.5rem;
  border-radius: 7px;
  cursor: pointer;
}
.tree-item.on {
  background: #171c26;
  color: #e7eaf0;
}
.tree-item.section {
  font-weight: 600;
  color: #5c6675;
  text-transform: uppercase;
  font-size: 0.68rem;
  letter-spacing: 0.05em;
  cursor: default;
  margin-top: 0.5rem;
}
.editor {
  flex: 1;
  min-width: 0;
  padding: 1rem 1.25rem;
  overflow: auto;
}
.field h2 {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #5c6675;
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}
.field-id {
  margin-left: auto;
  font-family: ui-monospace, monospace;
  font-size: 0.65rem;
  color: #3fd0be;
  text-transform: none;
  letter-spacing: 0;
}
.inp {
  width: 100%;
  box-sizing: border-box;
  background: #171c26;
  border: 1px solid #232a36;
  border-radius: 7px;
  padding: 0.55rem 0.65rem;
  color: #e7eaf0;
  font-size: 0.85rem;
  line-height: 1.5;
  margin-top: 0.5rem;
}
.row2 {
  display: flex;
  gap: 0.5rem;
}
.row2 .inp {
  flex: 1;
}
.contact-block {
  margin-top: 1.25rem;
  padding-top: 1rem;
  border-top: 1px solid #232a36;
}
.contact-block h3 {
  font-size: 0.7rem;
  text-transform: uppercase;
  color: #5c6675;
  margin: 0 0 0.35rem;
}
.meta-panel {
  width: 280px;
  flex-shrink: 0;
  border-left: 1px solid #232a36;
  background: #11151d;
  padding: 0.75rem 1rem;
}
.prov {
  list-style: none;
  padding: 0;
  margin: 0 0 1rem;
  font-size: 0.8rem;
  color: #8b95a5;
}
.prov li {
  padding: 0.25rem 0;
}
.hist-placeholder .dim {
  font-size: 0.8rem;
  color: #5c6675;
  margin: 0;
}
.meta {
  color: #8b95a5;
}
</style>
