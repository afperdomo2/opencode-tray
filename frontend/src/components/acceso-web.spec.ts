import { describe, it, expect, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import AccesoWeb from './AccesoWeb.vue'

const mocks = vi.hoisted(() => ({
  urls: vi.fn(),
  username: vi.fn(),
  password: vi.fn(),
  changePassword: vi.fn(),
  applyMode: vi.fn(),
}))

vi.mock('../../wailsjs/go/main/App', () => ({
  WebAccessURLs: mocks.urls,
  WebAccessUsername: mocks.username,
  WebAccessCurrentPassword: mocks.password,
  WebAccessChangePassword: mocks.changePassword,
  WebAccessApplyMode: mocks.applyMode,
}))

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

describe('AccesoWeb', () => {
  it('usa un panel grid y el botón Guardar tiene el estilo estándar', async () => {
    mocks.urls.mockResolvedValue(['http://127.0.0.1:49374'])
    mocks.username.mockResolvedValue('opencode')
    mocks.password.mockResolvedValue('secreto')

    const w = mount(AccesoWeb, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    expect(w.find('.panel-grid').exists()).toBe(true)
    expect(w.findAll('.panel-grid > .card').length).toBe(2)
    expect(w.find('.pw-change button.btn').exists()).toBe(true)
    expect(w.text()).toContain('http://127.0.0.1:49374')
    expect(w.find('.icon-btn').attributes('title')).toBe('Mostrar')
  })

  it('muestra skeleton de URLs mientras cargan', async () => {
    mocks.urls.mockReturnValue(new Promise(() => {}))
    mocks.username.mockResolvedValue('opencode')
    mocks.password.mockResolvedValue('secreto')

    const w = mount(AccesoWeb, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    expect(w.find('.urls').exists()).toBe(false)
    expect(w.findAll('.skeleton').length).toBeGreaterThan(0)
  })
})
