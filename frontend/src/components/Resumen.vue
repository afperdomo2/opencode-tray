<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Skeleton from './Skeleton.vue'
import { ServiceStatus, UsageFetch, ConfigListMCP } from '../../wailsjs/go/main/App'

// Dashboard (landing): solo resúmenes, sin acciones (las acciones están en cada menú).
// Widgets: Consumo · Servicio + Config. Cada widget carga por separado para que
// no haya que esperar al más lento (consumo = REST; servicio = OpenCode).
const { t } = useI18n()

const running = ref(false)
const url = ref('')
const rolling = ref(0)
const weekly = ref(0)
const monthly = ref(0)
const mcp = ref<{ name: string; status: string }[]>([])

const loadingUsage = ref(true)
const loadingService = ref(true)
const loadingMcp = ref(true)
const errorUsage = ref('')
const errorService = ref('')
const errorMcp = ref('')

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

async function loadUsage() {
  loadingUsage.value = true
  errorUsage.value = ''
  try {
    const usage = await UsageFetch()
    rolling.value = usage.rolling.percent
    weekly.value = usage.weekly.percent
    monthly.value = usage.monthly.percent
  } catch (e) {
    errorUsage.value = String(e)
  } finally {
    loadingUsage.value = false
  }
}

async function loadService() {
  loadingService.value = true
  errorService.value = ''
  try {
    const svc = await ServiceStatus()
    running.value = svc.running
    url.value = svc.url
  } catch (e) {
    errorService.value = String(e)
  } finally {
    loadingService.value = false
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

onMounted(() => {
  void loadUsage()
  void loadService()
  void loadMcp()
})
</script>

<template>
  <section class="section resumen">
    <h2>{{ t('nav.resumen') }}</h2>

    <div class="dashboard">
      <!-- Widget: Consumo -->
      <div class="card">
        <div class="card-head">
          <Icon name="chart" :size="16" />
          <h3>{{ t('nav.consumo') }}</h3>
        </div>

        <div v-if="loadingUsage" class="skeleton-stack" role="status">
          <span class="sr-only">{{ t('common.loading') }}</span>
          <div v-for="i in 3" :key="i" class="skeleton-row">
            <Skeleton width="64px" height="12px" />
            <Skeleton height="10px" radius="5px" />
            <Skeleton width="36px" height="12px" />
          </div>
        </div>
        <p v-else-if="errorUsage" class="error">{{ errorUsage }}</p>
        <template v-else>
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
        </template>
      </div>

      <!-- Widget: Servicio + Config (solo lectura) -->
      <div class="card">
        <div class="card-head">
          <Icon name="server" :size="16" />
          <h3>{{ t('nav.servicio') }}</h3>
        </div>

        <div v-if="loadingService" class="skeleton-stack" role="status">
          <span class="sr-only">{{ t('common.loading') }}</span>
          <Skeleton width="120px" height="14px" />
          <Skeleton width="75%" height="12px" />
        </div>
        <p v-else-if="errorService" class="error">{{ errorService }}</p>
        <template v-else>
          <div class="status-row">
            <span class="dot" :class="running ? 'on' : 'off'" aria-hidden="true"></span>
            <span>{{ running ? t('service.running') : t('service.stopped') }}</span>
          </div>
          <p v-if="url" class="url">{{ url }}</p>
        </template>

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
      </div>
    </div>
  </section>
</template>
