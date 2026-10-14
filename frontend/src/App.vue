<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetSettings, SetLanguage } from '../wailsjs/go/main/App'
import Section from './components/Section.vue'
import Ajustes from './components/Ajustes.vue'

const { t, locale } = useI18n()

const active = ref('resumen')
const sections = ['resumen', 'servicio', 'accesoWeb', 'consumo', 'config', 'ajustes']

onMounted(async () => {
  try {
    const s = await GetSettings()
    locale.value = s.language
  } catch {
    // Sin backend (tests): mantiene 'es' por defecto.
  }
})

async function onLocale(e: Event) {
  const lang = (e.target as HTMLSelectElement).value
  locale.value = lang
  try {
    await SetLanguage(lang)
  } catch {
    // Sin backend (tests): se ignora.
  }
}
</script>

<template>
  <div class="app">
    <header class="app-header">
      <h1>{{ t('app.title') }}</h1>
      <select
        class="lang-select"
        :value="locale"
        :aria-label="t('common.language')"
        @change="onLocale"
      >
        <option value="es">ES</option>
        <option value="en">EN</option>
      </select>
    </header>

    <nav class="app-nav" aria-label="Secciones">
      <button
        v-for="id in sections"
        :key="id"
        class="nav-btn"
        :class="{ active: active === id }"
        :aria-current="active === id ? 'page' : undefined"
        @click="active = id"
      >
        {{ t('nav.' + id) }}
      </button>
    </nav>

    <main class="app-content">
      <Ajustes v-if="active === 'ajustes'" />
      <Section v-else :name="active" />
    </main>
  </div>
</template>
