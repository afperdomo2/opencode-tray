<script setup lang="ts">
// Control segmentado genérico. Se usa para opciones cortas (intervalo de
// refresco) evitando el <select> nativo: en Linux/WebKitGTK abre una ventana
// GTK que roba el foco y ocultaba el panel al instante.
const props = defineProps<{
  modelValue: number
  options: number[]
  label: string
  unit?: string
  // formatLabel arma el texto accesible de cada opción (p. ej. "5 min").
  formatLabel?: (value: number) => string
}>()

const emit = defineEmits<{ (e: 'change', value: number): void }>()

function pick(value: number) {
  if (value === props.modelValue) return
  emit('change', value)
}
</script>

<template>
  <div class="seg" role="group" :aria-label="label">
    <button
      v-for="value in options"
      :key="value"
      type="button"
      class="seg-opt"
      :class="{ active: value === modelValue }"
      :aria-pressed="value === modelValue"
      :aria-label="formatLabel ? formatLabel(value) : String(value)"
      @click="pick(value)"
    >
      {{ value }}
    </button>
    <span v-if="unit" class="seg-unit" aria-hidden="true">{{ unit }}</span>
  </div>
</template>
