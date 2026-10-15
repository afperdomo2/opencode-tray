<script setup lang="ts">
// Switch estilizado: el checkbox nativo queda accesible (oculto visualmente) y el
// aspecto (track + knob) se pinta con CSS según :checked.
defineProps<{ checked: boolean }>()
defineEmits<{ (e: 'change', checked: boolean): void }>()
</script>

<template>
  <label class="switch">
    <input
      type="checkbox"
      :checked="checked"
      @change="$emit('change', ($event.target as HTMLInputElement).checked)"
    />
    <span class="track"><span class="thumb"></span></span>
    <span class="switch-label"><slot /></span>
  </label>
</template>
