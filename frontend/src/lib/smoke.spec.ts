import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'

// Componente mínimo para verificar que vitest + jsdom + @vue/test-utils
// funcionan correctamente antes de tener componentes reales.
const Trivial = defineComponent({
  props: { name: { type: String, required: true } },
  setup(props) {
    return () => h('div', `Hola ${props.name}`)
  },
})

describe('test setup', () => {
  it('runs vitest', () => {
    expect(1 + 1).toBe(2)
  })

  it('mounts a Vue component (jsdom + @vue/test-utils)', () => {
    const wrapper = mount(Trivial, { props: { name: 'OpenCode' } })
    expect(wrapper.text()).toBe('Hola OpenCode')
  })
})
