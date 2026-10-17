import { en, type Messages } from './locales/en'
import { formatMessage, resolveMessage, type MessageParams } from './resolve'

export type LocaleId = 'en'

const catalogs: Record<LocaleId, Messages> = {
  en,
}

export function getCatalog(locale: LocaleId): Messages {
  return catalogs[locale] ?? en
}

/** Non-Vue helper (composables, activity log). */
export function msg(path: string, params?: MessageParams, locale: LocaleId = 'en'): string {
  const text = resolveMessage(getCatalog(locale), path)
  if (!text) return path
  return formatMessage(text, params)
}

export { en, formatMessage, resolveMessage }
export type { MessageParams, Messages }
