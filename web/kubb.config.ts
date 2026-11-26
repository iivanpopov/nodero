import { env } from 'node:process'

import { adapterOas } from '@kubb/adapter-oas'
import { pluginBarrel } from '@kubb/plugin-barrel'
import { pluginFetch } from '@kubb/plugin-fetch'
import { pluginSvelteQuery } from '@nodero/plugin-svelte-query'
import { pluginTs } from '@kubb/plugin-ts'
import { pluginValibot } from '@kubb/plugin-valibot'
import { defineConfig } from 'kubb'

export default defineConfig({
  input: `${env.BACKEND_URL}/openapi.json`,
  output: {
    path: './generated',
    clean: true,
    barrel: {
      type: 'named',
    },
  },
  adapter: adapterOas(),
  plugins: [pluginTs(), pluginFetch(), pluginValibot(), pluginSvelteQuery({ hooks: true }), pluginBarrel()],
})
