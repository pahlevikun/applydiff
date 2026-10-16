<script setup lang="ts">
const route = useRoute()
const jobId = computed(() => route.params.id as string)

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

type CvDoc = {
  contact: { name: string; headline: string; email: string; location?: string }
  summary: { id: string; text: string }
  skills: { id: string; label: string; items: string[] }[]
  experience: { id: string; company: string; title: string; start: string; bullets: { id: string; text: string }[] }[]
}

const proposal = ref<Proposal | null>(null)
const master = ref<CvDoc | null>(null)
const selectedHunkId = ref<string | null>(null)
const error = ref('')
const loading = ref(false)

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

async function loadAll() {
  error.value = ''
  loading.value = true
  try {
    const [propRes, masterRes] = await Promise.all([
      fetch(`/api/tailor/${jobId.value}`),
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
    if (!selectedHunkId.value && proposal.value?.hunks.length) {
      selectedHunkId.value = proposal.value.hunks[0].id
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function generate() {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`/api/tailor/${jobId.value}`, { method: 'POST' })
    if (!res.ok) throw new Error(await res.text())
    proposal.value = await res.json()
    if (proposal.value?.hunks.length) {
      selectedHunkId.value = proposal.value.hunks[0].id
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function setStatus(hunk: Hunk, action: 'accept' | 'reject') {
  const res = await fetch(`/api/hunks/${hunk.id}/${action}`, { method: 'POST' })
  if (!res.ok) {
    error.value = await res.text()
    return
  }
  proposal.value = await res.json()
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

onMounted(loadAll)
watch(jobId, loadAll)
</script>

<template>
  <div class="page">
    <header class="topbar">
      <NuxtLink to="/applications" class="brand">← ApplyDiff</NuxtLink>
      <span class="muted">Job {{ jobId }}</span>
      <div class="grow" />
      <button class="btn ghost" :disabled="loading" @click="generate">Tailor for this JD</button>
    </header>

    <p v-if="error" class="err">{{ error }}</p>

    <div v-if="!proposal && !loading" class="empty">
      <p>No tailoring proposal yet.</p>
      <button class="btn" @click="generate">Generate hunks</button>
    </div>

    <div v-else-if="master" class="rows">
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

      <div class="sheetwrap">
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
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg, #12151a);
  color: var(--fg, #e9eaed);
}
.topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid #2a2f37;
  background: #191d23;
}
.brand {
  color: #6bc9ba;
  text-decoration: none;
  font-weight: 600;
}
.grow {
  flex: 1;
}
.btn {
  background: #6bc9ba;
  color: #05221f;
  border: none;
  border-radius: 7px;
  padding: 8px 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn.ghost {
  background: transparent;
  color: #e9eaed;
  border: 1px solid #353b44;
}
.rows {
  flex: 1;
  display: flex;
  min-height: 0;
}
.changes {
  width: 264px;
  border-right: 1px solid #2a2f37;
  background: #191d23;
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
  color: #e9eaed;
  cursor: pointer;
}
.ci.on {
  background: #20252c;
  border-color: #2a2f37;
}
.ci .s {
  font-size: 10px;
  color: #6e7785;
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
  color: #d6ad6a;
}
.st.accepted {
  color: #79c392;
}
.sheetwrap {
  flex: 1;
  overflow: auto;
  background: #101317;
  display: flex;
  justify-content: center;
  padding: 16px 0;
}
.sheet {
  width: min(840px, 100%);
  background: #fbfaf7;
  color: #1c1f25;
  box-shadow: 0 24px 70px #000a;
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
  color: #606772;
  margin-top: 4px;
}
.cvc {
  font-size: 11.5px;
  color: #606772;
  margin-top: 6px;
}
.cvsec {
  font-weight: 700;
  font-size: 10.5px;
  letter-spacing: 1.1px;
  text-transform: uppercase;
  border-bottom: 1.4px solid #1c1f25;
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
  color: #57a97b;
}
.chg::before {
  content: '';
  position: absolute;
  left: 12px;
  top: 1px;
  bottom: 1px;
  width: 2px;
  background: #6bc9ba88;
  border-radius: 2px;
}
ins {
  background: #6bc9ba20;
  border-bottom: 1px solid #a0d5cb;
  text-decoration: none;
}
.ats {
  font-size: 10px;
  color: #2c554d;
  background: #6bc9ba30;
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
  border: 1px solid #d8d4ca;
  cursor: pointer;
  background: #fff;
}
.ctl .ok {
  color: #1a5c3a;
  border-color: #79c392;
}
.ctl .no {
  color: #8a3f32;
  border-color: #db8c7c;
}
.rait {
  width: 290px;
  border-left: 1px solid #2a2f37;
  background: #191d23;
  padding: 14px;
  overflow: auto;
}
.lbl {
  font-size: 10.5px;
  text-transform: uppercase;
  letter-spacing: 0.9px;
  color: #6e7785;
  font-weight: 600;
  margin-bottom: 8px;
}
.ev {
  font-size: 11px;
  margin: 4px 0;
  font-family: ui-monospace, monospace;
}
.muted {
  color: #9ba3ae;
}
.err {
  color: #db8c7c;
  padding: 8px 16px;
}
.empty {
  padding: 40px;
  text-align: center;
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
  color: #606772;
  font-weight: 400;
}
.accepted {
  opacity: 0.85;
}
</style>
