import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import Ajustes from './Ajustes.vue'
import { clearQueryCache } from '../lib/query'
import { settings } from '../lib/settings'

const mocks = vi.hoisted(() => ({
  usage: vi.fn(),
  tokenStatus: vi.fn(),
  reload: vi.fn(),
  setToken: vi.fn(),
  clearToken: vi.fn(),
}))

// Mock del puente con el backend para que el test no dependa de Wails.
vi.mock('../../wailsjs/go/main/App', () => ({
  GetSettings: vi.fn().mockResolvedValue({ language: 'es', autoStart: false, refreshMin: 15 }),
  SetAutoStart: vi.fn().mockResolvedValue(undefined),
  SetLanguage: vi.fn().mockResolvedValue(undefined),
  SetRefreshMin: vi.fn().mockResolvedValue(undefined),
  NotifySetEnabled: vi.fn(),
  NotifyEvaluate: vi.fn(),
  UsageFetch: mocks.usage,
  UsageTokenStatus: mocks.tokenStatus,
  UsageReloadToken: mocks.reload,
  UsageSetManualToken: mocks.setToken,
  UsageClearManualToken: mocks.clearToken,
}))

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

const sampleUsage = {
  rolling: { percent: 10, resetsAt: '' },
  weekly: { percent: 20, resetsAt: '' },
  monthly: { percent: 30, resetsAt: '' },
}

async function mountAjustes() {
  const wrapper = mount(Ajustes, { global: { plugins: [makeI18n()] } })
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  vi.clearAllMocks()
  clearQueryCache()
  settings.language = 'es'
  settings.autoStart = false
  settings.refreshMin = 15

  mocks.usage.mockResolvedValue(sampleUsage)
  mocks.tokenStatus.mockResolvedValue({ configured: true, source: 'opencode-export' })
  mocks.reload.mockResolvedValue({ configured: true, source: 'opencode-export' })
  mocks.setToken.mockResolvedValue(undefined)
  mocks.clearToken.mockResolvedValue(undefined)
})

describe('Ajustes', () => {
  it('renderiza auto-inicio, notificaciones y la tarjeta de OpenCode Go', async () => {
    const w = await mountAjustes()
    const toggles = w.findAll('input[type="checkbox"]')
    // auto-inicio + 3 tipos de notificación
    expect(toggles.length).toBe(4)
    expect(w.text()).toContain('Iniciar con el sistema')
    expect(w.text()).toContain('Ritmo de consumo')
    expect(w.text()).toContain('Casi sin saldo')
    expect(w.text()).toContain('Servicio detenido')
    expect(w.text()).toContain('OpenCode Go')
    expect(w.text()).toContain('Token vigente de OpenCode')
    expect(w.find('.settings-grid').exists()).toBe(true)
    expect(w.findAll('.settings-grid > .card').length).toBe(4)
  })

  it('cambia el intervalo de refresco', async () => {
    const w = await mountAjustes()
    await w.find('select').setValue('5')
    await flushPromises()
    const { SetRefreshMin } = await import('../../wailsjs/go/main/App')
    expect(SetRefreshMin).toHaveBeenCalledWith(5)
  })

  it('relee el token desde OpenCode con el botón', async () => {
    const w = await mountAjustes()
    await w.find('.btn-full').trigger('click')
    await flushPromises()
    expect(mocks.reload).toHaveBeenCalled()
  })

  it('guarda una API key manual y muestra confirmación', async () => {
    const w = await mountAjustes()
    await w.find('.token-form input').setValue('sk-manual')
    await w.find('.token-form').trigger('submit')
    await flushPromises()

    expect(mocks.setToken).toHaveBeenCalledWith('sk-manual')
    expect(mocks.usage).toHaveBeenCalled()
    expect(w.text()).toContain('API key guardada y verificada')
  })

  it('elimina la API key manual', async () => {
    const w = await mountAjustes()
    const clear = w.findAll('button').find((b) => b.text().includes('Eliminar API key manual'))
    expect(clear).toBeTruthy()
    await clear!.trigger('click')
    await flushPromises()
    expect(mocks.clearToken).toHaveBeenCalled()
  })
})
