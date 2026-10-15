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
}))

vi.mock('../../wailsjs/go/main/App', () => ({
  ServiceStatus: mocks.status,
  UsageFetch: mocks.usage,
  ConfigListMCP: mocks.mcp,
}))

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

const usage = (r: number, w: number, m: number) => ({
  rolling: { percent: r, resetsAt: '' },
  weekly: { percent: w, resetsAt: '' },
  monthly: { percent: m, resetsAt: '' },
})

async function mountResumen(r: number, w: number, m: number, running = true) {
  mocks.status.mockResolvedValue({ running, url: running ? 'http://0.0.0.0:49374' : '' })
  mocks.usage.mockResolvedValue(usage(r, w, m))
  mocks.mcp.mockResolvedValue([{ name: 'context7', status: 'connected' }])
  const wrapper = mount(Resumen, { global: { plugins: [makeI18n()] } })
  await flushPromises()
  return wrapper
}

describe('Resumen', () => {
  it('muestra los widgets de Consumo y Servicio (con MCP)', async () => {
    const w = await mountResumen(10, 50, 20)
    const text = w.text()
    expect(text).toContain('Consumo')
    expect(text).toContain('Servicio')
    expect(text).toContain('context7')
    expect(text).toContain('http://0.0.0.0:49374')
  })

  it('es solo resumen: no tiene botones de acción', async () => {
    const w = await mountResumen(10, 50, 20)
    expect(w.find('.actions').exists()).toBe(false)
    expect(w.findAll('button').length).toBe(0)
  })

  it('colorea el consumo de forma graduada (crítico >= 98)', async () => {
    const w = await mountResumen(0, 40, 99)
    expect(w.find('.bar-fill.crit').exists()).toBe(true)
    expect(w.find('.bar-pct.crit').exists()).toBe(true)
  })

  it('usa ámbar/naranja según proximidad al límite', async () => {
    const w = await mountResumen(72, 90, 10)
    expect(w.find('.bar-fill.warn').exists()).toBe(true) // 72% -> ámbar
    expect(w.find('.bar-fill.high').exists()).toBe(true) // 90% -> naranja
  })
})
