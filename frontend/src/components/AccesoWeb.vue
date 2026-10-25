<script setup lang="ts">
import { ref, computed, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Skeleton from './Skeleton.vue'
import Modal from './Modal.vue'
import {
  WebAccessState,
  WebAccessUsername,
  WebAccessCurrentPassword,
  WebAccessChangePassword,
  WebAccessApplyMode,
  WebAccessOpenURL,
} from '../../wailsjs/go/main/App'
import { ClipboardSetText } from '../../wailsjs/runtime/runtime'
import { clearQueries, useQuery } from '../lib/query'
import {
  SERVICE_KEY,
  WEB_ACCESS_KEY,
  WEB_CREDENTIALS_KEY,
} from '../lib/query-keys'

interface URLItem {
  url: string
  kind: string
}

interface AccessState {
  mode: string
  running: boolean
  urls: URLItem[]
}

interface Credentials {
  user: string
  pass: string
}

const { t } = useI18n()

const access = useQuery<AccessState>({
  key: WEB_ACCESS_KEY,
  fetcher: async () => (await WebAccessState()) as AccessState,
  staleTime: 15_000,
})
const credentials = useQuery<Credentials>({
  key: WEB_CREDENTIALS_KEY,
  fetcher: async () => {
    const [user, pass] = await Promise.all([
      WebAccessUsername(),
      WebAccessCurrentPassword(),
    ])
    return { user, pass }
  },
  staleTime: 60_000,
})

const busy = ref(false)
const mutating = ref(false)
const manualRefreshing = ref(false)
const showPassword = ref(false)
const newPw = ref('')
const saved = ref(false)
const confirmRemote = ref(false)
const copiedUrl = ref('')
const actionError = ref('')

// Último modo conocido: mantiene resaltado el botón correcto mientras el estado
// se descarta durante una mutación (evita que "salte" a Local).
const lastMode = ref('local')

let copyTimer: ReturnType<typeof setTimeout> | undefined

watch(
  () => access.data.value?.mode,
  (mode) => {
    if (mode) lastMode.value = mode
  },
)

const mode = computed(() => access.data.value?.mode ?? lastMode.value)
const running = computed(() => access.data.value?.running ?? false)
const urls = computed(() => access.data.value?.urls ?? [])

const modeDescription = computed(() => t(`web.modeHint.${mode.value}`))
const showWarning = computed(
  () => running.value && (mode.value === 'network' || mode.value === 'remote'),
)

const loadingState = computed(
  () => access.isInitialLoading.value || mutating.value,
)
const loadingCredentials = computed(
  () => credentials.isInitialLoading.value || mutating.value,
)
const refreshing = computed(
  () => manualRefreshing.value || mutating.value,
)

const hasStateData = computed(() => access.data.value !== undefined)
const errorState = computed(() =>
  access.error.value ? String(access.error.value) : '',
)
const hasCredentialsData = computed(
  () => credentials.data.value !== undefined,
)
const errorCredentials = computed(() =>
  credentials.error.value ? String(credentials.error.value) : '',
)

const username = computed(() => credentials.data.value?.user ?? 'opencode')
const password = computed(() => credentials.data.value?.pass ?? '')

// Anchos escalonados para que el skeleton no se vea como una lista uniforme.
const skeletonWidths = ['88%', '64%']

onBeforeUnmount(() => {
  if (copyTimer !== undefined) clearTimeout(copyTimer)
})

async function refresh() {
  actionError.value = ''
  manualRefreshing.value = true
  try {
    await Promise.all([access.refetchHard(), credentials.refetchHard()])
  } finally {
    manualRefreshing.value = false
  }
}

function selectMode(next: string) {
  if (next === 'remote') {
    confirmRemote.value = true
    return
  }
  void applyMode(next)
}

async function applyMode(next: string) {
  busy.value = true
  actionError.value = ''
  saved.value = false
  confirmRemote.value = false
  // Refresco duro desde el inicio: el estado anterior deja de ser válido.
  lastMode.value = next
  mutating.value = true
  clearQueries([WEB_ACCESS_KEY, SERVICE_KEY])
  try {
    await WebAccessApplyMode(next)
  } catch (e) {
    actionError.value = String(e)
  } finally {
    busy.value = false
  }
  await access.refetch()
  mutating.value = false
}

async function openUrl(url: string) {
  actionError.value = ''
  try {
    await WebAccessOpenURL(url)
  } catch (e) {
    actionError.value = String(e)
  }
}

async function copyUrl(url: string) {
  actionError.value = ''
  try {
    const ok = await ClipboardSetText(url)
    if (!ok) {
      actionError.value = t('web.copyError')
      return
    }
    copiedUrl.value = url
    if (copyTimer !== undefined) clearTimeout(copyTimer)
    copyTimer = setTimeout(() => {
      copiedUrl.value = ''
    }, 1500)
  } catch (e) {
    actionError.value = String(e)
  }
}

async function changePassword() {
  if (!newPw.value) return
  busy.value = true
  actionError.value = ''
  saved.value = false
  mutating.value = true
  clearQueries([WEB_ACCESS_KEY, WEB_CREDENTIALS_KEY, SERVICE_KEY])
  try {
    await WebAccessChangePassword(newPw.value)
    newPw.value = ''
  } catch (e) {
    actionError.value = String(e)
  } finally {
    busy.value = false
  }
  await Promise.all([access.refetch(), credentials.refetch()])
  mutating.value = false
  if (!actionError.value) saved.value = true
}
</script>

<template>
  <section class="section">
    <h2>{{ t('nav.accesoWeb') }}</h2>

    <div class="panel-grid">
      <div class="card">
        <div class="card-head">
          <Icon name="globe" :size="16" />
          <h3>{{ t('web.mode') }}</h3>
          <button
            type="button"
            class="icon-btn card-action"
            :title="t('web.refreshHint')"
            :aria-label="t('web.refreshHint')"
            :disabled="busy || refreshing"
            :aria-busy="refreshing"
            @click="refresh"
          >
            <Icon name="refresh" :size="15" :class="{ spinning: refreshing }" />
          </button>
        </div>

        <div class="mode-actions" role="group" :aria-label="t('web.mode')">
          <button
            class="btn mode-btn"
            :class="{ active: mode === 'local' }"
            :disabled="busy"
            :title="t('web.modeHint.local')"
            :aria-label="t('web.modeHint.local')"
            :aria-pressed="mode === 'local'"
            @click="selectMode('local')"
          >
            <Icon name="server" :size="14" />{{ t('web.local') }}
          </button>
          <button
            class="btn mode-btn"
            :class="{ active: mode === 'network' }"
            :disabled="busy"
            :title="t('web.modeHint.network')"
            :aria-label="t('web.modeHint.network')"
            :aria-pressed="mode === 'network'"
            @click="selectMode('network')"
          >
            <Icon name="globe" :size="14" />{{ t('web.network') }}
          </button>
          <button
            class="btn mode-btn"
            :class="{ active: mode === 'remote' }"
            :disabled="busy"
            :title="t('web.modeHint.remote')"
            :aria-label="t('web.modeHint.remote')"
            :aria-pressed="mode === 'remote'"
            @click="selectMode('remote')"
          >
            <Icon name="cloud" :size="14" />{{ t('web.remote') }}
          </button>
        </div>

        <p class="mode-hint">
          <span v-if="mode" class="mode-active">{{ t('web.active') }}</span>
          {{ modeDescription }}
        </p>

        <p v-if="showWarning" class="warning" role="status">
          <Icon name="zap" :size="14" />
          {{ mode === 'remote' ? t('web.remoteWarning') : t('web.networkWarning') }}
        </p>

        <h3>{{ t('web.urls') }}</h3>

        <div v-if="loadingState" class="skeleton-stack" role="status" :aria-label="t('common.loading')">
          <div v-for="(width, i) in skeletonWidths" :key="i" class="url-row skeleton-url-row">
            <Skeleton width="46px" height="18px" radius="9px" />
            <Skeleton :width="width" height="14px" />
            <Skeleton width="30px" height="30px" radius="6px" />
            <Skeleton width="30px" height="30px" radius="6px" />
          </div>
        </div>
        <p v-else-if="errorState && !hasStateData" class="error" role="alert">{{ errorState }}</p>
        <template v-else>
          <p v-if="urls.length === 0" class="muted">{{ t('web.noUrls') }}</p>
          <ul v-else class="urls url-list">
            <li v-for="item in urls" :key="item.url" class="url-row">
              <span class="url-badge" :class="'kind-' + item.kind">
                {{ t('web.kind.' + item.kind) }}
              </span>
              <code class="url-text" :title="item.url">{{ item.url }}</code>
              <button
                type="button"
                class="icon-btn url-action"
                :disabled="!running"
                :title="running ? t('web.open') : t('web.startToOpen')"
                :aria-label="running ? t('web.open') : t('web.startToOpen')"
                @click="openUrl(item.url)"
              >
                <Icon name="externalLink" :size="15" />
              </button>
              <button
                type="button"
                class="icon-btn url-action"
                :title="copiedUrl === item.url ? t('web.copied') : t('web.copy')"
                :aria-label="copiedUrl === item.url ? t('web.copied') : t('web.copy')"
                @click="copyUrl(item.url)"
              >
                <Icon :name="copiedUrl === item.url ? 'check' : 'copy'" :size="15" />
              </button>
            </li>
          </ul>
          <p v-if="copiedUrl" class="ok-msg" role="status" aria-live="polite">
            {{ t('web.copied') }}
          </p>
          <p v-if="errorState" class="error" role="alert">{{ errorState }}</p>
          <p v-if="actionError" class="error" role="alert">{{ actionError }}</p>
          <p v-if="!running" class="muted">{{ t('web.stoppedHint') }}</p>
        </template>
      </div>

      <div class="card">
        <div class="card-head">
          <Icon name="eye" :size="16" />
          <h3>{{ t('web.username') }} / {{ t('web.password') }}</h3>
        </div>

        <div v-if="loadingCredentials" class="skeleton-stack" role="status">
          <span class="sr-only">{{ t('common.loading') }}</span>
          <Skeleton width="60%" height="14px" />
        </div>
        <p v-else-if="errorCredentials && !hasCredentialsData" class="error">
          {{ errorCredentials }}
        </p>
        <template v-else>
          <p class="muted">
            {{ username }} ·
            <code>{{ showPassword ? password : '••••••' }}</code>
            <button
              type="button"
              class="icon-btn"
              :aria-label="showPassword ? t('web.hide') : t('web.show')"
              :title="showPassword ? t('web.hide') : t('web.show')"
              @click="showPassword = !showPassword"
            >
              <Icon :name="showPassword ? 'eyeOff' : 'eye'" :size="16" />
            </button>
          </p>
          <p v-if="errorCredentials" class="error">{{ errorCredentials }}</p>
        </template>

        <p class="muted session-hint">{{ t('web.sessionHint') }}</p>

        <div class="pw-change">
          <input
            v-model="newPw"
            type="password"
            :placeholder="t('web.newPassword')"
            :aria-label="t('web.newPassword')"
          />
          <button
            class="btn"
            :disabled="busy || loadingCredentials || !newPw"
            @click="changePassword"
          >
            <Icon name="check" :size="14" />{{ t('web.save') }}
          </button>
        </div>
        <p v-if="saved" class="ok-msg">{{ t('web.saved') }}</p>
      </div>
    </div>

    <Modal
      :open="confirmRemote"
      :title="t('web.remoteConfirmTitle')"
      @close="confirmRemote = false"
    >
      <p class="warning" role="alert">
        <Icon name="zap" :size="14" />
        {{ t('web.remoteConfirmBody') }}
      </p>
      <p class="muted">{{ t('web.remoteConfirmHint') }}</p>
      <div class="modal-actions">
        <button type="button" class="btn" @click="confirmRemote = false">
          {{ t('common.cancel') }}
        </button>
        <button type="button" class="btn btn-danger" :disabled="busy" @click="applyMode('remote')">
          <Icon name="cloud" :size="14" />{{ t('web.remoteConfirmEnable') }}
        </button>
      </div>
    </Modal>
  </section>
</template>
