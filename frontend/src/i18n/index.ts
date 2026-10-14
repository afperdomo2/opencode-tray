import { createI18n } from 'vue-i18n'
import es from './es.json'
import en from './en.json'

// i18n de la app: español e inglés. El idioma se carga desde los ajustes
// y se cambia con el selector del encabezado.
export const i18n = createI18n({
  legacy: false,
  locale: 'es',
  fallbackLocale: 'en',
  messages: { es, en },
})

export type Locale = 'es' | 'en'
