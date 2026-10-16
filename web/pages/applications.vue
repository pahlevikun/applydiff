<script setup lang="ts">
type Job = { id: string; company: string; title: string; status: string }

const jobs = ref<Job[]>([])
const error = ref('')

onMounted(async () => {
  try {
    const res = await fetch('/api/jobs')
    if (!res.ok) throw new Error(await res.text())
    jobs.value = await res.json()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  }
})
</script>

<template>
  <div class="wrap">
    <header class="topbar">
      <NuxtLink to="/" class="brand">ApplyDiff</NuxtLink>
      <span class="on">Applications</span>
    </header>
    <p v-if="error" class="err">{{ error }}</p>
    <ul class="list">
      <li v-for="j in jobs" :key="j.id">
        <NuxtLink :to="`/diff/${j.id}`" class="row">
          <strong>{{ j.title }}</strong>
          <span class="muted">{{ j.company }}</span>
          <span class="chip">{{ j.status }}</span>
        </NuxtLink>
      </li>
    </ul>
    <p v-if="!jobs.length && !error" class="muted empty">No jobs yet — run a scan from the API.</p>
  </div>
</template>

<style scoped>
.wrap {
  min-height: 100vh;
  background: #12151a;
  color: #e9eaed;
}
.topbar {
  display: flex;
  gap: 16px;
  padding: 14px 16px;
  border-bottom: 1px solid #2a2f37;
}
.brand {
  color: #6bc9ba;
  text-decoration: none;
  font-weight: 600;
}
.on {
  font-weight: 600;
}
.list {
  list-style: none;
  margin: 0;
  padding: 16px;
}
.row {
  display: flex;
  gap: 12px;
  align-items: baseline;
  padding: 12px;
  border: 1px solid #2a2f37;
  border-radius: 7px;
  margin-bottom: 8px;
  text-decoration: none;
  color: inherit;
}
.row:hover {
  background: #191d23;
}
.muted {
  color: #9ba3ae;
}
.chip {
  margin-left: auto;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: #20252c;
}
.err {
  color: #db8c7c;
  padding: 8px 16px;
}
.empty {
  padding: 24px;
}
</style>
