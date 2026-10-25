import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import Resumen from './Resumen.vue'
import { clearQueryCache, clearQueryData, prefetchQuery } from '../lib/query'
import { SERVICE_KEY } from '../lib/query-keys'
import { USAGE_KEY } from '../lib/usage'

// Mocks de los bindings de Wails (hoisted para poder referenciarlos en el describe).
const mocks = vi.hoisted(() => ({
  status: vi.fn(),
  usage: vi.fn(),
  mcp: vi.fn(),
  access: vi.fn(),
  tokenStatus: vi.fn(),
  reload: vi.fn(),
  setToken: vi.fn(),
  clearToken: vi.fn(),
  getSettings: vi.fn(),
  setLanguage: vi.fn(),
  setAutoStart: vi.fn(),
  setRefreshMin: vi.fn(),
}))

vi.mock('../../wailsjs/go/main/App', () => ({
  ServiceStatus: mocks.status,
  UsageFetch: mocks.usage,
  ConfigListMCP: mocks.mcp,
  WebAccessState: mocks.access,
  UsageTokenStatus: mocks.tokenStatus,
  UsageReloadToken: mocks.reload,
  UsageSetManualToken: mocks.setToken,
  UsageClearManualToken: mocks.clearToken,
  GetSettings: mocks.getSettings,
  SetLanguage: mocks.setLanguage,
  SetAutoStart: mocks.setAutoStart,
  SetRefreshMin: mocks.setRefreshMin,
}))

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

const usage = (r: number, w: number, m: number) => ({
  rolling: { percent: r, resetsAt: '', hasReset: false },
  weekly: { percent: w, resetsAt: '', hasReset: false },
  monthly: { percent: m, resetsAt: '', hasReset: false },
})

function setDefaults() {
  mocks.status.mockResolvedValue({ running: true, url: 'http://0.0.0.0:49374' })
  mocks.usage.mockResolvedValue(usage(10, 50, 20))
  mocks.mcp.mockResolvedValue([{ name: 'context7', status: 'connected' }])
  mocks.access.mockResolvedValue({
    mode: 'local',
    running: true,
    urls: [{ url: 'http://127.0.0.1:49374', kind: 'local' }],
  })
  mocks.tokenStatus.mockResolvedValue({ configured: true, source: 'opencode-export' })
  mocks.reload.mockResolvedValue({ configured: true, source: 'opencode-export' })
  mocks.setToken.mockResolvedValue(undefined)
  mocks.clearToken.mockResolvedValue(undefined)
  mocks.getSettings.mockResolvedValue({ language: 'es', autoStart: false, refreshMin: 15 })
  mocks.setLanguage.mockResolvedValue(undefined)
  mocks.setAutoStart.mockResolvedValue(undefined)
  mocks.setRefreshMin.mockResolvedValue(undefined)
}

async function mountResumen() {
  const wrapper = mount(Resumen, { global: { plugins: [makeI18n()] } })
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  vi.clearAllMocks()
  clearQueryCache()
  setDefaults()
})

