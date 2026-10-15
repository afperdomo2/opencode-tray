<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Accordion from './Accordion.vue'
import {
  ServiceStatus,
  ServiceStart,
  ServiceStop,
  ServiceForceStop,
  UsageFetch,
  ConfigListMCP,
  ConfigJSON,
} from '../../wailsjs/go/main/App'

// Vista de overview (landing): compone datos ya existentes.
// Orden de bloques: Consumo → MCP/Config → Servicio.
const { t } = useI18n()

const running = ref(false)
const url = ref('')
const rolling = ref(0)
const weekly = ref(0)
const monthly = ref(0)
const mcp = ref<{ name: string; status: string }[]>([])
const configText = ref('')
const loading = ref(true)
const busy = ref(false)
const error = ref('')

// Severidad por umbral (misma regla que Consumo): >=98 crítico, >=90 aviso.
function sev(pct: number): string {
  return pct >= 98 ? 'crit' : pct >= 90 ? 'warn' : 'ok'
}

const windows = computed(() => [
  { key: 'rolling', label: t('usage.rolling'), pct: rolling.value },
  { key: 'weekly', label: t('usage.weekly'), pct: weekly.value },
  { key: 'monthly', label: t('usage.monthly'), pct: monthly.value },
])

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const [svc, usage] = await Promise.all([ServiceStatus(), UsageFetch()])
    running.value = svc.running
    url.value = svc.url
    rolling.value = usage.rolling.percent
    weekly.value = usage.weekly.percent
    monthly.value = usage.monthly.percent
  } catch (e) {
    error.value = String(e)
  }
  try {
    mcp.value = await ConfigListMCP()
    configText.value = await ConfigJSON()
  } catch {
    // La config es opcional en el Resumen: si falla, se conservan servicio/consumo.
  } finally {
    loading.value = false
  }
}
onMounted(refresh)

async function svcAction(fn: () => Promise<void>) {
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
  <section class="section resumen">
    <h2>{{ t('nav.resumen') }}</h2>

    <div v-if="loading" class="muted">{{ t('usage.loading') }}</div>

    <template v-else>
      <!-- 1 · Consumo -->
      <div class="card">
        <div class="card-head">
          <Icon name="chart" :size="16" />
          <h3>{{ t('nav.consumo') }}</h3>
        </div>
        <div v-for="w in windows" :key="w.key" class="bar-row">
          <span class="bar-label">{{ w.label }}</span>
          <div class="bar-track">
            <div
              class="bar-fill"
              :class="sev(w.pct)"
              :style="{ width: Math.min(100, w.pct) + '%' }"
            ></div>
          </div>
          <span class="bar-pct">{{ Math.round(w.pct) }}%</span>
        </div>
      </div>

      <!-- 2 · MCP / Config -->
      <div class="card">
        <div class="card-head">
          <Icon name="sliders" :size="16" />
          <h3>{{ t('nav.config') }}</h3>
        </div>
        <p v-if="mcp.length === 0" class="muted">{{ t('config.empty') }}</p>
        <ul class="mcp-list">
          <li v-for="s in mcp" :key="s.name">
            <span
              class="dot"
              :class="s.status === 'connected' ? 'on' : 'off'"
              aria-hidden="true"
            ></span>
            {{ s.name }}
            <span class="muted">{{ s.status }}</span>
          </li>
        </ul>
        <Accordion :title="t('config.json')">
          <pre class="config-view">{{ configText }}</pre>
        </Accordion>
      </div>

      <!-- 3 · Servicio -->
      <div class="card">
        <div class="card-head">
          <Icon name="server" :size="16" />
          <h3>{{ t('nav.servicio') }}</h3>
        </div>
        <div class="status-row">
          <span class="dot" :class="running ? 'on' : 'off'" aria-hidden="true"></span>
          <span>{{ running ? t('service.running') : t('service.stopped') }}</span>
        </div>
        <p v-if="url" class="url">{{ url }}</p>
        <div class="actions">
          <button :disabled="busy || running" @click="svcAction(() => ServiceStart())">
            <Icon name="play" :size="16" />{{ t('service.start') }}
          </button>
          <button :disabled="busy || !running" @click="svcAction(() => ServiceStop())">
            <Icon name="stop" :size="16" />{{ t('service.stop') }}
          </button>
          <button :disabled="busy" @click="svcAction(() => ServiceForceStop())">
            <Icon name="zap" :size="16" />{{ t('service.forceStop') }}
          </button>
          <button :disabled="busy" @click="refresh">
            <Icon name="refresh" :size="16" />{{ t('service.refresh') }}
          </button>
        </div>
      </div>

      <p v-if="error" class="error">{{ error }}</p>
    </template>
  </section>
</template>
