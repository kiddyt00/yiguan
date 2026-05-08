import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
// import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  base: '/admin/',
  plugins: [vue()], // tailwindcss() removed temporarily,
  server: {
    port: 3001,
    proxy: { '/api': 'http://localhost:9090' }
  }
})
