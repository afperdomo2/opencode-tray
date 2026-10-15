<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'

// Control segmentado de idioma (solo 2 opciones): ES | EN.
// Emite 'change' para que el contenedor persista el idioma (SetLanguage).
const emit = defineEmits<{ (e: 'change', lang: string): void }>()

const { t, locale } = useI18n()
const langs = ['es', 'en'] as const

function pick(lang: string) {
  if (locale.value === lang) return
  locale.value = lang
  emit('change', lang)
}
</script>

<template>
  <div class="lang-switch" role="group" :aria-label="t('common.language')">
    <Icon name="languages" :size="16" />
    <button
      v-for="l in langs"
      :key="l"
      type="button"
      class="lang-opt"
      :class="{ active: locale === l }"
      :aria-pressed="locale === l"
      @click="pick(l)"
    >
      {{ l.toUpperCase() }}
    </button>
  </div>
</template>
