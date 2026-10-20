<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Skeleton from './Skeleton.vue'
import { ServiceStatus, ConfigListMCP } from '../../wailsjs/go/main/App'
import { useUsageQuery, usageErrorKey, isTokenError } from '../lib/usage'

// Dashboard (landing): resúmenes. El consumo usa caché compartida; si falta el
// token, muestra un único CTA para releerlo desde OpenCode (excepción aprobada).
const { t } = useI18n()
const emit = defineEmits<{ (e: 'navigate', section: string): void }>()

const running = ref(false)
const url = ref('')
const mcp = ref<{ name: string; status: string }[]>([])

const usage = useUsageQuery()
const tokenBusy = usage.tokenBusy

const loadingService = ref(true)
const loadingMcp = ref(true)
const errorService = ref('')
const errorMcp = ref('')

const usageError = computed(() => usage.error.value)
const usageErrorLabel = computed(() =>
  usageError.value ? usageErrorKey(usageError.value) : '',
)
const showTokenRetry = computed(() => isTokenError(usageError.value))
const usingManualToken = computed(() => usage.tokenStatus.value?.source === 'manual')
const hasUsageData = computed(() => usage.data.value !== undefined)
const isUpdating = computed(() => usage.isFetching.value && !usageError.value)

// Severidad graduada por proximidad al límite (ver SPEC-ui-ux).
function sev(pct: number): string {
  if (pct >= 98) return 'crit'
  if (pct >= 85) return 'high'
  if (pct >= 70) return 'warn'
  return 'ok'
}

const windows = computed(() => {
  const u = usage.data.value
  return [
    { key: 'rolling', label: t('usage.rolling'), pct: u?.rolling.percent ?? 0 },
    { key: 'weekly', label: t('usage.weekly'), pct: u?.weekly.percent ?? 0 },
    { key: 'monthly', label: t('usage.monthly'), pct: u?.monthly.percent ?? 0 },
  ]
})

function retryToken() {
  void usage.reloadToken()
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

        <div v-if="usage.isInitialLoading.value" class="skeleton-stack" role="status">
          <span class="sr-only">{{ t('common.loading') }}</span>
          <div v-for="i in 3" :key="i" class="skeleton-row">
            <Skeleton width="64px" height="12px" />
            <Skeleton height="10px" radius="5px" />
            <Skeleton width="36px" height="12px" />
          </div>
        </div>
        <template v-else>
          <template v-if="hasUsageData">
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
            <p v-if="isUpdating" class="usage-status" role="status">{{ t('usage.updating') }}</p>
          </template>

          <div v-if="usageError" class="error-block">
            <p class="error" role="alert">
              {{ hasUsageData ? t('usage.staleError') : t(usageErrorLabel) }}
            </p>
            <div v-if="showTokenRetry" class="error-ctas">
              <button
                type="button"
                class="btn token-retry"
                :disabled="tokenBusy"
                :aria-busy="tokenBusy"
                @click="retryToken"
              >
                <Icon name="refresh" :size="14" />
                {{ tokenBusy ? t('usage.reloading') : t('usage.reloadFromOpenCode') }}
              </button>
              <button type="button" class="btn" @click="emit('navigate', 'ajustes')">
                <Icon name="key" :size="14" />
                {{ t('usage.configureKey') }}
              </button>
            </div>
            <p v-if="usingManualToken" class="muted">{{ t('usage.manualHint') }}</p>
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
            <span v-if="url" class="url" :title="url">{{ url }}</span>
          </div>
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
