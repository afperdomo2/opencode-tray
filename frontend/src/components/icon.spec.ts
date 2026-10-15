import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import Icon from './Icon.vue'

describe('Icon', () => {
  it('renderiza un <svg> para un nombre conocido', () => {
    const w = mount(Icon, { props: { name: 'globe' } })
    expect(w.find('svg').exists()).toBe(true)
    expect(w.find('svg').attributes('viewBox')).toBe('0 0 24 24')
  })

  it('es decorativo por defecto (aria-hidden)', () => {
    const w = mount(Icon, { props: { name: 'chart' } })
    expect(w.attributes('aria-hidden')).toBe('true')
  })

  it('un nombre desconocido no rompe (svg vacío pero presente)', () => {
    const w = mount(Icon, { props: { name: 'no-existe' } })
    expect(w.find('svg').exists()).toBe(true)
  })
})
