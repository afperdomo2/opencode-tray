import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { nextTick } from 'vue'
import es from './i18n/es.json'
import en from './i18n/en.json'
import App from './App.vue'

const mocks = vi.hoisted(() => ({
  settings: vi.fn(),
  height: vi.fn(),
  hide: vi.fn(),
  blurred: vi.fn(),
}))

vi.mock('../wailsjs/go/main/App', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../wailsjs/go/main/App')>()
  return {
    ...actual,
    GetSettings: mocks.settings,
    SetFlyoutHeight: mocks.height,
    HideWindow: mocks.hide,
    WindowBlurred: mocks.blurred,
  }
})

function makeI18n() {
  return createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })
}

function mountApp() {
  return mount(App, {
    global: {
      plugins: [makeI18n()],
      stubs: { Resumen: true, Servicio: true, AccesoWeb: true, Ajustes: true },
    },
  })
}

describe('App (flyout)', () => {
  beforeEach(() => {
    mocks.settings.mockResolvedValue({ language: 'es' })
    mocks.height.mockResolvedValue(undefined)
    mocks.hide.mockResolvedValue(undefined)
    mocks.blurred.mockResolvedValue(undefined)
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('coloca la navegación después del contenido', async () => {
    const w = mountApp()
    await flushPromises()
    const children = Array.from(w.find('.app').element.children).map((el) => el.tagName.toLowerCase())
    expect(children).toEqual(['main', 'nav'])
    expect(w.findAll('.app-nav .nav-btn').length).toBe(4)
    w.unmount()
  })

  it('mantiene el alto fijo al cambiar de sección', async () => {
    const w = mountApp()
    await flushPromises()

    await w.findAll('.nav-btn')[1].trigger('click')
    await nextTick()
    await flushPromises()
    expect(mocks.height).not.toHaveBeenCalled()
    w.unmount()
  })

  it('oculta el panel con Escape y al perder el foco', async () => {
    const w = mountApp()
    await flushPromises()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(mocks.hide).toHaveBeenCalled()

    window.dispatchEvent(new Event('blur'))
    expect(mocks.blurred).toHaveBeenCalled()
    w.unmount()
  })
})
