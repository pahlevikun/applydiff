<script setup lang="ts">
const jobId = useJobIdFromRoute()
</script>

<template>
  <DiffJobWorkspace v-if="jobId" :key="jobId" :job-id="jobId" />
  <div v-else class="page muted">
    Open <strong>Diff</strong> from a job row on <NuxtLink to="/jobs">Jobs</NuxtLink>, or after you
    <strong>Tailor</strong> on the Match screen.
    <p class="hint">
      Diff shows CV hunks once tailoring has run. You can also click <strong>Generate hunks</strong> here if you already
      shortlisted the role.
    </p>
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
