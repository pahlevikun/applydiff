import { getCatalog, formatMessage, resolveMessage, type LocaleId, type MessageParams } from '~/i18n'

export function useI18n() {
  const locale = useState<LocaleId>('i18n-locale', () => 'en')

  const messages = computed(() => getCatalog(locale.value))

  function t(path: string, params?: MessageParams): string {
    const text = resolveMessage(messages.value, path)
    if (!text) return path
    return formatMessage(text, params)
  }

  function setLocale(next: LocaleId) {
    locale.value = next
  }

  return { locale, t, setLocale }
}
