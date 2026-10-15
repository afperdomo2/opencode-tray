<script setup lang="ts">
import { ref } from 'vue'
import Icon from './Icon.vue'

// Disclosure accesible: cabecera <button> con aria-expanded + panel asociado
// vía aria-controls. Cerrado por defecto (salvo defaultOpen).
const props = defineProps<{ title: string; defaultOpen?: boolean }>()
const open = ref(props.defaultOpen ?? false)
const panelId = `acc-${Math.random().toString(36).slice(2, 9)}`
</script>

<template>
  <div class="accordion">
    <button
      type="button"
      class="acc-head"
      :aria-expanded="open"
      :aria-controls="panelId"
      @click="open = !open"
    >
      <Icon name="chevronDown" :size="16" />
      <span>{{ title }}</span>
    </button>
    <div v-show="open" :id="panelId" class="acc-panel">
      <slot />
    </div>
  </div>
</template>
