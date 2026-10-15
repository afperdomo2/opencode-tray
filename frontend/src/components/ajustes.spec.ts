import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import Ajustes from './Ajustes.vue'

// Mock del puente con el backend para que el test no dependa de Wails.
vi.mock('../../wailsjs/go/main/App', () => ({
  GetSettings: vi.fn().mockResolvedValue({ language: 'es', autoStart: false, refreshMin: 15 }),
  SetAutoStart: vi.fn().mockResolvedValue(undefined),
  NotifySetEnabled: vi.fn(),
  NotifyEvaluate: vi.fn(),
}))

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

describe('Ajustes', () => {
  it('renderiza el toggle de auto-inicio y los de notificación', () => {
    const w = mount(Ajustes, { global: { plugins: [makeI18n()] } })
    const toggles = w.findAll('input[type="checkbox"]')
    // auto-inicio + 3 tipos de notificación
    expect(toggles.length).toBe(4)
    expect(w.text()).toContain('Iniciar con Windows')
    expect(w.text()).toContain('Ritmo de consumo')
    expect(w.text()).toContain('Casi sin saldo')
    expect(w.text()).toContain('Servicio detenido')
  })
})
