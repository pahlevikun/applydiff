<script setup lang="ts">
const jobId = useJobIdFromRoute()
</script>

<template>
  <MatchJobWorkspace v-if="jobId" :key="jobId" :job-id="jobId" />
  <div v-else class="page muted">
    Select a job from <NuxtLink to="/jobs">Jobs</NuxtLink> and click <strong>Match</strong>, or open
    <NuxtLink to="/applications">Applications</NuxtLink> for roles already in your pipeline.
    <p class="hint">Flow: Jobs → Match → Shortlist → Tailor → Diff → Apply.</p>
  </div>
</template>

<style scoped>
.page {
  padding: 24px;
}
.muted {
  color: var(--muted);
}
.hint {
  margin-top: 0.75rem;
  font-size: 0.9rem;
}
a {
  color: var(--accent);
}
</style>
