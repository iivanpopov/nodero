import { resolveOperationName } from '@internals/shared'
import { createMutationResolver, createQueryResolver } from '@internals/tanstack-query'
import { createResolver, operationKey } from '@kubb/kit'
import type { PluginSvelteQuery } from '../types.ts'

/**
 * Default resolver used by `@kubb/plugin-svelte-query`. Decides the names and
 * file paths for every generated TanStack Query hook (`createFoo`,
 * `createFooInfinite`) and its companion helpers.
 */
export const resolverSvelteQuery = createResolver<PluginSvelteQuery>({
  pluginName: 'plugin-svelte-query',
  name: resolveOperationName,
  query: createQueryResolver('', 'create'),
  infiniteQuery: createQueryResolver('Infinite', 'create'),
  mutation: {
    ...createMutationResolver('create'),
    optionsName(node) {
      const name = `${this.name(node.operationId)}MutationOptions`
      return this.resolveName({ key: operationKey(node, 'mutation', 'optionsName'), name })
    },
  },
})
