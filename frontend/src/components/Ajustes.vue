<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Switch from './Switch.vue'
import LanguageSwitch from './LanguageSwitch.vue'
import SegmentedControl from './SegmentedControl.vue'
import Skeleton from './Skeleton.vue'
import { NotifySetEnabled, NotifyEvaluate } from '../../wailsjs/go/main/App'
import {
  loadSettings,
  setAutoStart as persistAutoStart,
  setLanguage,
  setRefreshMin as persistRefreshMin,
  settings,
} from '../lib/settings'
import { useUsageQuery, usageErrorKey } from '../lib/usage'

const { t } = useI18n()
const burnrate = ref(true)
const exhaustion = ref(true)
const service = ref(true)
const loadingSettings = ref(true)
const tokenInput = ref('')
const tokenSaved = ref(false)

const usage = useUsageQuery()
const tokenStatus = usage.tokenStatus
const tokenBusy = usage.tokenBusy
const tokenError = usage.tokenError

const tokenConfigured = computed(() => tokenStatus.value?.configured === true)
const tokenErrorLabel = computed(() => {
  if (!tokenError.value) return ''
  return String(tokenError.value).includes('USAGE_')
    ? usageErrorKey(tokenError.value)
    : 'usage.tokenError'
})
const tokenSource = computed(() => {
  const s = tokenStatus.value
  if (!s?.configured) return t('usage.sourceNone')
  if (s.source === 'opencode-export') return t('usage.sourceExport')
  if (s.source === 'opencode-json') return t('usage.sourceJson')
  return t('usage.sourceManual')
})

const intervals = [1, 5, 15, 30, 60]

onMounted(async () => {
  try {
    await loadSettings()
    await usage.loadTokenStatus()
  } finally {
    loadingSettings.value = false
  }
})

async function onLocale(lang: string) {
  await setLanguage(lang)
}

async function onAutoStart(on: boolean) {
  await persistAutoStart(on)
}

async function onRefresh(min: number) {
  await persistRefreshMin(min)
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

function retryFromOpenCode() {
  tokenSaved.value = false
  void usage.reloadToken()
}

async function saveToken() {
  if (!tokenInput.value.trim()) return
  tokenSaved.value = false
  const ok = await usage.saveManualToken(tokenInput.value)
  if (ok) {
    tokenInput.value = ''
    tokenSaved.value = true
  }
}

async function clearToken() {
  tokenSaved.value = false
  await usage.clearManualToken()
}
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
        <template v-else>
          <Switch :checked="settings.autoStart" @change="onAutoStart">
            {{ t('common.autoStart') }}
          </Switch>

          <div class="field">
            <span class="field-label">{{ t('usage.refreshInterval') }}</span>
            <SegmentedControl
              :model-value="settings.refreshMin"
              :options="intervals"
              :label="t('usage.refreshInterval')"
              :unit="t('usage.minuteUnit')"
              :format-label="(n: number) => t('usage.everyMinutes', { n })"
              @change="onRefresh"
            />
          </div>
        </template>
      </div>

      <div class="card">
        <div class="card-head">
          <Icon name="key" :size="16" />
          <h3>{{ t('usage.tokenTitle') }}</h3>
        </div>

        <div v-if="loadingSettings" class="skeleton-stack" role="status">
          <span class="sr-only">{{ t('common.loading') }}</span>
          <Skeleton width="60%" height="14px" />
          <Skeleton width="90%" height="32px" />
        </div>
        <template v-else>
          <div class="status-row">
            <span class="dot" :class="tokenConfigured ? 'on' : 'off'" aria-hidden="true"></span>
            <span>{{ tokenSource }}</span>
          </div>
          <p class="muted">{{ t('usage.tokenDesc') }}</p>

          <div class="actions">
            <button
              type="button"
              class="btn btn-full"
              :disabled="tokenBusy"
              :aria-busy="tokenBusy"
              @click="retryFromOpenCode"
            >
              <Icon name="refresh" :size="14" />
              {{ tokenBusy ? t('usage.reloading') : t('usage.reloadFromOpenCode') }}
            </button>
          </div>

          <form class="token-form" @submit.prevent="saveToken">
            <input
              v-model="tokenInput"
              type="password"
              class="input"
              autocomplete="off"
              :placeholder="t('usage.tokenPlaceholder')"
              :aria-label="t('usage.tokenPlaceholder')"
            />
            <button type="submit" class="btn" :disabled="tokenBusy || !tokenInput.trim()">
              {{ t('usage.tokenSave') }}
            </button>
          </form>
          <div class="actions">
            <button type="button" class="btn" :disabled="tokenBusy" @click="clearToken">
              {{ t('usage.tokenClear') }}
            </button>
          </div>

          <p v-if="tokenSaved" class="ok-msg" role="status">{{ t('usage.tokenSaved') }}</p>
          <p v-if="tokenErrorLabel" class="error" role="alert">{{ t(tokenErrorLabel) }}</p>
        </template>
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
