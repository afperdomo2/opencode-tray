<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { HideWindow, Platform, WindowBlurred } from '../wailsjs/go/main/App'
import Icon from './components/Icon.vue'
import Resumen from './components/Resumen.vue'
import Servicio from './components/Servicio.vue'
import AccesoWeb from './components/AccesoWeb.vue'
import Ajustes from './components/Ajustes.vue'
import { loadSettings, settings } from './lib/settings'

const { t, locale } = useI18n()

const active = ref('resumen')

// En Linux/macOS el panel no está anclado al tray: se muestra una franja
// superior para arrastrarlo (Wails mueve la ventana con --wails-draggable).
const canDrag = ref(false)

// Secciones (Servicio ya incluye la configuración; Resumen es el dashboard).
const sections = [
  { id: 'resumen', icon: 'gauge' },
  { id: 'servicio', icon: 'server' },
  { id: 'accesoWeb', icon: 'globe' },
  { id: 'ajustes', icon: 'settings' },
]

// El alto lo fija el backend y no depende del contenido de cada pestaña.
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    void HideWindow().catch(() => {})
  }
}

function onBlur() {
  void WindowBlurred().catch(() => {})
}

onMounted(async () => {
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('blur', onBlur)

  try {
    canDrag.value = (await Platform()) !== 'windows'
  } catch {
    // Sin puente (tests o modo dev): se asume Windows y no se muestra la franja.
    canDrag.value = false
  }

  await loadSettings()
  locale.value = settings.language
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('blur', onBlur)
})
</script>

<template>
  <div class="app">
    <div v-if="canDrag" class="drag-strip" aria-hidden="true">
      <span class="drag-grip" />
    </div>

    <main class="app-content">
      <div class="app-content-inner">
        <Resumen v-if="active === 'resumen'" @navigate="active = $event" />
        <Servicio v-else-if="active === 'servicio'" />
        <AccesoWeb v-else-if="active === 'accesoWeb'" />
        <Ajustes v-else-if="active === 'ajustes'" />
      </div>
    </main>

    <nav class="app-nav" :aria-label="t('nav.label')">
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
  </div>
</template>
