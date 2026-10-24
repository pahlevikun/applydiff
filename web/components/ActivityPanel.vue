<script setup lang="ts">
import { useActivity } from '~/composables/useActivity'

const { open, lines, activeRun, elapsedMs, clear, formatDuration, refreshHistory } = useActivity()
const { t } = useI18n()

let historyTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  void refreshHistory()
  historyTimer = setInterval(() => {
    void refreshHistory()
  }, 20_000)
})

onUnmounted(() => {
  if (historyTimer) clearInterval(historyTimer)
})

const height = ref(168)

function toggle() {
  open.value = !open.value
}

function formatTime(at: number) {
  return new Date(at).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

const statusLabel = computed(() => {
  const s = activeRun.value?.status
  if (!s || s === 'running') return t('activity.running')
  if (s === 'success') return t('activity.success')
  if (s === 'timeout') return t('activity.timedOut')
  return t('activity.failed')
})

const statusClass = computed(() => activeRun.value?.status ?? 'idle')

const timeLine = computed(() => {
  const run = activeRun.value
  if (!run) return ''
  const elapsed = formatDuration(elapsedMs.value)
  if (run.status === 'running') {
    const left = Math.max(0, run.timeoutMs - elapsedMs.value)
    return t('activity.timeoutLeft', { elapsed, left: formatDuration(left) })
  }
  return run.phase
})
</script>

<template>
  <section class="panel" :class="{ collapsed: !open }" aria-label="Activity">
    <header class="bar" @click="toggle">
      <span class="grip" aria-hidden="true" />
      <span class="title">{{ t('activity.title') }}</span>
      <span v-if="activeRun" class="pill" :class="statusClass">{{ statusLabel }}</span>
      <span v-if="activeRun?.status === 'running'" class="bar-elapsed">{{ formatDuration(elapsedMs) }}</span>
      <span class="grow" />
      <button type="button" class="icon-btn" @click.stop="clear">{{ t('common.clear') }}</button>
      <span class="chev" aria-hidden="true">{{ open ? '▾' : '▴' }}</span>
    </header>
    <div v-show="open" class="body" :style="{ height: `${height}px` }">
      <div v-if="activeRun" class="current" :class="statusClass">
        <div class="current-head">
          <span class="current-label">{{ activeRun.label }}</span>
          <span class="current-phase">{{ timeLine }}</span>
        </div>
        <div class="track" role="progressbar" :aria-valuenow="activeRun.progress" aria-valuemin="0" aria-valuemax="100">
          <div
            class="fill"
            :class="{ indeterminate: activeRun.status === 'running' && activeRun.progress < 90 }"
            :style="{ width: `${activeRun.progress}%` }"
          />
        </div>
        <p v-if="activeRun.detail && activeRun.status === 'success'" class="current-detail ok">{{ activeRun.detail }}</p>
        <p v-if="activeRun.error && activeRun.status !== 'success'" class="current-detail err">{{ activeRun.error }}</p>
      </div>

      <p v-if="!lines.length && !activeRun" class="empty">{{ t('activity.empty') }}</p>
      <ul v-else class="log">
        <li v-for="line in lines" :key="line.id" :class="['line', line.level]">
          <span class="ts">{{ formatTime(line.at) }}</span>
          <span class="dot" aria-hidden="true" />
          <span class="msg">{{ line.text }}</span>
        </li>
      </ul>
    </div>
  </section>
</template>

<style scoped>
.panel {
  flex-shrink: 0;
  border-top: 1px solid var(--line);
  background: var(--surface);
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 11.5px;
}

.panel.collapsed .bar {
  border-bottom: none;
}

.panel.collapsed .body {
  display: none;
}

.bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  height: 40px;
  cursor: pointer;
  user-select: none;
  background: var(--surface);
  border-bottom: 1px solid transparent;
}

.panel:not(.collapsed) .bar {
  border-bottom-color: var(--line);
}

