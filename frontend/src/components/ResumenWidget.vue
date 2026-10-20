<script setup lang="ts">
import Icon from './Icon.vue'

// Shell visual de los widgets del Resumen. Es el único punto que maneja el
// clic/teclado cuando el widget es navegable.
const props = withDefaults(
  defineProps<{
    title: string
    icon: string
    tone?:
      | 'neutral'
      | 'running'
      | 'stopped'
      | 'local'
      | 'network'
      | 'remote'
      | 'mcp-ok'
      | 'mcp-warn'
    clickable?: boolean
    clickLabel?: string
  }>(),
  {
    tone: 'neutral',
    clickable: false,
    clickLabel: '',
  },
)

const emit = defineEmits<{ (e: 'open'): void }>()

function open() {
  if (props.clickable) emit('open')
}

function onKeydown(event: KeyboardEvent) {
  if (!props.clickable) return
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    open()
  }
}
</script>

<template>
  <article
    class="widget card"
    :class="[`tone-${tone}`, { 'widget-clickable': clickable }]"
    :role="clickable ? 'link' : undefined"
    :tabindex="clickable ? 0 : undefined"
    :aria-label="clickable ? clickLabel || title : undefined"
    :title="clickable ? clickLabel || title : undefined"
    @click="open"
    @keydown="onKeydown"
  >
    <header class="widget-head">
      <span class="widget-icon">
        <Icon :name="icon" :size="16" />
      </span>
      <h3>{{ title }}</h3>
      <span class="widget-tail">
        <slot name="meta" />
        <Icon v-if="clickable" class="widget-chevron" name="chevronDown" :size="16" />
      </span>
    </header>

    <div class="widget-body">
      <slot />
    </div>
  </article>
</template>
