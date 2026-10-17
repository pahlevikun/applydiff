<script setup lang="ts">
const props = defineProps<{
  hard: boolean
  boost?: number
  showBoost?: boolean
}>()

const emit = defineEmits<{
  'update:hard': [value: boolean]
  'update:boost': [value: number]
}>()

const boostVal = computed({
  get: () => props.boost ?? 5,
  set: (v: number) => emit('update:boost', v),
})
</script>

<template>
  <div class="mode">
    <div class="seg" role="group" aria-label="Rule strictness">
      <button type="button" class="seg-btn" :class="{ on: hard }" @click="emit('update:hard', true)">Hard</button>
      <button type="button" class="seg-btn" :class="{ on: !hard }" @click="emit('update:hard', false)">Soft</button>
    </div>
    <p class="mode-hint">
      <template v-if="hard">Hard — listings that fail this rule are <strong>hidden</strong> (skipped in the feed).</template>
      <template v-else>Soft — matches add to your <strong>ranking score</strong> instead of being removed.</template>
    </p>
    <div v-if="showBoost && !hard" class="boost">
      <label class="boost-label">
        <span>Boost</span>
        <span class="boost-val">{{ boostVal }}</span>
      </label>
      <input
        v-model.number="boostVal"
        type="range"
        min="1"
        max="15"
        step="1"
        class="boost-range"
        aria-label="Boost points when rule matches"
      />
      <p class="boost-hint">Higher boost moves matching jobs up in the Jobs list.</p>
    </div>
  </div>
</template>

<style scoped>
.mode {
  width: 100%;
  max-width: 28rem;
}

.seg {
  display: inline-flex;
  border: 1px solid var(--line2);
  border-radius: var(--r-sm);
  overflow: hidden;
  background: var(--surface);
}

.seg-btn {
  border: none;
  background: transparent;
  color: var(--muted);
  font-size: 12px;
  font-weight: 500;
  padding: 7px 14px;
  cursor: pointer;
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}

.seg-btn.on {
  background: var(--accent-dim);
  color: var(--accent);
}

.mode-hint {
  margin: 8px 0 0;
  font-size: 11.5px;
  line-height: 1.45;
  color: var(--dim);
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}

.mode-hint strong {
  color: var(--muted);
  font-weight: 600;
}

.boost {
  margin-top: 10px;
  padding: 10px 12px;
  border-radius: var(--r-sm);
  border: 1px solid var(--line);
  background: var(--bg);
}

.boost-label {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--dim);
  margin-bottom: 6px;
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}

.boost-val {
  color: var(--accent);
  font-variant-numeric: tabular-nums;
}

.boost-range {
  width: 100%;
  accent-color: var(--accent);
}

.boost-hint {
  margin: 6px 0 0;
  font-size: 11px;
  color: var(--dim);
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}
</style>
