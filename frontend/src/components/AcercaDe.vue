<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Skeleton from './Skeleton.vue'
import { GetAppVersion } from '../../wailsjs/go/main/App'
import { ClipboardSetText } from '../../wailsjs/runtime/runtime'

// Contrato que expone App.GetAppVersion (buildinfo.go).
interface AppVersion {
  version: string
  commit: string
  date: string
  dirty: boolean
  platform: string
}

const { t, locale } = useI18n()
const info = ref<AppVersion | null>(null)
const loading = ref(true)
const copied = ref(false)
const copyError = ref('')
let copyTimer: ReturnType<typeof setTimeout> | undefined

onMounted(async () => {
  try {
    info.value = await GetAppVersion()
  } catch {
    // Sin puente (tests/dev) se muestra "No disponible".
    info.value = null
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => {
  if (copyTimer !== undefined) clearTimeout(copyTimer)
})

const unavailable = computed(() => t('about.unavailable'))
const versionLabel = computed(() => {
  const version = info.value?.version
  return version ? `v${version}` : unavailable.value
})
const commitLabel = computed(() => {
  const commit = info.value?.commit
  if (!commit) return unavailable.value
  return info.value?.dirty ? `${commit} *` : commit
})
const dateLabel = computed(() => {
  const raw = info.value?.date
  if (!raw) return unavailable.value
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return new Intl.DateTimeFormat(locale.value, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
})
const platformLabel = computed(() => info.value?.platform || unavailable.value)

function buildCopyText(): string {
  return [
    `${t('app.title')} ${versionLabel.value}`,
    `${t('about.commit')}: ${commitLabel.value}`,
    `${t('about.date')}: ${dateLabel.value}`,
    `${t('about.platform')}: ${platformLabel.value}`,
  ].join('\n')
}

async function copyInfo() {
  copyError.value = ''
  try {
    const ok = await ClipboardSetText(buildCopyText())
    if (!ok) {
      copyError.value = t('about.copyError')
      return
    }
    copied.value = true
    if (copyTimer !== undefined) clearTimeout(copyTimer)
    copyTimer = setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch {
    copyError.value = t('about.copyError')
  }
}
</script>

<template>
  <section class="card about-card">
    <div class="card-head">
      <Icon name="info" :size="16" />
      <h3>{{ t('about.title') }}</h3>
    </div>

    <div v-if="loading" class="skeleton-stack" role="status">
      <span class="sr-only">{{ t('common.loading') }}</span>
      <Skeleton width="40%" height="14px" />
      <Skeleton width="70%" height="14px" />
    </div>
    <template v-else>
      <dl class="about-rows">
        <div class="about-row">
          <dt class="about-label">{{ t('about.version') }}</dt>
          <dd class="about-value about-version">{{ versionLabel }}</dd>
        </div>
        <div class="about-row">
          <dt class="about-label">{{ t('about.commit') }}</dt>
          <dd
            class="about-value about-commit"
            :title="info?.dirty ? t('about.dirtyHint') : undefined"
          >
            {{ commitLabel }}
          </dd>
        </div>
        <div class="about-row">
          <dt class="about-label">{{ t('about.date') }}</dt>
          <dd class="about-value about-date">{{ dateLabel }}</dd>
        </div>
        <div class="about-row">
          <dt class="about-label">{{ t('about.platform') }}</dt>
          <dd class="about-value about-platform">{{ platformLabel }}</dd>
        </div>
      </dl>

      <div class="actions">
        <button type="button" class="btn" @click="copyInfo">
          <Icon :name="copied ? 'check' : 'copy'" :size="14" />
          {{ copied ? t('about.copied') : t('about.copy') }}
        </button>
      </div>

      <p v-if="copied" class="ok-msg" role="status" aria-live="polite">
        {{ t('about.copied') }}
      </p>
      <p v-if="copyError" class="error" role="alert">{{ copyError }}</p>
    </template>
  </section>
</template>