.grip {
  width: 28px;
  height: 3px;
  border-radius: 2px;
  background: var(--line2);
  opacity: 0.85;
  flex-shrink: 0;
}

.title {
  font-weight: 600;
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
  font-size: 12px;
  color: var(--fg);
  letter-spacing: -0.01em;
}

.bar-elapsed {
  font-size: 11px;
  color: var(--muted);
  font-variant-numeric: tabular-nums;
}

.pill {
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 999px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
  border: 1px solid var(--line2);
  color: var(--muted);
}

.pill.running {
  background: var(--amber-bg);
  color: var(--accent);
  border-color: var(--accent-dim);
}

.pill.success {
  background: var(--add-bg);
  color: var(--add);
  border-color: var(--line2);
}

.pill.failed,
.pill.timeout {
  background: var(--del-bg);
  color: var(--del);
  border-color: var(--line2);
}

.grow {
  flex: 1;
}

.icon-btn {
  border: none;
  background: transparent;
  color: var(--muted);
  font-size: 11px;
  cursor: pointer;
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
  padding: 4px 8px;
  border-radius: var(--r-sm);
}

.icon-btn:hover {
  color: var(--fg);
  background: var(--surface2);
}

.chev {
  color: var(--dim);
  font-size: 10px;
  margin-left: 2px;
}

.body {
  overflow: auto;
  padding: 10px 16px 12px;
  background: var(--bg);
  color: var(--muted);
}

.current {
  margin-bottom: 10px;
  padding: 10px 12px;
  border-radius: var(--r-sm);
  border: 1px solid var(--line);
  background: var(--surface);
}

.current.running {
  border-color: var(--accent-dim);
}

.current.success {
  border-color: var(--line2);
}

.current.failed,
.current.timeout {
  border-color: var(--line2);
}

.current-head {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 12px;
  align-items: baseline;
  margin-bottom: 8px;
}

.current-label {
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
  font-weight: 600;
  font-size: 12px;
  color: var(--fg);
}

.current-phase {
  font-size: 11px;
  color: var(--muted);
  flex: 1;
  min-width: 12rem;
}

.track {
  height: 4px;
  border-radius: 2px;
  background: var(--line);
  overflow: hidden;
}

.fill {
  height: 100%;
  border-radius: 2px;
  background: var(--accent);
  transition: width 0.4s ease;
}

.current.success .fill {
  background: var(--add);
}

.current.failed .fill,
.current.timeout .fill {
  background: var(--del);
}

.fill.indeterminate::after {
  content: '';
  display: block;
  height: 100%;
  width: 40%;
  background: linear-gradient(90deg, transparent, color-mix(in srgb, var(--fg) 13%, transparent), transparent);
  animation: shimmer 1.4s ease-in-out infinite;
}

@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(280%);
  }
}

.current-detail {
  margin: 8px 0 0;
  font-size: 11px;
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}

.current-detail.ok {
  color: var(--add);
}

.current-detail.err {
  color: var(--del);
}

.empty {
  margin: 0;
  color: var(--dim);
  font-size: 11.5px;
  line-height: 1.5;
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}

.log {
  list-style: none;
  margin: 0;
  padding: 0;
}

.line {
  display: flex;
  gap: 8px;
  align-items: baseline;
  padding: 3px 0;
  line-height: 1.45;
}

.ts {
  color: var(--dim);
  flex-shrink: 0;
  font-size: 10.5px;
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 0.35em;
  background: var(--line2);
}

.line.info .dot {
  background: var(--line2);
}

.line.run .dot {
  background: var(--accent);
  box-shadow: 0 0 0 2px var(--amber-bg);
}

.line.ok .dot {
  background: var(--add);
}

.line.err .dot {
  background: var(--del);
}

.msg {
  color: var(--fg);
  flex: 1;
  min-width: 0;
}

.line.info .msg {
  color: var(--muted);
}

.line.run .msg {
  color: var(--fg);
}

.line.ok .msg {
  color: var(--add);
}

.line.err .msg {
  color: var(--del);
}
</style>
