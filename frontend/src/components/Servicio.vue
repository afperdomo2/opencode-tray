<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Modal from './Modal.vue'
import Skeleton from './Skeleton.vue'
import {
  ServiceStatus,
  ServiceStart,
  ServiceStop,
  ServiceForceStop,
  ConfigListMCP,
  ConfigJSON,
  ConfigOpenInEditor,
  ConfigOpenDir,
} from '../../wailsjs/go/main/App'
import { clearQueries, useQuery } from '../lib/query'
import { MCP_KEY, SERVICE_KEY, WEB_ACCESS_KEY } from '../lib/query-keys'

// Sección unificada: control del servicio + configuración (MCP / JSON).
// El estado y los MCP usan el caché compartido con Resumen; el JSON se carga
// de forma diferida al abrir el modal.
const { t } = useI18n()

const status = useQuery({
  key: SERVICE_KEY,
  fetcher: ServiceStatus,
  staleTime: 10_000,
})
const mcp = useQuery({
  key: MCP_KEY,
  fetcher: ConfigListMCP,
  staleTime: 60_000,
})

const busy = ref(false)
const mutating = ref(false)
const manualRefreshing = ref(false)
const actionError = ref('')

const showJson = ref(false)
const configText = ref('')
const loadingJson = ref(false)
const errorJson = ref('')
let jsonLoaded = false

// El estado y el servicio mutado se descartan al iniciar una acción: durante
// toda la operación (puede tardar segundos) se muestra skeleton, nunca el
// estado anterior ya inválido.
const loadingStatus = computed(
  () => status.isInitialLoading.value || mutating.value,
)
const loadingMcp = computed(() => mcp.isInitialLoading.value || mutating.value)
const refreshing = computed(
  () => manualRefreshing.value || mutating.value,
)

const running = computed(() => status.data.value?.running ?? false)
const url = computed(() => status.data.value?.url ?? '')
const hasStatusData = computed(() => status.data.value !== undefined)
const errorStatus = computed(() =>
  status.error.value ? String(status.error.value) : '',
)

const servers = computed(() => mcp.data.value ?? [])
const hasMcpData = computed(() => mcp.data.value !== undefined)
const errorMcp = computed(() => (mcp.error.value ? String(mcp.error.value) : ''))

async function loadJson(force = false) {
  if (loadingJson.value) return
  if (jsonLoaded && !force) return
  loadingJson.value = true
  errorJson.value = ''
  try {
    configText.value = await ConfigJSON()
    jsonLoaded = true
  } catch (e) {
    errorJson.value = String(e)
  } finally {
    loadingJson.value = false
  }
}

function openJson() {
  showJson.value = true
  void loadJson()
}

async function refresh() {
  actionError.value = ''
  manualRefreshing.value = true
  try {
    await Promise.all([
      status.refetchHard(),
      mcp.refetchHard(),
      jsonLoaded ? loadJson(true) : Promise.resolve(),
    ])
  } finally {
    manualRefreshing.value = false
  }
}

async function run(fn: () => Promise<void>) {
  busy.value = true
  actionError.value = ''
  mutating.value = true
  clearQueries([SERVICE_KEY, WEB_ACCESS_KEY, MCP_KEY])
  try {
    await fn()
  } catch (e) {
    actionError.value = String(e)
  } finally {
    busy.value = false
  }
  // Leemos el estado real: los datos estaban descartados, así que esto pinta
  // skeleton hasta que llega la respuesta.
  await Promise.all([status.refetch(), mcp.refetch()])
  mutating.value = false
}
</script>

