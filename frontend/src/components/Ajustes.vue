<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Switch from './Switch.vue'
import LanguageSwitch from './LanguageSwitch.vue'
import Skeleton from './Skeleton.vue'
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
const loadingSettings = ref(true)

onMounted(async () => {
  try {
    const s = await GetSettings()
    autoStart.value = s.autoStart
  } catch {
    // Sin backend (tests): mantiene el valor por defecto.
  } finally {
    loadingSettings.value = false
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

    <div class="settings-grid">
      <div class="card">
        <div class="card-head">
          <Icon name="languages" :size="16" />
          <h3>{{ t('common.language') }}</h3>
        </div>
        <LanguageSwitch @change="onLocale" />
      </div>

      <div class="card">
        <div class="card-head">
          <Icon name="settings" :size="16" />
          <h3>{{ t('common.general') }}</h3>
        </div>

        <div v-if="loadingSettings" class="skeleton-stack" role="status">
          <span class="sr-only">{{ t('common.loading') }}</span>
          <div class="status-row">
            <Skeleton width="40px" height="22px" radius="999px" />
            <Skeleton width="140px" height="14px" />
          </div>
        </div>
        <Switch v-else :checked="autoStart" @change="onAutoStart">{{ t('common.autoStart') }}</Switch>
      </div>

      <div class="card">
        <div class="card-head">
          <Icon name="bell" :size="16" />
          <h3>{{ t('notify.title') }}</h3>
        </div>
        <Switch :checked="burnrate" @change="onNotifyBurnrate">{{ t('notify.burnrate') }}</Switch>
        <Switch :checked="exhaustion" @change="onNotifyExhaustion">
          {{ t('notify.exhaustion') }}
        </Switch>
        <Switch :checked="service" @change="onNotifyService">{{ t('notify.service') }}</Switch>

        <div class="actions">
          <button class="btn" @click="NotifyEvaluate">
            <Icon name="bell" :size="14" />{{ t('notify.test') }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>
