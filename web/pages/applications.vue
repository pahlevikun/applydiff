<script setup lang="ts">
type ApplicationRow = {
  job_id: string
  company: string
  title: string
  url: string
  status: string
  follow_up_note: string
}

type ExportResult = {
  dir: string
  cv_tex: string
  pdf_compiled: boolean
  compile_note?: string
}

const rows = ref<ApplicationRow[]>([])
const error = ref('')
const exporting = ref<string | null>(null)
const lastExport = ref<ExportResult | null>(null)

async function load() {
  error.value = ''
  try {
    rows.value = await $fetch<ApplicationRow[]>('/api/applications')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load applications'
  }
}

async function runExport(jobId: string) {
  exporting.value = jobId
  error.value = ''
  try {
    lastExport.value = await $fetch<ExportResult>(`/api/export/${jobId}`, { method: 'POST' })
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Export failed'
  } finally {
    exporting.value = null
  }
}

onMounted(load)
</script>

<template>
  <div class="wrap">
    <header class="topbar">
      <NuxtLink to="/jobs" class="brand"><span class="dot" /> ApplyDiff</NuxtLink>
      <nav class="nav">
        <NuxtLink to="/jobs">Jobs</NuxtLink>
        <NuxtLink to="/applications" class="on">Applications</NuxtLink>
        <NuxtLink to="/cv">My CV</NuxtLink>
        <NuxtLink to="/sources">Sources</NuxtLink>
      </nav>
    </header>

    <p v-if="error" class="err">{{ error }}</p>
    <p v-if="lastExport?.compile_note" class="note">{{ lastExport.compile_note }}</p>

    <ul class="list">
      <li v-for="row in rows" :key="row.job_id" class="item">
        <div class="main">
          <NuxtLink :to="`/diff/${row.job_id}`" class="title">{{ row.title }}</NuxtLink>
          <span class="muted">{{ row.company }}</span>
          <span class="chip">{{ row.status }}</span>
        </div>
        <div class="actions">
          <NuxtLink :to="`/diff/${row.job_id}`" class="btn ghost">Diff</NuxtLink>
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
    <p v-if="!rows.length && !error" class="muted empty">No applications yet — scan jobs and match roles first.</p>
  </div>
</template>

<style scoped>
.wrap {
  min-height: 100vh;
  background: var(--bg, #12151a);
  color: var(--fg, #e9eaed);
}
.topbar {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--line, #2a2f37);
}
.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--accent, #6bc9ba);
  text-decoration: none;
  font-weight: 600;
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent, #6bc9ba);
}
.nav {
  display: flex;
  gap: 14px;
}
.nav a {
  color: var(--muted, #9ba3ae);
  text-decoration: none;
  font-size: 13px;
}
.nav a.on,
.nav a:hover {
  color: var(--fg, #e9eaed);
}
.list {
  list-style: none;
  margin: 0;
  padding: 16px;
  max-width: 960px;
}
.item {
  border: 1px solid var(--line, #2a2f37);
  border-radius: var(--r-sm, 7px);
  padding: 12px 14px;
  margin-bottom: 10px;
  background: var(--surface, #191d23);
}
.main {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: baseline;
}
.title {
  font-weight: 600;
  color: inherit;
  text-decoration: none;
}
.title:hover {
  color: var(--accent, #6bc9ba);
}
.muted {
  color: var(--muted, #9ba3ae);
}
.chip {
  margin-left: auto;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--surface2, #20252c);
  text-transform: lowercase;
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
  border: 1px solid var(--line2, #353b44);
  background: var(--accent-dim, #2c554d);
  color: var(--fg, #e9eaed);
  cursor: pointer;
}
.btn.ghost {
  background: transparent;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
}
.btn:disabled {
  opacity: 0.6;
  cursor: default;
}
.follow {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--amber, #d6ad6a);
}
.err {
  color: var(--del, #db8c7c);
  padding: 8px 16px;
}
.note {
  padding: 8px 16px;
  color: var(--muted, #9ba3ae);
  font-size: 12px;
}
.empty {
  padding: 24px;
}
</style>
