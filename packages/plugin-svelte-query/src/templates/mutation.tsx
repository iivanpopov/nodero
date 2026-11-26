import { ast } from '@internals/ast'
import { resolveClientOperation } from '@internals/client'
import { classifyOperation, HookFile, MutationKey, resolveHookFile } from '@internals/tanstack-query'
import { defineTemplate } from '@kubb/kit'
import { File } from '@kubb/renderer-jsx'
import { Mutation } from '../components/Mutation.tsx'
import { MutationOptions } from '../components/MutationOptions.tsx'
import type { PluginSvelteQuery } from '../types'

/**
 * Built-in template for `createMutation` hooks. Emits one
 * `createFoo` hook per POST/PUT/DELETE operation (configurable
 * via `mutation.methods`) plus the matching `fooMutationKey` / `fooMutationOptions` helpers.
 */
export const mutationOperation = defineTemplate<PluginSvelteQuery>({
  kind: ast.node.kinds.operation,
  render(node, ctx) {
    if (!ast.http.isOperation(node)) return null
    const { query, mutation, client, hooks } = ctx.options

    const { isMutation } = classifyOperation(node, { query, mutation })

    if (!isMutation) return null

    const importPath = mutation ? mutation.importPath : '@tanstack/svelte-query'

    const contractOp = resolveClientOperation({ node, pluginName: client.pluginName })
    if (!contractOp) return null

    const mutationHookName = ctx.resolver.mutation.name(node)
    const mutationOptionsName = ctx.resolver.mutation.optionsName(node)
    const mutationKeyName = ctx.resolver.mutation.keyName(node)

    const hook = resolveHookFile(node, { variant: 'mutation', contractOp, returnType: ctx.options.returnType })

    return (
      <HookFile hook={hook}>
        <MutationKey name={mutationKeyName} node={node} transformer={ctx.options.mutationKey} />

        <File.Import names={[{ name: 'mutationOptions', space: 'value' }]} path={importPath} />

        <MutationOptions
          name={mutationOptionsName}
          clientName={contractOp.name}
          mutationKeyName={mutationKeyName}
          node={node}
          tsResolver={hook.tsResolver}
          clientReturnType={contractOp.returnType}
          returnType={hook.returnType}
          typedHeaders={hook.contractOp.typedHeaders}
        />

        {mutation && hooks && (
          <>
            <File.Import
              names={[
                { name: 'createMutation', space: 'value' },
                { name: 'Accessor', space: 'type' },
                { name: 'CreateMutationOptions', space: 'type' },
                { name: 'QueryClient', space: 'type' },
              ]}
              path={importPath}
            />
            <Mutation
              name={mutationHookName}
              mutationOptionsName={mutationOptionsName}
              node={node}
              tsResolver={hook.tsResolver}
              returnType={hook.returnType}
              typedHeaders={hook.contractOp.typedHeaders}
            />
          </>
        )}
      </HookFile>
    )
  },
})
