import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { clearQueryCache, invalidateQuery, prefetchQuery, useQuery } from './query'

function probe(key: string, fetcher: () => Promise<{ v: number }>) {
  return defineComponent({
    setup() {
      const q = useQuery({ key, fetcher, staleTime: 60_000 })
      return () =>
        h('div', [
          h('span', { class: 'data' }, q.data.value ? String(q.data.value.v) : 'empty'),
          h('span', { class: 'error' }, q.error.value ? q.error.value.message : ''),
          h('span', { class: 'loading' }, q.isInitialLoading.value ? 'loading' : 'ready'),
        ])
    },
  })
}

beforeEach(() => {
  clearQueryCache()
  vi.useRealTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

describe('query cache', () => {
  it('reutiliza datos frescos al remontar y no vuelve a pedir', async () => {
    const fetcher = vi.fn().mockResolvedValue({ v: 1 })
    const first = mount(probe('k1', fetcher))
    await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(1)
    expect(first.find('.data').text()).toBe('1')
    first.unmount()

    const second = mount(probe('k1', fetcher))
    await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(1)
    expect(second.find('.data').text()).toBe('1')
    expect(second.find('.loading').text()).toBe('ready')
  })

  it('invalidateQuery fuerza un refetch al remontar', async () => {
    const fetcher = vi.fn().mockResolvedValue({ v: 1 })
    const first = mount(probe('k2', fetcher))
    await flushPromises()
    first.unmount()

    invalidateQuery('k2')
    mount(probe('k2', fetcher))
    await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(2)
  })

  it('deduplica peticiones concurrentes', async () => {
    let resolve: (value: { v: number }) => void = () => {}
    const fetcher = vi.fn(
      () =>
        new Promise<{ v: number }>((r) => {
          resolve = r
        }),
    )
    const p1 = prefetchQuery('k3', fetcher)
    const p2 = prefetchQuery('k3', fetcher)
    await Promise.resolve()
    expect(fetcher).toHaveBeenCalledTimes(1)
    resolve({ v: 3 })
    await Promise.all([p1, p2])
  })

  it('expone el error y conserva el estado para reintentar', async () => {
    const fetcher = vi.fn().mockRejectedValue(new Error('boom'))
    const w = mount(probe('k4', fetcher))
    await flushPromises()
    expect(w.find('.error').text()).toBe('boom')
    expect(w.find('.data').text()).toBe('empty')

    fetcher.mockResolvedValue({ v: 4 })
    await prefetchQuery('k4', fetcher)
    await flushPromises()
    expect(w.find('.error').text()).toBe('')
    expect(w.find('.data').text()).toBe('4')
  })

  it('refetch manual vuelve a pedir aunque los datos estén frescos', async () => {
    const fetcher = vi.fn().mockResolvedValue({ v: 1 })
    const probeWithRefetch = defineComponent({
      setup() {
        const q = useQuery({ key: 'k5', fetcher, staleTime: 60_000 })
        return () => h('button', { onClick: () => void q.refetch() }, 'refetch')
      },
    })
    const w = mount(probeWithRefetch)
    await flushPromises()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(2)
  })

  it('el refetch duro descarta los datos y muestra carga hasta recibir los nuevos', async () => {
    let resolve: (value: { v: number }) => void = () => {}
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce({ v: 1 })
      .mockImplementationOnce(
        () =>
          new Promise<{ v: number }>((r) => {
            resolve = r
          }),
      )
    const hardProbe = defineComponent({
      setup() {
        const q = useQuery<{ v: number }>({ key: 'k7', fetcher, staleTime: 60_000 })
        return () =>
          h('div', [
            h('span', { class: 'data' }, q.data.value ? String(q.data.value.v) : 'empty'),
            h('span', { class: 'loading' }, q.isInitialLoading.value ? 'loading' : 'ready'),
            h('button', { onClick: () => void q.refetchHard() }, 'hard'),
          ])
      },
    })
    const w = mount(hardProbe)
    await flushPromises()
    expect(w.find('.data').text()).toBe('1')

    await w.find('button').trigger('click')
    await flushPromises()
    expect(w.find('.data').text()).toBe('empty')
    expect(w.find('.loading').text()).toBe('loading')

    resolve({ v: 2 })
    await flushPromises()
    expect(w.find('.data').text()).toBe('2')
    expect(w.find('.loading').text()).toBe('ready')
  })

  it('hace polling con el intervalo configurado y lo detiene al desmontar', async () => {
    vi.useFakeTimers()
    const fetcher = vi.fn().mockResolvedValue({ v: 1 })
    const polling = defineComponent({
      setup() {
        const q = useQuery<{ v: number }>({
          key: 'k6',
          fetcher,
          staleTime: 60_000,
          refetchInterval: () => 1000,
        })
        return () => h('div', q.data.value ? String(q.data.value.v) : 'empty')
      },
    })
    const w = mount(polling)
    await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(1000)
    expect(fetcher).toHaveBeenCalledTimes(2)

    w.unmount()
    await vi.advanceTimersByTimeAsync(3000)
    expect(fetcher).toHaveBeenCalledTimes(2)
  })
})
