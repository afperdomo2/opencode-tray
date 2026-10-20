import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import Skeleton from './Skeleton.vue'

describe('Skeleton', () => {
  it('es decorativo y aplica las dimensiones recibidas', () => {
    const w = mount(Skeleton, { props: { width: '50%', height: '10px' } })
    expect(w.attributes('aria-hidden')).toBe('true')
    expect(w.attributes('style')).toContain('width: 50%')
    expect(w.attributes('style')).toContain('height: 10px')
  })
})
