<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
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
const error = ref('')
const saved = ref(false)

async function refresh() {
  error.value = ''
  try {
    urls.value = await WebAccessURLs()
    username.value = await WebAccessUsername()
    password.value = await WebAccessCurrentPassword()
  } catch (e) {
    error.value = String(e)
  }
}
onMounted(refresh)

async function applyMode(mode: string) {
  busy.value = true
  error.value = ''
  saved.value = false
  try {
    await WebAccessApplyMode(mode)
    await refresh()
  } catch (e) {
    error.value = String(e)
  } finally {
    busy.value = false
  }
}

async function changePassword() {
  if (!newPw.value) return
  busy.value = true
  error.value = ''
  saved.value = false
  try {
    await WebAccessChangePassword(newPw.value)
    newPw.value = ''
    saved.value = true
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
    <h2>{{ t('nav.accesoWeb') }}</h2>

    <h3>{{ t('web.mode') }}</h3>
    <div class="actions">
      <button :disabled="busy" @click="applyMode('local')">
        <Icon name="server" :size="16" />{{ t('web.local') }}
      </button>
      <button :disabled="busy" @click="applyMode('network')">
        <Icon name="globe" :size="16" />{{ t('web.network') }}
      </button>
      <button :disabled="busy" @click="applyMode('remote')">
        <Icon name="cloud" :size="16" />{{ t('web.remote') }}
      </button>
    </div>

    <h3>{{ t('web.urls') }}</h3>
    <p v-if="urls.length === 0" class="muted">{{ t('web.noUrls') }}</p>
    <ul class="urls">
      <li v-for="u in urls" :key="u">{{ u }}</li>
    </ul>

    <h3>{{ t('web.username') }} / {{ t('web.password') }}</h3>
    <p class="muted">
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
      <button :disabled="busy || !newPw" @click="changePassword">
        <Icon name="check" :size="16" />{{ t('web.save') }}
      </button>
    </div>
    <p v-if="saved" class="ok-msg">{{ t('web.saved') }}</p>
    <p v-if="error" class="error">{{ error }}</p>
  </section>
</template>
