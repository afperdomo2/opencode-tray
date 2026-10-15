import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import Accordion from './Accordion.vue'

// `v-show` aplica display:none inline; se comprueba el estilo directamente
// (más robusto que isVisible() en jsdom).
const display = (w: ReturnType<typeof mount>) =>
  (w.find('.acc-panel').element as HTMLElement).style.display

describe('Accordion', () => {
  it('está colapsado por defecto', () => {
    const w = mount(Accordion, {
      props: { title: 'Config' },
      slots: { default: '<pre>json</pre>' },
    })
    expect(w.find('.acc-head').attributes('aria-expanded')).toBe('false')
    expect(display(w)).toBe('none')
  })

  it('abre y cierra al hacer clic', async () => {
    const w = mount(Accordion, { props: { title: 'Config' } })
    await w.find('.acc-head').trigger('click')
    expect(w.find('.acc-head').attributes('aria-expanded')).toBe('true')
    expect(display(w)).not.toBe('none')
    await w.find('.acc-head').trigger('click')
    expect(w.find('.acc-head').attributes('aria-expanded')).toBe('false')
    expect(display(w)).toBe('none')
  })

  it('aria-controls apunta al id del panel', () => {
    const w = mount(Accordion, { props: { title: 'Config' } })
    const controls = w.find('.acc-head').attributes('aria-controls')
    expect(controls).toBeTruthy()
    expect(w.find('.acc-panel').attributes('id')).toBe(controls)
  })
})
