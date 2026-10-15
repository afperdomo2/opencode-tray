<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Modal from './Modal.vue'
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
const { t } = useI18n()

const running = ref(false)
const url = ref('')
const busy = ref(false)
const mcp = ref<{ name: string; status: string }[]>([])
const configText = ref('')
const showJson = ref(false)
const error = ref('')

async function refresh() {
  error.value = ''
  try {
    const s = await ServiceStatus()
    running.value = s.running
    url.value = s.url
  } catch (e) {
    error.value = String(e)
  }
  try {
    mcp.value = await ConfigListMCP()
    configText.value = await ConfigJSON()
  } catch {
    // La config es opcional aquí: si falla, se conserva el control del servicio.
  }
}
onMounted(refresh)

async function run(fn: () => Promise<void>) {
  busy.value = true
  error.value = ''
  try {
    await fn()
    const s = await ServiceStatus()
    running.value = s.running
    url.value = s.url
  } catch (e) {
    error.value = String(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="section">
    <h2>{{ t('nav.servicio') }}</h2>

    <div class="status-row">
      <span class="dot" :class="running ? 'on' : 'off'" aria-hidden="true"></span>
      <span>{{ running ? t('service.running') : t('service.stopped') }}</span>
    </div>
    <p v-if="url" class="url">{{ url }}</p>

    <div class="actions">
      <button :disabled="busy || running" @click="run(() => ServiceStart())">
        <Icon name="play" :size="14" />{{ t('service.start') }}
      </button>
      <button :disabled="busy || !running" @click="run(() => ServiceStop())">
        <Icon name="stop" :size="14" />{{ t('service.stop') }}
      </button>
      <button :disabled="busy" @click="run(() => ServiceForceStop())">
        <Icon name="zap" :size="14" />{{ t('service.forceStop') }}
      </button>
      <button :disabled="busy" @click="refresh">
        <Icon name="refresh" :size="14" />{{ t('service.refresh') }}
      </button>
    </div>

    <h3>{{ t('config.mcp') }}</h3>
    <p v-if="mcp.length === 0" class="muted">{{ t('config.empty') }}</p>
    <ul class="mcp-list">
      <li v-for="s in mcp" :key="s.name">
        <span class="dot" :class="s.status === 'connected' ? 'on' : 'off'" aria-hidden="true"></span>
        {{ s.name }}
        <span class="muted">{{ s.status }}</span>
      </li>
    </ul>

    <div class="actions">
      <button @click="ConfigOpenInEditor">
        <Icon name="externalLink" :size="14" />{{ t('config.openEditor') }}
      </button>
      <button @click="ConfigOpenDir">
        <Icon name="folder" :size="14" />{{ t('config.openDir') }}
      </button>
      <button @click="showJson = true">
        <Icon name="code" :size="14" />{{ t('config.json') }}
      </button>
    </div>

    <Modal :open="showJson" :title="t('config.json')" @close="showJson = false">
      <pre class="config-view">{{ configText }}</pre>
    </Modal>

    <p v-if="error" class="error">{{ error }}</p>
  </section>
</template>
