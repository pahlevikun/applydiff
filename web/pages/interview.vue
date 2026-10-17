<script setup lang="ts">
type Job = { id: string; title: string; company: string }

const jobs = ref<Job[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const rows = await $fetch<{ job: Job }[]>('/api/jobs')
    jobs.value = rows.map((r) => r.job)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="page">
    <h1>Interview prep</h1>
    <p class="muted">Pick a job to open kit, mock runner, rubric, and story bank. No auto-submit.</p>
    <p v-if="loading" class="dim">Loading jobs…</p>
    <ul v-else class="jobs">
      <li v-for="j in jobs" :key="j.id">
        <NuxtLink :to="`/interview/${encodeURIComponent(j.id)}`" class="job-link">
          {{ j.title }} <span class="dim">@ {{ j.company }}</span>
        </NuxtLink>
      </li>
      <li v-if="!jobs.length" class="dim">No jobs yet — run a scan from Jobs.</li>
    </ul>
  </div>
</template>

<style scoped>
.page {
  max-width: 960px;
  margin: 0 auto;
  padding: 24px;
}
h1 {
  font-size: 1.25rem;
  margin: 0 0 8px;
}
.muted {
  color: var(--muted);
  margin-bottom: 16px;
}
.jobs {
  list-style: none;
  padding: 0;
  margin: 0;
}
.job-link {
  display: block;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: var(--r-sm);
  margin-bottom: 8px;
  text-decoration: none;
  color: var(--fg);
  background: var(--surface);
}
.job-link:hover {
  border-color: var(--accent-dim);
}
.dim {
  color: var(--dim);
}
</style>
