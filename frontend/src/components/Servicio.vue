<script setup lang="ts">
import { ref, onMounted } from 'vue'
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

// Sección unificada: control del servicio + configuración (MCP / JSON).
// Servicio y configuración cargan en paralelo, cada uno con su propio estado.
const { t } = useI18n()

const running = ref(false)
const url = ref('')
const busy = ref(false)
const mcp = ref<{ name: string; status: string }[]>([])
const configText = ref('')
const showJson = ref(false)

const loadingStatus = ref(true)
const loadingMcp = ref(true)
const loadingJson = ref(true)
const errorStatus = ref('')
const errorMcp = ref('')
const errorJson = ref('')

async function loadStatus() {
  loadingStatus.value = true
  errorStatus.value = ''
  try {
    const s = await ServiceStatus()
    running.value = s.running
    url.value = s.url
  } catch (e) {
    errorStatus.value = String(e)
  } finally {
    loadingStatus.value = false
  }
}

async function loadMcp() {
  loadingMcp.value = true
  errorMcp.value = ''
  try {
    mcp.value = await ConfigListMCP()
  } catch (e) {
    errorMcp.value = String(e)
  } finally {
    loadingMcp.value = false
  }
}

async function loadJson() {
  loadingJson.value = true
  errorJson.value = ''
  try {
    configText.value = await ConfigJSON()
  } catch (e) {
    errorJson.value = String(e)
  } finally {
    loadingJson.value = false
  }
}

async function refresh() {
  await Promise.all([loadStatus(), loadMcp(), loadJson()])
}
onMounted(refresh)

async function run(fn: () => Promise<void>) {
  busy.value = true
  errorStatus.value = ''
  try {
    await fn()
    const s = await ServiceStatus()
    running.value = s.running
    url.value = s.url
  } catch (e) {
    errorStatus.value = String(e)
  } finally {
    busy.value = false
  }
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
        <p v-else-if="errorStatus" class="error">{{ errorStatus }}</p>
        <template v-else>
          <div class="status-row">
            <span class="dot" :class="running ? 'on' : 'off'" aria-hidden="true"></span>
            <span>{{ running ? t('service.running') : t('service.stopped') }}</span>
          </div>
          <p v-if="url" class="url">{{ url }}</p>
        </template>

        <div class="actions">
          <button
            class="btn"
            :disabled="busy || loadingStatus || running"
            @click="run(() => ServiceStart())"
          >
            <Icon name="play" :size="14" />{{ t('service.start') }}
          </button>
          <button
            class="btn"
            :disabled="busy || loadingStatus || !running"
            @click="run(() => ServiceStop())"
          >
            <Icon name="stop" :size="14" />{{ t('service.stop') }}
          </button>
          <button class="btn" :disabled="busy || loadingStatus" @click="run(() => ServiceForceStop())">
            <Icon name="zap" :size="14" />{{ t('service.forceStop') }}
          </button>
          <button class="btn" :disabled="busy || loadingStatus" @click="refresh">
            <Icon name="refresh" :size="14" />{{ t('service.refresh') }}
          </button>
        </div>
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
        <p v-else-if="errorMcp" class="error">{{ errorMcp }}</p>
        <template v-else>
          <p v-if="mcp.length === 0" class="muted">{{ t('config.empty') }}</p>
          <ul class="mcp-list">
            <li v-for="s in mcp" :key="s.name">
              <span
                class="dot"
                :class="s.status === 'connected' ? 'on' : 'off'"
                aria-hidden="true"
              ></span>
              <span class="mcp-name">{{ s.name }}</span>
              <span class="muted mcp-status">{{ s.status }}</span>
            </li>
          </ul>
        </template>

        <div class="actions">
          <button class="btn" @click="ConfigOpenInEditor">
            <Icon name="externalLink" :size="14" />{{ t('config.openEditor') }}
          </button>
          <button class="btn" @click="ConfigOpenDir">
            <Icon name="folder" :size="14" />{{ t('config.openDir') }}
          </button>
          <button class="btn" :disabled="loadingJson" @click="showJson = true">
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
