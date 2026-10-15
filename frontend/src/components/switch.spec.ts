import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import Switch from './Switch.vue'

describe('Switch', () => {
  it('refleja el estado checked', () => {
    const w = mount(Switch, { props: { checked: true } })
    expect((w.find('input').element as HTMLInputElement).checked).toBe(true)
  })

  it('emite "change" con el nuevo valor', async () => {
    const w = mount(Switch, { props: { checked: false } })
    const input = w.find('input')
    ;(input.element as HTMLInputElement).checked = true
    await input.trigger('change')
    expect(w.emitted('change')?.[0]).toEqual([true])
  })

  it('muestra el label por slot', () => {
    const w = mount(Switch, {
      props: { checked: false },
      slots: { default: 'Iniciar con Windows' },
    })
    expect(w.text()).toContain('Iniciar con Windows')
  })
})
