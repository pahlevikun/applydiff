import type { Ref } from 'vue'

export function useInboxKeys(
  selected: Ref<number>,
  length: Ref<number>,
  opts: {
    onOpen: () => void
    onSkip: () => void
    onJumpEnd: (toEnd: boolean) => void
  },
) {
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
    if (e.key === ' ' || e.code === 'Space') {
      e.preventDefault()
      opts.onOpen()
    }
    if (e.key === 'x') {
      e.preventDefault()
      opts.onSkip()
    }
    if (e.key === 'g' && !e.shiftKey) {
      e.preventDefault()
      opts.onJumpEnd(false)
    }
    if (e.key === 'G' || (e.key === 'g' && e.shiftKey)) {
      e.preventDefault()
      opts.onJumpEnd(true)
    }
  }
  onMounted(() => window.addEventListener('keydown', onKey))
  onUnmounted(() => window.removeEventListener('keydown', onKey))
}