<template>
  <section class="section">
    <h2>{{ t('nav.servicio') }}</h2>

    <div class="panel-grid">
      <div class="card">
        <div class="card-head">
          <Icon name="server" :size="16" />
          <h3>{{ t('service.control') }}</h3>
        </div>

        <div v-if="loadingStatus" class="skeleton-stack" role="status">
          <span class="sr-only">{{ t('common.loading') }}</span>
          <Skeleton width="120px" height="14px" />
          <Skeleton width="75%" height="12px" />
        </div>
        <p v-else-if="errorStatus && !hasStatusData" class="error">{{ errorStatus }}</p>
        <template v-else>
          <div class="status-row">
            <span class="dot" :class="running ? 'on' : 'off'" aria-hidden="true"></span>
            <span>{{ running ? t('service.running') : t('service.stopped') }}</span>
            <span v-if="url" class="url" :title="url">{{ url }}</span>
          </div>
          <p v-if="errorStatus" class="error">{{ errorStatus }}</p>
        </template>

        <div class="actions actions-grid cols-2">
          <button
            class="btn"
            :title="t('service.startHint')"
            :disabled="busy || loadingStatus || running"
            @click="run(() => ServiceStart())"
          >
            <Icon name="play" :size="14" />{{ t('service.start') }}
          </button>
          <button
            class="btn"
            :title="t('service.stopHint')"
            :disabled="busy || loadingStatus || !running"
            @click="run(() => ServiceStop())"
          >
            <Icon name="stop" :size="14" />{{ t('service.stop') }}
          </button>
          <button
            class="btn"
            :title="t('service.forceStopHint')"
            :disabled="busy || loadingStatus"
            @click="run(() => ServiceForceStop())"
          >
            <Icon name="zap" :size="14" />{{ t('service.forceStop') }}
          </button>
          <button
            class="btn"
            :title="t('service.refreshHint')"
            :disabled="busy || refreshing"
            :aria-busy="refreshing"
            @click="refresh"
          >
            <Icon name="refresh" :size="14" :class="{ spinning: refreshing }" />{{
              t('service.refresh')
            }}
          </button>
        </div>
        <p v-if="actionError" class="error" role="alert">{{ actionError }}</p>
      </div>

      <div class="card">
        <div class="card-head">
          <Icon name="code" :size="16" />
          <h3>{{ t('config.title') }}</h3>
        </div>

        <h3>{{ t('config.mcp') }}</h3>

        <div v-if="loadingMcp" class="skeleton-stack" role="status">
          <span class="sr-only">{{ t('common.loading') }}</span>
          <div v-for="i in 2" :key="i" class="status-row">
            <Skeleton width="10px" height="10px" radius="50%" />
            <Skeleton width="120px" height="14px" />
            <Skeleton width="56px" height="12px" />
          </div>
        </div>
        <p v-else-if="errorMcp && !hasMcpData" class="error">{{ errorMcp }}</p>
        <template v-else>
          <p v-if="servers.length === 0" class="muted">{{ t('config.empty') }}</p>
          <ul v-else class="mcp-list">
            <li v-for="s in servers" :key="s.name">
              <span
                class="dot"
                :class="s.status === 'connected' ? 'on' : 'off'"
                aria-hidden="true"
              ></span>
              <span class="mcp-name">{{ s.name }}</span>
              <span class="muted mcp-status">{{ s.status }}</span>
            </li>
          </ul>
          <p v-if="errorMcp" class="error">{{ errorMcp }}</p>
        </template>

        <div class="actions actions-grid cols-3">
          <button class="btn" :title="t('config.openEditorHint')" @click="ConfigOpenInEditor">
            <Icon name="externalLink" :size="14" />{{ t('config.openEditor') }}
          </button>
          <button class="btn" :title="t('config.openDirHint')" @click="ConfigOpenDir">
            <Icon name="folder" :size="14" />{{ t('config.openDir') }}
          </button>
          <button class="btn" :title="t('config.jsonHint')" @click="openJson">
            <Icon name="code" :size="14" />{{ t('config.json') }}
          </button>
        </div>
      </div>
    </div>

    <Modal :open="showJson" :title="t('config.json')" @close="showJson = false">
      <div v-if="loadingJson" class="skeleton-stack" role="status">
        <span class="sr-only">{{ t('common.loading') }}</span>
        <Skeleton v-for="i in 6" :key="i" :width="i === 6 ? '60%' : '100%'" height="12px" />
      </div>
      <p v-else-if="errorJson" class="error">{{ errorJson }}</p>
      <pre v-else class="config-view">{{ configText }}</pre>
    </Modal>
  </section>
</template>
