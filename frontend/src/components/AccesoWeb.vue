<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
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

interface URLItem {
  url: string
  kind: string
}

interface AccessState {
  mode: string
  running: boolean
  urls: URLItem[]
}

const { t } = useI18n()

const state = ref<AccessState>({ mode: 'local', running: false, urls: [] })
const username = ref('opencode')
const password = ref('')
const showPassword = ref(false)
const newPw = ref('')
const busy = ref(false)
const saved = ref(false)
const confirmRemote = ref(false)
const copiedUrl = ref('')
const actionError = ref('')

const loadingState = ref(true)
const loadingCredentials = ref(true)
const errorState = ref('')
const errorCredentials = ref('')

let copyTimer: ReturnType<typeof setTimeout> | undefined

const modeDescription = computed(() => t(`web.modeHint.${state.value.mode}`))
const showWarning = computed(
  () => state.value.running && (state.value.mode === 'network' || state.value.mode === 'remote'),
)

// Anchos escalonados para que el skeleton no se vea como una lista uniforme.
const skeletonWidths = ['88%', '64%']

async function loadState() {
  loadingState.value = true
  errorState.value = ''
  try {
    state.value = (await WebAccessState()) as AccessState
  } catch (e) {
    errorState.value = String(e)
  } finally {
    loadingState.value = false
  }
}

async function loadCredentials() {
  loadingCredentials.value = true
  errorCredentials.value = ''
  try {
    const [user, pass] = await Promise.all([WebAccessUsername(), WebAccessCurrentPassword()])
    username.value = user
    password.value = pass
  } catch (e) {
    errorCredentials.value = String(e)
  } finally {
    loadingCredentials.value = false
  }
}

async function refresh() {
  await Promise.all([loadState(), loadCredentials()])
}
onMounted(refresh)

onBeforeUnmount(() => {
  if (copyTimer !== undefined) clearTimeout(copyTimer)
})

function selectMode(mode: string) {
  if (mode === 'remote') {
    confirmRemote.value = true
    return
  }
  void applyMode(mode)
}

async function applyMode(mode: string) {
  busy.value = true
  errorState.value = ''
  actionError.value = ''
  saved.value = false
  confirmRemote.value = false
  try {
    await WebAccessApplyMode(mode)
    await loadState()
  } catch (e) {
    errorState.value = String(e)
  } finally {
    busy.value = false
  }
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
  errorCredentials.value = ''
  saved.value = false
  try {
    await WebAccessChangePassword(newPw.value)
    newPw.value = ''
    saved.value = true
    await loadCredentials()
  } catch (e) {
    errorCredentials.value = String(e)
  } finally {
    busy.value = false
  }
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
        </div>

        <div class="mode-actions" role="group" :aria-label="t('web.mode')">
          <button
            class="btn mode-btn"
            :class="{ active: state.mode === 'local' }"
            :disabled="busy"
            :title="t('web.modeHint.local')"
            :aria-label="t('web.modeHint.local')"
            :aria-pressed="state.mode === 'local'"
            @click="selectMode('local')"
          >
            <Icon name="server" :size="14" />{{ t('web.local') }}
          </button>
          <button
            class="btn mode-btn"
            :class="{ active: state.mode === 'network' }"
            :disabled="busy"
            :title="t('web.modeHint.network')"
            :aria-label="t('web.modeHint.network')"
            :aria-pressed="state.mode === 'network'"
            @click="selectMode('network')"
          >
            <Icon name="globe" :size="14" />{{ t('web.network') }}
          </button>
          <button
            class="btn mode-btn"
            :class="{ active: state.mode === 'remote' }"
            :disabled="busy"
            :title="t('web.modeHint.remote')"
            :aria-label="t('web.modeHint.remote')"
            :aria-pressed="state.mode === 'remote'"
            @click="selectMode('remote')"
          >
            <Icon name="cloud" :size="14" />{{ t('web.remote') }}
          </button>
        </div>

        <p class="mode-hint">
          <span v-if="state.mode" class="mode-active">{{ t('web.active') }}</span>
          {{ modeDescription }}
        </p>

        <p v-if="showWarning" class="warning" role="status">
          <Icon name="zap" :size="14" />
          {{ state.mode === 'remote' ? t('web.remoteWarning') : t('web.networkWarning') }}
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
        <p v-else-if="errorState" class="error" role="alert">{{ errorState }}</p>
        <template v-else>
          <p v-if="state.urls.length === 0" class="muted">{{ t('web.noUrls') }}</p>
          <ul v-else class="urls url-list">
            <li v-for="item in state.urls" :key="item.url" class="url-row">
              <span class="url-badge" :class="'kind-' + item.kind">
                {{ t('web.kind.' + item.kind) }}
              </span>
              <code class="url-text" :title="item.url">{{ item.url }}</code>
              <button
                type="button"
                class="icon-btn url-action"
                :disabled="!state.running"
                :title="state.running ? t('web.open') : t('web.startToOpen')"
                :aria-label="state.running ? t('web.open') : t('web.startToOpen')"
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
          <p v-if="actionError" class="error" role="alert">{{ actionError }}</p>
          <p v-if="!state.running" class="muted">{{ t('web.stoppedHint') }}</p>
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
        <p v-else-if="errorCredentials" class="error">{{ errorCredentials }}</p>
        <p v-else class="muted">
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
