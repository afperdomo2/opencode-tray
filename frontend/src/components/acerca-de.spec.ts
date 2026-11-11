import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import AcercaDe from './AcercaDe.vue'

const mocks = vi.hoisted(() => ({
  version: vi.fn(),
  clipboard: vi.fn(),
}))

vi.mock('../../wailsjs/go/main/App', () => ({
  GetAppVersion: mocks.version,
}))

vi.mock('../../wailsjs/runtime/runtime', () => ({
  ClipboardSetText: mocks.clipboard,
}))

const makeI18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

const sampleVersion = {
  version: '1.4.0',
  commit: 'a1b2c3d',
  date: '2026-11-11T15:04:05Z',
  dirty: false,
  platform: 'linux/amd64',
}

async function mountAbout() {
  const wrapper = mount(AcercaDe, { global: { plugins: [makeI18n()] } })
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  vi.clearAllMocks()
  mocks.version.mockResolvedValue(sampleVersion)
  mocks.clipboard.mockResolvedValue(true)
})

describe('AcercaDe', () => {
  it('muestra versión, compilación, fecha y plataforma', async () => {
    const w = await mountAbout()
    expect(w.find('.about-version').text()).toBe('v1.4.0')
    expect(w.find('.about-commit').text()).toBe('a1b2c3d')
    expect(w.find('.about-platform').text()).toBe('linux/amd64')
    expect(w.find('.about-date').text()).not.toBe('No disponible')
  })

  it('marca los builds con cambios sin commitear', async () => {
    mocks.version.mockResolvedValue({ ...sampleVersion, dirty: true })
    const w = await mountAbout()
    const commit = w.find('.about-commit')
    expect(commit.text()).toBe('a1b2c3d *')
    expect(commit.attributes('title')).toBe('Hay cambios sin commitear')
  })

  it('copia la información de la compilación', async () => {
    const w = await mountAbout()
    const button = w.findAll('button').find((b) => b.text().includes('Copiar información'))
    expect(button).toBeTruthy()
    await button!.trigger('click')
    await flushPromises()

    expect(mocks.clipboard).toHaveBeenCalledTimes(1)
    const copied = mocks.clipboard.mock.calls[0][0] as string
    expect(copied).toContain('OpenCode Tray v1.4.0')
    expect(copied).toContain('Compilación: a1b2c3d')
    expect(copied).toContain('Plataforma: linux/amd64')
    expect(w.text()).toContain('Información copiada')
  })

  it('muestra error si el portapapeles falla', async () => {
    mocks.clipboard.mockResolvedValue(false)
    const w = await mountAbout()
    const button = w.findAll('button').find((b) => b.text().includes('Copiar información'))
    await button!.trigger('click')
    await flushPromises()
    expect(w.text()).toContain('No se pudo copiar la información.')
  })

  it('usa "No disponible" si el backend no responde', async () => {
    mocks.version.mockRejectedValue(new Error('sin puente'))
    const w = await mountAbout()
    expect(w.find('.about-version').text()).toBe('No disponible')
    expect(w.find('.about-commit').text()).toBe('No disponible')
    expect(w.find('.about-platform').text()).toBe('No disponible')
  })
})
