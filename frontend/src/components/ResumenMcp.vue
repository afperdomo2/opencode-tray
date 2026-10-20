<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Skeleton from './Skeleton.vue'
import ResumenWidget from './ResumenWidget.vue'

interface MCPServer {
  name: string
  status: string
}

const props = defineProps<{
  servers: MCPServer[]
  loading: boolean
  error: string
}>()

const emit = defineEmits<{ (e: 'open'): void }>()
const { t } = useI18n()

const connected = computed(
  () => props.servers.filter((s) => s.status === 'connected').length,
)
const total = computed(() => props.servers.length)
const hasDown = computed(() => total.value > 0 && connected.value < total.value)
const tone = computed(() => {
  if (total.value === 0) return 'neutral'
  return hasDown.value ? 'mcp-warn' : 'mcp-ok'
})

function statusLabel(server: MCPServer): string {
  return server.status === 'connected'
    ? t('resumen.mcpConnected', { name: server.name })
    : t('resumen.mcpDisconnected', { name: server.name })
}
</script>

<template>
  <ResumenWidget
    :title="t('config.mcp')"
    icon="code"
    :tone="tone"
    clickable
    :click-label="t('resumen.openMcp')"
    class="widget-mcp"
    @open="emit('open')"
  >
    <div v-if="loading" class="skeleton-stack" role="status">
      <span class="sr-only">{{ t('common.loading') }}</span>
      <div class="mcp-chips">
        <Skeleton v-for="i in 3" :key="i" width="92px" height="26px" radius="999px" />
      </div>
    </div>
    <p v-else-if="error" class="error" role="alert">{{ error }}</p>
    <p v-else-if="servers.length === 0" class="muted">{{ t('config.empty') }}</p>
    <ul v-else class="mcp-chips">
      <li v-for="server in servers" :key="server.name" class="mcp-chip">
        <span
          class="dot"
          :class="server.status === 'connected' ? 'on' : 'off'"
          :title="statusLabel(server)"
          :aria-label="statusLabel(server)"
          role="img"
        ></span>
        <span class="mcp-chip-name">{{ server.name }}</span>
      </li>
    </ul>
  </ResumenWidget>
</template>
