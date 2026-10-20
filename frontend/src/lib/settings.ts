import { reactive } from 'vue'
import {
  GetSettings,
  SetAutoStart,
  SetLanguage,
  SetRefreshMin,
} from '../../wailsjs/go/main/App'

// Estado de ajustes compartido por App, Resumen y Ajustes. Evita que cada
// sección vuelva a pedir GetSettings al montarse (App.vue desmonta con v-if).

export interface AppSettings {
  language: string
  autoStart: boolean
  refreshMin: number
}

export const settings = reactive<AppSettings>({
  language: 'es',
  autoStart: false,
  refreshMin: 15,
})

let loadPromise: Promise<void> | null = null
let loaded = false

export function loadSettings(force = false): Promise<void> {
  if (loadPromise) return loadPromise
  if (loaded && !force) return Promise.resolve()

  loadPromise = GetSettings()
    .then((remote) => {
      settings.language = remote.language || 'es'
      settings.autoStart = remote.autoStart
      settings.refreshMin = remote.refreshMin > 0 ? remote.refreshMin : 15
      loaded = true
    })
    .catch(() => {
      // Sin backend (tests/dev): se mantienen los valores por defecto.
    })
    .finally(() => {
      loadPromise = null
    })

  return loadPromise
}

export async function setLanguage(lang: string): Promise<void> {
  settings.language = lang
  try {
    await SetLanguage(lang)
  } catch {
    // Sin backend (tests): se conserva el cambio local.
  }
}

export async function setAutoStart(on: boolean): Promise<void> {
  settings.autoStart = on
  try {
    await SetAutoStart(on)
  } catch {
    // Sin backend (tests): se conserva el cambio local.
  }
}

export async function setRefreshMin(min: number): Promise<void> {
  const value = [1, 5, 15, 30, 60].includes(min) ? min : 15
  settings.refreshMin = value
  try {
    await SetRefreshMin(value)
  } catch {
    // Sin backend (tests): se conserva el cambio local.
  }
}

export function refreshIntervalMs(): number {
  return settings.refreshMin * 60_000
}
