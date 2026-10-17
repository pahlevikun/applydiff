<script setup lang="ts">
const jobId = useJobIdFromRoute()
const { t } = useI18n()
</script>

<template>
  <MatchJobWorkspace v-if="jobId" :key="jobId" :job-id="jobId" />
  <div v-else class="page muted">
    {{ t('match.emptyLead') }}
    <p class="hint">{{ t('match.emptyHint') }}</p>
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
</style>
