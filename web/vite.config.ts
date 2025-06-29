import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:65605',
        changeOrigin: true,
      },
      '/auth': {
        target: 'http://localhost:65605',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:65605',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
}) 