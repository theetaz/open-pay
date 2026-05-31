import { defineConfig } from 'vite'
import tsconfigPaths from 'vite-tsconfig-paths'
import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    tsconfigPaths({ projects: ['./tsconfig.json'] }),
    tailwindcss(),
    viteReact(),
  ],
  server: {
    port: 4500,
    host: true,
    watch: {
      // Enabled inside Docker (bind mounts need polling on macOS/Windows).
      usePolling: process.env.VITE_USE_POLLING === 'true',
    },
  },
})
