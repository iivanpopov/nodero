import type { ast } from '@internals/ast'
import { ensureValidVarName } from '@internals/utils'
import { casing, operationKey } from '@kubb/kit'
import type { Resolver } from '@kubb/kit'

export type QueryVariant = '' | 'Suspense' | 'Infinite' | 'SuspenseInfinite'

export type QueryResolver = {
  name(this: Resolver, node: ast.OperationNode): string
  optionsName(this: Resolver, node: ast.OperationNode): string
  keyName(this: Resolver, node: ast.OperationNode): string
  keyTypeName(this: Resolver, node: ast.OperationNode): string
}

export type MutationResolver = {
  name(this: Resolver, node: ast.OperationNode): string
  keyName(this: Resolver, node: ast.OperationNode): string
}

export type HookFileVariant = QueryVariant | 'mutation'

export type HookFileNameOptions = {
  resolver: Resolver
  node: ast.OperationNode
  variant?: HookFileVariant
  prefix?: string
}

export function hookFileName({ resolver, node, variant = '', prefix }: HookFileNameOptions): string {
  const resolvedPrefix = prefix ?? (resolver.pluginName === 'plugin-svelte-query' ? 'create' : 'use')
  const name = `${resolvedPrefix}${casing.capitalize(ensureValidVarName(casing.camel(node.operationId)))}${variant === 'mutation' ? '' : variant}`
  return resolver.resolveName({ key: operationKey(node, 'hookFile', variant), name, scope: 'file' })
}

export function createQueryResolver(variant: QueryVariant = '', prefix = 'use'): QueryResolver {
  return {
    name(node) {
      const name = `${prefix}${casing.capitalize(this.name(node.operationId))}${variant}`
      return this.resolveName({ key: operationKey(node, 'query', variant, 'name'), name })
    },
    optionsName(node) {
      const name = `${this.name(node.operationId)}${variant}QueryOptions`
      return this.resolveName({ key: operationKey(node, 'query', variant, 'optionsName'), name })
    },
    keyName(node) {
      const name = `${this.name(node.operationId)}${variant}QueryKey`
      return this.resolveName({ key: operationKey(node, 'query', variant, 'keyName'), name })
    },
    keyTypeName(node) {
      const name = `${casing.capitalize(this.name(node.operationId))}${variant}QueryKey`
      return this.resolveName({ key: operationKey(node, 'query', variant, 'keyTypeName'), name })
    },
  }
}

export function createMutationResolver(prefix = 'use'): MutationResolver {
  return {
    name(node) {
      const name = `${prefix}${casing.capitalize(this.name(node.operationId))}`
      return this.resolveName({ key: operationKey(node, 'mutation', 'name'), name })
    },
    keyName(node) {
      const name = `${this.name(node.operationId)}MutationKey`
      return this.resolveName({ key: operationKey(node, 'mutation', 'keyName'), name })
    },
  }
}
