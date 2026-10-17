<script setup lang="ts">
const note = ref('')
const saved = ref(false)

function save() {
  saved.value = true
}
</script>

<template>
  <div class="page">
    <h1>Interview prep</h1>
    <p class="muted">Notes stay local in this session. No auto-submit.</p>
    <textarea v-model="note" class="editor mono" rows="16" placeholder="STAR stories, questions to ask…" />
    <button type="button" class="btn" @click="save">Save notes</button>
    <p v-if="saved" class="dim">Saved locally.</p>
  </div>
</template>

<style scoped>
.page {
  max-width: 960px;
  margin: 0 auto;
  padding: 24px;
}
h1 {
  font-size: 1.25rem;
  margin: 0 0 8px;
}
.muted {
  color: var(--muted);
  margin-bottom: 16px;
}
.editor {
  width: 100%;
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--r-sm);
  color: var(--fg);
  padding: 12px;
  margin-bottom: 12px;
}
.btn {
  background: var(--accent);
  color: var(--ink);
  border: none;
  border-radius: var(--r-sm);
  padding: 8px 14px;
  font-weight: 600;
  cursor: pointer;
}
.dim {
  color: var(--dim);
  margin-top: 8px;
}
</style>
