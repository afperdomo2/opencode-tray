<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Skeleton from './Skeleton.vue'
import ResumenWidget from './ResumenWidget.vue'

interface URLItem {
  url: string
  kind: string
}

const props = defineProps<{
  mode: string
  running: boolean
  urls: URLItem[]
  loading: boolean
  error: string
}>()

const emit = defineEmits<{ (e: 'open'): void }>()
const { t } = useI18n()

const mode = computed(() => {
  if (props.mode === 'network' || props.mode === 'remote') return props.mode
  return 'local'
})

const icon = computed(() => {
  if (mode.value === 'remote') return 'cloud'
  if (mode.value === 'network') return 'globe'
  return 'server'
})

const tone = computed(() => mode.value as 'local' | 'network' | 'remote')

const lanCount = computed(
  () => props.urls.filter((u) => u.kind === 'lan').length,
)

const description = computed(() => {
  if (mode.value === 'network') return t('resumen.accessLan', { n: lanCount.value })
  if (mode.value === 'remote') return t('resumen.accessRemote')
  return t('resumen.accessLocal')
})
</script>

<template>
  <ResumenWidget
    :title="t('resumen.accessTitle')"
    :icon="icon"
    :tone="tone"
    clickable
    :click-label="t('resumen.openAccess')"
    class="widget-acceso"
    @open="emit('open')"
  >
    <template #meta>
      <span v-if="mode === 'remote'" class="access-chip">
        <Icon name="zap" :size="11" />
        {{ t('resumen.internet') }}
      </span>
    </template>

    <div v-if="loading" class="skeleton-stack" role="status">
      <span class="sr-only">{{ t('common.loading') }}</span>
      <Skeleton width="90px" height="18px" />
      <Skeleton width="75%" height="12px" />
    </div>
    <p v-else-if="error" class="error" role="alert">{{ error }}</p>
    <template v-else>
      <div class="widget-mode">
        <span class="widget-mode-label">
          {{ t('web.' + mode) }}
        </span>
      </div>
      <p class="widget-desc">{{ description }}</p>
      <p v-if="!running" class="widget-desc widget-desc-warn">
        {{ t('resumen.accessStopped') }}
      </p>
    </template>
  </ResumenWidget>
</template>
