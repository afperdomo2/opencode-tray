import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SegmentedControl from './SegmentedControl.vue'

function mountSeg(modelValue = 15) {
  return mount(SegmentedControl, {
    props: {
      modelValue,
      options: [1, 5, 15, 30, 60],
      label: 'Actualizar consumo cada',
      unit: 'min',
      formatLabel: (n: number) => `${n} min`,
    },
  })
}

describe('SegmentedControl', () => {
  it('renderiza una opción por valor y marca la activa', () => {
    const w = mountSeg()
    const opts = w.findAll('.seg-opt')
    expect(opts.length).toBe(5)
    expect(opts.map((o) => o.text())).toEqual(['1', '5', '15', '30', '60'])
    expect(w.find('.seg-opt.active').text()).toBe('15')
    expect(w.find('.seg-unit').text()).toBe('min')
    expect(w.find('[role="group"]').attributes('aria-label')).toBe('Actualizar consumo cada')
  })

  it('emite change con el valor elegido', async () => {
    const w = mountSeg()
    await w.findAll('.seg-opt')[1].trigger('click')
    expect(w.emitted('change')).toEqual([[5]])
  })

  it('no emite change al pulsar la opción activa', async () => {
    const w = mountSeg()
    await w.find('.seg-opt.active').trigger('click')
    expect(w.emitted('change')).toBeUndefined()
  })

  it('expone el texto accesible de cada opción', () => {
    const w = mountSeg()
    expect(w.findAll('.seg-opt')[2].attributes('aria-label')).toBe('15 min')
    expect(w.findAll('.seg-opt')[2].attributes('aria-pressed')).toBe('true')
  })
})
