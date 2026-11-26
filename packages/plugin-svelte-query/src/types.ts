import type { ast } from '@internals/ast'
import type { ClientSelector, ResolvedContractClient } from '@internals/client'
import type { InfiniteConfig, MutationConfig, QueryConfig, KeyTransformer } from '@internals/tanstack-query'
import type { PluginOptions, ResolvedPluginOptions } from '@internals/shared/types'
import type { PluginType, Resolver, ResolverInput } from '@kubb/kit'

export type { InfiniteConfig, KeyTransformer, Transformer } from '@internals/tanstack-query'

/**
 * Resolver for Svelte Query that provides naming methods for hook functions (`create*`).
 */
export type ResolverSvelteQuery = Resolver & {
  /**
   * Names for a query operation: its `createQuery` hook, key helpers, and options helper.
   */
  query: {
    name(node: ast.OperationNode): string
    keyName(node: ast.OperationNode): string
    keyTypeName(node: ast.OperationNode): string
    optionsName(node: ast.OperationNode): string
  }
  /**
   * Names for an infinite query operation: its `createInfiniteQuery` hook, key helpers, and
   * options helper.
   */
  infiniteQuery: {
    name(node: ast.OperationNode): string
    keyName(node: ast.OperationNode): string
    keyTypeName(node: ast.OperationNode): string
    optionsName(node: ast.OperationNode): string
  }
  /**
   * Names for a mutation operation: its `createMutation` hook, options helper, and key helper.
   */
  mutation: {
    name(node: ast.OperationNode): string
    optionsName(node: ast.OperationNode): string
    keyName(node: ast.OperationNode): string
  }
}

export type QueryKeyTransformer = KeyTransformer
export type MutationKeyTransformer = KeyTransformer

export type PluginSvelteQueryOptions = PluginOptions<PluginSvelteQueryResolvedOptions> & {
  client?: ClientSelector
  returnType?: 'data' | 'full'
  infinite?: Partial<InfiniteConfig> | false
  queryKey?: QueryKeyTransformer
  query?: Partial<QueryConfig> | false
  mutationKey?: MutationKeyTransformer
  mutation?: Partial<MutationConfig> | false
  resolver?: ResolverInput<ResolverSvelteQuery>
  /**
   * Set to `true` to also generate `create*` hook functions. By default only the
   * `queryOptions`, `mutationOptions`, `queryKey`, and `mutationKey` helpers are emitted.
   *
   * @default false
   */
  hooks?: boolean
  macros?: Array<ast.Macro>
}

export type PluginSvelteQueryResolvedOptions = ResolvedPluginOptions<PluginSvelteQueryResolvedOptions> & {
  client: ResolvedContractClient
  returnType?: 'data' | 'full'
  infinite: NonNullable<InfiniteConfig> | false
  queryKey: QueryKeyTransformer | null
  query: NonNullable<Required<QueryConfig>> | false
  mutationKey: MutationKeyTransformer | null
  mutation: NonNullable<Required<MutationConfig>> | false
  hooks: boolean
  resolver: ResolverSvelteQuery
}

export type PluginSvelteQuery = PluginType<'plugin-svelte-query', PluginSvelteQueryOptions, PluginSvelteQueryResolvedOptions, ResolverSvelteQuery>

declare global {
  namespace Kubb {
    interface PluginRegistry {
      'plugin-svelte-query': PluginSvelteQuery
    }
  }
}
