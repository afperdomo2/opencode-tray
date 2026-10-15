<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'

// Modal accesible: role=dialog + aria-modal; cierra con Esc, botón o clic en el
// backdrop; al abrir, el foco entra al diálogo.
const props = defineProps<{ open: boolean; title: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const { t } = useI18n()
const dialog = ref<HTMLElement | null>(null)

watch(
  () => props.open,
  async (open) => {
    if (open) {
      await nextTick()
      dialog.value?.focus()
    }
  },
)

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}
</script>

<template>
  <div v-if="open" class="modal-backdrop" @click.self="emit('close')">
    <div
      ref="dialog"
      class="modal"
      role="dialog"
      aria-modal="true"
      :aria-label="title"
      tabindex="-1"
      @keydown="onKeydown"
    >
      <div class="modal-head">
        <h3>{{ title }}</h3>
        <button type="button" class="icon-btn" :aria-label="t('common.close')" @click="emit('close')">
          <Icon name="x" :size="18" />
        </button>
      </div>
      <div class="modal-body">
        <slot />
      </div>
    </div>
  </div>
</template>
