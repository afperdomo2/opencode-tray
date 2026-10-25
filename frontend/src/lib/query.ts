import {
  computed,
  onMounted,
  onScopeDispose,
  reactive,
  watch,
  type ComputedRef,
} from 'vue'

// Caché nativo de datos de servidor (estilo TanStack Query, sin dependencias).
// El estado vive a nivel de módulo: sobrevive al desmontaje de los componentes
// (App.vue usa v-if), evitando skeletons y peticiones repetidas al navegar.
//
// Dos caminos de revalidación:
//   · refetch()      → suave: conserva los datos visibles (polling, foco, remontaje).
//   · refetchHard()  → duro: descarta los datos y muestra skeleton (botón o mutación).

interface QueryState<T> {
  data: T | undefined
  error: Error | null
  updatedAt: number
  isFetching: boolean
}

export interface QueryOptions<T> {
  key: string
  fetcher: () => Promise<T>
  staleTime?: number
  refetchInterval?: () => number | false
}

export interface QueryResult<T> {
  data: ComputedRef<T | undefined>
  error: ComputedRef<Error | null>
  isFetching: ComputedRef<boolean>
  isInitialLoading: ComputedRef<boolean>
  updatedAt: ComputedRef<number>
  refetch: () => Promise<void>
  refetchHard: () => Promise<void>
}

const DEFAULT_STALE_MS = 60_000

const states = new Map<string, QueryState<unknown>>()
const inflight = new Map<string, Promise<unknown>>()
// Generación por key: al descartar datos invalidamos la petición en vuelo para
// que su resultado (ya obsoleto) no sobrescriba el estado nuevo.
const generations = new Map<string, number>()

function stateFor<T>(key: string): QueryState<T> {
  let state = states.get(key)
  if (!state) {
    state = reactive({ data: undefined, error: null, updatedAt: 0, isFetching: false })
    states.set(key, state)
  }
  return state as QueryState<T>
}

function generationOf(key: string): number {
  return generations.get(key) ?? 0
}

function toError(e: unknown): Error {
  if (e instanceof Error) return e
  return new Error(String(e))
}

// execute deduplica: si ya hay una petición en vuelo para la key, devuelve la misma.
function execute<T>(key: string, fetcher: () => Promise<T>): Promise<T> {
  const current = inflight.get(key)
  if (current) return current as Promise<T>

  const state = stateFor<T>(key)
  const gen = generationOf(key)
  state.isFetching = true

  const promise = Promise.resolve()
    .then(() => fetcher())
    .then((data) => {
      if (gen === generationOf(key)) {
        state.data = data
        state.error = null
        state.updatedAt = Date.now()
      }
      return data
    })
    .catch((e: unknown) => {
      if (gen === generationOf(key)) state.error = toError(e)
      throw e
    })
    .finally(() => {
      if (gen === generationOf(key)) state.isFetching = false
      if (inflight.get(key) === promise) inflight.delete(key)
    })

  inflight.set(key, promise)
  return promise
}

async function refetchKey<T>(
  key: string,
  fetcher: () => Promise<T>,
  hard = false,
): Promise<void> {
  if (hard) clearQueryData(key)
  try {
    await execute(key, fetcher)
  } catch {
    // El error queda en el estado compartido; el componente lo muestra.
  }
}

export function useQuery<T>(options: QueryOptions<T>): QueryResult<T> {
  const { key, fetcher } = options
  const staleTime = options.staleTime ?? DEFAULT_STALE_MS
  const state = stateFor<T>(key)

  const intervalMs = computed(() => {
    const value = options.refetchInterval?.()
    return typeof value === 'number' && value > 0 ? value : 0
  })

  const isStale = () => Date.now() - state.updatedAt >= staleTime

  let timer: ReturnType<typeof setInterval> | undefined
  let mounted = false

  function stopTimer() {
    if (timer !== undefined) {
      clearInterval(timer)
      timer = undefined
    }
  }

  function startTimer() {
    stopTimer()
    if (!mounted || intervalMs.value <= 0) return
    timer = setInterval(() => {
      if (typeof document !== 'undefined' && document.visibilityState === 'hidden') return
      void refetchKey(key, fetcher)
    }, intervalMs.value)
  }

  function onVisible() {
    if (typeof document !== 'undefined' && document.visibilityState === 'hidden') return
    if (isStale()) void refetchKey(key, fetcher)
  }

  onMounted(() => {
    mounted = true
    if (isStale()) void refetchKey(key, fetcher)
    startTimer()
    document.addEventListener('visibilitychange', onVisible)
    window.addEventListener('focus', onVisible)
  })

  watch(intervalMs, startTimer)

  onScopeDispose(() => {
    mounted = false
    stopTimer()
    document.removeEventListener('visibilitychange', onVisible)
    window.removeEventListener('focus', onVisible)
  })

  return {
    data: computed(() => state.data),
    error: computed(() => state.error),
    isFetching: computed(() => state.isFetching),
    isInitialLoading: computed(() => state.data === undefined && state.isFetching),
    updatedAt: computed(() => state.updatedAt),
    refetch: () => refetchKey(key, fetcher),
    refetchHard: () => refetchKey(key, fetcher, true),
  }
}

// invalidateQuery marca los datos como stale; no borra el contenido visible.
export function invalidateQuery(key: string) {
  const state = states.get(key)
  if (state) state.updatedAt = 0
}

// invalidateQueries invalida varias keys a la vez.
export function invalidateQueries(keys: string[]) {
  keys.forEach(invalidateQuery)
}

// clearQueryData descarta los datos de una key: el próximo render con la query
// mostrará skeleton (primera carga o refresco duro). También invalida cualquier
// petición en vuelo para que no restaure datos obsoletos.
export function clearQueryData(key: string) {
  generations.set(key, generationOf(key) + 1)
  inflight.delete(key)

  const state = states.get(key)
  if (!state) return
  state.data = undefined
  state.error = null
  state.updatedAt = 0
  // Se considera "pendiente" hasta que arranque (o falle) la recarga: así la
  // vista que monte durante una mutación pinta skeleton, no un estado vacío.
  state.isFetching = true
}

// clearQueries descarta varias keys a la vez (mutaciones).
export function clearQueries(keys: string[]) {
  keys.forEach(clearQueryData)
}

// prefetch calienta la caché sin crear temporizadores ni suscripciones.
export function prefetchQuery<T>(key: string, fetcher: () => Promise<T>): Promise<void> {
  return refetchKey(key, fetcher)
}

// clearQueryCache se usa en tests y al cambiar de fuente de token.
export function clearQueryCache() {
  states.clear()
  inflight.clear()
  generations.clear()
}
