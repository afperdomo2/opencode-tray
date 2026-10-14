import {defineConfig} from 'vitest/config'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/  ·  https://vitest.dev/config/
export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'jsdom',
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
  },
})
