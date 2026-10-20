import { describe, it, expect, vi, beforeEach } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { clearQueryCache } from './query'
import { useUsageQuery } from './usage'

const mocks = vi.hoisted(() => ({
  fetch: vi.fn(),
  reload: vi.fn(),
  status: vi.fn(),
  setToken: vi.fn(),
  clearToken: vi.fn(),
}))

vi.mock('../../wailsjs/go/main/App', () => ({
  UsageFetch: mocks.fetch,
  UsageReloadToken: mocks.reload,
  UsageTokenStatus: mocks.status,
  UsageSetManualToken: mocks.setToken,
  UsageClearManualToken: mocks.clearToken,
}))

vi.mock('./settings', () => ({
  refreshIntervalMs: () => false,
}))

let api: ReturnType<typeof useUsageQuery> | undefined

const probe = defineComponent({
  setup() {
    api = useUsageQuery()
    return () => h('div', api?.error.value?.message ?? '')
  },
})

const tokenError = () => new Error('USAGE_NO_TOKEN: falta el token')
const usage = {
  rolling: { percent: 10, resetsAt: '' },
  weekly: { percent: 20, resetsAt: '' },
  monthly: { percent: 30, resetsAt: '' },
}

beforeEach(() => {
  vi.clearAllMocks()
  clearQueryCache()
  api = undefined
  mocks.reload.mockResolvedValue({ configured: true, source: 'opencode-export' })
})

describe('usage query', () => {
  it('reintenta una vez por carga y vuelve a permitirlo en la siguiente', async () => {
    mocks.fetch.mockRejectedValue(tokenError())
    const wrapper = mount(probe)
    await flushPromises()

    expect(mocks.fetch).toHaveBeenCalledTimes(2)
    expect(mocks.reload).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('USAGE_NO_TOKEN')

    await api!.refetch()
    await flushPromises()

    expect(mocks.fetch).toHaveBeenCalledTimes(4)
    expect(mocks.reload).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('no reintenta ante errores que no son de token', async () => {
    mocks.fetch.mockRejectedValue(new Error('USAGE_TRANSPORT: sin conexión'))
    mount(probe)
    await flushPromises()

    expect(mocks.fetch).toHaveBeenCalledTimes(1)
    expect(mocks.reload).not.toHaveBeenCalled()
  })

  it('conserva los datos previos si falla un refresco', async () => {
    mocks.fetch.mockResolvedValueOnce(usage).mockRejectedValue(tokenError())
    const wrapper = mount(probe)
    await flushPromises()

    await api!.refetch()
    await flushPromises()

    expect(api!.data.value?.monthly.percent).toBe(30)
    expect(api!.error.value).not.toBeNull()
    wrapper.unmount()
  })
})
