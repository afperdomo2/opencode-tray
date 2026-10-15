<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  GetSettings,
  SetAutoStart,
  NotifySetEnabled,
  NotifyEvaluate,
} from '../../wailsjs/go/main/App'

const { t } = useI18n()
const autoStart = ref(false)
const burnrate = ref(true)
const exhaustion = ref(true)
const service = ref(true)

onMounted(async () => {
  try {
    const s = await GetSettings()
    autoStart.value = s.autoStart
  } catch {
    // Sin backend (tests): mantiene el valor por defecto.
  }
})

async function onAutoStart(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  autoStart.value = on
  try {
    await SetAutoStart(on)
  } catch {
    // Sin backend (tests): se ignora.
  }
}

function onNotify(kind: string, e: Event) {
  const on = (e.target as HTMLInputElement).checked
  if (kind === 'burnrate') burnrate.value = on
  if (kind === 'exhaustion') exhaustion.value = on
  if (kind === 'service') service.value = on
  try {
    NotifySetEnabled(kind, on)
  } catch {
    // Sin backend (tests): se ignora.
  }
}
</script>

<template>
  <section class="section">
    <h2>{{ t('nav.ajustes') }}</h2>

    <label class="toggle">
      <input type="checkbox" :checked="autoStart" @change="onAutoStart" />
      <span>{{ t('common.autoStart') }}</span>
    </label>

    <h3>{{ t('notify.title') }}</h3>
    <label class="toggle">
      <input type="checkbox" :checked="burnrate" @change="onNotify('burnrate', $event)" />
      <span>{{ t('notify.burnrate') }}</span>
    </label>
    <label class="toggle">
      <input type="checkbox" :checked="exhaustion" @change="onNotify('exhaustion', $event)" />
      <span>{{ t('notify.exhaustion') }}</span>
    </label>
    <label class="toggle">
      <input type="checkbox" :checked="service" @change="onNotify('service', $event)" />
      <span>{{ t('notify.service') }}</span>
    </label>

    <button class="mini" @click="NotifyEvaluate">{{ t('notify.test') }}</button>
  </section>
</template>
