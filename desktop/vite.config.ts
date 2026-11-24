import adapter from '@sveltejs/adapter-static'
import { sveltekit } from '@sveltejs/kit/vite'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const { TAURI_DEV_HOST: host } = loadEnv(mode, '.', 'TAURI_')

  return {
    plugins: [
      sveltekit({
        compilerOptions: {
          runes: ({ filename }) =>
            filename.split(/[/\\]/).includes('node_modules') ? undefined : true,
        },
        adapter: adapter({
          fallback: 'index.html',
        }),
      }),
    ],
    clearScreen: false,
    server: {
      port: 1420,
      strictPort: true,
      host: host || '127.0.0.1',
      hmr: host
        ? {
            protocol: 'ws',
            host,
            port: 1421,
          }
        : undefined,
      watch: {
        ignored: ['**/src-tauri/**'],
      },
    },
  }
})
