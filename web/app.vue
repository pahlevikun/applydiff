<script setup lang="ts">
import { interviewPrepEnabled } from '~/utils/features'

const route = useRoute()
const { t } = useI18n()
const isMatchNav = computed(() => route.path === '/match' || route.path.startsWith('/match/'))
const isDiffNav = computed(() => route.path === '/diff' || route.path.startsWith('/diff/'))
</script>

<template>
  <div class="shell">
    <header class="topbar">
      <div class="brand">{{ t('brand') }}</div>
      <nav class="nav" aria-label="Main">
        <NuxtLink to="/jobs" class="nav-link">{{ t('nav.jobs') }}</NuxtLink>
        <NuxtLink to="/match" class="nav-link" :class="{ 'router-link-active': isMatchNav }">{{ t('nav.match') }}</NuxtLink>
        <NuxtLink to="/diff" class="nav-link" :class="{ 'router-link-active': isDiffNav }">{{ t('nav.diff') }}</NuxtLink>
        <NuxtLink to="/cv" class="nav-link">{{ t('nav.cv') }}</NuxtLink>
        <NuxtLink to="/applications" class="nav-link">{{ t('nav.applications') }}</NuxtLink>
        <NuxtLink v-if="interviewPrepEnabled" to="/interview" class="nav-link">{{ t('nav.interview') }}</NuxtLink>
        <NuxtLink to="/sources" class="nav-link">{{ t('nav.sources') }}</NuxtLink>
      </nav>
    </header>
    <main class="main">
      <NuxtPage />
    </main>
    <ActivityPanel />
  </div>
</template>

<style>
@import url('~/assets/theme.css');

.shell {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg);
  color: var(--fg);
}

.topbar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 0 16px;
  height: 52px;
  border-bottom: 1px solid var(--line);
  background: var(--surface);
}

.brand {
  font-weight: 600;
  font-size: 14.5px;
}

.nav {
  display: flex;
  gap: 2px;
  margin-left: 10px;
}

.nav-link {
  color: var(--muted);
  text-decoration: none;
  padding: 7px 12px;
  border-radius: var(--r-sm);
  font-size: 13px;
}

.nav-link:hover,
.nav-link.router-link-active {
  color: var(--fg);
  background: var(--surface2);
}

.main {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
</style>
