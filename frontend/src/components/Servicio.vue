<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import {
  ServiceStatus,
  ServiceStart,
  ServiceStop,
  ServiceForceStop,
} from '../../wailsjs/go/main/App'

const { t } = useI18n()
const running = ref(false)
const url = ref('')
const busy = ref(false)
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
}

onMounted(refresh)

async function run(fn: () => Promise<void>) {
  busy.value = true
  error.value = ''
  try {
    await fn()
    await refresh()
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
        <Icon name="play" :size="16" />{{ t('service.start') }}
      </button>
      <button :disabled="busy || !running" @click="run(() => ServiceStop())">
        <Icon name="stop" :size="16" />{{ t('service.stop') }}
      </button>
      <button :disabled="busy" @click="run(() => ServiceForceStop())">
        <Icon name="zap" :size="16" />{{ t('service.forceStop') }}
      </button>
      <button :disabled="busy" @click="refresh">
        <Icon name="refresh" :size="16" />{{ t('service.refresh') }}
      </button>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
  </section>
</template>
