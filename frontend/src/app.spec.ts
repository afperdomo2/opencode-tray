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
  platform: vi.fn(),
}))

vi.mock('../wailsjs/go/main/App', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../wailsjs/go/main/App')>()
  return {
    ...actual,
    GetSettings: mocks.settings,
    SetFlyoutHeight: mocks.height,
    HideWindow: mocks.hide,
    Platform: mocks.platform,
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
    mocks.platform.mockResolvedValue('windows')
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('coloca la cabecera arriba y la navegación al final', async () => {
    const w = mountApp()
    await flushPromises()
    const children = Array.from(w.find('.app').element.children).map((el) => el.tagName.toLowerCase())
    expect(children).toEqual(['header', 'main', 'nav'])
    expect(w.find('.app-header h1').text()).toBe('OpenCode Tray')
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

  it('activa el arrastre desde la cabecera solo fuera de Windows', async () => {
    const win = mountApp()
    await flushPromises()
    expect(win.find('.app-header').exists()).toBe(true)
    expect(win.find('.app-header').classes()).not.toContain('is-draggable')
    win.unmount()

    mocks.platform.mockResolvedValue('linux')
    const linux = mountApp()
    await flushPromises()
    expect(linux.find('.app-header').classes()).toContain('is-draggable')
    expect(linux.find('.drag-strip').exists()).toBe(false)
    linux.unmount()
  })

  it('cierra con la X y con Escape', async () => {
    const w = mountApp()
    await flushPromises()

    await w.find('.panel-close').trigger('click')
    expect(mocks.hide).toHaveBeenCalledTimes(1)

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(mocks.hide).toHaveBeenCalledTimes(2)
    w.unmount()
  })

  it('no se oculta al perder el foco', async () => {
    const w = mountApp()
    await flushPromises()

    window.dispatchEvent(new Event('blur'))
    await flushPromises()
    expect(mocks.hide).not.toHaveBeenCalled()
    w.unmount()
  })
})
