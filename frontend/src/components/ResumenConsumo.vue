<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Skeleton from './Skeleton.vue'
import ResumenWidget from './ResumenWidget.vue'
import { useUsageQuery, usageErrorKey, isTokenError } from '../lib/usage'

// Widget de consumo: conserva las tres barras actuales. El único caso con
// acciones es el token ausente/inválido (reintentar o configurar la API key).
const emit = defineEmits<{ (e: 'navigate', section: string): void }>()
const { t } = useI18n()

const usage = useUsageQuery()
const tokenBusy = usage.tokenBusy

const usageError = computed(() => usage.error.value)
const usageErrorLabel = computed(() =>
  usageError.value ? usageErrorKey(usageError.value) : '',
)
const showTokenRetry = computed(() => isTokenError(usageError.value))
const usingManualToken = computed(() => usage.tokenStatus.value?.source === 'manual')
const hasUsageData = computed(() => usage.data.value !== undefined)

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
</script>

<template>
  <ResumenWidget :title="t('nav.consumo')" icon="chart" class="widget-consumo">
    <template #meta>
      <span class="widget-meta">{{ t('usage.tokenTitle') }}</span>
    </template>

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
  </ResumenWidget>
</template>
