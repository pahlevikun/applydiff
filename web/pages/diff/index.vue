<script setup lang="ts">
const jobId = useJobIdFromRoute()
const { t } = useI18n()

if (jobId.value) {
  await navigateTo(jobDetailPath(jobId.value, 'diff'), { replace: true })
}
</script>

<template>
  <div v-if="!jobId" class="page muted">
    {{ t('diff.emptyLead') }}
    <p class="hint">{{ t('diff.emptyHint') }}</p>
  </div>
  <p v-else class="muted page">Redirecting…</p>
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
