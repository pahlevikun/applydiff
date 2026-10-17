// https://nuxt.com/docs/api/configuration/nuxt-config
import { interviewPrepEnabled } from './utils/features'

const interviewRouteRules = interviewPrepEnabled
  ? {}
  : {
      '/interview': { redirect: '/jobs' },
      '/interview/**': { redirect: '/jobs' },
    }

export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },
  ssr: false,
  nitro: {
    preset: 'static',
  },
  routeRules: {
    '/': { redirect: '/jobs' },
    ...interviewRouteRules,
  },
  css: ['~/assets/theme.css'],
  vite: {
    server: {
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
  },
})
