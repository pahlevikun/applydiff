<script setup lang="ts">
const jobId = useJobIdFromRoute()
const { t } = useI18n()

if (jobId.value) {
  await navigateTo(jobDetailPath(jobId.value, 'fit'), { replace: true })
}
</script>

<template>
  <div v-if="!jobId" class="page muted">
    {{ t('match.emptyLead') }}
    <p class="hint">{{ t('match.emptyHint') }}</p>
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
