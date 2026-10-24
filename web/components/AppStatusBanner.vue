<script setup lang="ts">
const { offline, isStale, lastScanAt, loading, refresh } = useAppStatus()
const { t } = useI18n()

const staleLabel = computed(() => {
  if (!lastScanAt.value) return ''
  return lastScanAt.value.toLocaleString()
})
</script>

<template>
  <div v-if="!loading" class="banners">
    <div v-if="offline" class="banner offline" role="status">
      <p>{{ t('status.offline') }}</p>
      <code class="cmd">{{ t('status.offlineCmd') }}</code>
      <button type="button" class="retry" @click="refresh">{{ t('status.retry') }}</button>
    </div>
    <div v-else-if="isStale" class="banner stale" role="status">
      {{ t('status.stale', { when: staleLabel }) }}
    </div>
  </div>
</template>

<style scoped>
.banners {
  display: contents;
}
.banner {
  padding: 0.65rem 1rem;
  font-size: 0.85rem;
  border-bottom: 1px solid var(--line);
  animation: fade-in 120ms ease;
}
.offline {
  background: var(--surface2);
  color: var(--fg);
}
.stale {
  background: var(--accent-dim);
  color: var(--fg);
}
.cmd {
  display: block;
  margin-top: 0.35rem;
  font-size: 0.8rem;
  color: var(--muted);
}
.retry {
  margin-top: 0.5rem;
  background: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: var(--r-sm);
  padding: 0.35rem 0.75rem;
  font-weight: 600;
  cursor: pointer;
}
@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
