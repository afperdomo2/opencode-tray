<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from './Icon.vue'
import Accordion from './Accordion.vue'
import {
  ConfigListMCP,
  ConfigJSON,
  ConfigOpenInEditor,
  ConfigOpenDir,
} from '../../wailsjs/go/main/App'

const { t } = useI18n()
const mcp = ref<{ name: string; status: string }[]>([])
const configText = ref('')
const loading = ref(true)
const error = ref('')

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    mcp.value = await ConfigListMCP()
    configText.value = await ConfigJSON()
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}
onMounted(refresh)
</script>

<template>
  <section class="section">
    <h2>{{ t('nav.config') }}</h2>
    <div v-if="loading" class="muted">{{ t('config.loading') }}</div>
    <template v-else>
      <h3>{{ t('config.mcp') }}</h3>
      <p v-if="mcp.length === 0" class="muted">{{ t('config.empty') }}</p>
      <ul class="mcp-list">
        <li v-for="s in mcp" :key="s.name">
          <span class="dot" :class="s.status === 'connected' ? 'on' : 'off'" aria-hidden="true"></span>
          {{ s.name }}
          <span class="muted">{{ s.status }}</span>
        </li>
      </ul>

      <div class="actions">
        <button @click="ConfigOpenInEditor">
          <Icon name="externalLink" :size="16" />{{ t('config.openEditor') }}
        </button>
        <button @click="ConfigOpenDir">
          <Icon name="folder" :size="16" />{{ t('config.openDir') }}
        </button>
      </div>

      <Accordion :title="t('config.json')">
        <pre class="config-view">{{ configText }}</pre>
      </Accordion>
      <p v-if="error" class="error">{{ error }}</p>
    </template>
  </section>
</template>
