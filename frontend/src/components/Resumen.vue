<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ServiceStatus,
  ConfigListMCP,
  WebAccessState,
} from '../../wailsjs/go/main/App'
import { useQuery } from '../lib/query'
import { MCP_KEY, SERVICE_KEY, WEB_ACCESS_KEY } from '../lib/query-keys'
import ResumenConsumo from './ResumenConsumo.vue'
import ResumenServicio from './ResumenServicio.vue'
import ResumenAcceso from './ResumenAcceso.vue'
import ResumenMcp from './ResumenMcp.vue'

// Dashboard (landing) en formato de widgets. Comparte claves de caché con las
// pestañas detalladas: una mutación en Servicio/Acceso invalida estos datos y
// el widget muestra skeleton hasta tener el estado nuevo.
const { t } = useI18n()
const emit = defineEmits<{ (e: 'navigate', section: string): void }>()

const service = useQuery({
  key: SERVICE_KEY,
  fetcher: ServiceStatus,
  staleTime: 10_000,
  // El servicio puede cambiar por fuera de la app: refresco silencioso
  // periódico solo mientras el panel está visible (query.ts lo pausa oculto).
  refetchInterval: () => 15_000,
})
const mcp = useQuery({
  key: MCP_KEY,
  fetcher: ConfigListMCP,
  staleTime: 60_000,
})
const access = useQuery({
  key: WEB_ACCESS_KEY,
  fetcher: WebAccessState,
  staleTime: 15_000,
})

// Los widgets conservan su contenido durante revalidaciones silenciosas; el
// error solo sustituye al contenido si aún no hay datos previos.
const serviceError = computed(() =>
  service.data.value === undefined && service.error.value
    ? String(service.error.value)
    : '',
)
const mcpError = computed(() =>
  mcp.data.value === undefined && mcp.error.value
    ? String(mcp.error.value)
    : '',
)
const accessError = computed(() =>
  access.data.value === undefined && access.error.value
    ? String(access.error.value)
    : '',
)

const serviceRunning = computed(
  () => service.data.value?.running ?? access.data.value?.running ?? false,
)
const accessMode = computed(() => access.data.value?.mode ?? 'local')
const accessUrls = computed(() => access.data.value?.urls ?? [])
</script>

<template>
  <section class="section resumen">
    <h2>{{ t('nav.resumen') }}</h2>

    <div class="dashboard widget-grid">
      <ResumenConsumo class="widget-consumo" @navigate="emit('navigate', $event)" />

      <ResumenServicio
        class="widget-servicio"
        :running="service.data.value?.running ?? false"
        :url="service.data.value?.url ?? ''"
        :loading="service.isInitialLoading.value"
        :error="serviceError"
        @open="emit('navigate', 'servicio')"
      />

      <ResumenAcceso
        class="widget-acceso"
        :mode="accessMode"
        :running="serviceRunning"
        :urls="accessUrls"
        :loading="access.isInitialLoading.value"
        :error="accessError"
        @open="emit('navigate', 'accesoWeb')"
      />

      <ResumenMcp
        class="widget-mcp"
        :servers="mcp.data.value ?? []"
        :loading="mcp.isInitialLoading.value"
        :error="mcpError"
        @open="emit('navigate', 'servicio')"
      />
    </div>
  </section>
</template>
