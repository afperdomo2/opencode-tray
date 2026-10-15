<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Switch from './Switch.vue'
import LanguageSwitch from './LanguageSwitch.vue'
import {
  GetSettings,
  SetAutoStart,
  SetLanguage,
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

async function onLocale(lang: string) {
  try {
    await SetLanguage(lang)
  } catch {
    // Sin backend (tests): se ignora.
  }
}

async function onAutoStart(on: boolean) {
  autoStart.value = on
  try {
    await SetAutoStart(on)
  } catch {
    // Sin backend (tests): se ignora.
  }
}

function onNotify(kind: string, on: boolean) {
  if (kind === 'burnrate') burnrate.value = on
  if (kind === 'exhaustion') exhaustion.value = on
  if (kind === 'service') service.value = on
  try {
    NotifySetEnabled(kind, on)
  } catch {
    // Sin backend (tests): se ignora.
  }
}

const onNotifyBurnrate = (v: boolean) => onNotify('burnrate', v)
const onNotifyExhaustion = (v: boolean) => onNotify('exhaustion', v)
const onNotifyService = (v: boolean) => onNotify('service', v)
</script>

<template>
  <section class="section">
    <h2>{{ t('nav.ajustes') }}</h2>

    <h3>{{ t('common.language') }}</h3>
    <LanguageSwitch @change="onLocale" />

    <h3>{{ t('common.general') }}</h3>
    <Switch :checked="autoStart" @change="onAutoStart">{{ t('common.autoStart') }}</Switch>

    <h3>{{ t('notify.title') }}</h3>
    <Switch :checked="burnrate" @change="onNotifyBurnrate">{{ t('notify.burnrate') }}</Switch>
    <Switch :checked="exhaustion" @change="onNotifyExhaustion">
      {{ t('notify.exhaustion') }}
    </Switch>
    <Switch :checked="service" @change="onNotifyService">{{ t('notify.service') }}</Switch>

    <div class="actions">
      <button @click="NotifyEvaluate">
        <Icon name="bell" :size="16" />{{ t('notify.test') }}
      </button>
    </div>
  </section>
</template>
