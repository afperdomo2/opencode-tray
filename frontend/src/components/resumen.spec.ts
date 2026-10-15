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
  ConfigOpenInEditor: vi.fn(),
  ConfigOpenDir: vi.fn(),
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
  mocks.json.mockResolvedValue('{}')
  const wrapper = mount(Resumen, { global: { plugins: [makeI18n()] } })
  await flushPromises()
  return wrapper
}

describe('Resumen', () => {
  it('muestra consumo, config y servicio (en ese orden de bloques)', async () => {
    const w = await mountResumen(10, 50, 20)
    const text = w.text()
    expect(text).toContain('Consumo')
    expect(text).toContain('Config')
    expect(text).toContain('Servicio')
    expect(text).toContain('http://0.0.0.0:49374')
  })

  it('el bloque de servicio usa botones "sm" con texto (4 botones)', async () => {
    const w = await mountResumen(10, 50, 20)
    const smGroups = w.findAll('.actions.sm')
    const serviceButtons = smGroups[smGroups.length - 1].findAll('button')
    expect(serviceButtons.length).toBe(4)
    expect(serviceButtons.map((b) => b.text())).toEqual([
      'Iniciar',
      'Detener',
      'Forzar detener',
      'Actualizar',
    ])
  })

  it('el bloque de config tiene 3 botones', async () => {
    const w = await mountResumen(10, 50, 20)
    const cards = w.findAll('.card')
    expect(cards[1].findAll('.actions button').length).toBe(3)
  })

  it('el botón de JSON abre el modal', async () => {
    const w = await mountResumen(10, 50, 20)
    expect(w.find('.modal').exists()).toBe(false)
    const configCard = w.findAll('.card')[1]
    await configCard.findAll('.actions button')[2].trigger('click')
    expect(w.find('.modal').exists()).toBe(true)
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
