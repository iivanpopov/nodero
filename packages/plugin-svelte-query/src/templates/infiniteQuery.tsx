import { ast } from '@internals/ast'
import { resolveClientOperation, resolveOperationTypes } from '@internals/client'
import {
  classifyOperation,
  getDefaultPageParamsWarning,
  HookFile,
  InfiniteQueryOptions,
  matchesInfinite,
  QueryKey,
  resolveHookFile,
  resolvePageParamType,
} from '@internals/tanstack-query'
import { defineTemplate } from '@kubb/kit'
import type { TemplateContext } from '@kubb/kit'
import { File } from '@kubb/renderer-jsx'
import { Query } from '../components/Query.tsx'
import type { PluginSvelteQuery } from '../types'

function matchesInfiniteQuery(node: ast.OperationNode, ctx: TemplateContext<PluginSvelteQuery>): boolean {
  if (!ast.http.isOperation(node)) return false
  const { query, mutation, infinite, hooks } = ctx.options

  const { isQuery, isMutation } = classifyOperation(node, { query, mutation })
  const infiniteOptions = infinite && typeof infinite === 'object' ? infinite : null
  if (!isQuery || isMutation || !infiniteOptions || !hooks) return false

  return matchesInfinite(node, infiniteOptions)
}

/**
 * Built-in template for `createInfiniteQuery` hooks. Enabled when
 * `pluginSvelteQuery({ infinite: { ... }, hooks: true })`.
 */
export const infiniteQueryOperation = defineTemplate<PluginSvelteQuery>({
  kind: ast.node.kinds.operation,
  render(node, ctx) {
    if (!matchesInfiniteQuery(node, ctx)) return null
    if (!ast.http.isOperation(node)) return null
    const { query, infinite, client } = ctx.options

    const infiniteOptions = infinite && typeof infinite === 'object' ? infinite : null
    if (!infiniteOptions) return null

    const pageParamsWarning = getDefaultPageParamsWarning(node, infiniteOptions)
    if (pageParamsWarning) ctx.warn(pageParamsWarning)

    const importPath = query ? query.importPath : '@tanstack/svelte-query'

    const contractOp = resolveClientOperation({ node, pluginName: client.pluginName })
    if (!contractOp) return null

    const queryName = ctx.resolver.infiniteQuery.name(node)
    const queryOptionsName = ctx.resolver.infiniteQuery.optionsName(node)
    const queryKeyName = ctx.resolver.infiniteQuery.keyName(node)
    const queryKeyTypeName = ctx.resolver.infiniteQuery.keyTypeName(node)

    const hook = resolveHookFile(node, {
      variant: 'Infinite',
      contractOp,
      returnType: ctx.options.returnType,
      keyTypeName: queryKeyTypeName,
      extraTypeNames: [
        resolvePageParamType(node, {
          resolver: resolveOperationTypes().resolver,
          initialPageParam: infiniteOptions.initialPageParam,
          queryParam: infiniteOptions.queryParam,
        }).queryParamsTypeName,
      ],
    })
    const { tsResolver } = hook

    return (
      <HookFile hook={hook}>
        <QueryKey
          name={queryKeyName}
          typeName={queryKeyTypeName}
          node={node}
          tsResolver={tsResolver}
          transformer={ctx.options.queryKey}
          variant="infiniteQuery"
          exportType
        />

        <File.Import
          names={[
            { name: 'infiniteQueryOptions', space: 'value' },
            { name: 'InfiniteData', space: 'type' },
          ]}
          path={importPath}
        />

        <InfiniteQueryOptions
          name={queryOptionsName}
          clientName={contractOp.name}
          queryKeyName={queryKeyName}
          node={node}
          tsResolver={tsResolver}
          cursorParam={infiniteOptions.cursorParam}
          nextParam={infiniteOptions.nextParam}
          previousParam={infiniteOptions.previousParam}
          getNextPageParam={infiniteOptions.getNextPageParam}
          getPreviousPageParam={infiniteOptions.getPreviousPageParam}
          initialPageParam={infiniteOptions.initialPageParam}
          hasExplicitInitialPageParam={infiniteOptions.hasExplicitInitialPageParam}
          queryParam={infiniteOptions.queryParam}
          clientReturnType={contractOp.returnType}
          returnType={hook.returnType}
          typedHeaders={hook.contractOp.typedHeaders}
        />

        <File.Import
          names={[
            { name: 'createInfiniteQuery', space: 'value' },
            { name: 'Accessor', space: 'type' },
            { name: 'QueryKey', space: 'type' },
            { name: 'QueryClient', space: 'type' },
            { name: 'CreateInfiniteQueryOptions', space: 'type' },
          ]}
          path={importPath}
        />

        <Query
          name={queryName}
          queryOptionsName={queryOptionsName}
          queryKeyTypeName={queryKeyTypeName}
          node={node}
          tsResolver={tsResolver}
          infinite={infiniteOptions}
          returnType={hook.returnType}
          typedHeaders={hook.contractOp.typedHeaders}
        />
      </HookFile>
    )
  },
})
