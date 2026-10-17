<script setup lang="ts">
defineProps<{
  label: string
  checked: boolean
}>()

defineEmits<{
  toggle: []
}>()
</script>

<template>
  <button
    type="button"
    class="chip-btn"
    :class="{ on: checked }"
    :aria-pressed="checked"
    @click="$emit('toggle')"
  >
    <span class="mark" aria-hidden="true">
      <svg v-if="checked" width="12" height="12" viewBox="0 0 12 12" fill="none">
        <path d="M2.5 6l2.5 2.5L9.5 4" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </span>
    <span class="lbl">{{ label }}</span>
  </button>
</template>

<style scoped>
.chip-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  border-radius: 999px;
  border: 1px solid var(--line2);
  background: var(--surface);
  color: var(--muted);
  font-size: 12.5px;
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
  cursor: pointer;
  transition:
    border-color 0.15s,
    background 0.15s,
    color 0.15s,
    box-shadow 0.15s;
}

.chip-btn:hover {
  border-color: var(--accent-dim);
  color: var(--fg);
}

.chip-btn.on {
  border-color: var(--accent);
  background: var(--amber-bg);
  color: var(--fg);
  box-shadow: 0 0 0 1px var(--accent-dim);
}

.mark {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  border: 1.5px solid var(--line2);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--accent-dim);
}

.chip-btn.on .mark {
  border-color: var(--accent);
  background: var(--accent);
  color: var(--surface);
}

.lbl {
  line-height: 1.2;
}
</style>
