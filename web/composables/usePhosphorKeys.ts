import { onMounted, onUnmounted, type Ref } from 'vue'

export function useSearchFocus(searchRef: Ref<HTMLInputElement | null>) {
  function onKey(e: KeyboardEvent) {
    if (e.key === '/' && !(e.target instanceof HTMLInputElement) && !(e.target instanceof HTMLTextAreaElement)) {
      e.preventDefault()
      searchRef.value?.focus()
    }
  }
  onMounted(() => window.addEventListener('keydown', onKey))
  onUnmounted(() => window.removeEventListener('keydown', onKey))
}

export function useListKeys(selected: Ref<number>, length: Ref<number>) {
  function onKey(e: KeyboardEvent) {
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
      return
    }
    if (e.key === 'j') {
      e.preventDefault()
      if (selected.value < length.value - 1) selected.value++
    }
    if (e.key === 'k') {
      e.preventDefault()
      if (selected.value > 0) selected.value--
    }
  }
  onMounted(() => window.addEventListener('keydown', onKey))
  onUnmounted(() => window.removeEventListener('keydown', onKey))
}

export function useHunkToggle(onToggle: () => void) {
  function onKey(e: KeyboardEvent) {
    if (e.code === 'Space' && !(e.target instanceof HTMLInputElement) && !(e.target instanceof HTMLTextAreaElement)) {
      e.preventDefault()
      onToggle()
    }
  }
  onMounted(() => window.addEventListener('keydown', onKey))
  onUnmounted(() => window.removeEventListener('keydown', onKey))
}
