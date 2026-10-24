<script setup lang="ts">
export type InboxCardModel = {
  job_key: string
  title: string
  company: string
  state: string
  reason: string
  updated_at: string
  region?: string
  risk?: number
  riskHard?: boolean
  verdict?: string
}

const props = defineProps<{
  item: InboxCardModel
  selected?: boolean
}>()

const { t } = useI18n()

const primaryLabel = computed(() => {
  switch (props.item.state) {
    case 'needs-decision':
      return t('inbox.actions.review')
    case 'in-progress':
      return t('inbox.actions.continue')
    case 'sent':
      return t('inbox.actions.track')
    case 'interview':
      return t('inbox.actions.prepare')
    default:
      return t('inbox.actions.review')
  }
})

function formatAge(iso: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  const days = Math.floor((Date.now() - d.getTime()) / 86400000)
  if (days <= 0) return t('inbox.age.today')
  if (days === 1) return t('inbox.age.oneDay')
  return t('inbox.age.days', { days })
}

const verdictLabel = computed(() => props.item.verdict || t('inbox.verdict.unknown'))
</script>

<template>
  <article class="card" :class="{ selected }">
    <div class="head">
      <h2 class="title">{{ item.title }}</h2>
      <span class="age">{{ formatAge(item.updated_at) }}</span>
    </div>
    <p class="sub">
      <span>{{ item.company }}</span>
      <span v-if="item.region" class="dot">·</span>
      <span v-if="item.region">{{ item.region }}</span>
    </p>
    <div class="badges">
      <span class="badge verdict">{{ verdictLabel }}</span>
      <span v-if="item.risk" :class="['badge', item.riskHard ? 'risk-hard' : 'risk-soft']">
        {{ t('jobs.risk', { score: item.risk }) }}
      </span>
    </div>
    <p v-if="item.reason" class="reason">{{ item.reason }}</p>
    <NuxtLink :to="jobDetailPath(item.job_key)" class="primary" @click.stop>
      {{ primaryLabel }}
    </NuxtLink>
  </article>
</template>

<style scoped>
.card {
  border: 1px solid var(--line);
  border-radius: var(--r);
  background: var(--surface);
  padding: 0.85rem 1rem;
  transition: border-color 120ms ease, opacity 120ms ease;
}
.card.selected {
  border-color: var(--accent);
  background: var(--surface2);
}
.head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
}
.title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}
.age {
  font-size: 0.75rem;
  color: var(--dim);
  font-variant-numeric: tabular-nums;
}
.sub {
  margin: 0.35rem 0 0;
  font-size: 0.85rem;
  color: var(--muted);
}
.dot {
  opacity: 0.5;
}
.badges {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.5rem;
}
.badge {
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
}
.verdict {
  background: var(--surface2);
  color: var(--muted);
  border: 1px solid var(--line);
}
.risk-soft {
  background: var(--amber-bg);
  color: var(--amber);
}
.risk-hard {
  background: var(--del-bg);
  color: var(--del);
}
.reason {
  margin: 0.45rem 0 0;
  font-size: 0.78rem;
  color: var(--dim);
}
.primary {
  display: inline-block;
  margin-top: 0.75rem;
  background: var(--accent);
  color: var(--bg);
  text-decoration: none;
  font-weight: 600;
  font-size: 0.85rem;
  padding: 0.45rem 0.9rem;
  border-radius: var(--r-sm);
}
</style>
