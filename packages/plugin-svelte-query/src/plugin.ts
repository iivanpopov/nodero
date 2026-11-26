import { addMacro, addTemplate, defineExtension, Resolver, setResolver, useConfig } from '@kubb/kit'
import { version } from '../package.json'
import { pluginTsName } from '@kubb/plugin-ts'
import { mutationKeyTransformer, queryKeyTransformer, resolveInfiniteConfig, resolveMutationConfig, resolveQueryConfig } from '@internals/tanstack-query'
import { resolveContractClient } from '@internals/client'
import { infiniteQueryOperation, mutationOperation, queryOperation } from './templates'
import { resolverSvelteQuery } from './resolvers/resolverSvelteQuery.ts'
import type { PluginSvelteQuery, ResolverSvelteQuery } from './types.ts'

/**
 * Canonical plugin name for `@kubb/plugin-svelte-query`. Used for driver lookups
 * and cross-plugin dependency references.
 */
export const pluginSvelteQueryName = 'plugin-svelte-query' satisfies PluginSvelteQuery['name']

/**
 * Generates one TanStack Query hook per OpenAPI operation for Svelte 5 (`@tanstack/svelte-query` v6).
 * Queries become `createFoo` (and optionally `createFooInfinite`). Mutations become `createFoo` as well.
 */
export const pluginSvelteQuery = defineExtension<PluginSvelteQuery>({
  kind: 'plugin',
  meta: { name: pluginSvelteQueryName, version, compatibility: { kubb: '>=5.19.0', node: '>=22.22' } },
  dependencies: { [pluginTsName]: { optional: true } },
  setup(options) {
    const {
      output = { path: 'queries', barrel: { type: 'named' } },
      group,
      exclude = [],
      include,
      override = [],
      infinite = false,
      mutation = {},
      query = {},
      mutationKey = mutationKeyTransformer,
      queryKey = queryKeyTransformer,
      hooks = false,
      client,
      returnType = 'data',
      resolver: userResolver,
      macros: userMacros,
    } = options

    const resolver = userResolver ? Resolver.merge<ResolverSvelteQuery>(resolverSvelteQuery, ...[userResolver].flat()) : resolverSvelteQuery

    setResolver(resolver)
    addMacro(...(userMacros ?? []))

    addTemplate(queryOperation)
    addTemplate(infiniteQueryOperation)
    addTemplate(mutationOperation)

    return {
      output,
      client: resolveContractClient({ plugin: pluginSvelteQueryName, client, plugins: useConfig().plugins }),
      returnType,
      queryKey,
      query: resolveQueryConfig(query, { importPath: '@tanstack/svelte-query' }),
      mutationKey,
      mutation: resolveMutationConfig(mutation, { importPath: '@tanstack/svelte-query' }),
      infinite: resolveInfiniteConfig(infinite),
      hooks,
      group: group ?? null,
      exclude,
      include,
      override,
      resolver,
    }
  },
})

export default pluginSvelteQuery
