<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Skeleton from './Skeleton.vue'
import {
  WebAccessURLs,
  WebAccessUsername,
  WebAccessCurrentPassword,
  WebAccessChangePassword,
  WebAccessApplyMode,
} from '../../wailsjs/go/main/App'

const { t } = useI18n()
const urls = ref<string[]>([])
const username = ref('opencode')
const password = ref('')
const showPassword = ref(false)
const newPw = ref('')
const busy = ref(false)
const saved = ref(false)

const loadingUrls = ref(true)
const loadingCredentials = ref(true)
const errorUrls = ref('')
const errorCredentials = ref('')

async function loadUrls() {
  loadingUrls.value = true
  errorUrls.value = ''
  try {
    urls.value = await WebAccessURLs()
  } catch (e) {
    errorUrls.value = String(e)
  } finally {
    loadingUrls.value = false
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
  await Promise.all([loadUrls(), loadCredentials()])
}
onMounted(refresh)

async function applyMode(mode: string) {
  busy.value = true
  errorUrls.value = ''
  saved.value = false
  try {
    await WebAccessApplyMode(mode)
    await refresh()
  } catch (e) {
    errorUrls.value = String(e)
  } finally {
    busy.value = false
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

        <div class="actions">
          <button class="btn" :disabled="busy" @click="applyMode('local')">
            <Icon name="server" :size="14" />{{ t('web.local') }}
          </button>
          <button class="btn" :disabled="busy" @click="applyMode('network')">
            <Icon name="globe" :size="14" />{{ t('web.network') }}
          </button>
          <button class="btn" :disabled="busy" @click="applyMode('remote')">
            <Icon name="cloud" :size="14" />{{ t('web.remote') }}
          </button>
        </div>

        <h3>{{ t('web.urls') }}</h3>

        <div v-if="loadingUrls" class="skeleton-stack" role="status">
          <span class="sr-only">{{ t('common.loading') }}</span>
          <Skeleton width="85%" height="14px" />
          <Skeleton width="65%" height="14px" />
        </div>
        <p v-else-if="errorUrls" class="error">{{ errorUrls }}</p>
        <template v-else>
          <p v-if="urls.length === 0" class="muted">{{ t('web.noUrls') }}</p>
          <ul class="urls">
            <li v-for="u in urls" :key="u">{{ u }}</li>
          </ul>
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
            @click="showPassword = !showPassword"
          >
            <Icon :name="showPassword ? 'eyeOff' : 'eye'" :size="16" />
          </button>
        </p>

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
  </section>
</template>
