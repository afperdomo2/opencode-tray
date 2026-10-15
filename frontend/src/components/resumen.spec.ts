import { describe, it, expect, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import Resumen from './Resumen.vue'

// Mocks de los bindings de Wails (hoisted para poder referenciarlos en el describe).
const mocks = vi.hoisted(() => ({
  status: vi.fn(),
  usage: vi.fn(),
  mcp: vi.fn(),
  json: vi.fn(),
}))

vi.mock('../../wailsjs/go/main/App', () => ({
  ServiceStatus: mocks.status,
  ServiceStart: vi.fn(),
  ServiceStop: vi.fn(),
  ServiceForceStop: vi.fn(),
  UsageFetch: mocks.usage,
  ConfigListMCP: mocks.mcp,
  ConfigJSON: mocks.json,
}))

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

const usage = (r: number, w: number, m: number) => ({
  rolling: { percent: r, resetsAt: '' },
  weekly: { percent: w, resetsAt: '' },
  monthly: { percent: m, resetsAt: '' },
})

describe('Resumen', () => {
  it('muestra consumo, config y servicio (en ese orden de bloques)', async () => {
    mocks.status.mockResolvedValue({ running: true, url: 'http://0.0.0.0:49374' })
    mocks.usage.mockResolvedValue(usage(10, 95, 50))
    mocks.mcp.mockResolvedValue([{ name: 'context7', status: 'connected' }])
    mocks.json.mockResolvedValue('{}')

    const w = mount(Resumen, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    const text = w.text()
    expect(text).toContain('Consumo')
    expect(text).toContain('Config')
    expect(text).toContain('Servicio')
    expect(text).toContain('95%')
    expect(text).toContain('http://0.0.0.0:49374')
  })

  it('marca la severidad del consumo (>=98 crítico)', async () => {
    mocks.status.mockResolvedValue({ running: false, url: '' })
    mocks.usage.mockResolvedValue(usage(0, 40, 99))
    mocks.mcp.mockResolvedValue([])
    mocks.json.mockResolvedValue('{}')

    const w = mount(Resumen, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    expect(w.find('.bar-fill.crit').exists()).toBe(true)
  })

  it('el JSON de config queda colapsado por defecto', async () => {
    mocks.status.mockResolvedValue({ running: true, url: '' })
    mocks.usage.mockResolvedValue(usage(1, 1, 1))
    mocks.mcp.mockResolvedValue([])
    mocks.json.mockResolvedValue('{"a":1}')

    const w = mount(Resumen, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    expect(w.find('.acc-head').attributes('aria-expanded')).toBe('false')
  })
})
