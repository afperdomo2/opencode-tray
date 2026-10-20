import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import AccesoWeb from './AccesoWeb.vue'

const mocks = vi.hoisted(() => ({
  state: vi.fn(),
  username: vi.fn(),
  password: vi.fn(),
  changePassword: vi.fn(),
  applyMode: vi.fn(),
  openUrl: vi.fn(),
  clipboard: vi.fn(),
}))

vi.mock('../../wailsjs/go/main/App', () => ({
  WebAccessState: mocks.state,
  WebAccessUsername: mocks.username,
  WebAccessCurrentPassword: mocks.password,
  WebAccessChangePassword: mocks.changePassword,
  WebAccessApplyMode: mocks.applyMode,
  WebAccessOpenURL: mocks.openUrl,
}))

vi.mock('../../wailsjs/runtime/runtime', () => ({
  ClipboardSetText: mocks.clipboard,
}))

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

function localState() {
  return {
    mode: 'local',
    running: true,
    urls: [{ url: 'http://127.0.0.1:45278', kind: 'local' }],
  }
}

beforeEach(() => {
  vi.clearAllMocks()
  mocks.username.mockResolvedValue('opencode')
  mocks.password.mockResolvedValue('secreto')
  mocks.clipboard.mockResolvedValue(true)
})

describe('AccesoWeb', () => {
  it('usa un panel grid, marca el modo activo y muestra las URLs', async () => {
    mocks.state.mockResolvedValue(localState())

    const w = mount(AccesoWeb, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    expect(w.find('.panel-grid').exists()).toBe(true)
    expect(w.findAll('.panel-grid > .card').length).toBe(2)
    expect(w.find('.mode-btn.active').text()).toContain('Local')
    expect(w.text()).toContain('http://127.0.0.1:45278')
    expect(w.findAll('.icon-btn').some((b) => b.attributes('title') === 'Mostrar')).toBe(true)
  })

  it('muestra un skeleton de 2 filas mientras cargan las URLs', async () => {
    mocks.state.mockReturnValue(new Promise(() => {}))

    const w = mount(AccesoWeb, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    expect(w.findAll('.skeleton-url-row').length).toBe(2)
    expect(w.findAll('.url-row').length).toBe(2)
  })

  it('abre la URL validada en el navegador', async () => {
    mocks.state.mockResolvedValue(localState())

    const w = mount(AccesoWeb, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    const buttons = w.findAll('.url-action')
    await buttons[0].trigger('click')
    await flushPromises()

    expect(mocks.openUrl).toHaveBeenCalledWith('http://127.0.0.1:45278')
  })

  it('copia la URL y muestra confirmación', async () => {
    mocks.state.mockResolvedValue(localState())

    const w = mount(AccesoWeb, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    const buttons = w.findAll('.url-action')
    await buttons[1].trigger('click')
    await flushPromises()

    expect(mocks.clipboard).toHaveBeenCalledWith('http://127.0.0.1:45278')
    expect(w.text()).toContain('Copiada')
  })

  it('pide confirmación antes de activar el modo remoto', async () => {
    mocks.state.mockResolvedValue(localState())
    mocks.applyMode.mockResolvedValue(undefined)

    const w = mount(AccesoWeb, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    await w.findAll('.mode-btn')[2].trigger('click')
    expect(w.find('.modal').exists()).toBe(true)
    expect(mocks.applyMode).not.toHaveBeenCalled()

    const confirm = w.findAll('.modal-actions .btn')[1]
    await confirm.trigger('click')
    await flushPromises()

    expect(mocks.applyMode).toHaveBeenCalledWith('remote')
  })

  it('aplica el modo local sin confirmación', async () => {
    mocks.state.mockResolvedValue({ ...localState(), mode: 'network' })
    mocks.applyMode.mockResolvedValue(undefined)

    const w = mount(AccesoWeb, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    await w.findAll('.mode-btn')[0].trigger('click')
    await flushPromises()

    expect(mocks.applyMode).toHaveBeenCalledWith('local')
  })

  it('deshabilita abrir si el servicio está detenido', async () => {
    mocks.state.mockResolvedValue({ ...localState(), running: false })

    const w = mount(AccesoWeb, { global: { plugins: [makeI18n()] } })
    await flushPromises()

    const open = w.findAll('.url-action')[0]
    expect(open.attributes('disabled')).toBeDefined()
    expect(w.text()).toContain('El servicio está detenido')
  })
})
