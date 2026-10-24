<script setup lang="ts">
/** Legacy `/match/:id` URLs → `/match?job=` (job ids may contain slashes). */
const route = useRoute()
const raw = route.params.id
const segment = Array.isArray(raw) ? raw.join('/') : String(raw ?? '')
const decoded = (() => {
  try {
    return decodeURIComponent(segment)
  } catch {
    return segment
  }
})()

await navigateTo(jobDetailPath(decoded, 'fit'), { replace: true })
</script>

<template>
  <p class="muted">Redirecting…</p>
</template>

<style scoped>
.muted {
  padding: 24px;
  color: var(--muted);
}
</style>
