<script setup lang="ts">
type Job = {
  id: string
  company: string
  title: string
  url: string
  source: string
  posted_at: string
  remote: string
  region: string
  status: string
}

type ScanResult = {
  fetched: number
  deduped: number
  dead: number
  stored: number
  skipped: number
}

const jobs = ref<Job[]>([])
const loading = ref(false)
const scanning = ref(false)
const error = ref<string | null>(null)
const lastScan = ref<ScanResult | null>(null)

async function loadJobs() {
  loading.value = true
  error.value = null
  try {
    jobs.value = await $fetch<Job[]>('/api/jobs')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load jobs'
  } finally {
    loading.value = false
  }
}

async function runScan() {
  scanning.value = true
  error.value = null
  try {
    lastScan.value = await $fetch<ScanResult>('/api/scan', { method: 'POST' })
    await loadJobs()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Scan failed'
  } finally {
    scanning.value = false
  }
}

function formatPosted(iso: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso.slice(0, 10)
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

onMounted(() => {
  loadJobs()
})
</script>

<template>
  <main class="page">
    <header class="bar">
      <h1>Jobs</h1>
      <button class="scan" type="button" :disabled="scanning" @click="runScan">
        {{ scanning ? 'Scanning…' : 'Scan' }}
      </button>
    </header>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="lastScan" class="meta">
      Last scan: fetched {{ lastScan.fetched }}, stored {{ lastScan.stored }}, dead {{ lastScan.dead }}
    </p>
    <p v-if="loading && !jobs.length" class="meta">Loading…</p>

    <ul v-if="jobs.length" class="feed">
      <li v-for="job in jobs" :key="job.id" class="row">
        <div class="main">
          <a class="title" :href="job.url" target="_blank" rel="noopener noreferrer">{{ job.title }}</a>
          <div class="sub">
            <span>{{ job.company }}</span>
            <span class="dot">·</span>
            <span class="source">{{ job.source }}</span>
            <span v-if="job.remote === 'yes'" class="badge remote">Remote</span>
            <span class="posted">{{ formatPosted(job.posted_at) }}</span>
          </div>
        </div>
      </li>
    </ul>
    <p v-else-if="!loading" class="meta">No jobs yet. Run a scan to fetch listings.</p>
  </main>
</template>

<style scoped>
.page {
  max-width: 52rem;
  margin: 0 auto;
  padding: 1.5rem 1rem 3rem;
  font-family: system-ui, sans-serif;
  color: #e7eaf0;
  background: #0b0e13;
  min-height: 100vh;
}
.bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}
h1 {
  font-size: 1.35rem;
  margin: 0;
  font-weight: 600;
}
.scan {
  background: #3fd0be;
  color: #0b0e13;
  border: none;
  border-radius: 7px;
  padding: 0.5rem 1rem;
  font-weight: 600;
  cursor: pointer;
}
.scan:disabled {
  opacity: 0.6;
  cursor: wait;
}
.error {
  color: #f0553f;
}
.meta {
  color: #8b95a5;
  font-size: 0.9rem;
}
.feed {
  list-style: none;
  margin: 1rem 0 0;
  padding: 0;
  border: 1px solid #232a36;
  border-radius: 10px;
  overflow: hidden;
}
.row {
  border-bottom: 1px solid #232a36;
  padding: 0.85rem 1rem;
  background: #11151d;
}
.row:last-child {
  border-bottom: none;
}
.title {
  color: #e7eaf0;
  font-weight: 600;
  text-decoration: none;
}
.title:hover {
  color: #3fd0be;
}
.sub {
  margin-top: 0.35rem;
  font-size: 0.85rem;
  color: #8b95a5;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
}
.dot {
  opacity: 0.5;
}
.source {
  text-transform: capitalize;
}
.badge.remote {
  background: #0e2a17;
  color: #3fb950;
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
}
.posted {
  margin-left: auto;
  font-variant-numeric: tabular-nums;
}
</style>
