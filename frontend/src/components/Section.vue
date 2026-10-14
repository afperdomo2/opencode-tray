<script setup lang="ts">
import { useI18n } from 'vue-i18n'

// Sección genérica: muestra título y descripción vía i18n.
// Los módulos futuros reemplazarán este contenido.
defineProps<{ name: string }>()
const { t } = useI18n()
</script>

<template>
  <section class="section">
    <h2>{{ t('nav.' + name) }}</h2>
    <p>{{ t('sections.' + name) }}</p>
    <p class="pending">{{ t('common.pending') }}</p>
  </section>
</template>
