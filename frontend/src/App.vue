<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetSettings, SetLanguage } from '../wailsjs/go/main/App'
import Icon from './components/Icon.vue'
import LanguageSwitch from './components/LanguageSwitch.vue'
import Resumen from './components/Resumen.vue'
import Servicio from './components/Servicio.vue'
import Consumo from './components/Consumo.vue'
import AccesoWeb from './components/AccesoWeb.vue'
import Config from './components/Config.vue'
import Ajustes from './components/Ajustes.vue'

const { t, locale } = useI18n()

const active = ref('resumen')
// Secciones con su icono (set de Icon.vue).
const sections = [
  { id: 'resumen', icon: 'gauge' },
  { id: 'servicio', icon: 'server' },
  { id: 'accesoWeb', icon: 'globe' },
  { id: 'consumo', icon: 'chart' },
  { id: 'config', icon: 'sliders' },
  { id: 'ajustes', icon: 'settings' },
]

onMounted(async () => {
  try {
    const s = await GetSettings()
    locale.value = s.language
  } catch {
    // Sin backend (tests): mantiene 'es' por defecto.
  }
})

async function onLocale(lang: string) {
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
      <LanguageSwitch @change="onLocale" />
    </header>

    <nav class="app-nav" aria-label="Secciones">
      <button
        v-for="s in sections"
        :key="s.id"
        class="nav-btn"
        :class="{ active: active === s.id }"
        :aria-current="active === s.id ? 'page' : undefined"
        @click="active = s.id"
      >
        <Icon :name="s.icon" :size="16" />
        <span>{{ t('nav.' + s.id) }}</span>
      </button>
    </nav>

    <main class="app-content">
      <Resumen v-if="active === 'resumen'" />
      <Servicio v-else-if="active === 'servicio'" />
      <Consumo v-else-if="active === 'consumo'" />
      <AccesoWeb v-else-if="active === 'accesoWeb'" />
      <Config v-else-if="active === 'config'" />
      <Ajustes v-else-if="active === 'ajustes'" />
    </main>
  </div>
</template>
