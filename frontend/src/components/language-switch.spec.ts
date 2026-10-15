import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import LanguageSwitch from './LanguageSwitch.vue'

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

describe('LanguageSwitch', () => {
  it('cambia el locale y emite "change"', async () => {
    const i18n = makeI18n()
    const w = mount(LanguageSwitch, { global: { plugins: [i18n] } })
    const opts = w.findAll('.lang-opt')
    expect(opts.length).toBe(2)
    await opts[1].trigger('click') // EN
    expect(i18n.global.locale.value).toBe('en')
    expect(w.emitted('change')?.[0]).toEqual(['en'])
  })

  it('marca la opción activa con aria-pressed', () => {
    const w = mount(LanguageSwitch, { global: { plugins: [makeI18n()] } })
    const opts = w.findAll('.lang-opt')
    expect(opts[0].attributes('aria-pressed')).toBe('true') // ES activo
    expect(opts[1].attributes('aria-pressed')).toBe('false')
  })
})
