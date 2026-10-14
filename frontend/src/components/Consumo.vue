<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { UsageFetch } from '../../wailsjs/go/main/App'

const { t } = useI18n()
const rolling = ref(0)
const weekly = ref(0)
const monthly = ref(0)
const loading = ref(true)
const error = ref('')

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const u = await UsageFetch()
    rolling.value = u.rolling.percent
    weekly.value = u.weekly.percent
    monthly.value = u.monthly.percent
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
    <h2>{{ t('nav.consumo') }}</h2>
    <div v-if="loading" class="muted">{{ t('usage.loading') }}</div>
    <template v-else>
      <div
        v-for="w in [
          { key: 'rolling', label: t('usage.rolling'), pct: rolling },
          { key: 'weekly', label: t('usage.weekly'), pct: weekly },
          { key: 'monthly', label: t('usage.monthly'), pct: monthly },
        ]"
        :key="w.key"
        class="bar-row"
      >
        <span class="bar-label">{{ w.label }}</span>
        <div class="bar-track">
          <div
            class="bar-fill"
            :class="w.pct >= 98 ? 'crit' : w.pct >= 90 ? 'warn' : 'ok'"
            :style="{ width: Math.min(100, w.pct) + '%' }"
          ></div>
        </div>
        <span class="bar-pct">{{ Math.round(w.pct) }}%</span>
      </div>
      <p v-if="error" class="error">{{ error }}</p>
    </template>
  </section>
</template>
