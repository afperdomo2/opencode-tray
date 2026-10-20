<script setup lang="ts">
// Bloque de carga (skeleton) con dimensiones configurables. Es decorativo: el
// contenedor que lo usa debe anunciar el estado con role="status" + texto sr-only.
withDefaults(defineProps<{ width?: string; height?: string; radius?: string }>(), {
  width: '100%',
  height: '12px',
  radius: 'var(--radius-sm)',
})
</script>

<template>
  <span
    class="skeleton"
    :style="{ width, height, borderRadius: radius }"
    aria-hidden="true"
  ></span>
</template>
