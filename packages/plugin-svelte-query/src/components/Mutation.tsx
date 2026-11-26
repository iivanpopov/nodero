import type { ast } from '@internals/ast'
import type { ResolverTs } from '@kubb/plugin-ts'
import { createFunctionParameter, createFunctionParameters, functionPrinter } from '@kubb/plugin-ts'
import { Function } from '@kubb/renderer-jsx'
import type { KubbReactNode } from '@kubb/renderer-jsx'
import { buildResponseTypes, resolveMutationRequestType } from '@internals/tanstack-query'
import type { HookResultProps } from '@internals/tanstack-query'
import { buildOperationComments, buildRequestConfigType } from '@internals/shared'

export type MutationProps = {
  name: string
  mutationOptionsName: string
  node: ast.OperationNode
  tsResolver: ResolverTs
} & HookResultProps

const declarationPrinter = functionPrinter({ mode: 'declaration' })

export function Mutation({ name, mutationOptionsName, node, tsResolver, returnType, typedHeaders }: MutationProps): KubbReactNode {
  const { TData, TError } = buildResponseTypes(node, tsResolver, { returnType, typedHeaders })
  const TRequest = resolveMutationRequestType(node, tsResolver)
  const generics = [TData, TError, TRequest, 'TContext'].join(', ')

  const paramsNode = createFunctionParameters({
    params: [
      createFunctionParameter({
        name: 'options',
        type: `Accessor<{
  mutation?: CreateMutationOptions<${generics}>
  client?: ${buildRequestConfigType(node)}
}>`,
        optional: true,
      }),
      createFunctionParameter({
        name: 'queryClient',
        type: 'Accessor<QueryClient>',
        optional: true,
      }),
    ],
  })
  const paramsSignature = declarationPrinter.print(paramsNode) ?? ''

  return (
    <Function name={name} export params={paramsSignature} JSDoc={{ comments: buildOperationComments(node) }} generics={['TContext = unknown']}>
      {`
        return createMutation(() => {
          const { mutation = {}, client: config = {} } = options?.() ?? {}
          return {
            ...${mutationOptionsName}<TContext>(config),
            ...mutation,
          }
        }, queryClient)
    `}
    </Function>
  )
}
