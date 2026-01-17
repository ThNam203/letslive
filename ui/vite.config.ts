import { defineConfig } from 'vite'
import { tanstackStart } from '@tanstack/react-start/plugin/vite'
import viteReact from '@vitejs/plugin-react'
import tsconfigPaths from 'vite-tsconfig-paths'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  server: {
    port: 5000,
  },
  plugins: [
    tailwindcss(),
    // enables Vite to resolve imports using path aliases.
    tsconfigPaths(),
    tanstackStart({
      srcDirectory: 'src', // this is the default
      router: {
        routesDirectory: 'routes', // relative to srcDirectory
      },
    }),
    viteReact(),
  ],
})