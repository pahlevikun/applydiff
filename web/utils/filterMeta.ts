import { getCatalog, resolveMessage } from '~/i18n'

export function filterTitle(key: string): string {
  const v = resolveMessage(getCatalog('en'), `filters.keys.${key}.title`)
  return v ?? key.replace(/_/g, ' ')
}

export function filterHint(key: string): string | undefined {
  return resolveMessage(getCatalog('en'), `filters.keys.${key}.hint`)
}
