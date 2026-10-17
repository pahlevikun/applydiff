# Web UI copy (i18n)

All user-facing strings for the Nuxt shell live under `locales/`.

- **Default:** `locales/en.ts` (plain, casual English)
- **In components:** `const { t } = useI18n()` then `t('nav.jobs')` or `t('actions.jobScanDetail', { fetched: 1, stored: 2, skipped: 0 })`
- **In scripts/composables:** `import { msg } from '~/i18n'`

## Add a language

1. Copy `locales/en.ts` to `locales/id.ts` (or `de.ts`, etc.) and translate values only. Keep keys identical.
2. Register the locale in `i18n/index.ts` (`catalogs` map).
3. Switch locale with `setLocale('id')` from `useI18n()` (wire a user setting when ready).

Do not put API payloads, config keys, or route paths in locale files.
