<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Skeleton from './Skeleton.vue'
import ResumenWidget from './ResumenWidget.vue'

// Widget de estado del servicio. El color es un refuerzo: el texto
// "Corriendo/Detenido" siempre está visible.
const props = defineProps<{
  running: boolean
  url: string
  loading: boolean
  error: string
}>()

const emit = defineEmits<{ (e: 'open'): void }>()
const { t } = useI18n()

const displayUrl = computed(() => props.url.replace(/^https?:\/\//i, ''))
</script>

<template>
  <ResumenWidget
    :title="t('nav.servicio')"
    icon="server"
    :tone="running ? 'running' : 'stopped'"
    clickable
    :click-label="t('resumen.openService')"
    class="widget-servicio"
    @open="emit('open')"
  >
    <div v-if="loading" class="skeleton-stack" role="status">
      <span class="sr-only">{{ t('common.loading') }}</span>
      <Skeleton width="110px" height="16px" />
      <Skeleton width="80%" height="12px" />
    </div>
    <p v-else-if="error" class="error" role="alert">{{ error }}</p>
    <template v-else>
      <div class="widget-state">
        <span class="dot" :class="running ? 'on' : 'off'" aria-hidden="true"></span>
        <strong>{{ running ? t('service.running') : t('service.stopped') }}</strong>
      </div>
      <p v-if="url" class="widget-url" :title="url">{{ displayUrl }}</p>
      <p v-else class="widget-desc">{{ t('resumen.noUrl') }}</p>
    </template>
  </ResumenWidget>
</template>
