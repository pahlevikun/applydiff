<script setup lang="ts">
const route = useRoute()
const { t } = useI18n()

const jobKey = computed(() => {
  const raw = route.params.key
  const segment = Array.isArray(raw) ? raw.join('/') : String(raw ?? '')
  try {
    return jobKeyFromRouteParam(decodeURIComponent(segment))
  } catch {
    return jobKeyFromRouteParam(segment)
  }
})

type Tab = 'fit' | 'diff' | 'letter' | 'interview'
const tab = ref<Tab>('fit')

watch(
  () => route.query.tab,
  (q) => {
    if (q === 'diff' || q === 'fit' || q === 'letter' || q === 'interview') tab.value = q
  },
  { immediate: true },
)

function setTab(next: Tab) {
  tab.value = next
  navigateTo({ path: route.path, query: { ...route.query, tab: next } }, { replace: true })
}

type MatchResult = { total: number }
type JobRow = {
  job: {
    id: string
    title: string
    company: string
    region: string
    risk?: number
    flags?: { severity: string }[]
  }
}

const job = ref<JobRow['job'] | null>(null)
const matchTotal = ref<number | null>(null)
const loading = ref(true)
const error = ref('')

async function loadMeta() {
  loading.value = true
  error.value = ''
  try {
    const jobs = await $fetch<JobRow[]>('/api/jobs')
    job.value = jobs.find((r) => r.job.id === jobKey.value)?.job ?? null
    try {
      const m = await $fetch<MatchResult>(`/api/match/${encodeJobIdForApi(jobKey.value)}`)
      matchTotal.value = m.total
    } catch {
      matchTotal.value = null
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('errors.loadJobs')
  } finally {
    loading.value = false
  }
}

watch(jobKey, () => {
  void loadMeta()
}, { immediate: true })

const interviewTo = computed(() => `/interview/${encodeURIComponent(jobKey.value)}`)
</script>

<template>
  <div class="detail">
    <div v-if="loading" class="sk-head" aria-busy="true" />
    <header v-else-if="job" class="head">
      <div>
        <h1>{{ job.title }}</h1>
        <p class="sub">{{ job.company }} · {{ job.region || '—' }}</p>
      </div>
      <div class="badges">
        <span v-if="matchTotal != null" class="score">{{ t('jobDetail.score', { total: matchTotal }) }}</span>
        <span v-if="job.risk" :class="['badge', job.flags?.some((f) => f.severity === 'hard') ? 'hard' : 'soft']">
          {{ t('jobs.risk', { score: job.risk }) }}
        </span>
      </div>
    </header>
    <p v-else-if="error" class="error">{{ error }}</p>

    <nav class="tabs" aria-label="Job sections">
      <button
        v-for="id in ['fit', 'diff', 'letter', 'interview']"
        :key="id"
        type="button"
        class="tab"
        :class="{ active: tab === id }"
        @click="setTab(id as Tab)"
      >
        {{ t(`jobDetail.tabs.${id}`) }}
      </button>
    </nav>

    <div class="panel">
      <MatchJobWorkspace v-if="tab === 'fit' && jobKey" :job-id="jobKey" />
      <DiffJobWorkspace v-else-if="tab === 'diff' && jobKey" :job-id="jobKey" />
      <section v-else-if="tab === 'letter'" class="letter">
        <p class="muted">{{ t('jobDetail.letterEmpty') }}</p>
        <p class="hint">{{ t('jobDetail.letterHint') }}</p>
      </section>
      <section v-else-if="tab === 'interview'" class="interview-link">
        <NuxtLink :to="interviewTo" class="primary">{{ t('jobDetail.openInterview') }}</NuxtLink>
      </section>
    </div>

    <footer class="bottom-bar">
      <span class="muted">{{ t('jobDetail.decisions') }}</span>
      <NuxtLink v-if="jobKey" :to="jobDetailPath(jobKey, tab === 'diff' ? 'diff' : 'fit')" class="primary-sm">
        {{ t('jobDetail.primaryContinue') }}
      </NuxtLink>
    </footer>
  </div>
</template>

<style scoped>
.detail {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  animation: fade-in 120ms ease;
}
.sk-head {
  height: 4rem;
  margin: 1rem;
  background: var(--surface);
  border-radius: var(--r);
  border: 1px solid var(--line);
}
.head {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--line);
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}
h1 {
  margin: 0;
  font-size: 1.2rem;
}
.sub {
  margin: 0.35rem 0 0;
  color: var(--muted);
  font-size: 0.9rem;
}
.badges {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.score {
  font-size: 0.85rem;
  color: var(--accent);
  font-weight: 600;
}
.badge {
  font-size: 0.7rem;
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
}
.badge.soft {
  background: var(--amber-bg);
  color: var(--amber);
}
.badge.hard {
  background: var(--del-bg);
  color: var(--del);
}
.tabs {
  display: flex;
  gap: 2px;
  padding: 0 1rem;
  border-bottom: 1px solid var(--line);
}
.tab {
  background: transparent;
  border: none;
  color: var(--muted);
  padding: 0.65rem 0.9rem;
  cursor: pointer;
  font-size: 0.9rem;
  border-bottom: 2px solid transparent;
}
.tab.active {
  color: var(--fg);
  border-bottom-color: var(--accent);
}
.panel {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
.letter,
.interview-link {
  padding: 1.5rem;
  max-width: 40rem;
}
.muted {
  color: var(--muted);
}
.hint {
  margin-top: 0.75rem;
  font-size: 0.9rem;
  color: var(--dim);
}
.primary {
  display: inline-block;
  margin-top: 1rem;
  background: var(--accent);
  color: var(--bg);
  padding: 0.5rem 1rem;
  border-radius: var(--r-sm);
  text-decoration: none;
  font-weight: 600;
}
.bottom-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 1rem;
  border-top: 1px solid var(--line);
  background: var(--surface);
}
.primary-sm {
  background: var(--accent);
  color: var(--bg);
  padding: 0.4rem 0.85rem;
  border-radius: var(--r-sm);
  text-decoration: none;
  font-weight: 600;
  font-size: 0.85rem;
}
.error {
  color: var(--del);
  padding: 1rem;
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