describe('Resumen', () => {
  it('muestra los cuatro widgets del dashboard', async () => {
    const w = await mountResumen()

    expect(w.findAll('.widget-grid .widget').length).toBe(4)
    expect(w.find('.widget-consumo').findAll('.bar-fill').length).toBe(3)
    expect(w.find('.widget-servicio').text()).toContain('Corriendo')
    expect(w.find('.widget-acceso').text()).toContain('Local')
    expect(w.find('.widget-mcp').text()).toContain('context7')
  })

  it('muestra estado y URL en el widget de Servicio', async () => {
    const w = await mountResumen()
    const service = w.find('.widget-servicio')

    expect(service.find('.widget-state').text()).toContain('Corriendo')
    expect(service.find('.widget-url').text()).toBe('0.0.0.0:49374')
    expect(service.find('.widget-url').attributes('title')).toBe('http://0.0.0.0:49374')
  })

  it('muestra el modo de acceso y el aviso de Internet en remoto', async () => {
    mocks.access.mockResolvedValue({
      mode: 'remote',
      running: true,
      urls: [{ url: 'https://abc.opentunnel.xyz', kind: 'remote' }],
    })

    const w = await mountResumen()
    const access = w.find('.widget-acceso')

    expect(access.text()).toContain('Remoto')
    expect(access.find('.access-chip').text()).toContain('Internet')
    expect(access.classes()).toContain('tone-remote')
  })

  it('muestra los chips de MCP sin el texto de estado por item', async () => {
    mocks.mcp.mockResolvedValue([
      { name: 'context7', status: 'connected' },
      { name: 'github', status: 'error' },
    ])

    const w = await mountResumen()
    const mcp = w.find('.widget-mcp')

    expect(mcp.findAll('.mcp-chip').length).toBe(2)
    expect(mcp.text()).not.toContain('connected')
    expect(mcp.text()).not.toContain('error')
    expect(mcp.findAll('.dot')[0].attributes('aria-label')).toBe('context7: conectado')
  })

  it('navega al hacer clic o pulsar Enter en Servicio', async () => {
    const w = await mountResumen()
    const service = w.find('.widget-servicio')

    await service.trigger('click')
    await service.trigger('keydown', { key: 'Enter' })

    const emitted = w.emitted('navigate')
    expect(emitted?.[0]).toEqual(['servicio'])
    expect(emitted?.[1]).toEqual(['servicio'])
  })

  it('navega a Acceso web y a Servicio desde MCP', async () => {
    const w = await mountResumen()

    await w.find('.widget-acceso').trigger('click')
    await w.find('.widget-mcp').trigger('click')

    const emitted = w.emitted('navigate')
    expect(emitted?.[0]).toEqual(['accesoWeb'])
    expect(emitted?.[1]).toEqual(['servicio'])
  })

  it('renderiza consumo aunque el servicio siga cargando (skeleton independiente)', async () => {
    let resolveStatus: (value: { running: boolean; url: string }) => void = () => {}
    mocks.status.mockReturnValue(
      new Promise((resolve) => {
        resolveStatus = resolve
      }),
    )
    mocks.usage.mockResolvedValue(usage(10, 50, 20))
    mocks.mcp.mockResolvedValue([])

    const w = mount(Resumen, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    expect(w.find('.widget-consumo').findAll('.bar-fill').length).toBe(3)
    expect(w.find('.widget-servicio').find('.skeleton').exists()).toBe(true)

    resolveStatus({ running: true, url: 'http://127.0.0.1:49374' })
    await flushPromises()

    expect(w.find('.widget-servicio').find('.skeleton').exists()).toBe(false)
    expect(w.find('.widget-servicio').text()).toContain('Corriendo')
  })

  it('muestra skeleton independiente en Acceso mientras carga', async () => {
    mocks.access.mockReturnValue(new Promise(() => {}))

    const w = await mountResumen()

    expect(w.find('.widget-acceso').find('.skeleton').exists()).toBe(true)
    expect(w.find('.widget-servicio').find('.skeleton').exists()).toBe(false)
  })

  it('es solo resumen: sin error no muestra botones de acción', async () => {
    const w = await mountResumen()
    expect(w.find('.actions').exists()).toBe(false)
    expect(w.findAll('button').length).toBe(0)
  })

  it('colorea el consumo de forma graduada (crítico >= 98)', async () => {
    mocks.usage.mockResolvedValue(usage(0, 40, 99))
    const w = await mountResumen()
    expect(w.find('.widget-consumo .bar-fill.crit').exists()).toBe(true)
    expect(w.find('.widget-consumo .bar-pct.crit').exists()).toBe(true)
  })

  it('usa ámbar/naranja según proximidad al límite', async () => {
    mocks.usage.mockResolvedValue(usage(72, 90, 10))
    const w = await mountResumen()
    expect(w.find('.widget-consumo .bar-fill.warn').exists()).toBe(true)
    expect(w.find('.widget-consumo .bar-fill.high').exists()).toBe(true)
  })

  it('reutiliza el consumo cacheado al remontar (sin skeleton ni refetch)', async () => {
    const first = await mountResumen()
    expect(mocks.usage).toHaveBeenCalledTimes(1)
    first.unmount()

    const second = await mountResumen()
    expect(mocks.usage).toHaveBeenCalledTimes(1)
    expect(second.findAll('.bar-fill').length).toBe(3)
    expect(second.find('.skeleton').exists()).toBe(false)
  })

  it('reintenta una vez leyendo el token de OpenCode si falta', async () => {
    mocks.usage
      .mockRejectedValueOnce('USAGE_NO_TOKEN: sin token')
      .mockResolvedValueOnce(usage(1, 2, 3))

    const w = await mountResumen()

    expect(mocks.reload).toHaveBeenCalledTimes(1)
    expect(mocks.usage).toHaveBeenCalledTimes(2)
    expect(w.findAll('.bar-fill').length).toBe(3)
    expect(w.find('.token-retry').exists()).toBe(false)
  })

  it('si el reintento falla muestra el error y el CTA de token', async () => {
    mocks.usage.mockRejectedValue('USAGE_NO_TOKEN: sin token')
    const w = await mountResumen()

    expect(w.text()).toContain('No hay token de OpenCode Go')
    expect(w.find('.token-retry').exists()).toBe(true)
    expect(w.find('.token-retry').text()).toContain('Reintentar leyendo de OpenCode')

    await w.find('.token-retry').trigger('click')
    await flushPromises()
    expect(mocks.reload).toHaveBeenCalled()
  })

  it('sustituye los datos en silencio durante un refresco con datos previos', async () => {
    mocks.usage.mockResolvedValueOnce(usage(10, 20, 30))
    const first = await mountResumen()
    first.unmount()

    let finish: (u: ReturnType<typeof usage>) => void = () => {}
    const pending = new Promise<ReturnType<typeof usage>>((resolve) => {
      finish = resolve
    })
    void prefetchQuery(USAGE_KEY, () => pending)
    await Promise.resolve()

    const second = await mountResumen()
    expect(second.findAll('.bar-fill').length).toBe(3)
    expect(second.text()).not.toContain('Actualizando')

    finish(usage(11, 21, 31))
    await flushPromises()
    expect(second.find('.widget-consumo .bar-pct').text()).toBe('11%')
    second.unmount()
  })

  it('comparte el estado del servicio al remontar (sin skeleton ni refetch)', async () => {
    const first = await mountResumen()
    expect(mocks.status).toHaveBeenCalledTimes(1)
    first.unmount()

    const second = await mountResumen()
    expect(mocks.status).toHaveBeenCalledTimes(1)
    expect(second.find('.widget-servicio').find('.skeleton').exists()).toBe(false)
    second.unmount()
  })

  it('muestra skeleton al invalidar el estado (mutación en otra pestaña)', async () => {
    const first = await mountResumen()
    first.unmount()

    clearQueryData(SERVICE_KEY)
    mocks.status.mockReturnValueOnce(new Promise(() => {}))

    const second = mount(Resumen, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    expect(second.find('.widget-servicio').find('.skeleton').exists()).toBe(true)
    second.unmount()
  })

  it('conserva las barras y avisa si falla un refresco con datos previos', async () => {
    mocks.usage.mockResolvedValueOnce(usage(10, 20, 30))
    const first = await mountResumen()
    expect(first.findAll('.bar-fill').length).toBe(3)
    first.unmount()

    await prefetchQuery(USAGE_KEY, vi.fn().mockRejectedValue('USAGE_TRANSPORT: sin red'))
    const second = await mountResumen()

    expect(second.findAll('.bar-fill').length).toBe(3)
    expect(second.text()).toContain('Mostrando los últimos datos')
  })

  it('el CTA de configurar API key navega a Ajustes', async () => {
    mocks.usage.mockRejectedValue('USAGE_NO_TOKEN: sin token')
    const w = await mountResumen()

    const configure = w.findAll('button').find((b) => b.text().includes('Configurar API key'))
    expect(configure).toBeTruthy()
    await configure!.trigger('click')
    const emitted = w.emitted('navigate')
    expect(emitted?.[0]).toEqual(['ajustes'])
  })
})
