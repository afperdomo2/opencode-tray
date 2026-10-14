<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetSettings, SetAutoStart } from '../../wailsjs/go/main/App'

const { t } = useI18n()
const autoStart = ref(false)

onMounted(async () => {
  try {
    const s = await GetSettings()
    autoStart.value = s.autoStart
  } catch {
    // Sin backend (tests): mantiene el valor por defecto.
  }
})

async function onChange(e: Event) {
  const on = (e.target as HTMLInputElement).checked
  autoStart.value = on
  try {
    await SetAutoStart(on)
  } catch {
    // Sin backend (tests): se ignora.
  }
}
</script>

<template>
  <section class="section">
    <h2>{{ t('nav.ajustes') }}</h2>
    <label class="toggle">
      <input type="checkbox" :checked="autoStart" @change="onChange" />
      <span>{{ t('common.autoStart') }}</span>
    </label>
  </section>
</template>
