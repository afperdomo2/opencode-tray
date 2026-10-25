import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import Servicio from './Servicio.vue'
import { clearQueryCache } from '../lib/query'

const mocks = vi.hoisted(() => ({
  status: vi.fn(),
  mcp: vi.fn(),
  json: vi.fn(),
  openEditor: vi.fn(),
  openDir: vi.fn(),
  start: vi.fn(),
  stop: vi.fn(),
  forceStop: vi.fn(),
}))

vi.mock('../../wailsjs/go/main/App', () => ({
  ServiceStatus: mocks.status,
  ServiceStart: mocks.start,
  ServiceStop: mocks.stop,
  ServiceForceStop: mocks.forceStop,
  ConfigListMCP: mocks.mcp,
  ConfigJSON: mocks.json,
  ConfigOpenInEditor: mocks.openEditor,
  ConfigOpenDir: mocks.openDir,
}))

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

beforeEach(() => {
  vi.clearAllMocks()
  clearQueryCache()
})

async function mountServicio() {
  mocks.status.mockResolvedValue({ running: true, url: 'http://127.0.0.1:49374' })
  mocks.mcp.mockResolvedValue([{ name: 'context7', status: 'connected' }])
  mocks.json.mockResolvedValue('{}')
  const w = mount(Servicio, { global: { plugins: [makeI18n()] } })
  await flushPromises()
  return w
}

describe('Servicio', () => {
  it('usa un panel grid con control y configuración', async () => {
    const w = await mountServicio()
    expect(w.find('.panel-grid').exists()).toBe(true)
    expect(w.findAll('.panel-grid > .card').length).toBe(2)
    expect(w.text()).toContain('Corriendo')
    expect(w.text()).toContain('context7')

    const statusRow = w.find('.status-row')
    expect(statusRow.find('.url').text()).toBe('http://127.0.0.1:49374')
  })

  it('usa etiquetas cortas y tooltips en las acciones', async () => {
    const w = await mountServicio()

    expect(w.findAll('.actions-grid.cols-2 .btn').length).toBe(4)
    expect(w.findAll('.actions-grid.cols-3 .btn').length).toBe(3)

    const buttons = w.findAll('.actions-grid .btn')
    expect(buttons.map((b) => b.text())).toEqual([
      'Iniciar',
      'Detener',
      'Forzar',
      'Refrescar',
      'Editor',
      'Carpeta',
      'JSON',
    ])
    expect(buttons.map((b) => b.attributes('title'))).toEqual([
      'Iniciar servicio',
      'Detener servicio',
      'Forzar detención del servicio',
      'Actualizar estado y configuración',
      'Abrir en editor',
      'Abrir carpeta',
      'Ver configuración JSON',
    ])
  })

  it('muestra skeleton mientras el estado del servicio carga', async () => {
    mocks.status.mockReturnValue(new Promise(() => {}))
    mocks.mcp.mockResolvedValue([])
    mocks.json.mockResolvedValue('{}')

    const w = mount(Servicio, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    const cards = w.findAll('.panel-grid > .card')
    expect(cards[0].find('.skeleton').exists()).toBe(true)
    expect(cards[0].find('.status-row').exists()).toBe(false)
  })

  it('reutiliza el estado cacheado al remontar (sin skeleton ni refetch)', async () => {
    const first = await mountServicio()
    expect(mocks.status).toHaveBeenCalledTimes(1)
    first.unmount()

    const second = await mountServicio()
    expect(mocks.status).toHaveBeenCalledTimes(1)
    expect(second.find('.panel-grid > .card').find('.skeleton').exists()).toBe(false)
    expect(second.text()).toContain('Corriendo')
  })

  it('el refresco manual descarta los datos y muestra skeleton', async () => {
    const w = await mountServicio()
    mocks.status.mockReturnValue(new Promise(() => {}))

    await w.findAll('.actions-grid.cols-2 .btn')[3].trigger('click')
    await flushPromises()

    expect(w.findAll('.panel-grid > .card')[0].find('.skeleton').exists()).toBe(true)
  })

  it('muestra skeleton desde el inicio de una mutación', async () => {
    mocks.stop.mockReturnValue(new Promise(() => {}))
    const w = await mountServicio()

    await w.findAll('.actions-grid.cols-2 .btn')[1].trigger('click')
    await flushPromises()

    expect(w.findAll('.panel-grid > .card')[0].find('.skeleton').exists()).toBe(true)
  })

  it('muestra los MCP sin esperar a que cargue el JSON', async () => {
    mocks.status.mockResolvedValue({ running: true, url: 'http://127.0.0.1:49374' })
    mocks.mcp.mockResolvedValue([{ name: 'context7', status: 'connected' }])
    mocks.json.mockReturnValue(new Promise(() => {}))

    const w = mount(Servicio, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    const configCard = w.findAll('.panel-grid > .card')[1]
    expect(configCard.text()).toContain('context7')
    expect(configCard.find('.skeleton').exists()).toBe(false)
  })
})
