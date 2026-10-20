<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ServiceStatus,
  ConfigListMCP,
  WebAccessState,
} from '../../wailsjs/go/main/App'
import { useQuery } from '../lib/query'
import ResumenConsumo from './ResumenConsumo.vue'
import ResumenServicio from './ResumenServicio.vue'
import ResumenAcceso from './ResumenAcceso.vue'
import ResumenMcp from './ResumenMcp.vue'

// Dashboard (landing) en formato de widgets. Cada bloque tiene su propio
// skeleton/error y comparte caché para no repetir llamadas al volver.
const { t } = useI18n()
const emit = defineEmits<{ (e: 'navigate', section: string): void }>()

const service = useQuery({
  key: 'resumen-service',
  fetcher: ServiceStatus,
  staleTime: 60_000,
})
const mcp = useQuery({
  key: 'resumen-mcp',
  fetcher: ConfigListMCP,
  staleTime: 60_000,
})
const access = useQuery({
  key: 'resumen-access',
  fetcher: WebAccessState,
  staleTime: 60_000,
})

const serviceError = computed(() =>
  service.error.value ? String(service.error.value) : '',
)
const mcpError = computed(() => (mcp.error.value ? String(mcp.error.value) : ''))
const accessError = computed(() =>
  access.error.value ? String(access.error.value) : '',
)

const accessMode = computed(() => access.data.value?.mode ?? 'local')
const accessRunning = computed(() => access.data.value?.running ?? false)
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
        :running="accessRunning"
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
