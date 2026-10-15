import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import es from '../i18n/es.json'
import en from '../i18n/en.json'
import Modal from './Modal.vue'

const i18n = () =>
  createI18n({ legacy: false, locale: 'es', fallbackLocale: 'en', messages: { es, en } })

describe('Modal', () => {
  it('no renderiza nada si está cerrado', () => {
    const w = mount(Modal, {
      props: { open: false, title: 'JSON' },
      global: { plugins: [i18n()] },
    })
    expect(w.find('.modal').exists()).toBe(false)
  })

  it('abre con role=dialog y aria-modal', () => {
    const w = mount(Modal, {
      props: { open: true, title: 'JSON' },
      global: { plugins: [i18n()] },
    })
    const dlg = w.find('.modal')
    expect(dlg.exists()).toBe(true)
    expect(dlg.attributes('role')).toBe('dialog')
    expect(dlg.attributes('aria-modal')).toBe('true')
  })

  it('emite "close" al pulsar Esc y con el botón de cerrar', async () => {
    const w = mount(Modal, {
      props: { open: true, title: 'JSON' },
      global: { plugins: [i18n()] },
    })
    await w.find('.modal').trigger('keydown', { key: 'Escape' })
    await w.find('.modal-head .icon-btn').trigger('click')
    expect(w.emitted('close')?.length).toBe(2)
  })
})
