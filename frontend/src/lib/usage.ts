import { ref } from 'vue'
import {
  UsageClearManualToken,
  UsageFetch,
  UsageReloadToken,
  UsageSetManualToken,
  UsageTokenStatus,
} from '../../wailsjs/go/main/App'
import type { usage as UsageModels } from '../../wailsjs/go/models'
import { invalidateQuery, prefetchQuery, useQuery } from './query'
import { refreshIntervalMs } from './settings'

export const USAGE_KEY = 'usage'
export const USAGE_STALE_MS = 60_000

const tokenStatus = ref<UsageModels.TokenStatus | null>(null)
const tokenBusy = ref(false)
const tokenError = ref('')

let autoReloaded = false

export function isTokenError(e: unknown): boolean {
  const msg = String(e)
  return msg.includes('USAGE_NO_TOKEN') || msg.includes('USAGE_INVALID_TOKEN')
}

export function usageErrorKey(e: unknown): string {
  const msg = String(e)
  if (msg.includes('USAGE_NO_TOKEN')) return 'usage.errorNoToken'
  if (msg.includes('USAGE_INVALID_TOKEN')) return 'usage.errorInvalidToken'
  if (msg.includes('USAGE_NO_ENTITLEMENT')) return 'usage.errorNoEntitlement'
  if (msg.includes('USAGE_FORBIDDEN')) return 'usage.errorForbidden'
  if (msg.includes('USAGE_TRANSPORT')) return 'usage.errorTransport'
  if (msg.includes('USAGE_PARSE')) return 'usage.errorParse'
  if (msg.includes('USAGE_HTTP_')) return 'usage.errorHttp'
  return 'usage.error'
}

async function fetchUsage(): Promise<UsageModels.Usage> {
  try {
    return await UsageFetch()
  } catch (e) {
    if (autoReloaded || !isTokenError(e)) throw e

    // Un solo reintento automático: relee el token vigente desde OpenCode.
    autoReloaded = true
    try {
      tokenStatus.value = await UsageReloadToken()
      return await UsageFetch()
    } catch {
      throw e
    }
  } finally {
    // Tras esta carga el flag vuelve a permitir un reintento en la siguiente.
    autoReloaded = false
  }
}

export function useUsageQuery() {
  const query = useQuery<UsageModels.Usage>({
    key: USAGE_KEY,
    fetcher: fetchUsage,
    staleTime: USAGE_STALE_MS,
    refetchInterval: refreshIntervalMs,
  })

  return {
    ...query,
    tokenStatus,
    tokenBusy,
    tokenError,
    loadTokenStatus,
    reloadToken,
    saveManualToken,
    clearManualToken,
  }
}

export async function loadTokenStatus(): Promise<void> {
  try {
    tokenStatus.value = await UsageTokenStatus()
  } catch {
    tokenStatus.value = null
  }
}

async function afterTokenChange(): Promise<void> {
  invalidateQuery(USAGE_KEY)
  await prefetchQuery(USAGE_KEY, fetchUsage)
}

export async function reloadToken(): Promise<void> {
  tokenBusy.value = true
  tokenError.value = ''
  try {
    tokenStatus.value = await UsageReloadToken()
    // El reintento explícito ya ejecutó el export: evita repetirlo en el fetch.
    autoReloaded = true
    await afterTokenChange()
  } catch (e) {
    tokenError.value = String(e)
  } finally {
    tokenBusy.value = false
  }
}

export async function saveManualToken(token: string): Promise<boolean> {
  tokenBusy.value = true
  tokenError.value = ''
  try {
    await UsageSetManualToken(token.trim())
    tokenStatus.value = await UsageTokenStatus()
    autoReloaded = true
    await afterTokenChange()
    return true
  } catch (e) {
    tokenError.value = String(e)
    return false
  } finally {
    tokenBusy.value = false
  }
}

export async function clearManualToken(): Promise<void> {
  tokenBusy.value = true
  tokenError.value = ''
  try {
    await UsageClearManualToken()
    tokenStatus.value = await UsageTokenStatus()
    autoReloaded = false
    await afterTokenChange()
  } catch (e) {
    tokenError.value = String(e)
  } finally {
    tokenBusy.value = false
  }
}
