<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import { ServiceStatus, UsageFetch, ConfigListMCP } from '../../wailsjs/go/main/App'

// Dashboard (landing): solo resúmenes, sin acciones (las acciones están en cada menú).
// Widgets: Consumo · Servicio + Config.
const { t } = useI18n()

const running = ref(false)
const url = ref('')
const rolling = ref(0)
const weekly = ref(0)
const monthly = ref(0)
const mcp = ref<{ name: string; status: string }[]>([])
const loading = ref(true)
const error = ref('')

// Severidad graduada por proximidad al límite (ver SPEC-ui-ux).
function sev(pct: number): string {
  if (pct >= 98) return 'crit'
  if (pct >= 85) return 'high'
  if (pct >= 70) return 'warn'
  return 'ok'
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
  } catch {
    // La config es opcional en el resumen: si falla, se conserva servicio/consumo.
  } finally {
    loading.value = false
  }
}
onMounted(refresh)
</script>

<template>
  <section class="section resumen">
    <h2>{{ t('nav.resumen') }}</h2>

    <div v-if="loading" class="muted">{{ t('usage.loading') }}</div>

    <template v-else>
      <!-- Widget: Consumo -->
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
          <span class="bar-pct" :class="sev(w.pct)">{{ Math.round(w.pct) }}%</span>
        </div>
      </div>

      <!-- Widget: Servicio + Config (solo lectura) -->
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

        <h3>{{ t('config.mcp') }}</h3>
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
      </div>

      <p v-if="error" class="error">{{ error }}</p>
    </template>
  </section>
</template>
