import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import Section from './Section.vue'

// i18n de prueba independiente (sin app real) para montar componentes.
const makeI18n = (locale: 'es' | 'en') =>
  createI18n({ legacy: false, locale, fallbackLocale: 'en', messages: { es, en } })

describe('Section', () => {
  it('renderiza el título en español', () => {
    const w = mount(Section, {
      props: { name: 'resumen' },
      global: { plugins: [makeI18n('es')] },
    })
    expect(w.text()).toContain('Resumen')
  })

  it('renderiza el título en inglés', () => {
    const w = mount(Section, {
      props: { name: 'resumen' },
      global: { plugins: [makeI18n('en')] },
    })
    expect(w.text()).toContain('Summary')
  })

  it('renderiza las 6 secciones con su clave i18n', () => {
    const ids = ['resumen', 'servicio', 'accesoWeb', 'consumo', 'config', 'ajustes']
    const i18n = makeI18n('es')
    for (const id of ids) {
      const w = mount(Section, { props: { name: id }, global: { plugins: [i18n] } })
      expect(w.find('h2').exists()).toBe(true)
    }
  })
})
